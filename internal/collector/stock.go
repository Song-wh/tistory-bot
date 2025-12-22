package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
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

// CryptoData 코인 데이터 (확장)
type CryptoData struct {
	Symbol        string    `json:"symbol"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	Change1h      float64   `json:"change_1h"`
	Change24h     float64   `json:"change_24h"`
	Change7d      float64   `json:"change_7d"`
	Volume24h     float64   `json:"volume_24h"`
	MarketCap     float64   `json:"market_cap"`
	ATH           float64   `json:"ath"`             // 역대 최고가
	ATHChangePerc float64   `json:"ath_change_perc"` // ATH 대비 변동률
	Sparkline     []float64 `json:"sparkline"`       // 7일 가격 데이터
	UpdatedAt     time.Time `json:"updated_at"`
}

// MarketData 시장 전체 데이터
type MarketData struct {
	TotalMarketCap     float64 `json:"total_market_cap"`
	TotalVolume        float64 `json:"total_volume"`
	BTCDominance       float64 `json:"btc_dominance"`
	ETHDominance       float64 `json:"eth_dominance"`
	MarketCapChange24h float64 `json:"market_cap_change_24h"`
}

// FearGreedData 공포탐욕지수
type FearGreedData struct {
	Value      int    `json:"value"`
	ValueClass string `json:"value_class"` // Extreme Fear, Fear, Neutral, Greed, Extreme Greed
}

// CryptoRecommendation 추천 종목
type CryptoRecommendation struct {
	Coin       CryptoData
	Score      float64 // 추천 점수
	Reason     string  // 추천 이유
	SignalType string  // BUY, HOLD, WATCH
}

func NewStockCollector() *StockCollector {
	return &StockCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetTopCryptos 상위 코인 정보 수집 (확장 버전)
func (s *StockCollector) GetTopCryptos(ctx context.Context, limit int) ([]CryptoData, error) {
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/coins/markets?vs_currency=krw&order=market_cap_desc&per_page=%d&page=1&sparkline=true&price_change_percentage=1h,24h,7d",
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
		ID                           string    `json:"id"`
		Symbol                       string    `json:"symbol"`
		Name                         string    `json:"name"`
		CurrentPrice                 float64   `json:"current_price"`
		MarketCap                    float64   `json:"market_cap"`
		TotalVolume                  float64   `json:"total_volume"`
		PriceChangePercentage1h      float64   `json:"price_change_percentage_1h_in_currency"`
		PriceChangePercentage24h     float64   `json:"price_change_percentage_24h_in_currency"`
		PriceChangePercentage7d      float64   `json:"price_change_percentage_7d_in_currency"`
		ATH                          float64   `json:"ath"`
		ATHChangePercentage          float64   `json:"ath_change_percentage"`
		SparklineIn7d                struct {
			Price []float64 `json:"price"`
		} `json:"sparkline_in_7d"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, err
	}

	var result []CryptoData
	for _, c := range coins {
		sparkline := []float64{}
		if len(c.SparklineIn7d.Price) > 0 {
			// 7일 데이터를 24개 포인트로 축약
			step := len(c.SparklineIn7d.Price) / 24
			if step < 1 {
				step = 1
			}
			for i := 0; i < len(c.SparklineIn7d.Price); i += step {
				sparkline = append(sparkline, c.SparklineIn7d.Price[i])
			}
		}

		result = append(result, CryptoData{
			Symbol:        strings.ToUpper(c.Symbol),
			Name:          c.Name,
			Price:         c.CurrentPrice,
			Change1h:      c.PriceChangePercentage1h,
			Change24h:     c.PriceChangePercentage24h,
			Change7d:      c.PriceChangePercentage7d,
			Volume24h:     c.TotalVolume,
			MarketCap:     c.MarketCap,
			ATH:           c.ATH,
			ATHChangePerc: c.ATHChangePercentage,
			Sparkline:     sparkline,
			UpdatedAt:     time.Now(),
		})
	}

	return result, nil
}

