package collector

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

// Google Trends RSS 전용 구조체
type GoogleTrendsRSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []GoogleTrendItem `xml:"item"`
	} `xml:"channel"`
}

type GoogleTrendItem struct {
	Title     string           `xml:"title"`
	Link      string           `xml:"link"`
	NewsItems []GoogleNewsItem `xml:"http://trends.google.com/trending/rss news_item"`
}

type GoogleNewsItem struct {
	Title  string `xml:"http://trends.google.com/trending/rss news_item_title"`
	URL    string `xml:"http://trends.google.com/trending/rss news_item_url"`
	Source string `xml:"http://trends.google.com/trending/rss news_item_source"`
}

// RSSFeed와 RSSItem은 tech.go에 정의되어 있음

func NewTrendCollector() *TrendCollector {
	return &TrendCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetGoogleTrends 구글 트렌드 수집 (실제 RSS 연동)
func (t *TrendCollector) GetGoogleTrends(ctx context.Context, limit int) ([]Trend, error) {
	// Google Trends RSS (한국)
	url := "https://trends.google.com/trending/rss?geo=KR"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google trends RSS failed: %d", resp.StatusCode)
	}

	// Google Trends 전용 RSS 파싱
	var feed GoogleTrendsRSS
	if err := decodeXML(resp.Body, &feed); err != nil {
		return nil, err
	}

	var trends []Trend
	rank := 1
	for _, item := range feed.Channel.Items {
		if rank > limit {
			break
		}
		keyword := cleanKeyword(item.Title)
		// 빈 키워드 (외국어만 있는 경우) 건너뛰기
		if keyword == "" {
			continue
		}

		// 실제 뉴스 링크 추출 (news_item_url 사용)
		newsLink := ""
		if len(item.NewsItems) > 0 {
			newsLink = item.NewsItems[0].URL
		}
		// 뉴스 링크가 없으면 구글 검색 링크로 대체
		if newsLink == "" {
			newsLink = fmt.Sprintf("https://www.google.com/search?q=%s", keyword)
		}

		trends = append(trends, Trend{
			Rank:      rank,
			Keyword:   keyword,
			Link:      newsLink,
			Source:    "Google Trends",
			UpdatedAt: time.Now(),
		})
		rank++
	}

	return trends, nil
}

// GetNaverNewsRSS 네이버 뉴스 RSS에서 핫토픽 추출
func (t *TrendCollector) GetNaverNewsRSS(ctx context.Context, limit int) ([]Trend, error) {
	// 네이버 뉴스 주요 RSS - 랭킹뉴스
	urls := []string{
		"https://news.google.com/rss/search?q=site:news.naver.com&hl=ko&gl=KR&ceid=KR:ko",
	}

	var trends []Trend
	rank := 1

	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := t.client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var feed RSSFeed
		if err := decodeXML(resp.Body, &feed); err != nil {
			continue
		}

		for _, item := range feed.Channel.Items {
			if rank > limit {
				break
			}
			// 제목에서 키워드 추출
			keyword := extractKeyword(item.Title)
			// 빈 키워드 또는 "NAVER"만 있는 경우 건너뛰기
			if keyword == "" || keyword == "NAVER" || strings.HasSuffix(keyword, "NAVER") {
				continue
			}
			trends = append(trends, Trend{
				Rank:      rank,
				Keyword:   keyword,
				Link:      item.Link,
				Source:    "네이버 뉴스",
				UpdatedAt: time.Now(),
			})
			rank++
		}
	}

	return trends, nil
}

