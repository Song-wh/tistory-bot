package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// StockCollector 주식/코인 정보 수집기
type StockCollector struct {
	client *http.Client
}

// StockData 주식 데이터
type StockData struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Volume        int64     `json:"volume"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CryptoData 코인 데이터
type CryptoData struct {
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Change24h float64   `json:"change_24h"`
	Volume24h float64   `json:"volume_24h"`
	MarketCap float64   `json:"market_cap"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewStockCollector() *StockCollector {
	return &StockCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTopCryptos 상위 코인 정보 수집 (CoinGecko API - 무료)
func (s *StockCollector) GetTopCryptos(ctx context.Context, limit int) ([]CryptoData, error) {
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/markets?vs_currency=krw&order=market_cap_desc&per_page=%d&page=1",
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var coins []struct {
		ID             string  `json:"id"`
		Symbol         string  `json:"symbol"`
		Name           string  `json:"name"`
		CurrentPrice   float64 `json:"current_price"`
		PriceChange24h float64 `json:"price_change_percentage_24h"`
		TotalVolume    float64 `json:"total_volume"`
		MarketCap      float64 `json:"market_cap"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, err
	}

	var result []CryptoData
	for _, c := range coins {
		result = append(result, CryptoData{
			Symbol:    c.Symbol,
			Name:      c.Name,
			Price:     c.CurrentPrice,
			Change24h: c.PriceChange24h,
			Volume24h: c.TotalVolume,
			MarketCap: c.MarketCap,
			UpdatedAt: time.Now(),
		})
	}

	return result, nil
}

// GenerateCryptoPost 코인 정보 포스트 생성
func (s *StockCollector) GenerateCryptoPost(cryptos []CryptoData) *Post {
	now := time.Now()
	title := fmt.Sprintf("[%s] 오늘의 코인 시세 TOP 10", now.Format("2006-01-02"))

	content := `<h2>🪙 오늘의 암호화폐 시세</h2>
<p>실시간 업데이트: ` + now.Format("2006년 01월 02일 15:04") + `</p>

<table border="1" style="border-collapse: collapse; width: 100%;">
<tr style="background-color: #f2f2f2;">
<th>순위</th><th>코인</th><th>현재가(원)</th><th>24시간 변동</th><th>거래량</th>
</tr>
`

	for i, c := range cryptos {
		changeColor := "green"
		changeSign := "▲"
		if c.Change24h < 0 {
			changeColor = "red"
			changeSign = "▼"
		}

		content += fmt.Sprintf(`<tr>
<td>%d</td>
<td><strong>%s</strong> (%s)</td>
<td>₩%s</td>
<td style="color: %s;">%s %.2f%%</td>
<td>₩%s</td>
</tr>
`, i+1, c.Name, c.Symbol, formatNumber(c.Price), changeColor, changeSign, c.Change24h, formatNumber(c.Volume24h))
	}

	content += `</table>

<h3>📊 시장 분석</h3>
<p>위 데이터는 CoinGecko API를 통해 실시간으로 수집된 정보입니다.</p>

<p><em>※ 투자에 대한 책임은 투자자 본인에게 있습니다.</em></p>
`

	return &Post{
		Title:    title,
		Content:  content,
		Category: "주식/코인",
		Tags:     []string{"비트코인", "이더리움", "코인시세", "암호화폐", "가상화폐"},
	}
}

func formatNumber(n float64) string {
	if n >= 1000000000000 {
		return fmt.Sprintf("%.1f조", n/1000000000000)
	}
	if n >= 100000000 {
		return fmt.Sprintf("%.1f억", n/100000000)
	}
	if n >= 10000 {
		return fmt.Sprintf("%.1f만", n/10000)
	}
	return fmt.Sprintf("%.0f", n)
}
