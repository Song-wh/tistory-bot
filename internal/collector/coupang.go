package collector

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CoupangCollector 쿠팡 파트너스 수집기
type CoupangCollector struct {
	client    *http.Client
	partnerID string
}

// CoupangProduct 쿠팡 상품 정보
type CoupangProduct struct {
	Title        string  `json:"title"`
	Price        int     `json:"price"`
	OrigPrice    int     `json:"orig_price"`
	DiscountRate int     `json:"discount_rate"`
	ImageURL     string  `json:"image_url"`
	ProductURL   string  `json:"product_url"`
	ProductID    string  `json:"product_id"`
	Category     string  `json:"category"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
	IsRocket     bool    `json:"is_rocket"`
}

// CoupangCategory 쿠팡 카테고리
type CoupangCategory struct {
	Name string
	URL  string
}

// NewCoupangCollector 쿠팡 수집기 생성
func NewCoupangCollector(partnerID string) *CoupangCollector {
	return &CoupangCollector{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		partnerID: partnerID,
	}
}

// GetGoldboxProducts 쿠팡 골드박스 상품 수집
func (c *CoupangCollector) GetGoldboxProducts(ctx context.Context, limit int) ([]CoupangProduct, error) {
	url := "https://www.coupang.com/np/goldbox"
	return c.scrapeProducts(ctx, url, limit, "골드박스")
}

// GetBestProducts 쿠팡 베스트 상품 수집
func (c *CoupangCollector) GetBestProducts(ctx context.Context, limit int) ([]CoupangProduct, error) {
	url := "https://www.coupang.com/np/campaigns/82"
	return c.scrapeProducts(ctx, url, limit, "베스트")
}

// GetRocketDeals 로켓배송 특가 수집
func (c *CoupangCollector) GetRocketDeals(ctx context.Context, limit int) ([]CoupangProduct, error) {
	url := "https://www.coupang.com/np/campaigns/82"
	products, err := c.scrapeProducts(ctx, url, limit*2, "로켓특가")
	if err != nil {
		return nil, err
	}

	// 로켓배송 상품만 필터링
	var rocketProducts []CoupangProduct
	for _, p := range products {
		if p.IsRocket && len(rocketProducts) < limit {
			rocketProducts = append(rocketProducts, p)
		}
	}
	return rocketProducts, nil
}

// scrapeProducts 상품 스크래핑
func (c *CoupangCollector) scrapeProducts(ctx context.Context, url string, limit int, category string) ([]CoupangProduct, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 브라우저처럼 헤더 설정
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP 에러: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var products []CoupangProduct

	// 골드박스/베스트 상품 파싱
	doc.Find(".baby-product, .product-item, li.baby-product-wrap").Each(func(i int, s *goquery.Selection) {
		if len(products) >= limit {
			return
		}

		product := c.parseProductItem(s, category)
		if product != nil && product.Title != "" {
			products = append(products, *product)
		}
	})

	// 대체 셀렉터
	if len(products) == 0 {
		doc.Find("[class*='product'], [class*='item']").Each(func(i int, s *goquery.Selection) {
			if len(products) >= limit {
				return
			}

			product := c.parseProductItem(s, category)
			if product != nil && product.Title != "" {
				products = append(products, *product)
			}
		})
	}

	return products, nil
}

// parseProductItem 개별 상품 파싱
func (c *CoupangCollector) parseProductItem(s *goquery.Selection, category string) *CoupangProduct {
	product := &CoupangProduct{
		Category: category,
	}

	// 제목
	product.Title = strings.TrimSpace(s.Find(".name, .product-name, [class*='name']").First().Text())
	if product.Title == "" {
		product.Title = strings.TrimSpace(s.Find("a").First().AttrOr("title", ""))
	}

	// 가격
	priceText := s.Find(".price-value, .sale-price, [class*='price']").First().Text()
	product.Price = c.parsePrice(priceText)

	// 원가
	origPriceText := s.Find(".base-price, .origin-price, [class*='origin']").First().Text()
	product.OrigPrice = c.parsePrice(origPriceText)

	// 할인율
	discountText := s.Find(".discount-rate, .discount-percentage, [class*='discount']").First().Text()
	product.DiscountRate = c.parseDiscount(discountText)

	// 할인율 계산 (없는 경우)
	if product.DiscountRate == 0 && product.OrigPrice > 0 && product.Price > 0 {
		product.DiscountRate = int((1 - float64(product.Price)/float64(product.OrigPrice)) * 100)
	}

	// 이미지
	product.ImageURL, _ = s.Find("img").First().Attr("src")
	if product.ImageURL == "" {
		product.ImageURL, _ = s.Find("img").First().Attr("data-src")
	}
	if !strings.HasPrefix(product.ImageURL, "http") && product.ImageURL != "" {
		product.ImageURL = "https:" + product.ImageURL
	}

	// 상품 URL
	productURL, exists := s.Find("a").First().Attr("href")
	if exists {
		if !strings.HasPrefix(productURL, "http") {
			productURL = "https://www.coupang.com" + productURL
		}
		product.ProductURL = productURL
		product.ProductID = c.extractProductID(productURL)
	}

	// 로켓배송 여부
	rocketBadge := s.Find("[class*='rocket'], .badge-rocket").Length()
	product.IsRocket = rocketBadge > 0

	// 리뷰
	reviewText := s.Find(".rating-total-count, [class*='review']").First().Text()
	product.ReviewCount = c.parseReviewCount(reviewText)

	// 평점
	ratingText := s.Find(".rating, [class*='star']").First().AttrOr("style", "")
	product.Rating = c.parseRating(ratingText)

	return product
}

// parsePrice 가격 파싱
func (c *CoupangCollector) parsePrice(text string) int {
	re := regexp.MustCompile(`[\d,]+`)
	match := re.FindString(text)
	if match == "" {
		return 0
	}
	match = strings.ReplaceAll(match, ",", "")
	price, _ := strconv.Atoi(match)
	return price
}

// parseDiscount 할인율 파싱
func (c *CoupangCollector) parseDiscount(text string) int {
	re := regexp.MustCompile(`(\d+)%`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		discount, _ := strconv.Atoi(matches[1])
		return discount
	}
	return 0
}

// parseReviewCount 리뷰 수 파싱
func (c *CoupangCollector) parseReviewCount(text string) int {
	re := regexp.MustCompile(`[\d,]+`)
	match := re.FindString(text)
	if match == "" {
		return 0
	}
	match = strings.ReplaceAll(match, ",", "")
	count, _ := strconv.Atoi(match)
	return count
}

// parseRating 평점 파싱
func (c *CoupangCollector) parseRating(styleText string) float64 {
	re := regexp.MustCompile(`width:\s*([\d.]+)%`)
	matches := re.FindStringSubmatch(styleText)
	if len(matches) >= 2 {
		percent, _ := strconv.ParseFloat(matches[1], 64)
		return percent / 20 // 100% = 5점
	}
	return 0
}

// extractProductID 상품 ID 추출
func (c *CoupangCollector) extractProductID(url string) string {
	re := regexp.MustCompile(`/products/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// GeneratePartnerLink 파트너스 링크 생성
func (c *CoupangCollector) GeneratePartnerLink(productURL string) string {
	// 쿠팡 파트너스 딥링크 형식
	// 상품 URL을 파트너스 추적 링크로 변환
	if c.partnerID == "" {
		return productURL
	}

	// URL에 파트너스 추적 파라미터 추가
	separator := "?"
	if strings.Contains(productURL, "?") {
		separator = "&"
	}

	return fmt.Sprintf("%s%swPcid=%s&sfrn=AFFILIATE", productURL, separator, c.partnerID)
}

// GenerateCoupangPost 쿠팡 특가 포스트 생성
func (c *CoupangCollector) GenerateCoupangPost(products []CoupangProduct) *Post {
	now := time.Now()
	title := fmt.Sprintf("[%s] 오늘의 쿠팡 특가 🛒 최대 %d%% 할인", now.Format("01/02"), c.getMaxDiscount(products))

	var content strings.Builder

	// 스타일 정의
	content.WriteString(`
<style>
.coupang-container { max-width: 800px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
.coupang-header { background: linear-gradient(135deg, #00A0E4 0%, #0075C4 100%); color: white; padding: 30px; border-radius: 16px; text-align: center; margin-bottom: 30px; }
.coupang-header h1 { margin: 0 0 10px 0; font-size: 28px; }
.coupang-header p { margin: 0; opacity: 0.9; }
.product-grid { display: grid; gap: 20px; }
.product-card { background: #fff; border: 1px solid #e5e5e5; border-radius: 12px; overflow: hidden; transition: all 0.3s; }
.product-card:hover { box-shadow: 0 8px 25px rgba(0,0,0,0.1); transform: translateY(-2px); }
.product-image { width: 100%; height: 200px; object-fit: cover; background: #f5f5f5; }
.product-info { padding: 16px; }
.product-title { font-size: 15px; font-weight: 600; color: #111; line-height: 1.4; margin-bottom: 12px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.price-section { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.discount-badge { background: #f03e3e; color: white; padding: 4px 8px; border-radius: 4px; font-weight: 700; font-size: 14px; }
.current-price { font-size: 22px; font-weight: 700; color: #111; }
.original-price { font-size: 14px; color: #999; text-decoration: line-through; }
.badges { display: flex; gap: 6px; margin-bottom: 12px; }
.badge { font-size: 11px; padding: 3px 8px; border-radius: 4px; }
.badge-rocket { background: #0073e9; color: white; }
.badge-best { background: #ff6b35; color: white; }
.buy-button { display: block; width: 100%; background: #00a0e4; color: white; text-align: center; padding: 14px; text-decoration: none; font-weight: 600; border-radius: 8px; transition: background 0.2s; }
.buy-button:hover { background: #0085c4; color: white; }
.footer-notice { background: #f9f9f9; padding: 20px; border-radius: 12px; margin-top: 30px; font-size: 13px; color: #666; }
.footer-notice p { margin: 5px 0; }
</style>
`)

	content.WriteString(`<div class="coupang-container">`)

	// 헤더
	content.WriteString(fmt.Sprintf(`
<div class="coupang-header">
	<h1>🛒 오늘의 쿠팡 특가</h1>
	<p>%s 업데이트 | 놓치면 후회할 핫딜 모음!</p>
</div>
`, now.Format("2006년 01월 02일 15:04")))

	// 상품 그리드
	content.WriteString(`<div class="product-grid">`)

	for i, product := range products {
		partnerLink := c.GeneratePartnerLink(product.ProductURL)

		// 이미지 URL 처리
		imageURL := product.ImageURL
		if imageURL == "" {
			imageURL = "https://via.placeholder.com/300x200?text=No+Image"
		}

		content.WriteString(fmt.Sprintf(`
<div class="product-card">
	<a href="%s" target="_blank" rel="noopener">
		<img src="%s" alt="%s" class="product-image" loading="lazy" onerror="this.src='https://via.placeholder.com/300x200?text=Image'">
	</a>
	<div class="product-info">
		<div class="product-title">%d. %s</div>
		<div class="badges">
`, partnerLink, imageURL, product.Title, i+1, product.Title))

		// 뱃지
		if product.IsRocket {
			content.WriteString(`<span class="badge badge-rocket">🚀 로켓배송</span>`)
		}
		if product.DiscountRate >= 50 {
			content.WriteString(`<span class="badge badge-best">🔥 초특가</span>`)
		}

		content.WriteString(`</div>`)

		// 가격 정보
		content.WriteString(`<div class="price-section">`)
		if product.DiscountRate > 0 {
			content.WriteString(fmt.Sprintf(`<span class="discount-badge">%d%%</span>`, product.DiscountRate))
		}
		if product.Price > 0 {
			content.WriteString(fmt.Sprintf(`<span class="current-price">%s원</span>`, c.formatPrice(product.Price)))
		}
		content.WriteString(`</div>`)

		if product.OrigPrice > 0 && product.OrigPrice != product.Price {
			content.WriteString(fmt.Sprintf(`<div class="original-price">정가 %s원</div>`, c.formatPrice(product.OrigPrice)))
		}

		content.WriteString(fmt.Sprintf(`
		<a href="%s" target="_blank" rel="noopener" class="buy-button">👉 최저가 구매하기</a>
	</div>
</div>
`, partnerLink))
	}

	content.WriteString(`</div>`) // product-grid 끝

	// 푸터
	content.WriteString(`
<div class="footer-notice">
	<p>💡 <strong>Tip:</strong> 쿠팡은 가격이 수시로 변동됩니다. 마음에 드는 상품은 빨리 구매하세요!</p>
	<p>📦 로켓배송 상품은 오늘 주문하면 내일 도착!</p>
	<p>⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.</p>
</div>
`)

	content.WriteString(`</div>`) // container 끝

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryCoupang,
		Tags:     []string{"쿠팡", "쿠팡특가", "골드박스", "핫딜", "오늘의특가", "로켓배송", "최저가"},
	}
}