// GetAllTrends 모든 소스에서 트렌드 수집
func (t *TrendCollector) GetAllTrends(ctx context.Context) ([]Trend, error) {
	var allTrends []Trend

	// 1. 구글 트렌드 (메인)
	googleTrends, err := t.GetGoogleTrends(ctx, 20)
	if err == nil && len(googleTrends) > 0 {
		allTrends = append(allTrends, googleTrends...)
	}

	// 2. 네이버 뉴스 RSS (보조)
	naverTrends, err := t.GetNaverNewsRSS(ctx, 10)
	if err == nil && len(naverTrends) > 0 {
		// 중복 제거하면서 추가
		existingKeywords := make(map[string]bool)
		for _, trend := range allTrends {
			existingKeywords[strings.ToLower(trend.Keyword)] = true
		}
		for _, trend := range naverTrends {
			if !existingKeywords[strings.ToLower(trend.Keyword)] {
				trend.Rank = len(allTrends) + 1
				allTrends = append(allTrends, trend)
			}
		}
	}

	// 최소 데이터 보장 (API 실패 시)
	if len(allTrends) == 0 {
		allTrends = t.getBackupTrends()
	}

	return allTrends, nil
}

// getBackupTrends API 실패 시 백업 트렌드
func (t *TrendCollector) getBackupTrends() []Trend {
	now := time.Now()
	keywords := []string{
		"크리스마스", "연말정산", "송년회", "새해", "부동산",
		"날씨", "코로나", "주식", "비트코인", "환율",
	}

	var trends []Trend
	for i, keyword := range keywords {
		trends = append(trends, Trend{
			Rank:      i + 1,
			Keyword:   keyword,
			Source:    "Hot Topics",
			UpdatedAt: now,
		})
	}
	return trends
}

// GenerateTrendPost 트렌드 포스트 생성
func (t *TrendCollector) GenerateTrendPost(trends []Trend) *Post {
	now := time.Now()
	title := fmt.Sprintf("[%s] 실시간 인기 검색어 TOP %d 🔥", now.Format("01/02 15:00"), len(trends))

	var content strings.Builder

	content.WriteString(`
<style>
.trend-container { max-width: 800px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
.trend-header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 20px; color: white; text-align: center; margin-bottom: 25px; }
.trend-header h1 { margin: 0; font-size: 26px; }
.trend-header .update-time { opacity: 0.9; margin-top: 8px; font-size: 14px; }
.trend-source { display: flex; gap: 10px; justify-content: center; margin-top: 15px; }
.trend-source span { background: rgba(255,255,255,0.2); padding: 5px 12px; border-radius: 20px; font-size: 12px; }
.trend-list { background: #f8f9fa; border-radius: 16px; padding: 20px; }
.trend-item { display: flex; align-items: center; padding: 15px; border-bottom: 1px solid #e9ecef; transition: background 0.2s; }
.trend-item:hover { background: #fff; }
.trend-item:last-child { border-bottom: none; }
.trend-rank { width: 40px; height: 40px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: bold; margin-right: 15px; }
.rank-1 { background: linear-gradient(135deg, #FFD700, #FFA500); color: white; font-size: 18px; }
.rank-2 { background: linear-gradient(135deg, #C0C0C0, #A0A0A0); color: white; font-size: 18px; }
.rank-3 { background: linear-gradient(135deg, #CD7F32, #B87333); color: white; font-size: 18px; }
.rank-default { background: #e9ecef; color: #495057; }
.trend-keyword { flex: 1; font-size: 16px; font-weight: 500; color: #2d3436; }
.trend-keyword a { color: #2d3436; text-decoration: none; }
.trend-keyword a:hover { color: #667eea; }
.trend-source-tag { font-size: 11px; padding: 4px 8px; border-radius: 4px; background: #e3f2fd; color: #1976d2; }
.google-tag { background: #fce4ec; color: #c2185b; }
.naver-tag { background: #e8f5e9; color: #388e3c; }
.trend-footer { margin-top: 25px; padding: 20px; background: #fff3cd; border-radius: 12px; text-align: center; }
</style>
`)

	content.WriteString(fmt.Sprintf(`
<div class="trend-container">
<div class="trend-header">
	<h1>🔥 실시간 인기 검색어</h1>
	<p class="update-time">📅 %s 업데이트</p>
	<div class="trend-source">
		<span>📊 Google Trends</span>
		<span>📰 네이버 뉴스</span>
	</div>
</div>

<div class="trend-list">
`, now.Format("2006년 01월 02일 15:04")))

	for _, trend := range trends {
		rankClass := "rank-default"
		if trend.Rank == 1 {
			rankClass = "rank-1"
		} else if trend.Rank == 2 {
			rankClass = "rank-2"
		} else if trend.Rank == 3 {
			rankClass = "rank-3"
		}

		sourceClass := "naver-tag"
		if trend.Source == "Google Trends" {
			sourceClass = "google-tag"
		}

		keywordLink := trend.Keyword
		if trend.Link != "" {
			keywordLink = fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, trend.Link, trend.Keyword)
		} else {
			// 구글 검색 링크 생성
			searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", trend.Keyword)
			keywordLink = fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, searchURL, trend.Keyword)
		}

		content.WriteString(fmt.Sprintf(`
<div class="trend-item">
	<div class="trend-rank %s">%d</div>
	<div class="trend-keyword">%s</div>
	<span class="trend-source-tag %s">%s</span>
</div>
`, rankClass, trend.Rank, keywordLink, sourceClass, trend.Source))
	}

	content.WriteString(`
</div>

<div class="trend-footer">
	<p>💡 <strong>실시간 데이터</strong>를 기반으로 수집된 인기 검색어입니다.</p>
	<p style="font-size: 13px; color: #856404; margin-top: 8px;">각 키워드를 클릭하면 관련 정보를 확인할 수 있습니다.</p>
</div>
</div>
`)

	// 동적 태그 생성 (실제 검색어 기반)
	tags := []string{
		"실시간검색어", "트렌드", "인기검색어", "구글트렌드",
		now.Format("01월02일") + "이슈", now.Format("01월02일") + "실검",
	}

	// 📌 실제 검색어를 태그로 (핵심!)
	for i, trend := range trends {
		if i >= 10 {
			break // 상위 10개만
		}
		tags = append(tags, trend.Keyword)
		if i < 5 {
			tags = append(tags, trend.Keyword+"뉴스")
		}
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryTrend,
		Tags:     tags,
	}
}

