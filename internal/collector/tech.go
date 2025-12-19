package collector

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TechCollector IT/테크 뉴스 수집기
type TechCollector struct {
	client *http.Client
}

// TechNews 테크 뉴스
type TechNews struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Source      string    `json:"source"`
	PubDate     time.Time `json:"pub_date"`
}

// RSS Feed 구조체
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title string    `xml:"title"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func NewTechCollector() *TechCollector {
	return &TechCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// 테크 뉴스 RSS 피드 목록
var techRSSFeeds = map[string]string{
	"지디넷코리아": "https://www.zdnet.co.kr/rss/newsall.xml",
	"IT조선":   "http://it.chosun.com/rss/rss.xml",
	"블로터":    "https://www.bloter.net/feed",
	"테크크런치":  "https://techcrunch.com/feed/",
}

// GetTechNews 테크 뉴스 수집
func (t *TechCollector) GetTechNews(ctx context.Context, limit int) ([]TechNews, error) {
	var allNews []TechNews

	for source, feedURL := range techRSSFeeds {
		news, err := t.fetchRSS(ctx, feedURL, source)
		if err != nil {
			continue // 에러 무시하고 다음 피드로
		}
		allNews = append(allNews, news...)
	}

	// 최신순 정렬 후 limit 적용
	if len(allNews) > limit {
		allNews = allNews[:limit]
	}

	return allNews, nil
}

func (t *TechCollector) fetchRSS(ctx context.Context, feedURL, source string) ([]TechNews, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TistoryBot/1.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var feed RSSFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	var news []TechNews
	for _, item := range feed.Channel.Items {
		pubDate, _ := time.Parse(time.RFC1123Z, item.PubDate)
		news = append(news, TechNews{
			Title:       item.Title,
			Description: stripHTML(item.Description),
			Link:        item.Link,
			Source:      source,
			PubDate:     pubDate,
		})
	}

	return news, nil
}

func stripHTML(s string) string {
	// 간단한 HTML 태그 제거
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")

	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// GenerateTechPost 테크 뉴스 포스트 생성
func (t *TechCollector) GenerateTechPost(news []TechNews) *Post {
	now := time.Now()
	title := fmt.Sprintf("[%s] IT/테크 뉴스 브리핑 💻", now.Format("01/02"))

	var content strings.Builder
	content.WriteString(`<h2>💻 오늘의 IT/테크 뉴스</h2>
<p>업데이트: ` + now.Format("2006년 01월 02일 15:04") + `</p>
`)

	for i, n := range news {
		content.WriteString(fmt.Sprintf(`
<div style="border-left: 4px solid #007bff; padding: 10px 15px; margin: 15px 0; background: #f8f9fa;">
<h3>%d. %s</h3>
<p>%s</p>
<p style="color: #666; font-size: 0.9em;">출처: %s | <a href="%s" target="_blank">원문 보기</a></p>
</div>
`, i+1, n.Title, truncate(n.Description, 200), n.Source, n.Link))
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryTech,
		Tags:     []string{"IT뉴스", "테크", "기술", "AI", "스마트폰"},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

