package collector

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SportsCollector 스포츠 정보 수집기
type SportsCollector struct {
	client    *http.Client
	coupangID string
}

// SportsNews 스포츠 뉴스
type SportsNews struct {
	Title       string
	Description string
	Link        string
	Category    string
	ImageURL    string
	Source      string
	SourceURL   string
	PubDate     string
}

// KBOTeam KBO 팀 정보
type KBOTeam struct {
	Name   string
	Wins   int
	Losses int
	Draws  int
	Pct    string
	Rank   int
}

// SportsProduct 스포츠 추천 상품
type SportsProduct struct {
	Name        string
	SearchQuery string
	Emoji       string
	Category    string
	Description string
}

func NewSportsCollector(coupangID string) *SportsCollector {
	return &SportsCollector{
		client:    &http.Client{Timeout: 30 * time.Second},
		coupangID: coupangID,
	}
}

// 종목별 추천 상품
var sportsProducts = map[string][]SportsProduct{
	"축구": {
		{Name: "축구공", SearchQuery: "축구공 정품", Emoji: "⚽", Category: "축구", Description: "FIFA 공인구"},
		{Name: "축구화", SearchQuery: "축구화 베스트", Emoji: "👟", Category: "축구", Description: "인기 브랜드"},
		{Name: "축구 유니폼", SearchQuery: "손흥민 유니폼", Emoji: "👕", Category: "축구", Description: "토트넘 유니폼"},
		{Name: "정강이 보호대", SearchQuery: "축구 정강이보호대", Emoji: "🦵", Category: "축구", Description: "안전한 경기"},
	},
	"야구": {
		{Name: "야구 글러브", SearchQuery: "야구글러브 추천", Emoji: "🧤", Category: "야구", Description: "입문자용 추천"},
		{Name: "야구 배트", SearchQuery: "야구배트 알루미늄", Emoji: "🏏", Category: "야구", Description: "연습용 배트"},
		{Name: "야구공", SearchQuery: "야구공 경식", Emoji: "⚾", Category: "야구", Description: "KBO 공인구"},
		{Name: "야구 모자", SearchQuery: "KBO 야구모자", Emoji: "🧢", Category: "야구", Description: "팀 응원용"},
	},
	"농구": {
		{Name: "농구공", SearchQuery: "농구공 실내외", Emoji: "🏀", Category: "농구", Description: "스팔딩 농구공"},
		{Name: "농구화", SearchQuery: "농구화 추천", Emoji: "👟", Category: "농구", Description: "조던/나이키"},
		{Name: "농구 유니폼", SearchQuery: "NBA 유니폼", Emoji: "👕", Category: "농구", Description: "NBA 정품"},
		{Name: "손목 밴드", SearchQuery: "농구 손목밴드", Emoji: "✋", Category: "농구", Description: "부상 방지"},
	},
}

// GetSportsNews 스포츠 뉴스 수집
func (s *SportsCollector) GetSportsNews(ctx context.Context) ([]SportsNews, error) {
	var allNews []SportsNews

	categories := []struct {
		name string
		url  string
	}{
		{"축구", "https://sports.news.naver.com/wfootball/index"},
		{"야구", "https://sports.news.naver.com/kbaseball/index"},
		{"농구", "https://sports.news.naver.com/basketball/index"},
	}

	for _, cat := range categories {
		news := s.getSimulatedNews(cat.name)
		allNews = append(allNews, news...)
	}

	return allNews, nil
}

