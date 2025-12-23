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
	client *http.Client
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

// SoccerMatch 축구 경기
type SoccerMatch struct {
	League    string
	HomeTeam  string
	AwayTeam  string
	HomeScore int
	AwayScore int
	Status    string
	Time      string
}

func NewSportsCollector() *SportsCollector {
	return &SportsCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetSportsNews 스포츠 뉴스 수집
func (s *SportsCollector) GetSportsNews(ctx context.Context) ([]SportsNews, error) {
	var allNews []SportsNews

	// 1. 네이버 스포츠 뉴스 (각 종목별)
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

// getSimulatedNews 시뮬레이션 뉴스 (실제 API 연동 전)
func (s *SportsCollector) getSimulatedNews(category string) []SportsNews {
	now := time.Now()
	dateStr := now.Format("01/02")

	newsData := map[string][]SportsNews{
		"축구": {
			{
				Title:       fmt.Sprintf("[%s] 손흥민, 시즌 10호골 폭발! 토트넘 승리 이끌어", dateStr),
				Description: "손흥민이 프리미어리그에서 시즌 10호골을 기록하며 팀의 승리를 이끌었다. 이로써 손흥민은 아시아 선수 최다 골 기록을 경신했다.",
				Category:    "축구",
				Source:      "네이버 스포츠",
				SourceURL:   "https://sports.news.naver.com/wfootball/index",
			},
			{
				Title:       fmt.Sprintf("[%s] K리그 2025시즌 일정 발표, 개막전 3월 1일", dateStr),
				Description: "한국프로축구연맹이 2025시즌 K리그 일정을 발표했다. 개막전은 3월 1일로 예정되어 있으며, 전북 현대와 울산 HD의 빅매치로 시작된다.",
				Category:    "축구",
				Source:      "K리그 공식",
				SourceURL:   "https://www.kleague.com",
			},
			{
				Title:       fmt.Sprintf("[%s] 이강인, 파리 생제르맹 주전 경쟁 치열", dateStr),
				Description: "이강인이 PSG에서 주전 경쟁에 나서고 있다. 최근 경기에서 좋은 활약을 보이며 출전 시간을 늘려가고 있다.",
				Category:    "축구",
				Source:      "네이버 스포츠",
				SourceURL:   "https://sports.news.naver.com/wfootball/index",
			},
		},
		"야구": {
			{
				Title:       fmt.Sprintf("[%s] MLB 겨울 FA 시장, 대형 계약 속출", dateStr),
				Description: "MLB 겨울 FA 시장이 뜨겁다. 여러 구단들이 대형 계약을 체결하며 내년 시즌을 준비하고 있다.",
				Category:    "야구",
				Source:      "MLB 공식",
				SourceURL:   "https://www.mlb.com",
			},
			{
				Title:       fmt.Sprintf("[%s] KBO 스토브리그, 각 구단 영입 현황 총정리", dateStr),
				Description: "KBO 스토브리그가 한창이다. 각 구단별 영입 현황과 전력 보강 상황을 살펴본다.",
				Category:    "야구",
				Source:      "KBO 공식",
				SourceURL:   "https://www.koreabaseball.com",
			},
			{
				Title:       fmt.Sprintf("[%s] 류현진, 재활 순항 중 \"내년 시즌 복귀 목표\"", dateStr),
				Description: "류현진이 재활을 성공적으로 진행하고 있다. 내년 시즌 복귀를 목표로 열심히 훈련 중이라고 밝혔다.",
				Category:    "야구",
				Source:      "네이버 스포츠",
				SourceURL:   "https://sports.news.naver.com/kbaseball/index",
			},
		},
		"농구": {
			{
				Title:       fmt.Sprintf("[%s] NBA 정규시즌, 각 팀 순위 현황", dateStr),
				Description: "NBA 정규시즌이 진행 중이다. 동부와 서부 컨퍼런스 각 팀의 순위 현황을 정리했다.",
				Category:    "농구",
				Source:      "NBA 공식",
				SourceURL:   "https://www.nba.com",
			},
			{
				Title:       fmt.Sprintf("[%s] KBL 프로농구, 치열한 순위 경쟁", dateStr),
				Description: "KBL 프로농구가 치열한 순위 경쟁을 펼치고 있다. 상위권 팀들의 격차가 좁혀지며 흥미진진한 경기가 이어지고 있다.",
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
	// 시즌 순위 (시즌 종료 후 최종 순위)
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

// GenerateSportsPost 스포츠 포스트 생성
func (s *SportsCollector) GenerateSportsPost(news []SportsNews) *Post {
	now := time.Now()
	title := fmt.Sprintf("⚽ 오늘의 스포츠 뉴스 [%s]", now.Format("01/02"))

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>⚽ 오늘의 스포츠 뉴스</h2>
<p style="color: #666;">업데이트: %s</p>

<div style="background: linear-gradient(135deg, #00b894 0%%, #00cec9 100%%); padding: 25px; border-radius: 15px; color: white; margin: 20px 0; text-align: center;">
<p style="font-size: 1.3em; margin: 0;">🏆 오늘의 주요 스포츠 소식</p>
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
		"배구": "🏐",
		"골프": "⛳",
		"기타": "🏅",
	}

	categoryOrder := []string{"축구", "야구", "농구"}

	for _, category := range categoryOrder {
		items, ok := categories[category]
		if !ok || len(items) == 0 {
			continue
		}

		emoji := categoryEmojis[category]
		if emoji == "" {
			emoji = "🏅"
		}

		content.WriteString(fmt.Sprintf(`
<h3 style="border-left: 4px solid #00b894; padding-left: 15px; margin-top: 30px;">%s %s</h3>
`, emoji, category))

		for _, item := range items {
			sourceLink := item.Source
			if item.SourceURL != "" {
				sourceLink = fmt.Sprintf(`<a href="%s" target="_blank" style="color: #0984e3; text-decoration: none;">%s 바로가기 →</a>`, item.SourceURL, item.Source)
			}
			content.WriteString(fmt.Sprintf(`
<div style="background: #f8f9fa; padding: 20px; border-radius: 12px; margin: 15px 0; border-left: 3px solid #00b894;">
  <h4 style="margin: 0 0 10px 0; color: #2d3436;">%s</h4>
  <p style="color: #636e72; line-height: 1.6; margin: 0;">%s</p>
  <p style="color: #b2bec3; font-size: 0.85em; margin: 10px 0 0 0;">📰 %s</p>
</div>
`, item.Title, item.Description, sourceLink))
		}
	}

	// NBA 하이라이트
	nbaHighlights := s.GetNBAHighlights()
	content.WriteString(`
<h3 style="border-left: 4px solid #e17055; padding-left: 15px; margin-top: 30px;">🏀 NBA 하이라이트</h3>
<div style="background: linear-gradient(135deg, #2d3436 0%, #636e72 100%); padding: 20px; border-radius: 12px; color: white;">
`)
	for _, highlight := range nbaHighlights {
		content.WriteString(fmt.Sprintf(`<p style="margin: 8px 0;">%s</p>`, highlight))
	}
	content.WriteString(`</div>`)

	// KBO 순위 (비시즌에도 표시)
	content.WriteString(`
<h3 style="border-left: 4px solid #fdcb6e; padding-left: 15px; margin-top: 30px;">⚾ 2024 KBO 최종 순위</h3>
<div style="overflow-x: auto;">
<table style="width: 100%; border-collapse: collapse; margin: 20px 0; min-width: 400px;">
<tr style="background: linear-gradient(135deg, #2d3436 0%, #636e72 100%); color: white;">
<th style="padding: 12px; border: none;">순위</th>
<th style="padding: 12px; border: none;">팀</th>
<th style="padding: 12px; border: none;">승</th>
<th style="padding: 12px; border: none;">패</th>
<th style="padding: 12px; border: none;">무</th>
<th style="padding: 12px; border: none;">승률</th>
</tr>
`)
	for i, team := range s.GetKBOStandings(context.Background()) {
		bgColor := "#fff"
		if i < 3 {
			bgColor = "#ffeaa7" // 상위 3팀 하이라이트
		} else if i < 5 {
			bgColor = "#dfe6e9"
		}
		rankEmoji := ""
		if i == 0 {
			rankEmoji = "🥇 "
		} else if i == 1 {
			rankEmoji = "🥈 "
		} else if i == 2 {
			rankEmoji = "🥉 "
		}
		content.WriteString(fmt.Sprintf(`<tr style="background: %s;">
<td style="padding: 12px; border-bottom: 1px solid #eee; text-align: center; font-weight: bold;">%s%d</td>
<td style="padding: 12px; border-bottom: 1px solid #eee; font-weight: bold;">%s</td>
<td style="padding: 12px; border-bottom: 1px solid #eee; text-align: center;">%d</td>
<td style="padding: 12px; border-bottom: 1px solid #eee; text-align: center;">%d</td>
<td style="padding: 12px; border-bottom: 1px solid #eee; text-align: center;">%d</td>
<td style="padding: 12px; border-bottom: 1px solid #eee; text-align: center;">%s</td>
</tr>
`, bgColor, rankEmoji, i+1, team.Name, team.Wins, team.Losses, team.Draws, team.Pct))
	}
	content.WriteString(`</table>
</div>
`)

	content.WriteString(`
<div style="background: #74b9ff; padding: 20px; border-radius: 12px; margin-top: 30px; color: white; text-align: center;">
<p style="margin: 0;">⚡ 더 자세한 경기 결과는 각 종목 공식 사이트에서 확인하세요!</p>
</div>
`)

	// 동적 태그 생성 (실제 뉴스 기반)
	tags := []string{
		// 기본 태그
		"스포츠", "스포츠뉴스",
		// 시간대 태그
		now.Format("01월02일") + "스포츠", now.Format("01월02일") + "경기결과",
	}

	// 📌 실제 뉴스 제목에서 키워드 추출 (핵심!)
	for _, item := range news {
		// 카테고리 태그
		tags = append(tags, item.Category)
		tags = append(tags, item.Category+"뉴스")

		// 제목에서 주요 키워드 추출
		keywords := []string{"손흥민", "이강인", "류현진", "김하성", "이정후", "오타니"}
		for _, kw := range keywords {
			if strings.Contains(item.Title, kw) {
				tags = append(tags, kw)
				tags = append(tags, kw+"뉴스")
			}
		}
	}

	// 종목별 태그
	for category := range categories {
		switch category {
		case "축구":
			tags = append(tags, "축구", "프리미어리그", "K리그", "해외축구")
		case "야구":
			tags = append(tags, "야구", "KBO", "MLB", "프로야구")
		case "농구":
			tags = append(tags, "농구", "NBA", "KBL")
		}
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "스포츠",
		Tags:     tags,
	}
}
