package collector

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// FortuneCollector 운세 정보 수집기
type FortuneCollector struct {
	coupangID string
}

// ZodiacFortune 띠별 운세
type ZodiacFortune struct {
	Zodiac       string
	Emoji        string
	Overall      int // 1-5
	Love         int
	Money        int
	Health       int
	Work         int
	LuckyItem    LuckyItem
	LuckyColor   string
	LuckyNumber  int
	Message      string
	Advice       string
}

// LuckyItem 행운의 아이템 (쿠팡 연동)
type LuckyItem struct {
	Name        string
	SearchQuery string // 쿠팡 검색어
	Emoji       string
	Category    string
}

func NewFortuneCollector(coupangID string) *FortuneCollector {
	return &FortuneCollector{coupangID: coupangID}
}

// 띠 목록
var zodiacs = []struct {
	Name      string
	Emoji     string
	Element   string // 오행
	Character string // 성격
}{
	{"쥐띠", "🐭", "수(水)", "지혜롭고 민첩함"},
	{"소띠", "🐮", "토(土)", "성실하고 인내심 강함"},
	{"호랑이띠", "🐯", "목(木)", "용감하고 자신감 넘침"},
	{"토끼띠", "🐰", "목(木)", "온화하고 섬세함"},
	{"용띠", "🐲", "토(土)", "카리스마 있고 야망적"},
	{"뱀띠", "🐍", "화(火)", "지혜롭고 신비로움"},
	{"말띠", "🐴", "화(火)", "활동적이고 자유로움"},
	{"양띠", "🐑", "토(土)", "온순하고 예술적"},
	{"원숭이띠", "🐵", "금(金)", "영리하고 재치있음"},
	{"닭띠", "🐔", "금(金)", "부지런하고 용감함"},
	{"개띠", "🐶", "토(土)", "충성스럽고 정직함"},
	{"돼지띠", "🐷", "수(水)", "관대하고 성실함"},
}

// 행운의 아이템 풀 (실제 상품으로 연결 가능)
var luckyItemPool = []LuckyItem{
	// 패션/액세서리
	{Name: "빨간색 머플러", SearchQuery: "빨간 머플러", Emoji: "🧣", Category: "패션"},
	{Name: "골드 목걸이", SearchQuery: "골드 목걸이", Emoji: "📿", Category: "액세서리"},
	{Name: "가죽 지갑", SearchQuery: "가죽 지갑 남성", Emoji: "👛", Category: "패션"},
	{Name: "실크 스카프", SearchQuery: "실크 스카프", Emoji: "🎀", Category: "패션"},
	{Name: "행운의 팔찌", SearchQuery: "행운 팔찌", Emoji: "📿", Category: "액세서리"},

	// 음료/식품
	{Name: "프리미엄 커피", SearchQuery: "원두커피 선물세트", Emoji: "☕", Category: "음료"},
	{Name: "녹차 세트", SearchQuery: "녹차 선물세트", Emoji: "🍵", Category: "음료"},
	{Name: "꿀 한 병", SearchQuery: "천연 벌꿀", Emoji: "🍯", Category: "식품"},
	{Name: "비타민", SearchQuery: "종합비타민", Emoji: "💊", Category: "건강"},

	// 인테리어/생활
	{Name: "미니 화분", SearchQuery: "미니 화분 세트", Emoji: "🌱", Category: "인테리어"},
	{Name: "아로마 캔들", SearchQuery: "아로마 캔들", Emoji: "🕯️", Category: "인테리어"},
	{Name: "행운목", SearchQuery: "행운목 화분", Emoji: "🌿", Category: "인테리어"},
	{Name: "수정 장식", SearchQuery: "수정 인테리어", Emoji: "💎", Category: "인테리어"},
	{Name: "풍수 거울", SearchQuery: "풍수 거울", Emoji: "🪞", Category: "인테리어"},

	// 문구/소품
	{Name: "고급 만년필", SearchQuery: "만년필 선물", Emoji: "🖋️", Category: "문구"},
	{Name: "다이어리", SearchQuery: "2025 다이어리", Emoji: "📔", Category: "문구"},
	{Name: "행운의 열쇠고리", SearchQuery: "행운 키링", Emoji: "🔑", Category: "소품"},

	// 건강/운동
	{Name: "요가 매트", SearchQuery: "요가매트", Emoji: "🧘", Category: "운동"},
	{Name: "마사지 볼", SearchQuery: "마사지볼", Emoji: "⚽", Category: "건강"},
	{Name: "아이마스크", SearchQuery: "수면 안대", Emoji: "😴", Category: "건강"},
}