// GenerateCategoryPost 카테고리별 포스트 생성
func (c *CoupangCollector) GenerateCategoryPost(products []CoupangProduct, categoryName string) *Post {
	now := time.Now()
	
	// 카테고리별 이모지
	emoji := "🛒"
	switch categoryName {
	case "가전/디지털":
		emoji = "📱"
	case "패션":
		emoji = "👗"
	case "식품":
		emoji = "🍎"
	case "생활":
		emoji = "🏠"
	case "뷰티":
		emoji = "💄"
	}

	title := fmt.Sprintf("[%s] %s %s 베스트 특가 TOP %d", now.Format("01/02"), emoji, categoryName, len(products))

	var content strings.Builder
	
	content.WriteString(fmt.Sprintf(`
<h2>%s %s 카테고리 인기 특가</h2>
<p>📅 %s 기준 | 실시간 베스트 상품</p>
<hr>
`, emoji, categoryName, now.Format("2006년 01월 02일 15:04")))

	for i, product := range products {
		partnerLink := c.GeneratePartnerLink(product.ProductURL)
		
		content.WriteString(fmt.Sprintf(`
<div style="border: 2px solid #00a0e4; border-radius: 12px; padding: 20px; margin: 15px 0; background: #fafafa;">
	<h3 style="margin: 0 0 10px 0; color: #333;">%d위. %s</h3>
`, i+1, product.Title))

		if product.ImageURL != "" {
			content.WriteString(fmt.Sprintf(`
	<div style="text-align: center; margin: 15px 0;">
		<img src="%s" alt="%s" style="max-width: 100%%; height: auto; border-radius: 8px;">
	</div>
`, product.ImageURL, product.Title))
		}

		// 가격 정보
		content.WriteString(`<div style="background: #fff; padding: 15px; border-radius: 8px; margin: 10px 0;">`)
		
		if product.DiscountRate > 0 {
			content.WriteString(fmt.Sprintf(`<span style="background: #f03e3e; color: white; padding: 5px 10px; border-radius: 5px; font-weight: bold; margin-right: 10px;">%d%% 할인</span>`, product.DiscountRate))
		}
		
		if product.Price > 0 {
			content.WriteString(fmt.Sprintf(`<span style="font-size: 24px; font-weight: bold; color: #111;">%s원</span>`, c.formatPrice(product.Price)))
		}
		
		if product.OrigPrice > 0 && product.OrigPrice != product.Price {
			content.WriteString(fmt.Sprintf(`<br><span style="text-decoration: line-through; color: #999;">정가 %s원</span>`, c.formatPrice(product.OrigPrice)))
		}
		
		content.WriteString(`</div>`)

		// 뱃지들
		if product.IsRocket {
			content.WriteString(`<span style="background: #0073e9; color: white; padding: 3px 8px; border-radius: 4px; font-size: 12px; margin-right: 5px;">🚀 로켓배송</span>`)
		}

		// 구매 버튼
		content.WriteString(fmt.Sprintf(`
	<div style="margin-top: 15px;">
		<a href="%s" target="_blank" style="display: inline-block; background: #00a0e4; color: white; padding: 12px 30px; border-radius: 8px; text-decoration: none; font-weight: bold;">👉 최저가 확인하기</a>
	</div>
</div>
`, partnerLink))
	}

	content.WriteString(`
<hr>
<p style="background: #f5f5f5; padding: 15px; border-radius: 8px; font-size: 13px; color: #666;">
⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.<br>
💡 가격 및 재고는 수시로 변동될 수 있으니 구매 전 확인해주세요.
</p>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryCoupang,
		Tags:     []string{"쿠팡", categoryName, "특가", "베스트", "추천", "할인"},
	}
}

// formatPrice 가격 포맷팅 (천단위 콤마)
func (c *CoupangCollector) formatPrice(price int) string {
	str := strconv.Itoa(price)
	n := len(str)
	if n <= 3 {
		return str
	}
	
	var result strings.Builder
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(str[:remainder])
		if n > 3 {
			result.WriteString(",")
		}
	}
	
	for i := remainder; i < n; i += 3 {
		if i > remainder {
			result.WriteString(",")
		}
		result.WriteString(str[i : i+3])
	}
	
	return result.String()
}

// getMaxDiscount 최대 할인율 반환
func (c *CoupangCollector) getMaxDiscount(products []CoupangProduct) int {
	maxDiscount := 0
	for _, p := range products {
		if p.DiscountRate > maxDiscount {
			maxDiscount = p.DiscountRate
		}
	}
	if maxDiscount == 0 {
		maxDiscount = 50 // 기본값
	}
	return maxDiscount
}

// GetMockProducts 테스트용 모의 상품 데이터
func (c *CoupangCollector) GetMockProducts(limit int) []CoupangProduct {
	mockProducts := []CoupangProduct{
		{
			Title:        "삼성전자 갤럭시 버즈3 프로 무선 이어폰",
			Price:        189000,
			OrigPrice:    289000,
			DiscountRate: 35,
			ImageURL:     "https://thumbnail7.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/07/11/15/1/5b99a2c2-69f5-4c7f-a5a4-0c9e4d6a6a8e.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7012345678",
			ProductID:    "7012345678",
			Category:     "디지털/가전",
			IsRocket:     true,
		},
		{
			Title:        "애플 에어팟 프로 2세대 USB-C",
			Price:        298000,
			OrigPrice:    359000,
			DiscountRate: 17,
			ImageURL:     "https://thumbnail6.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2023/09/20/11/8/a3e6b7c8-1234-5678-9abc-def012345678.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7023456789",
			ProductID:    "7023456789",
			Category:     "디지털/가전",
			IsRocket:     true,
		},
		{
			Title:        "LG 스탠바이미 GO 27인치 휴대용 스마트 TV",
			Price:        890000,
			OrigPrice:    1190000,
			DiscountRate: 25,
			ImageURL:     "https://thumbnail8.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/03/15/10/5/lg-standbyme.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7034567890",
			ProductID:    "7034567890",
			Category:     "디지털/가전",
			IsRocket:     true,
		},
		{
			Title:        "다이슨 V15 디텍트 컴플리트 무선청소기",
			Price:        999000,
			OrigPrice:    1290000,
			DiscountRate: 23,
			ImageURL:     "https://thumbnail9.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/01/10/14/2/dyson-v15.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7045678901",
			ProductID:    "7045678901",
			Category:     "가전",
			IsRocket:     true,
		},
		{
			Title:        "나이키 에어맥스 97 남성 운동화",
			Price:        129000,
			OrigPrice:    199000,
			DiscountRate: 35,
			ImageURL:     "https://thumbnail10.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/02/20/09/3/nike-airmax.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7056789012",
			ProductID:    "7056789012",
			Category:     "패션",
			IsRocket:     true,
		},
		{
			Title:        "곰곰 GAP 냉동 블루베리 1kg",
			Price:        12900,
			OrigPrice:    18900,
			DiscountRate: 32,
			ImageURL:     "https://thumbnail11.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/04/05/11/1/blueberry.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7067890123",
			ProductID:    "7067890123",
			Category:     "식품",
			IsRocket:     true,
		},
		{
			Title:        "에스티로더 갈색병 어드밴스드 나이트 리페어 세럼 50ml",
			Price:        89000,
			OrigPrice:    142000,
			DiscountRate: 37,
			ImageURL:     "https://thumbnail12.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/05/10/16/4/esteelauder.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7078901234",
			ProductID:    "7078901234",
			Category:     "뷰티",
			IsRocket:     true,
		},
		{
			Title:        "코멧 올인원 캡슐 식기세척기 세제 100개입",
			Price:        19900,
			OrigPrice:    35000,
			DiscountRate: 43,
			ImageURL:     "https://thumbnail13.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/06/15/13/2/comet-dish.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7089012345",
			ProductID:    "7089012345",
			Category:     "생활",
			IsRocket:     true,
		},
		{
			Title:        "샤오미 미밴드 8 프로 스마트밴드",
			Price:        59000,
			OrigPrice:    79000,
			DiscountRate: 25,
			ImageURL:     "https://thumbnail14.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/07/20/10/5/miband8.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7090123456",
			ProductID:    "7090123456",
			Category:     "디지털",
			IsRocket:     true,
		},
		{
			Title:        "오뚜기 진라면 매운맛 120g x 40봉",
			Price:        23900,
			OrigPrice:    32000,
			DiscountRate: 25,
			ImageURL:     "https://thumbnail15.coupangcdn.com/thumbnails/remote/230x230ex/image/retail/images/2024/08/01/09/1/jinramen.jpg",
			ProductURL:   "https://www.coupang.com/vp/products/7101234567",
			ProductID:    "7101234567",
			Category:     "식품",
			IsRocket:     true,
		},
	}

	if limit > len(mockProducts) {
		limit = len(mockProducts)
	}

	return mockProducts[:limit]
}

