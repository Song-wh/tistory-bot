package collector

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// DealsCollector 핫딜 정보 수집기
type DealsCollector struct {
	client *http.Client
}

// Deal 할인 정보
type Deal struct {
	Title       string    `json:"title"`
	Price       string    `json:"price"`
	OrigPrice   string    `json:"orig_price"`
	Discount    string    `json:"discount"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	CollectedAt time.Time `json:"collected_at"`
}

func NewDealsCollector() *DealsCollector {
	return &DealsCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetDealsFromPpomppu 뽐뿌 핫딜 수집
func (d *DealsCollector) GetDealsFromPpomppu(ctx context.Context, limit int) ([]Deal, error) {
	// 뽐뿌 핫딜 게시판
	url := "https://www.ppomppu.co.kr/zboard/zboard.php?id=ppomppu"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var deals []Deal
	// HTML 파싱하여 핫딜 정보 추출
	// 실제 구현에서는 더 정교한 파싱 필요
	d.extractDeals(doc, &deals, limit)

	return deals, nil
}

func (d *DealsCollector) extractDeals(n *html.Node, deals *[]Deal, limit int) {
	// 간단한 예시 - 실제로는 더 정교한 파싱 필요
	if len(*deals) >= limit {
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		d.extractDeals(c, deals, limit)
	}
}

// GenerateDealsPost 핫딜 정보 포스트 생성
func (d *DealsCollector) GenerateDealsPost(deals []Deal) *Post {
	now := time.Now()
	title := fmt.Sprintf("[%s] 오늘의 핫딜 모음 🔥", now.Format("01/02"))

	var content strings.Builder
	content.WriteString(`<h2>🛒 오늘의 핫딜 모음</h2>
<p>업데이트: ` + now.Format("2006년 01월 02일 15:04") + `</p>
`)

	for i, deal := range deals {
		content.WriteString(fmt.Sprintf(`
<div style="border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 8px;">
<h3>%d. %s</h3>
<p><strong style="color: red; font-size: 1.2em;">%s</strong> <del>%s</del></p>
<p>할인율: %s | 출처: %s</p>
<p><a href="%s" target="_blank">👉 바로가기</a></p>
</div>
`, i+1, deal.Title, deal.Price, deal.OrigPrice, deal.Discount, deal.Source, deal.URL))
	}

	content.WriteString(`
<p><em>※ 가격 및 할인율은 변동될 수 있습니다. 구매 전 확인해주세요.</em></p>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryDeal,
		Tags:     []string{"핫딜", "특가", "할인", "쿠팡", "최저가"},
	}
}

// GetCoupangDeals 쿠팡 골드박스 수집 (쿠팡파트너스 API 필요)
func (d *DealsCollector) GetCoupangDeals(ctx context.Context, apiKey, secretKey string) ([]Deal, error) {
	// 쿠팡파트너스 API 구현
	// https://partners.coupang.com/ 에서 API 발급 필요
	return nil, fmt.Errorf("쿠팡파트너스 API 키가 필요합니다")
}
