package collector

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// FortuneCollector 운세 정보 수집기
type FortuneCollector struct{}

// ZodiacFortune 띠별 운세
type ZodiacFortune struct {
	Zodiac  string
	Emoji   string
	Overall int // 1-5
	Love    int
	Money   int
	Health  int
	Lucky   string
	Message string
}

func NewFortuneCollector() *FortuneCollector {
	return &FortuneCollector{}
}

// 띠 목록
var zodiacs = []struct {
	Name  string
	Emoji string
	Years []int
}{
	{"쥐띠", "🐭", []int{1960, 1972, 1984, 1996, 2008, 2020}},
	{"소띠", "🐮", []int{1961, 1973, 1985, 1997, 2009, 2021}},
	{"호랑이띠", "🐯", []int{1962, 1974, 1986, 1998, 2010, 2022}},
	{"토끼띠", "🐰", []int{1963, 1975, 1987, 1999, 2011, 2023}},
	{"용띠", "🐲", []int{1964, 1976, 1988, 2000, 2012, 2024}},
	{"뱀띠", "🐍", []int{1965, 1977, 1989, 2001, 2013, 2025}},
	{"말띠", "🐴", []int{1966, 1978, 1990, 2002, 2014, 2026}},
	{"양띠", "🐑", []int{1967, 1979, 1991, 2003, 2015, 2027}},
	{"원숭이띠", "🐵", []int{1968, 1980, 1992, 2004, 2016, 2028}},
	{"닭띠", "🐔", []int{1969, 1981, 1993, 2005, 2017, 2029}},
	{"개띠", "🐶", []int{1970, 1982, 1994, 2006, 2018, 2030}},
	{"돼지띠", "🐷", []int{1971, 1983, 1995, 2007, 2019, 2031}},
}

// 운세 메시지 풀
var fortuneMessages = []string{
	"오늘은 새로운 시작을 하기 좋은 날입니다. 용기를 내어 도전해보세요.",
	"주변 사람들과의 소통이 중요한 하루입니다. 경청하는 자세가 행운을 가져옵니다.",
	"예상치 못한 기회가 찾아올 수 있습니다. 열린 마음으로 받아들이세요.",
	"재정적인 결정은 신중하게 내리세요. 충동구매는 피하는 것이 좋습니다.",
	"건강에 신경 쓰는 하루가 되세요. 가벼운 운동이 도움이 됩니다.",
	"창의적인 아이디어가 떠오르는 날입니다. 메모해두면 좋겠습니다.",
	"인내심이 필요한 하루입니다. 조급해하지 마세요.",
	"긍정적인 에너지가 가득한 날입니다. 주변에 좋은 영향을 줄 수 있어요.",
	"자기 자신을 돌보는 시간을 가지세요. 휴식도 중요합니다.",
	"오래된 문제가 해결될 수 있는 날입니다. 포기하지 마세요.",
	"귀인이 나타날 수 있습니다. 주변을 잘 살펴보세요.",
	"학습과 성장에 좋은 하루입니다. 새로운 것을 배워보세요.",
}

var luckyItems = []string{
	"빨간색 옷", "파란색 소품", "커피", "숫자 7", "동쪽 방향",
	"흰색 액세서리", "꽃 향기", "음악", "숫자 3", "남쪽 방향",
	"노란색 물건", "차 한잔", "책", "숫자 8", "북쪽 방향",
	"초록색 식물", "향초", "물", "숫자 5", "서쪽 방향",
}

// GetTodayFortune 오늘의 띠별 운세 생성
func (f *FortuneCollector) GetTodayFortune() []ZodiacFortune {
	// 날짜 기반 시드 설정 (같은 날은 같은 운세)
	today := time.Now().Format("2006-01-02")
	seed := int64(0)
	for _, c := range today {
		seed += int64(c)
	}
	rng := rand.New(rand.NewSource(seed))

	var fortunes []ZodiacFortune
	for _, zodiac := range zodiacs {
		// 각 띠별로 다른 시드
		zodiacSeed := seed + int64(len(zodiac.Name))
		zodiacRng := rand.New(rand.NewSource(zodiacSeed))

		fortune := ZodiacFortune{
			Zodiac:  zodiac.Name,
			Emoji:   zodiac.Emoji,
			Overall: zodiacRng.Intn(5) + 1,
			Love:    zodiacRng.Intn(5) + 1,
			Money:   zodiacRng.Intn(5) + 1,
			Health:  zodiacRng.Intn(5) + 1,
			Lucky:   luckyItems[zodiacRng.Intn(len(luckyItems))],
			Message: fortuneMessages[zodiacRng.Intn(len(fortuneMessages))],
		}
		fortunes = append(fortunes, fortune)
	}

	_ = rng // 사용
	return fortunes
}

// GenerateFortunePost 운세 포스트 생성
func (f *FortuneCollector) GenerateFortunePost(fortunes []ZodiacFortune) *Post {
	now := time.Now()
	title := fmt.Sprintf("🔮 오늘의 띠별 운세 [%s]", now.Format("01/02"))

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>🔮 오늘의 띠별 운세</h2>
<p>%s</p>

<div style="background: linear-gradient(135deg, #6c5ce7 0%%, #a29bfe 100%%); padding: 20px; border-radius: 15px; color: white; margin: 20px 0;">
<p style="text-align: center; font-size: 1.2em;">오늘 하루도 행복하세요! ✨</p>
</div>
`, now.Format("2006년 01월 02일")))

	for _, fortune := range fortunes {
		stars := strings.Repeat("⭐", fortune.Overall) + strings.Repeat("☆", 5-fortune.Overall)

		content.WriteString(fmt.Sprintf(`
<div style="background: #f8f9fa; padding: 20px; border-radius: 10px; margin-bottom: 15px; border-left: 5px solid #6c5ce7;">
<h3 style="margin-top: 0;">%s %s</h3>
<p><strong>종합운:</strong> %s</p>
<p style="display: flex; gap: 20px;">
<span>💕 애정운: %s</span>
<span>💰 금전운: %s</span>
<span>💪 건강운: %s</span>
</p>
<p>🍀 <strong>행운의 아이템:</strong> %s</p>
<p style="color: #666; font-style: italic;">"%s"</p>
</div>
`, fortune.Emoji, fortune.Zodiac, stars,
			getStarRating(fortune.Love),
			getStarRating(fortune.Money),
			getStarRating(fortune.Health),
			fortune.Lucky, fortune.Message))
	}

	content.WriteString(`
<p style="color: #888; font-size: 0.9em; margin-top: 30px; text-align: center;">
※ 운세는 재미로만 봐주세요! 오늘 하루도 화이팅! 💪
</p>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "운세/점술",
		Tags:     []string{"오늘의운세", "띠별운세", "운세", now.Format("01월02일운세"), "매일운세"},
	}
}

func getStarRating(n int) string {
	return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
}
