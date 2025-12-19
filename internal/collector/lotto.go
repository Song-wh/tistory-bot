package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LottoCollector 로또 정보 수집기
type LottoCollector struct {
	client *http.Client
}

// LottoResult 로또 당첨 결과
type LottoResult struct {
	DrawNo      int    `json:"drwNo"`
	DrawDate    string `json:"drwNoDate"`
	Number1     int    `json:"drwtNo1"`
	Number2     int    `json:"drwtNo2"`
	Number3     int    `json:"drwtNo3"`
	Number4     int    `json:"drwtNo4"`
	Number5     int    `json:"drwtNo5"`
	Number6     int    `json:"drwtNo6"`
	BonusNumber int    `json:"bnusNo"`
	Prize1      int64  `json:"firstWinamnt"`
	Winner1     int    `json:"firstPrzwnerCo"`
}

func NewLottoCollector() *LottoCollector {
	return &LottoCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetLatestLotto 최신 로또 당첨번호 조회
func (l *LottoCollector) GetLatestLotto(ctx context.Context) (*LottoResult, error) {
	// 최신 회차 계산 (2002년 12월 7일 1회차 기준)
	startDate := time.Date(2002, 12, 7, 0, 0, 0, 0, time.Local)
	now := time.Now()
	weeks := int(now.Sub(startDate).Hours() / 24 / 7)
	
	return l.GetLottoByRound(ctx, weeks)
}

// GetLottoByRound 특정 회차 로또 당첨번호 조회
func (l *LottoCollector) GetLottoByRound(ctx context.Context, round int) (*LottoResult, error) {
	url := fmt.Sprintf("https://www.dhlottery.co.kr/common.do?method=getLottoNumber&drwNo=%d", round)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result LottoResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	if result.DrawNo == 0 {
		// 아직 추첨 전이면 이전 회차 조회
		return l.GetLottoByRound(ctx, round-1)
	}
	
	return &result, nil
}

// GenerateLottoPost 로또 포스트 생성
func (l *LottoCollector) GenerateLottoPost(result *LottoResult) *Post {
	title := fmt.Sprintf("🎰 %d회 로또 당첨번호 [%s]", result.DrawNo, result.DrawDate)
	
	// 번호 슬라이스
	numbers := []int{result.Number1, result.Number2, result.Number3, result.Number4, result.Number5, result.Number6}
	
	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>🎰 %d회 로또 당첨번호</h2>
<p>추첨일: %s</p>

<div style="background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); padding: 30px; border-radius: 15px; text-align: center; margin: 20px 0;">
<h3 style="color: #eee; margin-bottom: 20px;">당첨번호</h3>
<div style="display: flex; justify-content: center; gap: 10px; flex-wrap: wrap;">
`, result.DrawNo, result.DrawDate))

	// 번호별 색상
	for _, num := range numbers {
		color := getLottoBallColor(num)
		content.WriteString(fmt.Sprintf(`<span style="display: inline-block; width: 50px; height: 50px; border-radius: 50%%; background: %s; color: white; font-size: 20px; font-weight: bold; line-height: 50px; text-shadow: 1px 1px 2px rgba(0,0,0,0.5);">%d</span>
`, color, num))
	}
	
	content.WriteString(fmt.Sprintf(`<span style="color: #eee; font-size: 24px; line-height: 50px; margin: 0 10px;">+</span>
<span style="display: inline-block; width: 50px; height: 50px; border-radius: 50%%; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; font-size: 20px; font-weight: bold; line-height: 50px; border: 3px solid gold;">%d</span>
</div>
<p style="color: #aaa; margin-top: 10px;">보너스 번호</p>
</div>
`, result.BonusNumber))

	// 당첨금 정보
	content.WriteString(fmt.Sprintf(`
<h3>💰 1등 당첨 정보</h3>
<table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
<tr style="background: #f5f5f5;">
<td style="padding: 15px; border: 1px solid #ddd; text-align: center;"><strong>1등 당첨금</strong></td>
<td style="padding: 15px; border: 1px solid #ddd; text-align: center; font-size: 1.2em; color: #e74c3c;"><strong>%s원</strong></td>
</tr>
<tr>
<td style="padding: 15px; border: 1px solid #ddd; text-align: center;"><strong>1등 당첨자 수</strong></td>
<td style="padding: 15px; border: 1px solid #ddd; text-align: center;"><strong>%d명</strong></td>
</tr>
</table>
`, formatMoney(result.Prize1), result.Winner1))

	content.WriteString(`
<h3>📊 번호 분석</h3>
<ul>
<li>홀수/짝수 비율 분석</li>
<li>고저 번호 분포</li>
<li>연속 번호 여부</li>
</ul>

<p style="color: #888; font-size: 0.9em; margin-top: 30px;">
※ 로또는 확률 게임입니다. 무리한 구매는 삼가해주세요.<br>
※ 공식 결과는 동행복권 사이트에서 확인하세요.
</p>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "로또/복권",
		Tags:     []string{"로또", "로또당첨번호", fmt.Sprintf("%d회로또", result.DrawNo), "복권", "당첨번호"},
	}
}

// getLottoBallColor 로또 공 색상 반환
func getLottoBallColor(num int) string {
	switch {
	case num <= 10:
		return "#fbc400" // 노랑
	case num <= 20:
		return "#69c8f2" // 파랑
	case num <= 30:
		return "#ff7272" // 빨강
	case num <= 40:
		return "#aaa" // 회색
	default:
		return "#b0d840" // 초록
	}
}

// formatMoney 금액 포맷팅
func formatMoney(n int64) string {
	str := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

