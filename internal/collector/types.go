package collector

import "time"

// Post 블로그 포스트
type Post struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	Thumbnail string    `json:"thumbnail"`
	CreatedAt time.Time `json:"created_at"`
}

// Category 카테고리 정보
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// 카테고리 상수
const (
	CategoryStock = "주식/코인"
	CategoryDeal  = "핫딜/할인"
	CategoryTech  = "IT/테크"
	CategoryMovie = "영화/드라마"
	CategoryTrend = "트렌드/실검"
)