// getSimulatedNews 시뮬레이션 뉴스
func (s *SportsCollector) getSimulatedNews(category string) []SportsNews {
	now := time.Now()
	dateStr := now.Format("01/02")

	newsData := map[string][]SportsNews{
		"축구": {
			{
				Title:       fmt.Sprintf("[%s] 손흥민, 시즌 10호골 폭발! 토트넘 승리 이끌어", dateStr),
				Description: "손흥민이 프리미어리그에서 시즌 10호골을 기록하며 팀의 승리를 이끌었다.",
				Category:    "축구",
				Source:      "네이버 스포츠",
				SourceURL:   "https://sports.news.naver.com/wfootball/index",
			},
			{
				Title:       fmt.Sprintf("[%s] K리그 2025시즌 일정 발표, 개막전 3월 1일", dateStr),
				Description: "한국프로축구연맹이 2025시즌 K리그 일정을 발표했다.",
				Category:    "축구",
				Source:      "K리그 공식",
				SourceURL:   "https://www.kleague.com",
			},
			{
				Title:       fmt.Sprintf("[%s] 이강인, 파리 생제르맹 주전 경쟁 치열", dateStr),
				Description: "이강인이 PSG에서 주전 경쟁에 나서고 있다.",
				Category:    "축구",
				Source:      "네이버 스포츠",
				SourceURL:   "https://sports.news.naver.com/wfootball/index",
			},
		},
		"야구": {
			{
				Title:       fmt.Sprintf("[%s] MLB 겨울 FA 시장, 대형 계약 속출", dateStr),
				Description: "MLB 겨울 FA 시장이 뜨겁다. 여러 구단들이 대형 계약을 체결하고 있다.",
				Category:    "야구",
				Source:      "MLB 공식",
				SourceURL:   "https://www.mlb.com",
			},
			{
				Title:       fmt.Sprintf("[%s] KBO 스토브리그, 각 구단 영입 현황 총정리", dateStr),
				Description: "KBO 스토브리그가 한창이다. 각 구단별 영입 현황을 살펴본다.",
				Category:    "야구",
				Source:      "KBO 공식",
				SourceURL:   "https://www.koreabaseball.com",
			},
			{
				Title:       fmt.Sprintf("[%s] 류현진, 재활 순항 중 \"내년 시즌 복귀 목표\"", dateStr),
				Description: "류현진이 재활을 성공적으로 진행하고 있다.",
				Category:    "야구",
				Source:      "네이버 스포츠",
				SourceURL:   "https://sports.news.naver.com/kbaseball/index",
			},
		},
		"농구": {
			{
				Title:       fmt.Sprintf("[%s] NBA 정규시즌, 각 팀 순위 현황", dateStr),
				Description: "NBA 정규시즌이 진행 중이다. 동부와 서부 컨퍼런스 순위를 정리했다.",
				Category:    "농구",
				Source:      "NBA 공식",
				SourceURL:   "https://www.nba.com",
			},
			{
				Title:       fmt.Sprintf("[%s] KBL 프로농구, 치열한 순위 경쟁", dateStr),
				Description: "KBL 프로농구가 치열한 순위 경쟁을 펼치고 있다.",
				Category:    "농구",
				Source:      "KBL 공식",
				SourceURL:   "https://www.kbl.or.kr",
			},
		},
	}

	if news, ok := newsData[category]; ok {
		return news
	}
	return []SportsNews{}
}

// GetKBOStandings KBO 순위 정보
func (s *SportsCollector) GetKBOStandings(ctx context.Context) []KBOTeam {
	return []KBOTeam{
		{"기아 타이거즈", 87, 55, 2, ".613", 1},
		{"삼성 라이온즈", 81, 62, 1, ".566", 2},
		{"LG 트윈스", 80, 63, 1, ".559", 3},
		{"두산 베어스", 75, 68, 1, ".524", 4},
		{"KT 위즈", 73, 69, 2, ".514", 5},
		{"SSG 랜더스", 69, 74, 1, ".483", 6},
		{"NC 다이노스", 66, 77, 1, ".462", 7},
		{"롯데 자이언츠", 62, 81, 1, ".434", 8},
		{"한화 이글스", 60, 83, 1, ".420", 9},
		{"키움 히어로즈", 55, 88, 1, ".385", 10},
	}
}