// GetMarketData 전체 시장 데이터 수집
func (s *StockCollector) GetMarketData(ctx context.Context) (*MarketData, error) {
	url := "https://api.coingecko.com/api/v3/global"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getSimulatedMarketData(), nil
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalMarketCap         map[string]float64 `json:"total_market_cap"`
			TotalVolume            map[string]float64 `json:"total_volume"`
			MarketCapPercentage    map[string]float64 `json:"market_cap_percentage"`
			MarketCapChangePercent float64            `json:"market_cap_change_percentage_24h_usd"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getSimulatedMarketData(), nil
	}

	return &MarketData{
		TotalMarketCap:     result.Data.TotalMarketCap["krw"],
		TotalVolume:        result.Data.TotalVolume["krw"],
		BTCDominance:       result.Data.MarketCapPercentage["btc"],
		ETHDominance:       result.Data.MarketCapPercentage["eth"],
		MarketCapChange24h: result.Data.MarketCapChangePercent,
	}, nil
}

// getSimulatedMarketData 시뮬레이션 데이터
func (s *StockCollector) getSimulatedMarketData() *MarketData {
	return &MarketData{
		TotalMarketCap:     3500000000000000, // 3500조
		TotalVolume:        150000000000000,  // 150조
		BTCDominance:       52.5,
		ETHDominance:       17.3,
		MarketCapChange24h: 1.5,
	}
}

// GetFearGreedIndex 공포탐욕지수 수집
func (s *StockCollector) GetFearGreedIndex(ctx context.Context) (*FearGreedData, error) {
	url := "https://api.alternative.me/fng/?limit=1"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getSimulatedFearGreed(), nil
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Value               string `json:"value"`
			ValueClassification string `json:"value_classification"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getSimulatedFearGreed(), nil
	}

	if len(result.Data) == 0 {
		return s.getSimulatedFearGreed(), nil
	}

	value := 50
	fmt.Sscanf(result.Data[0].Value, "%d", &value)

	return &FearGreedData{
		Value:      value,
		ValueClass: result.Data[0].ValueClassification,
	}, nil
}

// getSimulatedFearGreed 시뮬레이션 데이터
func (s *StockCollector) getSimulatedFearGreed() *FearGreedData {
	return &FearGreedData{
		Value:      45,
		ValueClass: "Fear",
	}
}

// GetRecommendations 추천 종목 분석
func (s *StockCollector) GetRecommendations(cryptos []CryptoData, fearGreed *FearGreedData) []CryptoRecommendation {
	var recommendations []CryptoRecommendation

	for _, coin := range cryptos {
		score := 0.0
		var reasons []string
		signalType := "WATCH"

		// 1. 모멘텀 분석 (1h, 24h, 7d 상승 추세)
		if coin.Change1h > 0 && coin.Change24h > 0 && coin.Change7d > 0 {
			score += 25
			reasons = append(reasons, "상승 모멘텀 🚀")
		} else if coin.Change1h > 0 && coin.Change24h > 0 {
			score += 15
			reasons = append(reasons, "단기 상승세 📈")
		}

		// 2. ATH 대비 저평가 분석 (ATH 대비 -50% 이상 하락)
		if coin.ATHChangePerc < -50 && coin.ATHChangePerc > -80 {
			score += 20
			reasons = append(reasons, fmt.Sprintf("ATH 대비 %.0f%% 저평가", coin.ATHChangePerc))
		}

		// 3. 7일 상승률이 높은 경우
		if coin.Change7d > 10 {
			score += 15
			reasons = append(reasons, fmt.Sprintf("7일 +%.1f%% 급등", coin.Change7d))
		}

		// 4. 거래량 분석 (시총 대비 거래량 비율)
		volumeRatio := coin.Volume24h / coin.MarketCap * 100
		if volumeRatio > 10 {
			score += 15
			reasons = append(reasons, "거래량 폭발 🔥")
		} else if volumeRatio > 5 {
			score += 10
			reasons = append(reasons, "거래량 증가")
		}

		// 5. 안정성 (변동성이 적당한 경우)
		volatility := math.Abs(coin.Change24h)
		if volatility < 5 && coin.Change24h > 0 {
			score += 10
			reasons = append(reasons, "안정적 상승")
		}

		// 6. 공포탐욕지수 고려
		if fearGreed.Value < 30 && coin.Change24h < 0 {
			// 극도의 공포 상태에서 하락한 코인 = 저점 매수 기회
			score += 15
			reasons = append(reasons, "공포 속 기회 💎")
		}

		// 시그널 결정
		if score >= 50 {
			signalType = "BUY"
		} else if score >= 30 {
			signalType = "HOLD"
		}

		if len(reasons) > 0 {
			recommendations = append(recommendations, CryptoRecommendation{
				Coin:       coin,
				Score:      score,
				Reason:     strings.Join(reasons, ", "),
				SignalType: signalType,
			})
		}
	}

	// 점수순 정렬
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// 상위 5개만 반환
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return recommendations
}

