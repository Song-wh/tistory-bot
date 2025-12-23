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
	client  *http.Client
	tmdbKey string // TMDB API Key (무료)
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

func NewMovieCollector(tmdbKey string) *MovieCollector {
	return &MovieCollector{
		client:  &http.Client{Timeout: 30 * time.Second},
		tmdbKey: tmdbKey,
	}
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

// GenerateMoviePost 영화 정보 포스트 생성
func (m *MovieCollector) GenerateMoviePost(movies []Movie, postType string) *Post {
	now := time.Now()

	var title string
	var emoji string
	switch postType {
	case "now_playing":
		title = fmt.Sprintf("[%s] 현재 상영 영화 TOP 10 🎬", now.Format("01/02"))
		emoji = "🎬"
	case "upcoming":
		title = fmt.Sprintf("[%s] 개봉 예정 영화 🎥", now.Format("01/02"))
		emoji = "🎥"
	case "tv":
		title = fmt.Sprintf("[%s] 이번 주 인기 드라마 📺", now.Format("01/02"))
		emoji = "📺"
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>%s %s</h2>
<p>업데이트: %s</p>
`, emoji, title, now.Format("2006년 01월 02일 15:04")))

	for i, movie := range movies {
		posterURL := ""
		if movie.PosterPath != "" {
			posterURL = "https://image.tmdb.org/t/p/w300" + movie.PosterPath
		}

		content.WriteString(fmt.Sprintf(`
<div style="display: flex; border: 1px solid #ddd; margin: 15px 0; border-radius: 8px; overflow: hidden;">
`))
		if posterURL != "" {
			content.WriteString(fmt.Sprintf(`<img src="%s" alt="%s" style="width: 120px; object-fit: cover;">`, posterURL, movie.Title))
		}
		content.WriteString(fmt.Sprintf(`
<div style="padding: 15px; flex: 1;">
<h3>%d. %s</h3>
<p>⭐ 평점: %.1f/10</p>
<p>📅 개봉일: %s</p>
<p>%s</p>
</div>
</div>
`, i+1, movie.Title, movie.VoteAverage, movie.ReleaseDate, truncate(movie.Overview, 150)))
	}

	// 공격적인 태그 전략
	tags := []string{
		// 기본 태그
		"영화", "드라마", "영화추천", "드라마추천",
		"넷플릭스", "넷플릭스추천", "Netflix",
		// 박스오피스 태그
		"박스오피스", "현재상영", "개봉예정", "상영영화",
		// 플랫폼 태그
		"왓챠", "디즈니플러스", "티빙", "쿠팡플레이", "웨이브",
		// 시간대 태그
		now.Format("01월") + "영화", now.Format("2006년") + "영화추천",
		// 인기 키워드
		"영화순위", "드라마순위", "인기영화", "인기드라마",
		"이번주영화", "신작영화", "신작드라마",
		// 장르 태그
		"액션영화", "로맨스영화", "코미디영화", "스릴러영화",
	}
	// 영화 제목을 태그에 추가
	for i, movie := range movies {
		if i >= 5 {
			break
		}
		tags = append(tags, movie.Title)
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryMovie,
		Tags:     tags,
	}
}

