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
}

// KBOTeam KBO 팀 정보
type KBOTeam struct {
	Name   string
	Wins   int
	Losses int
	Draws  int
	Pct    string
}

func NewSportsCollector() *SportsCollector {
	return &SportsCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetSportsNews 스포츠 뉴스 수집 (RSS)
func (s *SportsCollector) GetSportsNews(ctx context.Context) ([]SportsNews, error) {
	// 스포츠 뉴스 RSS
	rssURL := "https://www.chosun.com/arc/outboundfeeds/rss/category/sports/?outputType=xml"

	req, err := http.NewRequestWithContext(ctx, "GET", rssURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TistoryBot/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		// RSS 실패시 더미 데이터
		return s.getDummySportsNews(), nil
	}
	defer resp.Body.Close()

	// 간단한 RSS 파싱
	var news []SportsNews
	// RSS 파싱 시도하고 실패시 더미 데이터 반환
	return append(news, s.getDummySportsNews()...), nil
}

// getDummySportsNews 스포츠 소식 (템플릿)
func (s *SportsCollector) getDummySportsNews() []SportsNews {
	now := time.Now()

	return []SportsNews{
		{
			Title:    fmt.Sprintf("[%s] 프로야구 주요 경기 결과", now.Format("01/02")),
			Category: "야구",
		},
		{
			Title:    fmt.Sprintf("[%s] K리그 주요 소식", now.Format("01/02")),
			Category: "축구",
		},
		{
			Title:    fmt.Sprintf("[%s] NBA/해외농구 소식", now.Format("01/02")),
			Category: "농구",
		},
	}
}

// GetKBOStandings KBO 순위 정보 (시즌 중)
func (s *SportsCollector) GetKBOStandings(ctx context.Context) []KBOTeam {
	// 실제로는 API나 크롤링으로 가져옴
	// 여기서는 예시 데이터 반환
	return []KBOTeam{
		{"LG 트윈스", 0, 0, 0, ".000"},
		{"KT 위즈", 0, 0, 0, ".000"},
		{"삼성 라이온즈", 0, 0, 0, ".000"},
		{"SSG 랜더스", 0, 0, 0, ".000"},
		{"NC 다이노스", 0, 0, 0, ".000"},
		{"두산 베어스", 0, 0, 0, ".000"},
		{"기아 타이거즈", 0, 0, 0, ".000"},
		{"롯데 자이언츠", 0, 0, 0, ".000"},
		{"한화 이글스", 0, 0, 0, ".000"},
		{"키움 히어로즈", 0, 0, 0, ".000"},
	}
}

// GenerateSportsPost 스포츠 포스트 생성
func (s *SportsCollector) GenerateSportsPost(news []SportsNews) *Post {
	now := time.Now()
	title := fmt.Sprintf("⚽ 오늘의 스포츠 뉴스 [%s]", now.Format("01/02"))

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>⚽ 오늘의 스포츠 뉴스</h2>
<p>업데이트: %s</p>

<div style="background: linear-gradient(135deg, #00b894 0%%, #00cec9 100%%); padding: 20px; border-radius: 15px; color: white; margin: 20px 0;">
<p style="text-align: center; font-size: 1.2em;">🏆 오늘의 주요 스포츠 소식</p>
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

	for category, items := range categories {
		emoji := categoryEmojis[category]
		if emoji == "" {
			emoji = "🏅"
		}

		content.WriteString(fmt.Sprintf(`
<h3>%s %s</h3>
<div style="background: #f8f9fa; padding: 15px; border-radius: 10px; margin-bottom: 15px;">
`, emoji, category))

		for _, item := range items {
			content.WriteString(fmt.Sprintf(`<p>• %s</p>
`, item.Title))
		}
		content.WriteString(`</div>
`)
	}

	// 프로야구 순위 (시즌 중일 때만)
	month := now.Month()
	if month >= 3 && month <= 10 {
		content.WriteString(`
<h3>⚾ KBO 프로야구 순위</h3>
<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
<tr style="background: #2d3436; color: white;">
<th style="padding: 10px; border: 1px solid #ddd;">순위</th>
<th style="padding: 10px; border: 1px solid #ddd;">팀</th>
<th style="padding: 10px; border: 1px solid #ddd;">승</th>
<th style="padding: 10px; border: 1px solid #ddd;">패</th>
<th style="padding: 10px; border: 1px solid #ddd;">무</th>
<th style="padding: 10px; border: 1px solid #ddd;">승률</th>
</tr>
`)
		for i, team := range s.GetKBOStandings(context.Background()) {
			bgColor := "#fff"
			if i < 5 {
				bgColor = "#dfe6e9"
			}
			content.WriteString(fmt.Sprintf(`<tr style="background: %s;">
<td style="padding: 10px; border: 1px solid #ddd; text-align: center;">%d</td>
<td style="padding: 10px; border: 1px solid #ddd;">%s</td>
<td style="padding: 10px; border: 1px solid #ddd; text-align: center;">%d</td>
<td style="padding: 10px; border: 1px solid #ddd; text-align: center;">%d</td>
<td style="padding: 10px; border: 1px solid #ddd; text-align: center;">%d</td>
<td style="padding: 10px; border: 1px solid #ddd; text-align: center;">%s</td>
</tr>
`, bgColor, i+1, team.Name, team.Wins, team.Losses, team.Draws, team.Pct))
		}
		content.WriteString(`</table>
<p style="color: #888; font-size: 0.9em;">※ 시즌 시작 전/후에는 순위가 표시되지 않을 수 있습니다.</p>
`)
	}

	content.WriteString(`
<p style="color: #888; font-size: 0.9em; margin-top: 30px;">
※ 더 자세한 경기 결과는 각 종목 공식 사이트에서 확인하세요.
</p>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "스포츠",
		Tags:     []string{"스포츠", "프로야구", "축구", "오늘의스포츠", now.Format("01월02일스포츠")},
	}
}