// GenerateCryptoPost 코인 정보 포스트 생성 (풀 버전)
func (s *StockCollector) GenerateCryptoPost(cryptos []CryptoData) *Post {
	ctx := context.Background()
	now := time.Now()

	// 추가 데이터 수집
	marketData, _ := s.GetMarketData(ctx)
	fearGreed, _ := s.GetFearGreedIndex(ctx)
	recommendations := s.GetRecommendations(cryptos, fearGreed)

	// 시장 분석
	upCount := 0
	downCount := 0
	for _, c := range cryptos {
		if c.Change24h > 0 {
			upCount++
		} else {
			downCount--
		}
	}

	// 공포탐욕 색상
	fgColor := "#888"
	fgEmoji := "😐"
	switch {
	case fearGreed.Value <= 25:
		fgColor = "#e53935"
		fgEmoji = "😱"
	case fearGreed.Value <= 45:
		fgColor = "#ff9800"
		fgEmoji = "😨"
	case fearGreed.Value <= 55:
		fgColor = "#9e9e9e"
		fgEmoji = "😐"
	case fearGreed.Value <= 75:
		fgColor = "#8bc34a"
		fgEmoji = "😊"
	default:
		fgColor = "#4caf50"
		fgEmoji = "🤑"
	}

	title := fmt.Sprintf("[%s] 코인 시세 분석 📊 공포탐욕 %d | BTC 도미넌스 %.1f%%",
		now.Format("01/02"), fearGreed.Value, marketData.BTCDominance)

	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.crypto-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
.crypto-header { background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); color: #fff; padding: 30px; border-radius: 16px; margin-bottom: 20px; }
.crypto-header h1 { margin: 0 0 10px 0; font-size: 24px; }
.market-stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-top: 20px; }
.stat-box { background: rgba(255,255,255,0.1); padding: 15px; border-radius: 8px; text-align: center; }
.stat-value { font-size: 24px; font-weight: 700; }
.stat-label { font-size: 12px; opacity: 0.8; margin-top: 5px; }
.fear-greed { text-align: center; padding: 20px; margin: 20px 0; background: #f5f5f5; border-radius: 12px; }
.fear-greed .value { font-size: 64px; font-weight: 700; }
.fear-greed .label { font-size: 18px; margin-top: 10px; }
.fear-greed .bar { height: 10px; background: linear-gradient(to right, #e53935, #ff9800, #9e9e9e, #8bc34a, #4caf50); border-radius: 5px; margin-top: 15px; position: relative; }
.fear-greed .pointer { position: absolute; top: -5px; width: 20px; height: 20px; background: #333; border-radius: 50%; transform: translateX(-50%); }
.coin-table { width: 100%; border-collapse: collapse; margin: 20px 0; font-size: 14px; }
.coin-table th { background: #1a1a2e; color: #fff; padding: 12px 8px; text-align: left; }
.coin-table td { padding: 12px 8px; border-bottom: 1px solid #eee; }
.coin-table tr:hover { background: #f9f9f9; }
.coin-name { font-weight: 600; }
.coin-symbol { color: #666; font-size: 12px; }
.change-up { color: #4caf50; font-weight: 600; }
.change-down { color: #e53935; font-weight: 600; }
.sparkline { display: flex; align-items: end; height: 30px; gap: 1px; }
.sparkline-bar { width: 4px; background: #4caf50; border-radius: 2px; }
.recommendations { background: #fff3e0; padding: 25px; border-radius: 12px; margin: 20px 0; }
.recommendations h2 { margin: 0 0 15px 0; color: #e65100; }
.rec-card { background: #fff; padding: 15px; border-radius: 8px; margin-bottom: 10px; display: flex; justify-content: space-between; align-items: center; border-left: 4px solid #ff9800; }
.rec-coin { font-weight: 600; font-size: 16px; }
.rec-reason { font-size: 13px; color: #666; margin-top: 5px; }
.rec-signal { padding: 5px 12px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.signal-buy { background: #4caf50; color: #fff; }
.signal-hold { background: #ff9800; color: #fff; }
.signal-watch { background: #9e9e9e; color: #fff; }
.analysis-section { background: #f5f5f5; padding: 20px; border-radius: 12px; margin: 20px 0; }
.analysis-section h3 { margin: 0 0 15px 0; }
.analysis-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 15px; }
.analysis-item { background: #fff; padding: 15px; border-radius: 8px; }
.analysis-item .label { font-size: 12px; color: #666; }
.analysis-item .value { font-size: 18px; font-weight: 600; margin-top: 5px; }
.footer-notice { margin-top: 20px; padding: 15px; background: #ffebee; border-radius: 8px; font-size: 12px; color: #c62828; }
</style>
`)

	content.WriteString(`<div class="crypto-container">`)

	// 헤더
	content.WriteString(fmt.Sprintf(`
<div class="crypto-header">
	<h1>🪙 실시간 암호화폐 시세 분석</h1>
	<p>%s 업데이트</p>
	<div class="market-stats">
		<div class="stat-box">
			<div class="stat-value">%s</div>
			<div class="stat-label">전체 시가총액</div>
		</div>
		<div class="stat-box">
			<div class="stat-value">%s</div>
			<div class="stat-label">24시간 거래량</div>
		</div>
		<div class="stat-box">
			<div class="stat-value">%.1f%%</div>
			<div class="stat-label">BTC 도미넌스</div>
		</div>
		<div class="stat-box">
			<div class="stat-value" style="color: %s;">%+.1f%%</div>
			<div class="stat-label">24시간 변동</div>
		</div>
	</div>
</div>
`, now.Format("2006년 01월 02일 15:04"),
		formatNumber(marketData.TotalMarketCap),
		formatNumber(marketData.TotalVolume),
		marketData.BTCDominance,
		getChangeColor(marketData.MarketCapChange24h),
		marketData.MarketCapChange24h))

	// 공포탐욕지수
	content.WriteString(fmt.Sprintf(`
<div class="fear-greed">
	<div class="value" style="color: %s;">%s %d</div>
	<div class="label">공포 & 탐욕 지수: <strong>%s</strong></div>
	<div class="bar">
		<div class="pointer" style="left: %d%%;"></div>
	</div>
	<p style="font-size: 12px; color: #666; margin-top: 15px;">0 = 극도의 공포 | 100 = 극도의 탐욕</p>
</div>
`, fgColor, fgEmoji, fearGreed.Value, getFearGreedKorean(fearGreed.ValueClass), fearGreed.Value))

	// 추천 종목
	if len(recommendations) > 0 {
		content.WriteString(`
<div class="recommendations">
	<h2>🎯 AI 추천 종목 TOP 5</h2>
`)
		for _, rec := range recommendations {
			signalClass := "signal-watch"
			if rec.SignalType == "BUY" {
				signalClass = "signal-buy"
			} else if rec.SignalType == "HOLD" {
				signalClass = "signal-hold"
			}

			content.WriteString(fmt.Sprintf(`
	<div class="rec-card">
		<div>
			<div class="rec-coin">%s (%s)</div>
			<div class="rec-reason">%s</div>
		</div>
		<div>
			<span class="rec-signal %s">%s</span>
			<div style="font-size: 12px; color: #666; margin-top: 5px;">점수: %.0f</div>
		</div>
	</div>
`, rec.Coin.Name, rec.Coin.Symbol, rec.Reason, signalClass, rec.SignalType, rec.Score))
		}
		content.WriteString(`</div>`)
	}

	// 코인 테이블
	content.WriteString(`
<h2>📊 시가총액 TOP 10</h2>
<table class="coin-table">
<tr>
	<th>#</th>
	<th>코인</th>
	<th>현재가</th>
	<th>1시간</th>
	<th>24시간</th>
	<th>7일</th>
	<th>시가총액</th>
	<th>ATH 대비</th>
</tr>
`)

	for i, c := range cryptos {
		content.WriteString(fmt.Sprintf(`
<tr>
	<td>%d</td>
	<td><span class="coin-name">%s</span> <span class="coin-symbol">%s</span></td>
	<td>₩%s</td>
	<td class="%s">%+.1f%%</td>
	<td class="%s">%+.1f%%</td>
	<td class="%s">%+.1f%%</td>
	<td>%s</td>
	<td class="%s">%+.1f%%</td>
</tr>
`, i+1, c.Name, c.Symbol,
			formatNumber(c.Price),
			getChangeClass(c.Change1h), c.Change1h,
			getChangeClass(c.Change24h), c.Change24h,
			getChangeClass(c.Change7d), c.Change7d,
			formatNumber(c.MarketCap),
			getChangeClass(c.ATHChangePerc), c.ATHChangePerc))
	}

	content.WriteString(`</table>`)

	// 시장 분석
	content.WriteString(fmt.Sprintf(`
<div class="analysis-section">
	<h3>📈 시장 분석 요약</h3>
	<div class="analysis-grid">
		<div class="analysis-item">
			<div class="label">상승 코인</div>
			<div class="value" style="color: #4caf50;">%d개</div>
		</div>
		<div class="analysis-item">
			<div class="label">하락 코인</div>
			<div class="value" style="color: #e53935;">%d개</div>
		</div>
		<div class="analysis-item">
			<div class="label">시장 심리</div>
			<div class="value">%s</div>
		</div>
		<div class="analysis-item">
			<div class="label">ETH 도미넌스</div>
			<div class="value">%.1f%%</div>
		</div>
	</div>
</div>
`, upCount, len(cryptos)-upCount, getFearGreedKorean(fearGreed.ValueClass), marketData.ETHDominance))

	// 푸터
	content.WriteString(`
<div class="footer-notice">
	<p>⚠️ <strong>투자 주의사항</strong></p>
	<p>본 분석은 참고용이며 투자 권유가 아닙니다. 암호화폐 투자는 원금 손실 위험이 있으며, 모든 투자 결정과 책임은 본인에게 있습니다.</p>
	<p>데이터 출처: CoinGecko, Alternative.me</p>
</div>
`)

	content.WriteString(`</div>`)

	// 태그
	tags := []string{"비트코인", "이더리움", "코인시세", "암호화폐", "가상화폐", "코인분석", "공포탐욕지수"}
	for _, rec := range recommendations[:min(3, len(recommendations))] {
		tags = append(tags, rec.Coin.Name)
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "주식/코인",
		Tags:     tags,
	}
}

// 헬퍼 함수들
func getChangeColor(change float64) string {
	if change >= 0 {
		return "#4caf50"
	}
	return "#e53935"
}

func getChangeClass(change float64) string {
	if change >= 0 {
		return "change-up"
	}
	return "change-down"
}

func getFearGreedKorean(class string) string {
	switch class {
	case "Extreme Fear":
		return "극도의 공포"
	case "Fear":
		return "공포"
	case "Neutral":
		return "중립"
	case "Greed":
		return "탐욕"
	case "Extreme Greed":
		return "극도의 탐욕"
	default:
		return class
	}
}

func formatNumber(n float64) string {
	if n >= 1000000000000000 {
		return fmt.Sprintf("%.0f경", n/10000000000000000)
	}
	if n >= 1000000000000 {
		return fmt.Sprintf("%.1f조", n/1000000000000)
	}
	if n >= 100000000 {
		return fmt.Sprintf("%.1f억", n/100000000)
	}
	if n >= 10000 {
		return fmt.Sprintf("%.1f만", n/10000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.2f", n)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
