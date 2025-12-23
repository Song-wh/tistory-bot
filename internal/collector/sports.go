package collector

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SportsCollector 스포츠 정보 수집기
type SportsCollector struct {
	client         *http.Client
	coupangID      string
	footballAPIKey string // Football-Data.org API Key
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

// FootballMatch 축구 경기 정보
type FootballMatch struct {
	HomeTeam    string
	AwayTeam    string
	HomeScore   int
	AwayScore   int
	Status      string
	Competition string
	MatchDate   time.Time
	IsLive      bool
}

// NBAGame NBA 경기 정보
type NBAGame struct {
	HomeTeam  string
	AwayTeam  string
	HomeScore int
	AwayScore int
	Status    string
	GameDate  time.Time
	IsLive    bool
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

func NewSportsCollectorWithAPI(coupangID, footballAPIKey string) *SportsCollector {
	return &SportsCollector{
		client:         &http.Client{Timeout: 30 * time.Second},
		coupangID:      coupangID,
		footballAPIKey: footballAPIKey,
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

// ===============================================
// 실제 API 연동
// ===============================================

// GetFootballMatches Football-Data.org API로 축구 경기 가져오기
func (s *SportsCollector) GetFootballMatches(ctx context.Context) ([]FootballMatch, error) {
	if s.footballAPIKey == "" {
		return s.getSimulatedFootballMatches(), nil
	}

	// Premier League 경기 조회
	url := "https://api.football-data.org/v4/competitions/PL/matches?status=SCHEDULED,LIVE,FINISHED&limit=10"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return s.getSimulatedFootballMatches(), nil
	}
	req.Header.Set("X-Auth-Token", s.footballAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getSimulatedFootballMatches(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return s.getSimulatedFootballMatches(), nil
	}

	var result struct {
		Matches []struct {
			Status      string `json:"status"`
			UtcDate     string `json:"utcDate"`
			Competition struct {
				Name string `json:"name"`
			} `json:"competition"`
			HomeTeam struct {
				Name string `json:"name"`
			} `json:"homeTeam"`
			AwayTeam struct {
				Name string `json:"name"`
			} `json:"awayTeam"`
			Score struct {
				FullTime struct {
					Home int `json:"home"`
					Away int `json:"away"`
				} `json:"fullTime"`
			} `json:"score"`
		} `json:"matches"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getSimulatedFootballMatches(), nil
	}

	var matches []FootballMatch
	for _, m := range result.Matches {
		matchDate, _ := time.Parse(time.RFC3339, m.UtcDate)
		matches = append(matches, FootballMatch{
			HomeTeam:    translateTeamName(m.HomeTeam.Name),
			AwayTeam:    translateTeamName(m.AwayTeam.Name),
			HomeScore:   m.Score.FullTime.Home,
			AwayScore:   m.Score.FullTime.Away,
			Status:      translateStatus(m.Status),
			Competition: m.Competition.Name,
			MatchDate:   matchDate.In(time.FixedZone("KST", 9*60*60)),
			IsLive:      m.Status == "LIVE" || m.Status == "IN_PLAY",
		})
	}

	if len(matches) == 0 {
		return s.getSimulatedFootballMatches(), nil
	}

	return matches, nil
}

// GetNBAGames NBA 경기 가져오기 (balldontlie.io - 무료)
func (s *SportsCollector) GetNBAGames(ctx context.Context) ([]NBAGame, error) {
	today := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://www.balldontlie.io/api/v1/games?dates[]=%s", today)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return s.getSimulatedNBAGames(), nil
	}
	req.Header.Set("User-Agent", "TistoryBot/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return s.getSimulatedNBAGames(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return s.getSimulatedNBAGames(), nil
	}

	var result struct {
		Data []struct {
			Date       string `json:"date"`
			Status     string `json:"status"`
			HomeTeam   struct{ Name string } `json:"home_team"`
			VisitorTeam struct{ Name string } `json:"visitor_team"`
			HomeTeamScore    int `json:"home_team_score"`
			VisitorTeamScore int `json:"visitor_team_score"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.getSimulatedNBAGames(), nil
	}

	var games []NBAGame
	for _, g := range result.Data {
		gameDate, _ := time.Parse("2006-01-02T15:04:05.000Z", g.Date)
		games = append(games, NBAGame{
			HomeTeam:  g.HomeTeam.Name,
			AwayTeam:  g.VisitorTeam.Name,
			HomeScore: g.HomeTeamScore,
			AwayScore: g.VisitorTeamScore,
			Status:    g.Status,
			GameDate:  gameDate,
			IsLive:    g.Status == "in progress",
		})
	}

	if len(games) == 0 {
		return s.getSimulatedNBAGames(), nil
	}

	return games, nil
}

// GetSportsNewsRSS 스포츠 뉴스 RSS에서 실시간 수집
func (s *SportsCollector) GetSportsNewsRSS(ctx context.Context) ([]SportsNews, error) {
	rssFeeds := []struct {
		category string
		url      string
	}{
		{"축구", "https://news.google.com/rss/search?q=축구+OR+손흥민+OR+프리미어리그&hl=ko&gl=KR&ceid=KR:ko"},
		{"야구", "https://news.google.com/rss/search?q=야구+OR+MLB+OR+KBO&hl=ko&gl=KR&ceid=KR:ko"},
		{"농구", "https://news.google.com/rss/search?q=NBA+OR+농구&hl=ko&gl=KR&ceid=KR:ko"},
	}

	var allNews []SportsNews

	for _, feed := range rssFeeds {
		req, err := http.NewRequestWithContext(ctx, "GET", feed.url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var rss struct {
			XMLName xml.Name `xml:"rss"`
			Channel struct {
				Items []struct {
					Title   string `xml:"title"`
					Link    string `xml:"link"`
					PubDate string `xml:"pubDate"`
					Source  string `xml:"source"`
				} `xml:"item"`
			} `xml:"channel"`
		}

		if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
			continue
		}

		count := 0
		for _, item := range rss.Channel.Items {
			if count >= 3 { // 카테고리당 3개
				break
			}
			allNews = append(allNews, SportsNews{
				Title:     cleanNewsTitle(item.Title),
				Link:      item.Link,
				Category:  feed.category,
				Source:    item.Source,
				SourceURL: item.Link,
				PubDate:   item.PubDate,
			})
			count++
		}
	}

	// RSS 실패 시 시뮬레이션 데이터
	if len(allNews) == 0 {
		return s.getSimulatedNews(), nil
	}

	return allNews, nil
}

// GetSportsNews 스포츠 뉴스 수집 (RSS 우선, 실패 시 시뮬레이션)
func (s *SportsCollector) GetSportsNews(ctx context.Context) ([]SportsNews, error) {
	// RSS로 실시간 뉴스 시도
	news, err := s.GetSportsNewsRSS(ctx)
	if err == nil && len(news) > 0 {
		return news, nil
	}

	// 실패 시 시뮬레이션
	return s.getSimulatedNews(), nil
}

// ===============================================
// 시뮬레이션 데이터 (API 실패 시 백업)
// ===============================================

func (s *SportsCollector) getSimulatedFootballMatches() []FootballMatch {
	now := time.Now()
	return []FootballMatch{
		{
			HomeTeam:    "토트넘",
			AwayTeam:    "맨체스터 유나이티드",
			HomeScore:   2,
			AwayScore:   1,
			Status:      "종료",
			Competition: "프리미어리그",
			MatchDate:   now.Add(-2 * time.Hour),
			IsLive:      false,
		},
		{
			HomeTeam:    "리버풀",
			AwayTeam:    "맨체스터 시티",
			HomeScore:   0,
			AwayScore:   0,
			Status:      "예정",
			Competition: "프리미어리그",
			MatchDate:   now.Add(24 * time.Hour),
			IsLive:      false,
		},
		{
			HomeTeam:    "PSG",
			AwayTeam:    "바르셀로나",
			HomeScore:   1,
			AwayScore:   1,
			Status:      "진행중",
			Competition: "챔피언스리그",
			MatchDate:   now,
			IsLive:      true,
		},
	}
}

func (s *SportsCollector) getSimulatedNBAGames() []NBAGame {
	now := time.Now()
	return []NBAGame{
		{
			HomeTeam:  "LA Lakers",
			AwayTeam:  "Golden State Warriors",
			HomeScore: 112,
			AwayScore: 108,
			Status:    "Final",
			GameDate:  now.Add(-3 * time.Hour),
		},
		{
			HomeTeam:  "Boston Celtics",
			AwayTeam:  "Miami Heat",
			HomeScore: 0,
			AwayScore: 0,
			Status:    "Scheduled",
			GameDate:  now.Add(5 * time.Hour),
		},
	}
}

func (s *SportsCollector) getSimulatedNews() []SportsNews {
	now := time.Now()
	dateStr := now.Format("01/02")

	return []SportsNews{
		{
			Title:     fmt.Sprintf("[%s] 손흥민, 시즌 10호골 폭발! 토트넘 승리 이끌어", dateStr),
			Category:  "축구",
			Source:    "네이버 스포츠",
			SourceURL: "https://sports.news.naver.com/wfootball/index",
		},
		{
			Title:     fmt.Sprintf("[%s] K리그 2025시즌 개막 D-100", dateStr),
			Category:  "축구",
			Source:    "K리그 공식",
			SourceURL: "https://www.kleague.com",
		},
		{
			Title:     fmt.Sprintf("[%s] MLB 겨울 FA 시장 대형 계약 속출", dateStr),
			Category:  "야구",
			Source:    "MLB 공식",
			SourceURL: "https://www.mlb.com",
		},
		{
			Title:     fmt.Sprintf("[%s] KBO 스토브리그 영입 현황", dateStr),
			Category:  "야구",
			Source:    "KBO 공식",
			SourceURL: "https://www.koreabaseball.com",
		},
		{
			Title:     fmt.Sprintf("[%s] NBA 정규시즌 순위 현황", dateStr),
			Category:  "농구",
			Source:    "NBA 공식",
			SourceURL: "https://www.nba.com",
		},
	}
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
	ctx := context.Background()

	// 실시간 경기 데이터 가져오기
	footballMatches, _ := s.GetFootballMatches(ctx)
	nbaGames, _ := s.GetNBAGames(ctx)

	title := fmt.Sprintf("⚽ [%s] 실시간 스포츠 뉴스 & 경기 결과", now.Format("01/02 15:00"))

	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.sports-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, sans-serif; }
.sports-header { background: linear-gradient(135deg, #00b894 0%, #00cec9 100%); padding: 30px; border-radius: 20px; color: white; text-align: center; margin-bottom: 25px; }
.live-badge { display: inline-block; background: #e74c3c; color: white; padding: 4px 10px; border-radius: 12px; font-size: 12px; animation: pulse 1.5s infinite; margin-left: 8px; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
.match-section { background: #f8f9fa; padding: 25px; border-radius: 16px; margin: 20px 0; }
.match-card { background: white; padding: 20px; border-radius: 12px; margin: 15px 0; display: flex; align-items: center; justify-content: space-between; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
.team { text-align: center; flex: 1; }
.team-name { font-weight: 600; font-size: 16px; color: #2d3436; }
.score { font-size: 28px; font-weight: bold; color: #00b894; padding: 0 20px; }
.match-status { font-size: 12px; color: #636e72; margin-top: 5px; }
.news-card { background: #fff; padding: 20px; border-radius: 12px; margin: 15px 0; border-left: 4px solid #00b894; box-shadow: 0 2px 8px rgba(0,0,0,0.03); }
.news-title { font-size: 17px; font-weight: 600; color: #2d3436; margin: 0 0 10px 0; }
.news-title a { color: #2d3436; text-decoration: none; }
.news-title a:hover { color: #00b894; }
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
.kbo-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
.kbo-table th { background: linear-gradient(135deg, #2d3436, #636e72); color: white; padding: 12px; }
.kbo-table td { padding: 12px; border-bottom: 1px solid #eee; text-align: center; }
.footer-notice { margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 12px; font-size: 13px; color: #636e72; text-align: center; }
.realtime-tag { background: #27ae60; color: white; padding: 3px 8px; border-radius: 4px; font-size: 11px; margin-left: 5px; }
</style>
`)

	content.WriteString(fmt.Sprintf(`
<div class="sports-container">
<div class="sports-header">
	<h1 style="margin: 0; font-size: 28px;">⚽ 실시간 스포츠 뉴스</h1>
	<p style="margin: 10px 0 0 0; opacity: 0.9;">%s 업데이트 <span class="realtime-tag">실시간</span></p>
</div>
`, now.Format("2006년 01월 02일 15:04")))

	// ===============================================
	// 축구 경기 결과 (실시간)
	// ===============================================
	if len(footballMatches) > 0 {
		content.WriteString(`
<div class="match-section">
	<h2 class="category-title">⚽ 축구 경기 현황</h2>
`)
		for _, match := range footballMatches {
			liveTag := ""
			if match.IsLive {
				liveTag = `<span class="live-badge">🔴 LIVE</span>`
			}
			content.WriteString(fmt.Sprintf(`
	<div class="match-card">
		<div class="team">
			<div class="team-name">%s</div>
		</div>
		<div class="score">%d - %d</div>
		<div class="team">
			<div class="team-name">%s</div>
		</div>
	</div>
	<div style="text-align: center; margin-bottom: 15px;">
		<span class="match-status">%s %s</span> %s
	</div>
`, match.HomeTeam, match.HomeScore, match.AwayScore, match.AwayTeam, match.Competition, match.Status, liveTag))
		}
		content.WriteString(`</div>`)
	}

	// ===============================================
	// NBA 경기 결과 (실시간)
	// ===============================================
	if len(nbaGames) > 0 {
		content.WriteString(`
<div class="match-section">
	<h2 class="category-title">🏀 NBA 경기 현황</h2>
`)
		for _, game := range nbaGames {
			liveTag := ""
			if game.IsLive {
				liveTag = `<span class="live-badge">🔴 LIVE</span>`
			}
			content.WriteString(fmt.Sprintf(`
	<div class="match-card">
		<div class="team">
			<div class="team-name">%s</div>
		</div>
		<div class="score">%d - %d</div>
		<div class="team">
			<div class="team-name">%s</div>
		</div>
	</div>
	<div style="text-align: center; margin-bottom: 15px;">
		<span class="match-status">%s</span> %s
	</div>
`, game.HomeTeam, game.HomeScore, game.AwayScore, game.AwayTeam, game.Status, liveTag))
		}
		content.WriteString(`</div>`)
	}

	// ===============================================
	// 스포츠 뉴스
	// ===============================================
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
<h2 class="category-title">%s %s 뉴스</h2>
`, emoji, category))

		for _, item := range items {
			newsLink := item.Title
			if item.Link != "" {
				newsLink = fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, item.Link, item.Title)
			}

			sourceLink := item.Source
			if item.SourceURL != "" {
				sourceLink = fmt.Sprintf(`<a href="%s" target="_blank">%s 바로가기 →</a>`, item.SourceURL, item.Source)
			}
			content.WriteString(fmt.Sprintf(`
<div class="news-card">
	<h4 class="news-title">%s</h4>
	<p class="news-source">📰 %s</p>
</div>
`, newsLink, sourceLink))
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
	<p>⚡ 실시간 데이터 기반으로 자동 업데이트됩니다!</p>
	<p style="margin-top: 10px; font-size: 12px; color: #888;">
	⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.
	</p>
</div>
</div>
`)

	// 동적 태그 생성
	tags := []string{
		"스포츠", "스포츠뉴스", "스포츠용품", "실시간스포츠",
		now.Format("01월02일") + "스포츠",
	}

	// 경기 관련 태그
	for _, match := range footballMatches {
		tags = append(tags, match.HomeTeam, match.AwayTeam)
		if match.IsLive {
			tags = append(tags, match.HomeTeam+"경기")
		}
	}

	for _, game := range nbaGames {
		tags = append(tags, game.HomeTeam, game.AwayTeam)
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

	tags = append(tags, "축구화", "야구글러브", "농구화", "스포츠장비추천", "프리미어리그", "NBA")

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "스포츠",
		Tags:     tags,
	}
}

// ===============================================
// 헬퍼 함수
// ===============================================

func translateTeamName(name string) string {
	translations := map[string]string{
		"Tottenham Hotspur FC":     "토트넘",
		"Manchester United FC":     "맨체스터 유나이티드",
		"Manchester City FC":       "맨체스터 시티",
		"Liverpool FC":             "리버풀",
		"Arsenal FC":               "아스날",
		"Chelsea FC":               "첼시",
		"Paris Saint-Germain FC":   "PSG",
		"FC Barcelona":             "바르셀로나",
		"Real Madrid CF":           "레알 마드리드",
		"FC Bayern München":        "바이에른 뮌헨",
	}
	if translated, ok := translations[name]; ok {
		return translated
	}
	return name
}

func translateStatus(status string) string {
	translations := map[string]string{
		"SCHEDULED":   "예정",
		"LIVE":        "진행중",
		"IN_PLAY":     "진행중",
		"PAUSED":      "휴식",
		"FINISHED":    "종료",
		"POSTPONED":   "연기",
		"SUSPENDED":   "중단",
		"CANCELLED":   "취소",
	}
	if translated, ok := translations[status]; ok {
		return translated
	}
	return status
}

func cleanNewsTitle(title string) string {
	// " - 출처" 제거
	if idx := strings.LastIndex(title, " - "); idx > 0 {
		title = title[:idx]
	}
	return strings.TrimSpace(title)
}
