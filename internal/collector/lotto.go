package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
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
	// 계산 오차를 방지하기 위해 여유있게 계산 후 실제 데이터로 확인
	startDate := time.Date(2002, 12, 7, 0, 0, 0, 0, time.Local)
	now := time.Now()
	weeks := int(now.Sub(startDate).Hours()/24/7) + 5 // 5회 여유분 추가

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

// LottoPrediction 예측 번호 세트
type LottoPrediction struct {
	Name    string
	Numbers []int
	Method  string
}

// NumberStats 번호 통계
type NumberStats struct {
	Number    int
	Frequency int
	LastDrawn int // 마지막 출현 회차
}

// GetRecentResults 최근 N회차 결과 조회
func (l *LottoCollector) GetRecentResults(ctx context.Context, count int) ([]LottoResult, error) {
	latest, err := l.GetLatestLotto(ctx)
	if err != nil {
		return nil, err
	}

	var results []LottoResult
	for i := 0; i < count && (latest.DrawNo-i) > 0; i++ {
		result, err := l.GetLottoByRound(ctx, latest.DrawNo-i)
		if err != nil {
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// AnalyzeNumbers 번호 분석 (최근 N회차 기준)
func (l *LottoCollector) AnalyzeNumbers(results []LottoResult) (hotNumbers []int, coldNumbers []int) {
	// 번호별 출현 빈도 계산
	frequency := make(map[int]int)
	for i := 1; i <= 45; i++ {
		frequency[i] = 0
	}

	for _, r := range results {
		numbers := []int{r.Number1, r.Number2, r.Number3, r.Number4, r.Number5, r.Number6}
		for _, n := range numbers {
			frequency[n]++
		}
	}

	// 정렬을 위한 슬라이스 생성
	var stats []NumberStats
	for num, freq := range frequency {
		stats = append(stats, NumberStats{Number: num, Frequency: freq})
	}

	// 빈도순 정렬
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Frequency > stats[j].Frequency
	})

	// 핫넘버 (상위 10개)
	for i := 0; i < 10 && i < len(stats); i++ {
		hotNumbers = append(hotNumbers, stats[i].Number)
	}

	// 콜드넘버 (하위 10개)
	for i := len(stats) - 1; i >= len(stats)-10 && i >= 0; i-- {
		coldNumbers = append(coldNumbers, stats[i].Number)
	}

	return hotNumbers, coldNumbers
}

// GeneratePredictions 예측 번호 생성 (5세트) - 계정별 다른 번호
func (l *LottoCollector) GeneratePredictions(hotNumbers, coldNumbers []int, accountName string) []LottoPrediction {
	// 날짜 + 계정 기반 시드 (같은 날 같은 계정은 같은 예측, 다른 계정은 다른 예측)
	today := time.Now().Format("2006-01-02")
	seed := int64(0)
	for _, c := range today {
		seed += int64(c)
	}
	// 계정 이름도 시드에 추가
	for _, c := range accountName {
		seed += int64(c) * 7 // 다른 가중치 적용
	}
	rng := rand.New(rand.NewSource(seed))

	var predictions []LottoPrediction

	// 1. 완전 랜덤
	predictions = append(predictions, LottoPrediction{
		Name:    "🎲 완전 랜덤",
		Numbers: generateRandomNumbers(rng),
		Method:  "1~45 중 무작위 6개",
	})

	// 2. 핫넘버 기반
	predictions = append(predictions, LottoPrediction{
		Name:    "🔥 핫넘버 조합",
		Numbers: generateFromPool(rng, hotNumbers, 45),
		Method:  "최근 자주 나온 번호 중심",
	})

	// 3. 콜드넘버 기반
	predictions = append(predictions, LottoPrediction{
		Name:    "❄️ 콜드넘버 조합",
		Numbers: generateFromPool(rng, coldNumbers, 45),
		Method:  "최근 안 나온 번호 중심",
	})

	// 4. 균형 조합 (핫+콜드)
	predictions = append(predictions, LottoPrediction{
		Name:    "⚖️ 균형 조합",
		Numbers: generateBalanced(rng, hotNumbers, coldNumbers),
		Method:  "핫넘버 3개 + 콜드넘버 3개",
	})

	// 5. 고저 균형
	predictions = append(predictions, LottoPrediction{
		Name:    "📊 고저 균형",
		Numbers: generateHighLowBalance(rng),
		Method:  "저번호(1-22) 3개 + 고번호(23-45) 3개",
	})

	return predictions
}

// generateRandomNumbers 완전 랜덤 6개
func generateRandomNumbers(rng *rand.Rand) []int {
	numbers := make(map[int]bool)
	var result []int

	for len(result) < 6 {
		n := rng.Intn(45) + 1
		if !numbers[n] {
			numbers[n] = true
			result = append(result, n)
		}
	}

	sort.Ints(result)
	return result
}

// generateFromPool 특정 풀에서 우선 선택
func generateFromPool(rng *rand.Rand, pool []int, max int) []int {
	numbers := make(map[int]bool)
	var result []int

	// 풀에서 4개 선택
	shuffled := make([]int, len(pool))
	copy(shuffled, pool)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i := 0; i < 4 && i < len(shuffled); i++ {
		numbers[shuffled[i]] = true
		result = append(result, shuffled[i])
	}

	// 나머지 2개는 랜덤
	for len(result) < 6 {
		n := rng.Intn(max) + 1
		if !numbers[n] {
			numbers[n] = true
			result = append(result, n)
		}
	}

	sort.Ints(result)
	return result
}

// generateBalanced 핫/콜드 균형
func generateBalanced(rng *rand.Rand, hot, cold []int) []int {
	numbers := make(map[int]bool)
	var result []int

	// 핫에서 3개
	shuffledHot := make([]int, len(hot))
	copy(shuffledHot, hot)
	rng.Shuffle(len(shuffledHot), func(i, j int) {
		shuffledHot[i], shuffledHot[j] = shuffledHot[j], shuffledHot[i]
	})
	for i := 0; i < 3 && i < len(shuffledHot); i++ {
		if !numbers[shuffledHot[i]] {
			numbers[shuffledHot[i]] = true
			result = append(result, shuffledHot[i])
		}
	}

	// 콜드에서 3개
	shuffledCold := make([]int, len(cold))
	copy(shuffledCold, cold)
	rng.Shuffle(len(shuffledCold), func(i, j int) {
		shuffledCold[i], shuffledCold[j] = shuffledCold[j], shuffledCold[i]
	})
	for i := 0; i < 3 && i < len(shuffledCold) && len(result) < 6; i++ {
		if !numbers[shuffledCold[i]] {
			numbers[shuffledCold[i]] = true
			result = append(result, shuffledCold[i])
		}
	}

	// 부족하면 랜덤 추가
	for len(result) < 6 {
		n := rng.Intn(45) + 1
		if !numbers[n] {
			numbers[n] = true
			result = append(result, n)
		}
	}

	sort.Ints(result)
	return result
}

// generateHighLowBalance 고저 균형
func generateHighLowBalance(rng *rand.Rand) []int {
	numbers := make(map[int]bool)
	var result []int

	// 저번호 (1-22) 3개
	for len(result) < 3 {
		n := rng.Intn(22) + 1
		if !numbers[n] {
			numbers[n] = true
			result = append(result, n)
		}
	}

	// 고번호 (23-45) 3개
	for len(result) < 6 {
		n := rng.Intn(23) + 23
		if !numbers[n] {
			numbers[n] = true
			result = append(result, n)
		}
	}

	sort.Ints(result)
	return result
}

// GeneratePredictionPost 예측 번호 포스트 생성
func (l *LottoCollector) GeneratePredictionPost(nextRound int, predictions []LottoPrediction, hotNumbers, coldNumbers []int) *Post {
	now := time.Now()
	title := fmt.Sprintf("🔮 %d회 로또 예측번호 [%s] AI 분석 추천", nextRound, now.Format("01/02"))

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>🔮 %d회 로또 예측번호</h2>
<p>분석일: %s</p>

<div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 20px; border-radius: 15px; color: white; margin: 20px 0; text-align: center;">
<p style="font-size: 1.3em; margin: 0;">✨ 이번 주 행운의 번호를 확인하세요! ✨</p>
</div>
`, nextRound, now.Format("2006년 01월 02일")))

	// 예측 번호 표시
	content.WriteString(`<h3>🎯 예측 번호 5세트</h3>`)

	for i, pred := range predictions {
		content.WriteString(fmt.Sprintf(`
<div style="background: #f8f9fa; padding: 20px; border-radius: 10px; margin-bottom: 15px; border-left: 5px solid %s;">
<h4 style="margin-top: 0;">%s</h4>
<p style="color: #666; font-size: 0.9em;">%s</p>
<div style="display: flex; gap: 8px; flex-wrap: wrap; margin-top: 10px;">
`, getPredictionColor(i), pred.Name, pred.Method))

		for _, num := range pred.Numbers {
			color := getLottoBallColor(num)
			content.WriteString(fmt.Sprintf(`<span style="display: inline-block; width: 45px; height: 45px; border-radius: 50%%; background: %s; color: white; font-size: 18px; font-weight: bold; line-height: 45px; text-align: center; text-shadow: 1px 1px 2px rgba(0,0,0,0.3);">%d</span>
`, color, num))
		}

		content.WriteString(`</div>
</div>
`)
	}

	// 번호 분석 정보
	content.WriteString(`
<h3>📈 최근 번호 분석 (20회차 기준)</h3>

<div style="display: flex; gap: 20px; flex-wrap: wrap;">
<div style="flex: 1; min-width: 200px; background: #fff3e0; padding: 15px; border-radius: 10px;">
<h4 style="color: #e65100; margin-top: 0;">🔥 핫넘버 (자주 출현)</h4>
<p style="font-size: 1.2em; font-weight: bold;">`)

	for i, n := range hotNumbers {
		if i > 0 {
			content.WriteString(", ")
		}
		content.WriteString(fmt.Sprintf("%d", n))
	}

	content.WriteString(`</p>
</div>

<div style="flex: 1; min-width: 200px; background: #e3f2fd; padding: 15px; border-radius: 10px;">
<h4 style="color: #1565c0; margin-top: 0;">❄️ 콜드넘버 (적게 출현)</h4>
<p style="font-size: 1.2em; font-weight: bold;">`)

	for i, n := range coldNumbers {
		if i > 0 {
			content.WriteString(", ")
		}
		content.WriteString(fmt.Sprintf("%d", n))
	}

	content.WriteString(`</p>
</div>
</div>

<h3>💡 로또 당첨 꿀팁</h3>
<ul>
<li>홀수/짝수 비율은 3:3 또는 4:2가 가장 많이 당첨</li>
<li>연속 번호는 1~2개 정도 포함되는 경우가 많음</li>
<li>같은 번호대(1~10, 11~20 등)에서 3개 이상은 드묾</li>
<li>총합이 100~175 사이인 경우가 가장 많음</li>
</ul>

<div style="background: #ffebee; padding: 15px; border-radius: 10px; margin-top: 20px;">
<p style="color: #c62828; margin: 0;">
⚠️ <strong>주의:</strong> 로또는 순수 확률 게임입니다. 예측 번호는 참고용이며, 당첨을 보장하지 않습니다.<br>
무리한 구매는 삼가해주시고, 즐거운 마음으로 참여하세요! 🍀
</p>
</div>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "로또/복권",
		Tags:     []string{"로또예측", "로또번호추천", fmt.Sprintf("%d회로또예측", nextRound), "로또분석", "행운의번호"},
	}
}

func getPredictionColor(index int) string {
	colors := []string{"#667eea", "#f093fb", "#4facfe", "#43e97b", "#fa709a"}
	return colors[index%len(colors)]
}