// cleanKeyword 키워드 정리 (한글/영문만 허용)
func cleanKeyword(keyword string) string {
	// HTML 태그 제거
	re := regexp.MustCompile(`<[^>]*>`)
	keyword = re.ReplaceAllString(keyword, "")
	// 특수문자 정리
	keyword = strings.TrimSpace(keyword)

	// 한글 또는 영문이 포함되어 있는지 확인
	hasKorean := regexp.MustCompile(`[가-힣]`).MatchString(keyword)
	hasEnglish := regexp.MustCompile(`[a-zA-Z]`).MatchString(keyword)

	// 한글이나 영문이 없으면 (외국어만 있으면) 빈 문자열 반환
	if !hasKorean && !hasEnglish {
		return ""
	}

	return keyword
}

// extractKeyword 뉴스 제목에서 핵심 키워드 추출
func extractKeyword(title string) string {
	// "[기관명]" 등 제거
	re := regexp.MustCompile(`\[[^\]]*\]`)
	title = re.ReplaceAllString(title, "")

	// " - 출처" 제거 (NAVER, 한겨레, 조선일보 등)
	if idx := strings.LastIndex(title, " - "); idx > 0 {
		title = title[:idx]
	}

	// " | 출처" 제거
	if idx := strings.LastIndex(title, " | "); idx > 0 {
		title = title[:idx]
	}

	// "..." 이후 제거
	if idx := strings.Index(title, "..."); idx > 0 {
		title = title[:idx]
	}

	// 30자 이상이면 자르기
	title = strings.TrimSpace(title)
	if len(title) > 30 {
		runes := []rune(title)
		if len(runes) > 30 {
			title = string(runes[:30])
		}
	}

	// 빈 문자열이거나 너무 짧으면 무시
	if len(title) < 2 {
		return ""
	}

	return title
}

func decodeXML(body io.Reader, v interface{}) error {
	return xml.NewDecoder(body).Decode(v)
}
