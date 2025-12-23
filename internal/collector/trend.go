package collector

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TrendCollector 트렌드/실검 수집기
type TrendCollector struct {
	client *http.Client
}

// Trend 트렌드 정보
type Trend struct {
	Rank      int       `json:"rank"`
	Keyword   string    `json:"keyword"`
	Link      string    `json:"link"`
	Category  string    `json:"category"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewTrendCollector() *TrendCollector {
	return &TrendCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetGoogleTrends 구글 트렌드 수집
func (t *TrendCollector) GetGoogleTrends(ctx context.Context, limit int) ([]Trend, error) {
	// Google Trends RSS (한국)
	url := "https://trends.google.co.kr/trending/rss?geo=KR"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TistoryBot/1.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// RSS 파싱
	var feed RSSFeed
	if err := decodeXML(resp.Body, &feed); err != nil {
		return nil, err
	}

	var trends []Trend
	for i, item := range feed.Channel.Items {
		if i >= limit {
			break
		}
		trends = append(trends, Trend{
			Rank:      i + 1,
			Keyword:   item.Title,
			Link:      item.Link,
			Source:    "Google Trends",
			UpdatedAt: time.Now(),
		})
	}

	return trends, nil
}

// GetNaverDataLab 네이버 데이터랩 (API 키 필요)
func (t *TrendCollector) GetNaverDataLab(ctx context.Context, clientID, clientSecret string) ([]Trend, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("네이버 API 키가 필요합니다. https://developers.naver.com 에서 발급받으세요")
	}

	// 네이버 검색 API 사용
	url := "https://openapi.naver.com/v1/datalab/search"

	// 요청 본문 구성
	reqBody := `{
		"startDate": "` + time.Now().AddDate(0, 0, -7).Format("2006-01-02") + `",
		"endDate": "` + time.Now().Format("2006-01-02") + `",
		"timeUnit": "date",
		"keywordGroups": [
			{"groupName": "트렌드", "keywords": ["인기검색어"]}
		]
	}`

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Naver-Client-Id", clientID)
	req.Header.Set("X-Naver-Client-Secret", clientSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 결과 파싱 (실제 구현에서는 더 상세히)
	return nil, nil
}

// GenerateTrendPost 트렌드 포스트 생성
func (t *TrendCollector) GenerateTrendPost(trends []Trend) *Post {
	now := time.Now()
	title := fmt.Sprintf("[%s] 실시간 인기 검색어 TOP 10 🔥", now.Format("01/02 15:00"))

	var content strings.Builder
	content.WriteString(`<h2>🔥 실시간 인기 검색어</h2>
<p>업데이트: ` + now.Format("2006년 01월 02일 15:04") + `</p>

<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 20px; border-radius: 10px; color: white;">
`)

	for _, trend := range trends {
		emoji := "🔹"
		if trend.Rank <= 3 {
			emoji = []string{"🥇", "🥈", "🥉"}[trend.Rank-1]
		}

		content.WriteString(fmt.Sprintf(`
<div style="padding: 10px 0; border-bottom: 1px solid rgba(255,255,255,0.2);">
<span style="font-size: 1.2em;">%s <strong>%d위</strong></span>
<span style="margin-left: 15px; font-size: 1.1em;">%s</span>
</div>
`, emoji, trend.Rank, trend.Keyword))
	}

	content.WriteString(`</div>

<h3>📊 트렌드 분석</h3>
<p>위 검색어들은 현재 가장 많이 검색되고 있는 키워드입니다.</p>
<p>실시간으로 변동되므로 참고용으로만 활용해주세요.</p>
`)

	// 공격적인 태그 전략
	tags := []string{
		// 기본 태그
		"실시간검색어", "트렌드", "인기검색어", "핫이슈", "화제",
		"실검", "실시간트렌드", "인기키워드",
		// 시간대 태그
		now.Format("01월02일"), now.Format("01월02일") + "실검",
		now.Format("2006년01월") + "트렌드",
		// 플랫폼 태그
		"네이버실검", "구글트렌드", "다음실검",
		// 인기 키워드
		"오늘화제", "지금인기", "핫키워드", "급상승검색어",
		"이슈", "오늘이슈", "실시간이슈",
	}
	// 검색어를 태그에 추가 (상위 5개)
	for i, trend := range trends {
		if i >= 5 {
			break
		}
		tags = append(tags, trend.Keyword)
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryTrend,
		Tags:     tags,
	}
}

func decodeXML(body io.Reader, v interface{}) error {
	return xml.NewDecoder(body).Decode(v)
}