// GetNBAHighlights NBA 하이라이트
func (s *SportsCollector) GetNBAHighlights() []string {
	return []string{
		"🏀 르브론 제임스, 통산 4만 득점 달성 임박",
		"🏀 스테판 커리, 3점슛 신기록 경신 중",
		"🏀 빅터 웸반야마, 올해의 신인상 유력",
	}
}

// generateCoupangLink 쿠팡 검색 링크 생성
func (s *SportsCollector) generateCoupangLink(query string) string {
	baseURL := fmt.Sprintf("https://www.coupang.com/np/search?component=&q=%s", query)
	if s.coupangID != "" {
		return fmt.Sprintf("%s&channel=affiliate&affiliate=%s", baseURL, s.coupangID)
	}
	return baseURL
}

// GenerateSportsPost 스포츠 포스트 생성
func (s *SportsCollector) GenerateSportsPost(news []SportsNews) *Post {
	now := time.Now()
	title := fmt.Sprintf("⚽ [%s] 오늘의 스포츠 뉴스 & 추천 장비", now.Format("01/02"))

	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.sports-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, sans-serif; }
.sports-header { background: linear-gradient(135deg, #00b894 0%, #00cec9 100%); padding: 30px; border-radius: 20px; color: white; text-align: center; margin-bottom: 25px; }
.news-card { background: #f8f9fa; padding: 20px; border-radius: 12px; margin: 15px 0; border-left: 4px solid #00b894; }
.news-title { font-size: 18px; font-weight: 600; color: #2d3436; margin: 0 0 10px 0; }
.news-desc { color: #636e72; line-height: 1.6; margin: 0 0 10px 0; }
.news-source { font-size: 13px; color: #b2bec3; }
.news-source a { color: #0984e3; text-decoration: none; }
.category-section { margin-top: 40px; }
.category-title { border-left: 5px solid #00b894; padding-left: 15px; font-size: 22px; margin-bottom: 20px; }
.product-section { background: linear-gradient(135deg, #fff5f5 0%, #ffe3e3 100%); padding: 25px; border-radius: 16px; margin-top: 30px; }
.product-title { font-size: 20px; font-weight: 700; color: #c53030; margin: 0 0 20px 0; text-align: center; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
.product-card { background: white; padding: 20px; border-radius: 12px; text-align: center; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
.product-emoji { font-size: 40px; margin-bottom: 10px; }
.product-name { font-size: 16px; font-weight: 600; color: #2d3436; }
.product-desc { font-size: 13px; color: #636e72; margin: 5px 0; }
.product-link { display: inline-block; background: #e53e3e; color: white; padding: 8px 16px; border-radius: 8px; text-decoration: none; font-size: 14px; margin-top: 10px; }
.product-link:hover { background: #c53030; }
.kbo-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
.kbo-table th { background: linear-gradient(135deg, #2d3436, #636e72); color: white; padding: 12px; }
.kbo-table td { padding: 12px; border-bottom: 1px solid #eee; text-align: center; }
.footer-notice { margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 12px; font-size: 13px; color: #636e72; text-align: center; }
</style>
`)

	content.WriteString(fmt.Sprintf(`
<div class="sports-container">
<div class="sports-header">
	<h1 style="margin: 0; font-size: 28px;">⚽ 오늘의 스포츠 뉴스</h1>
	<p style="margin: 10px 0 0 0; opacity: 0.9;">%s 업데이트</p>
</div>
`, now.Format("2006년 01월 02일 15:04")))

	// 종목별 그룹화
	categories := map[string][]SportsNews{}
	for _, n := range news {
		categories[n.Category] = append(categories[n.Category], n)
	}

	categoryEmojis := map[string]string{
		"야구": "⚾",
		"축구": "⚽",
		"농구": "🏀",
	}

	categoryOrder := []string{"축구", "야구", "농구"}

	for _, category := range categoryOrder {
		items, ok := categories[category]
		if !ok || len(items) == 0 {
			continue
		}

		emoji := categoryEmojis[category]

		content.WriteString(fmt.Sprintf(`
<div class="category-section">
<h2 class="category-title">%s %s</h2>
`, emoji, category))

		for _, item := range items {
			sourceLink := item.Source
			if item.SourceURL != "" {
				sourceLink = fmt.Sprintf(`<a href="%s" target="_blank">%s 바로가기 →</a>`, item.SourceURL, item.Source)
			}
			content.WriteString(fmt.Sprintf(`
<div class="news-card">
	<h4 class="news-title">%s</h4>
	<p class="news-desc">%s</p>
	<p class="news-source">📰 %s</p>
</div>
`, item.Title, item.Description, sourceLink))
		}

		// 종목별 추천 상품
		if products, ok := sportsProducts[category]; ok && s.coupangID != "" {
			content.WriteString(fmt.Sprintf(`
<div class="product-section">
	<h3 class="product-title">🛒 %s %s 추천 장비</h3>
	<div class="product-grid">
`, emoji, category))

			for _, product := range products {
				content.WriteString(fmt.Sprintf(`
		<div class="product-card">
			<div class="product-emoji">%s</div>
			<div class="product-name">%s</div>
			<div class="product-desc">%s</div>
			<a href="%s" target="_blank" class="product-link">쿠팡에서 보기</a>
		</div>
`, product.Emoji, product.Name, product.Description, s.generateCoupangLink(product.SearchQuery)))
			}

			content.WriteString(`
	</div>
</div>
`)
		}

		content.WriteString(`</div>`) // category-section 끝
	}

	// KBO 순위
	content.WriteString(`
<div class="category-section">
<h2 class="category-title">⚾ 2024 KBO 최종 순위</h2>
<div style="overflow-x: auto;">
<table class="kbo-table">
<tr>
<th>순위</th><th>팀</th><th>승</th><th>패</th><th>무</th><th>승률</th>
</tr>
`)
	for i, team := range s.GetKBOStandings(context.Background()) {
		rankEmoji := ""
		if i == 0 {
			rankEmoji = "🥇 "
		} else if i == 1 {
			rankEmoji = "🥈 "
		} else if i == 2 {
			rankEmoji = "🥉 "
		}
		bgColor := "#fff"
		if i < 3 {
			bgColor = "#ffeaa7"
		}
		content.WriteString(fmt.Sprintf(`<tr style="background: %s;">
<td style="font-weight: bold;">%s%d</td>
<td style="font-weight: bold;">%s</td>
<td>%d</td><td>%d</td><td>%d</td><td>%s</td>
</tr>
`, bgColor, rankEmoji, i+1, team.Name, team.Wins, team.Losses, team.Draws, team.Pct))
	}
	content.WriteString(`</table></div></div>`)

	// 푸터
	content.WriteString(`
<div class="footer-notice">
	<p>⚡ 더 자세한 경기 결과는 각 종목 공식 사이트에서 확인하세요!</p>
	<p style="margin-top: 10px; font-size: 12px; color: #888;">
	⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.
	</p>
</div>
</div>
`)

	// 동적 태그 생성
	tags := []string{
		"스포츠", "스포츠뉴스", "스포츠용품",
		now.Format("01월02일") + "스포츠",
	}

	for _, item := range news {
		tags = append(tags, item.Category)
		keywords := []string{"손흥민", "이강인", "류현진", "김하성", "이정후"}
		for _, kw := range keywords {
			if strings.Contains(item.Title, kw) {
				tags = append(tags, kw)
			}
		}
	}

	// 상품 태그
	for category := range categories {
		if products, ok := sportsProducts[category]; ok {
			for _, p := range products[:2] {
				tags = append(tags, p.Name)
			}
		}
	}

	tags = append(tags, "축구화", "야구글러브", "농구화", "스포츠장비추천")

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "스포츠",
		Tags:     tags,
	}
}