// 색상 풀
var luckyColors = []struct {
	Name  string
	Emoji string
}{
	{"빨간색", "🔴"},
	{"파란색", "🔵"},
	{"노란색", "🟡"},
	{"초록색", "🟢"},
	{"보라색", "🟣"},
	{"주황색", "🟠"},
	{"흰색", "⚪"},
	{"검정색", "⚫"},
	{"분홍색", "🩷"},
	{"하늘색", "🩵"},
}

// 운세 메시지 풀 (띠별 특성에 맞게)
var fortuneMessagePool = map[string][]string{
	"positive": {
		"오늘은 당신의 매력이 빛나는 날입니다. 자신감을 가지세요!",
		"좋은 기운이 가득합니다. 새로운 도전에 적극적으로 나서보세요.",
		"귀인이 나타날 수 있습니다. 인연을 소중히 여기세요.",
		"창의적인 아이디어가 샘솟는 날입니다. 놓치지 마세요!",
		"행운이 함께하는 하루입니다. 복권 구매도 좋은 날!",
	},
	"neutral": {
		"평온한 하루가 예상됩니다. 급하게 서두르지 마세요.",
		"주변을 돌아보는 여유가 필요한 날입니다.",
		"작은 것에서 행복을 찾아보세요.",
		"계획을 세우기 좋은 날입니다. 미래를 준비하세요.",
		"오늘은 휴식과 재충전에 집중하세요.",
	},
	"careful": {
		"급한 결정은 피하세요. 신중함이 필요합니다.",
		"건강에 특별히 신경 쓰는 것이 좋겠습니다.",
		"금전 관련 결정은 다음으로 미루세요.",
		"오해가 생기기 쉬운 날입니다. 말조심하세요.",
		"혼자만의 시간이 필요한 하루입니다.",
	},
}

// 조언 풀
var advicePool = []string{
	"아침에 따뜻한 물 한 잔으로 하루를 시작하세요.",
	"오늘은 감사 일기를 써보는 건 어떨까요?",
	"잠시 멈추고 심호흡을 해보세요.",
	"오늘 만나는 사람에게 먼저 인사해보세요.",
	"작은 목표 하나를 정하고 달성해보세요.",
	"퇴근 후 가벼운 산책을 추천합니다.",
	"좋아하는 음악을 들으며 휴식하세요.",
	"오랜만에 연락 못한 친구에게 안부 전해보세요.",
	"오늘의 작은 성취를 스스로 칭찬해주세요.",
	"자기 전 5분 명상으로 마음을 정리하세요.",
}

// GetTodayFortune 오늘의 띠별 운세 생성
func (f *FortuneCollector) GetTodayFortune() []ZodiacFortune {
	now := time.Now()

	// 날짜 + 시간 기반 시드 (시간대별로 다른 결과)
	baseSeed := now.Year()*10000 + int(now.Month())*100 + now.Day()

	var fortunes []ZodiacFortune
	for i, zodiac := range zodiacs {
		// 각 띠별로 완전히 다른 시드 생성
		// 띠 인덱스 * 큰 소수를 사용하여 확실히 다른 시드
		zodiacSeed := int64(baseSeed*1000 + i*127 + len(zodiac.Name)*31)
		rng := rand.New(rand.NewSource(zodiacSeed))

		// 점수 생성 (1-5, 띠별로 다른 분포)
		overall := f.generateScore(rng, i)
		love := f.generateScore(rng, i+1)
		money := f.generateScore(rng, i+2)
		health := f.generateScore(rng, i+3)
		work := f.generateScore(rng, i+4)

		// 행운의 아이템 (띠별로 다르게)
		luckyItem := luckyItemPool[(baseSeed+i*17)%len(luckyItemPool)]

		// 행운의 색상
		luckyColor := luckyColors[(baseSeed+i*13)%len(luckyColors)]

		// 행운의 숫자 (1-45, 로또 범위)
		luckyNumber := (baseSeed+i*7)%45 + 1

		// 메시지 선택 (점수에 따라)
		avgScore := (overall + love + money + health + work) / 5
		var messageType string
		if avgScore >= 4 {
			messageType = "positive"
		} else if avgScore >= 2 {
			messageType = "neutral"
		} else {
			messageType = "careful"
		}
		messages := fortuneMessagePool[messageType]
		message := messages[(baseSeed+i*23)%len(messages)]

		// 조언
		advice := advicePool[(baseSeed+i*11)%len(advicePool)]

		fortune := ZodiacFortune{
			Zodiac:      zodiac.Name,
			Emoji:       zodiac.Emoji,
			Overall:     overall,
			Love:        love,
			Money:       money,
			Health:      health,
			Work:        work,
			LuckyItem:   luckyItem,
			LuckyColor:  fmt.Sprintf("%s %s", luckyColor.Emoji, luckyColor.Name),
			LuckyNumber: luckyNumber,
			Message:     message,
			Advice:      advice,
		}
		fortunes = append(fortunes, fortune)
	}

	return fortunes
}

