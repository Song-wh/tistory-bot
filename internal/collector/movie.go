package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// MovieCollector 영화/드라마 정보 수집기
type MovieCollector struct {
	client    *http.Client
	tmdbKey   string // TMDB API Key (무료)
	coupangID string
}

// Movie 영화 정보
type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	OrigTitle   string  `json:"original_title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	PosterPath  string  `json:"poster_path"`
	VoteAverage float64 `json:"vote_average"`
	Popularity  float64 `json:"popularity"`
}

// TMDBResponse TMDB API 응답
type TMDBResponse struct {
	Results []Movie `json:"results"`
}

// MovieProduct 영화 관련 추천 상품
type MovieProduct struct {
	Name        string
	SearchQuery string
	Emoji       string
	Description string
}

func NewMovieCollector(tmdbKey, coupangID string) *MovieCollector {
	return &MovieCollector{
		client:    &http.Client{Timeout: 30 * time.Second},
		tmdbKey:   tmdbKey,
		coupangID: coupangID,
	}
}

// 영화 관람용 추천 상품
var movieProducts = []MovieProduct{
	{Name: "팝콘", SearchQuery: "전자레인지 팝콘", Emoji: "🍿", Description: "영화관 감성 그대로"},
	{Name: "담요", SearchQuery: "극세사 담요", Emoji: "🛋️", Description: "아늑한 영화 감상"},
	{Name: "빔프로젝터", SearchQuery: "가정용 빔프로젝터", Emoji: "📽️", Description: "홈시네마 필수템"},
	{Name: "사운드바", SearchQuery: "TV 사운드바", Emoji: "🔊", Description: "웅장한 사운드"},
}

// 드라마 시청용 추천 상품
var dramaProducts = []MovieProduct{
	{Name: "간식 세트", SearchQuery: "영화 간식세트", Emoji: "🍫", Description: "정주행 필수"},
	{Name: "쿠션", SearchQuery: "등쿠션", Emoji: "🛋️", Description: "편안한 시청"},
	{Name: "무선 이어폰", SearchQuery: "무선이어폰 추천", Emoji: "🎧", Description: "몰입 시청"},
	{Name: "테이블", SearchQuery: "노트북 테이블", Emoji: "🪑", Description: "침대에서 시청"},
}

// GetNowPlaying 현재 상영작 가져오기
func (m *MovieCollector) GetNowPlaying(ctx context.Context, limit int) ([]Movie, error) {
	if m.tmdbKey == "" {
		return nil, fmt.Errorf("TMDB API 키가 필요합니다. https://www.themoviedb.org/settings/api 에서 무료로 발급받으세요")
	}

	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/movie/now_playing?api_key=%s&language=ko-KR&region=KR&page=1",
		m.tmdbKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tmdbResp TMDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	if len(tmdbResp.Results) > limit {
		tmdbResp.Results = tmdbResp.Results[:limit]
	}

	return tmdbResp.Results, nil
}

// GetUpcoming 개봉 예정작 가져오기
func (m *MovieCollector) GetUpcoming(ctx context.Context, limit int) ([]Movie, error) {
	if m.tmdbKey == "" {
		return nil, fmt.Errorf("TMDB API 키가 필요합니다")
	}

	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/movie/upcoming?api_key=%s&language=ko-KR&region=KR&page=1",
		m.tmdbKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tmdbResp TMDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	if len(tmdbResp.Results) > limit {
		tmdbResp.Results = tmdbResp.Results[:limit]
	}

	return tmdbResp.Results, nil
}

// GetTrendingTV 인기 TV 프로그램 가져오기
func (m *MovieCollector) GetTrendingTV(ctx context.Context, limit int) ([]Movie, error) {
	if m.tmdbKey == "" {
		return nil, fmt.Errorf("TMDB API 키가 필요합니다")
	}

	url := fmt.Sprintf(
		"https://api.themoviedb.org/3/trending/tv/week?api_key=%s&language=ko-KR",
		m.tmdbKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tmdbResp TMDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	if len(tmdbResp.Results) > limit {
		tmdbResp.Results = tmdbResp.Results[:limit]
	}

	return tmdbResp.Results, nil
}

// generateCoupangLink 쿠팡 검색 링크 생성
func (m *MovieCollector) generateCoupangLink(query string) string {
	baseURL := fmt.Sprintf("https://www.coupang.com/np/search?component=&q=%s", query)
	if m.coupangID != "" {
		return fmt.Sprintf("%s&channel=affiliate&affiliate=%s", baseURL, m.coupangID)
	}
	return baseURL
}

// GenerateMoviePost 영화 정보 포스트 생성
func (m *MovieCollector) GenerateMoviePost(movies []Movie, postType string) *Post {
	now := time.Now()

	var title string
	var emoji string
	var products []MovieProduct

	switch postType {
	case "now_playing":
		title = fmt.Sprintf("🎬 [%s] 현재 상영 영화 TOP 10 & 홈시네마 추천", now.Format("01/02"))
		emoji = "🎬"
		products = movieProducts
	case "upcoming":
		title = fmt.Sprintf("🎥 [%s] 개봉 예정 영화 & 영화관 준비물", now.Format("01/02"))
		emoji = "🎥"
		products = movieProducts
	case "tv":
		title = fmt.Sprintf("📺 [%s] 이번 주 인기 드라마 & 정주행 필수템", now.Format("01/02"))
		emoji = "📺"
		products = dramaProducts
	}

	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.movie-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, sans-serif; }
.movie-header { background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%); padding: 30px; border-radius: 20px; color: white; text-align: center; margin-bottom: 25px; }
.movie-card { display: flex; background: white; border-radius: 12px; overflow: hidden; margin: 15px 0; box-shadow: 0 4px 15px rgba(0,0,0,0.1); }
.movie-poster { width: 140px; min-height: 200px; object-fit: cover; }
.movie-info { padding: 20px; flex: 1; }
.movie-rank { display: inline-block; background: #e74c3c; color: white; padding: 5px 12px; border-radius: 20px; font-weight: bold; margin-bottom: 10px; }
.movie-title { font-size: 20px; font-weight: 700; color: #2d3436; margin: 0 0 10px 0; }
.movie-meta { display: flex; gap: 15px; margin-bottom: 10px; color: #636e72; font-size: 14px; }
.movie-rating { color: #f39c12; font-weight: 600; }
.movie-desc { color: #636e72; line-height: 1.6; font-size: 14px; }
.product-section { background: linear-gradient(135deg, #fff5f5 0%, #ffe3e3 100%); padding: 30px; border-radius: 16px; margin-top: 40px; }
.product-title { font-size: 22px; font-weight: 700; color: #c53030; margin: 0 0 25px 0; text-align: center; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 15px; }
.product-card { background: white; padding: 20px; border-radius: 12px; text-align: center; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
.product-emoji { font-size: 40px; margin-bottom: 10px; }
.product-name { font-size: 16px; font-weight: 600; color: #2d3436; }
.product-desc { font-size: 13px; color: #636e72; margin: 5px 0 15px 0; }
.product-link { display: inline-block; background: #e53e3e; color: white; padding: 10px 20px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600; }
.product-link:hover { background: #c53030; }
.theater-links { display: flex; gap: 10px; justify-content: center; margin: 30px 0; flex-wrap: wrap; }
.theater-btn { padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; color: white; }
.cgv { background: #e74c3c; }
.megabox { background: #8e44ad; }
.lotte { background: #e74c3c; }
.footer-notice { margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 12px; font-size: 13px; color: #636e72; text-align: center; }
</style>
`)

	content.WriteString(fmt.Sprintf(`
<div class="movie-container">
<div class="movie-header">
	<h1 style="margin: 0; font-size: 28px;">%s %s</h1>
	<p style="margin: 10px 0 0 0; opacity: 0.9;">%s 업데이트</p>
</div>
`, emoji, title, now.Format("2006년 01월 02일")))

	// 영화 목록
	for i, movie := range movies {
		posterURL := ""
		if movie.PosterPath != "" {
			posterURL = "https://image.tmdb.org/t/p/w300" + movie.PosterPath
		}

		content.WriteString(`<div class="movie-card">`)
		if posterURL != "" {
			content.WriteString(fmt.Sprintf(`<img src="%s" alt="%s" class="movie-poster">`, posterURL, movie.Title))
		}
		content.WriteString(fmt.Sprintf(`
<div class="movie-info">
	<span class="movie-rank">%d위</span>
	<h3 class="movie-title">%s</h3>
	<div class="movie-meta">
		<span class="movie-rating">⭐ %.1f/10</span>
		<span>📅 %s</span>
	</div>
	<p class="movie-desc">%s</p>
</div>
</div>
`, i+1, movie.Title, movie.VoteAverage, movie.ReleaseDate, truncate(movie.Overview, 120)))
	}

	// 극장 예매 링크 (영화인 경우)
	if postType == "now_playing" || postType == "upcoming" {
		content.WriteString(`
<div class="theater-links">
	<a href="https://www.cgv.co.kr" target="_blank" class="theater-btn cgv">🎬 CGV 예매</a>
	<a href="https://www.megabox.co.kr" target="_blank" class="theater-btn megabox">🎬 메가박스 예매</a>
	<a href="https://www.lottecinema.co.kr" target="_blank" class="theater-btn lotte">🎬 롯데시네마 예매</a>
</div>
`)
	}

	// 추천 상품 섹션
	if m.coupangID != "" && len(products) > 0 {
		productTitle := "🍿 영화 감상 필수템"
		if postType == "tv" {
			productTitle = "📺 드라마 정주행 필수템"
		}

		content.WriteString(fmt.Sprintf(`
<div class="product-section">
	<h3 class="product-title">%s</h3>
	<div class="product-grid">
`, productTitle))

		for _, product := range products {
			content.WriteString(fmt.Sprintf(`
		<div class="product-card">
			<div class="product-emoji">%s</div>
			<div class="product-name">%s</div>
			<div class="product-desc">%s</div>
			<a href="%s" target="_blank" class="product-link">쿠팡에서 보기</a>
		</div>
`, product.Emoji, product.Name, product.Description, m.generateCoupangLink(product.SearchQuery)))
		}

		content.WriteString(`
	</div>
</div>
`)
	}

	// OTT 플랫폼 링크
	if postType == "tv" {
		content.WriteString(`
<div style="margin-top: 30px; text-align: center;">
	<h3>📱 OTT 플랫폼에서 시청하기</h3>
	<div style="display: flex; gap: 10px; justify-content: center; flex-wrap: wrap; margin-top: 15px;">
		<a href="https://www.netflix.com" target="_blank" style="padding: 10px 20px; background: #E50914; color: white; border-radius: 8px; text-decoration: none; font-weight: 600;">넷플릭스</a>
		<a href="https://www.tving.com" target="_blank" style="padding: 10px 20px; background: #FF0558; color: white; border-radius: 8px; text-decoration: none; font-weight: 600;">티빙</a>
		<a href="https://www.wavve.com" target="_blank" style="padding: 10px 20px; background: #1E2875; color: white; border-radius: 8px; text-decoration: none; font-weight: 600;">웨이브</a>
		<a href="https://watcha.com" target="_blank" style="padding: 10px 20px; background: #FF0558; color: white; border-radius: 8px; text-decoration: none; font-weight: 600;">왓챠</a>
	</div>
</div>
`)
	}

	// 푸터
	content.WriteString(`
<div class="footer-notice">
	<p>🎬 즐거운 영화/드라마 감상 되세요!</p>
	<p style="margin-top: 10px; font-size: 12px; color: #888;">
	⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.
	</p>
</div>
</div>
`)

	// 동적 태그 생성
	tags := []string{
		"영화", "영화추천", "박스오피스",
		now.Format("01월") + "영화", now.Format("01월02일") + "영화순위",
	}

	// 영화 제목 태그
	for _, movie := range movies {
		tags = append(tags, movie.Title)
		tags = append(tags, movie.Title+"리뷰")
	}

	// 상품 태그
	for _, p := range products[:2] {
		tags = append(tags, p.Name)
	}

	// 타입별 추가 태그
	switch postType {
	case "now_playing":
		tags = append(tags, "현재상영영화", "CGV", "메가박스", "롯데시네마", "극장영화")
	case "upcoming":
		tags = append(tags, "개봉예정영화", "신작영화", now.Format("01월")+"개봉영화")
	case "tv":
		tags = append(tags, "드라마", "드라마추천", "넷플릭스", "티빙", "웨이브", "왓챠", "정주행드라마")
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryMovie,
		Tags:     tags,
	}
}
