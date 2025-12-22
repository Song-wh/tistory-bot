package collector

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// CoupangCollector 쿠팡 파트너스 수집기
type CoupangCollector struct {
	partnerID string
	browser   *rod.Browser
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
		partnerID: partnerID,
	}
}

// Connect 브라우저 연결
func (c *CoupangCollector) Connect() error {
	l := launcher.New().
		Headless(true).
		Leakless(false). // Windows 호환성
		Set("disable-gpu").
		Set("no-sandbox")

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("브라우저 실행 실패: %w", err)
	}

	c.browser = rod.New().ControlURL(url)
	if err := c.browser.Connect(); err != nil {
		return fmt.Errorf("브라우저 연결 실패: %w", err)
	}

	return nil
}

// Close 브라우저 종료
func (c *CoupangCollector) Close() {
	if c.browser != nil {
		c.browser.MustClose()
	}
}

// GetGoldboxProducts 쿠팡 골드박스 상품 수집 (브라우저 크롤링)
func (c *CoupangCollector) GetGoldboxProducts(ctx context.Context, limit int) ([]CoupangProduct, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}
	defer c.Close()

	fmt.Println("    🌐 쿠팡 골드박스 페이지 로딩 중...")

	page, err := c.browser.Page(proto.TargetCreateTarget{URL: "https://www.coupang.com/np/goldbox"})
	if err != nil {
		return nil, fmt.Errorf("페이지 열기 실패: %w", err)
	}
	defer page.Close()

	// 페이지 로딩 대기
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("페이지 로딩 실패: %w", err)
	}

	// 추가 대기 (동적 컨텐츠)
	time.Sleep(3 * time.Second)

	// 스크롤 다운해서 더 많은 상품 로딩
	page.MustEval(`() => window.scrollTo(0, 1000)`)
	time.Sleep(1 * time.Second)

	fmt.Println("    📦 상품 정보 추출 중...")

	// JavaScript로 상품 정보 추출
	result := page.MustEval(`(limit) => {
		const products = [];
		
		// 골드박스 상품 셀렉터들
		const items = document.querySelectorAll('.product-item, .baby-product-wrap, [class*="product-"]');
		
		for (const item of items) {
			if (products.length >= limit) break;
			
			try {
				// 제목
				let title = '';
				const nameEl = item.querySelector('.name, .product-name, [class*="name"]');
				if (nameEl) title = nameEl.textContent.trim();
				if (!title) {
					const linkEl = item.querySelector('a');
					if (linkEl) title = linkEl.getAttribute('title') || '';
				}
				if (!title) continue;
				
				// 가격
				let price = 0;
				const priceEl = item.querySelector('.price-value, .sale-price, [class*="price"] strong, [class*="price"]');
				if (priceEl) {
					const priceText = priceEl.textContent.replace(/[^0-9]/g, '');
					price = parseInt(priceText) || 0;
				}
				
				// 원가
				let origPrice = 0;
				const origEl = item.querySelector('.base-price, .origin-price, del');
				if (origEl) {
					const origText = origEl.textContent.replace(/[^0-9]/g, '');
					origPrice = parseInt(origText) || 0;
				}
				
				// 할인율
				let discountRate = 0;
				const discountEl = item.querySelector('.discount-rate, .discount-percentage, [class*="discount"]');
				if (discountEl) {
					const discountText = discountEl.textContent.match(/(\d+)/);
					if (discountText) discountRate = parseInt(discountText[1]);
				}
				if (!discountRate && origPrice > 0 && price > 0) {
					discountRate = Math.round((1 - price / origPrice) * 100);
				}
				
				// 이미지
				let imageUrl = '';
				const imgEl = item.querySelector('img');
				if (imgEl) {
					imageUrl = imgEl.src || imgEl.getAttribute('data-src') || '';
					if (imageUrl && !imageUrl.startsWith('http')) {
						imageUrl = 'https:' + imageUrl;
					}
				}
				
				// 상품 URL
				let productUrl = '';
				let productId = '';
				const linkEl = item.querySelector('a');
				if (linkEl) {
					productUrl = linkEl.href || '';
					const match = productUrl.match(/\/products\/(\d+)/);
					if (match) productId = match[1];
				}
				
				// 로켓배송 여부
				const isRocket = item.querySelector('[class*="rocket"], .badge-rocket') !== null;
				
				products.push({
					title: title,
					price: price,
					origPrice: origPrice,
					discountRate: discountRate,
					imageUrl: imageUrl,
					productUrl: productUrl,
					productId: productId,
					isRocket: isRocket,
					category: '골드박스'
				});
			} catch (e) {
				console.error('상품 파싱 에러:', e);
			}
		}
		
		return products;
	}`, limit)

	// 결과 파싱
	var products []CoupangProduct
	arr := result.Arr()
	
	for _, item := range arr {
		m := item.Map()
		product := CoupangProduct{
			Title:        m["title"].Str(),
			Price:        int(m["price"].Num()),
			OrigPrice:    int(m["origPrice"].Num()),
			DiscountRate: int(m["discountRate"].Num()),
			ImageURL:     m["imageUrl"].Str(),
			ProductURL:   m["productUrl"].Str(),
			ProductID:    m["productId"].Str(),
			IsRocket:     m["isRocket"].Bool(),
			Category:     m["category"].Str(),
		}
		
		if product.Title != "" && product.Price > 0 {
			products = append(products, product)
		}
	}

	fmt.Printf("    ✅ %d개 상품 수집 완료\n", len(products))
	return products, nil
}

// GetBestProducts 쿠팡 베스트 상품 수집
func (c *CoupangCollector) GetBestProducts(ctx context.Context, limit int) ([]CoupangProduct, error) {
	// 골드박스와 동일한 방식 사용
	return c.GetGoldboxProducts(ctx, limit)
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