// generateScore 점수 생성 (약간의 변동성 추가)
func (f *FortuneCollector) generateScore(rng *rand.Rand, offset int) int {
	base := rng.Intn(5) + 1
	// 오프셋에 따라 조금씩 다르게
	adjustment := (offset % 3) - 1 // -1, 0, 1
	result := base + adjustment
	if result < 1 {
		result = 1
	}
	if result > 5 {
		result = 5
	}
	return result
}

// GenerateFortunePost 운세 포스트 생성
func (f *FortuneCollector) GenerateFortunePost(fortunes []ZodiacFortune) *Post {
	now := time.Now()
	title := fmt.Sprintf("🔮 [%s] 오늘의 띠별 운세 & 행운 아이템 추천", now.Format("01/02"))

	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.fortune-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, sans-serif; }
.fortune-header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 20px; color: white; text-align: center; margin-bottom: 25px; }
.fortune-header h1 { margin: 0; font-size: 28px; }
.fortune-header p { margin: 10px 0 0 0; opacity: 0.9; }
.zodiac-card { background: #fff; border-radius: 16px; padding: 25px; margin-bottom: 20px; box-shadow: 0 4px 15px rgba(0,0,0,0.08); border-left: 5px solid #667eea; }
.zodiac-header { display: flex; align-items: center; gap: 15px; margin-bottom: 15px; }
.zodiac-emoji { font-size: 48px; }
.zodiac-name { font-size: 24px; font-weight: 700; color: #2d3436; }
.zodiac-element { font-size: 14px; color: #636e72; }
.score-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 10px; margin: 20px 0; }
.score-item { text-align: center; padding: 15px 10px; background: #f8f9fa; border-radius: 10px; }
.score-label { font-size: 12px; color: #636e72; margin-bottom: 5px; }
.score-stars { font-size: 14px; color: #f1c40f; }
.lucky-section { display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; margin: 20px 0; }
.lucky-item { background: linear-gradient(135deg, #fff9e6 0%, #fff3cd 100%); padding: 15px; border-radius: 12px; text-align: center; }
.lucky-label { font-size: 12px; color: #856404; margin-bottom: 5px; }
.lucky-value { font-size: 16px; font-weight: 600; color: #533f03; }
.message-box { background: #e8f4fd; padding: 20px; border-radius: 12px; margin: 15px 0; }
.message-text { font-size: 16px; color: #1565c0; margin: 0; line-height: 1.6; }
.advice-box { background: #f0fff4; padding: 15px; border-radius: 10px; border-left: 4px solid #38a169; }
.advice-text { font-size: 14px; color: #276749; margin: 0; }
.product-recommend { background: linear-gradient(135deg, #fff5f5 0%, #ffe3e3 100%); padding: 20px; border-radius: 12px; margin-top: 15px; }
.product-title { font-size: 14px; color: #c53030; margin: 0 0 10px 0; }
.product-link { display: inline-block; background: #e53e3e; color: white; padding: 10px 20px; border-radius: 8px; text-decoration: none; font-weight: 600; }
.product-link:hover { background: #c53030; }
.footer-notice { margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 12px; text-align: center; }
</style>
`)

	content.WriteString(fmt.Sprintf(`
<div class="fortune-container">
<div class="fortune-header">
	<h1>🔮 오늘의 띠별 운세</h1>
	<p>%s | 행운의 아이템과 함께하는 특별한 하루</p>
</div>
`, now.Format("2006년 01월 02일 (Mon)")))

	for _, fortune := range fortunes {
		// 총점 계산
		avgScore := (fortune.Overall + fortune.Love + fortune.Money + fortune.Health + fortune.Work) / 5

		// 띠 특성 찾기
		var element string
		for _, z := range zodiacs {
			if z.Name == fortune.Zodiac {
				element = z.Element
				break
			}
		}

		content.WriteString(fmt.Sprintf(`
<div class="zodiac-card">
	<div class="zodiac-header">
		<span class="zodiac-emoji">%s</span>
		<div>
			<div class="zodiac-name">%s</div>
			<div class="zodiac-element">%s | 오늘의 종합운 %s</div>
		</div>
	</div>

	<div class="score-grid">
		<div class="score-item">
			<div class="score-label">종합운</div>
			<div class="score-stars">%s</div>
		</div>
		<div class="score-item">
			<div class="score-label">💕 애정</div>
			<div class="score-stars">%s</div>
		</div>
		<div class="score-item">
			<div class="score-label">💰 금전</div>
			<div class="score-stars">%s</div>
		</div>
		<div class="score-item">
			<div class="score-label">💪 건강</div>
			<div class="score-stars">%s</div>
		</div>
		<div class="score-item">
			<div class="score-label">💼 직장</div>
			<div class="score-stars">%s</div>
		</div>
	</div>

	<div class="lucky-section">
		<div class="lucky-item">
			<div class="lucky-label">🍀 행운의 아이템</div>
			<div class="lucky-value">%s %s</div>
		</div>
		<div class="lucky-item">
			<div class="lucky-label">🎨 행운의 색상</div>
			<div class="lucky-value">%s</div>
		</div>
		<div class="lucky-item">
			<div class="lucky-label">🔢 행운의 숫자</div>
			<div class="lucky-value">%d</div>
		</div>
	</div>

	<div class="message-box">
		<p class="message-text">💬 %s</p>
	</div>

	<div class="advice-box">
		<p class="advice-text">💡 오늘의 조언: %s</p>
	</div>

	<div class="product-recommend">
		<p class="product-title">%s 오늘의 행운 아이템 쇼핑하기</p>
		<a href="%s" target="_blank" class="product-link">🛒 %s 보러가기</a>
	</div>
</div>
`, fortune.Emoji, fortune.Zodiac, element, getGradeText(avgScore),
			getStarRating(fortune.Overall),
			getStarRating(fortune.Love),
			getStarRating(fortune.Money),
			getStarRating(fortune.Health),
			getStarRating(fortune.Work),
			fortune.LuckyItem.Emoji, fortune.LuckyItem.Name,
			fortune.LuckyColor,
			fortune.LuckyNumber,
			fortune.Message,
			fortune.Advice,
			fortune.Emoji,
			f.generateCoupangSearchLink(fortune.LuckyItem.SearchQuery),
			fortune.LuckyItem.Name,
		))
	}

	content.WriteString(`
<div class="footer-notice">
	<p>🔮 운세는 재미로만 봐주세요!</p>
	<p>오늘 하루도 행복하고 건강한 하루 되세요! ✨</p>
	<p style="font-size: 12px; color: #888; margin-top: 10px;">
	⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.
	</p>
</div>
</div>
`)

	// 동적 태그 생성
	tags := []string{
		"오늘의운세", "띠별운세", "운세",
		now.Format("01월02일") + "운세", now.Format("2006년") + "운세",
		"무료운세", "오늘운세", "일일운세",
	}
	// 띠별 태그 추가
	for _, fortune := range fortunes {
		tags = append(tags, fortune.Zodiac+"운세")
	}
	// 행운 아이템 태그
	for _, fortune := range fortunes[:min(3, len(fortunes))] {
		tags = append(tags, fortune.LuckyItem.Name)
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "운세/점술",
		Tags:     tags,
	}
}

// generateCoupangSearchLink 쿠팡 검색 링크 생성
func (f *FortuneCollector) generateCoupangSearchLink(query string) string {
	baseURL := fmt.Sprintf("https://www.coupang.com/np/search?component=&q=%s", query)
	if f.coupangID != "" {
		return fmt.Sprintf("%s&channel=affiliate&affiliate=%s", baseURL, f.coupangID)
	}
	return baseURL
}

func getStarRating(n int) string {
	return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
}

func getGradeText(score int) string {
	switch score {
	case 5:
		return "🌟 최고의 하루!"
	case 4:
		return "😊 좋은 하루"
	case 3:
		return "😐 평범한 하루"
	case 2:
		return "😓 조심하는 하루"
	default:
		return "🙏 신중한 하루"
	}
}
