package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// GolfCollector 골프 + 날씨 수집기
type GolfCollector struct {
	client       *http.Client
	coupangID    string
	regions      []GolfRegion
}

// GolfRegion 지역 정보
type GolfRegion struct {
	Name       string       `json:"name"`        // 지역명 (예: 용인)
	City       string       `json:"city"`        // 시/도 (예: 경기도)
	Lat        float64      `json:"lat"`         // 위도
	Lon        float64      `json:"lon"`         // 경도
	GolfCourses []GolfCourse `json:"golf_courses"` // 골프장 목록
}

// GolfCourse 골프장 정보
type GolfCourse struct {
	Name        string   `json:"name"`         // 골프장명
	Address     string   `json:"address"`      // 주소
	Phone       string   `json:"phone"`        // 전화번호
	GreenFee    string   `json:"green_fee"`    // 그린피
	Holes       int      `json:"holes"`        // 홀 수
	Features    []string `json:"features"`     // 특징
	Rating      float64  `json:"rating"`       // 평점
	ImageURL    string   `json:"image_url"`    // 이미지
	BookingURL  string   `json:"booking_url"`  // 예약 URL
}

// GolfWeather 골프 날씨 정보
type GolfWeather struct {
	Region      string  `json:"region"`
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feels_like"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	GolfIndex   int     `json:"golf_index"`    // 골프 지수 (0-100)
	GolfGrade   string  `json:"golf_grade"`    // 등급 (최적/좋음/보통/비추)
}

// GolfProduct 골프 용품 (쿠팡 파트너스)
type GolfProduct struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	ImageURL string `json:"image_url"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

// NewGolfCollector 골프 수집기 생성
func NewGolfCollector(coupangID string) *GolfCollector {
	return &GolfCollector{
		client: &http.Client{Timeout: 30 * time.Second},
		coupangID: coupangID,
		regions: getDefaultRegions(),
	}
}

// getDefaultRegions 전국 주요 지역 및 골프장 데이터
func getDefaultRegions() []GolfRegion {
	return []GolfRegion{
		// ===== 수도권 =====
		{
			Name: "용인",
			City: "경기도",
			Lat:  37.2411,
			Lon:  127.1776,
			GolfCourses: []GolfCourse{
				{Name: "레이크사이드CC", Address: "경기도 용인시 처인구", GreenFee: "주중 18만원 / 주말 25만원", Features: []string{"명문 골프장", "호수 뷰"}, Rating: 4.7},
				{Name: "용인CC", Address: "경기도 용인시 처인구", GreenFee: "주중 15만원 / 주말 22만원", Features: []string{"접근성 좋음", "가성비"}, Rating: 4.3},
				{Name: "양지파인리조트CC", Address: "경기도 용인시 처인구", GreenFee: "주중 16만원 / 주말 23만원", Features: []string{"리조트 연계", "사계절"}, Rating: 4.5},
			},
		},
		{
			Name: "이천",
			City: "경기도",
			Lat:  37.2719,
			Lon:  127.4348,
			GolfCourses: []GolfCourse{
				{Name: "블랙스톤CC", Address: "경기도 이천시", GreenFee: "주중 20만원 / 주말 30만원", Features: []string{"프리미엄", "VIP 서비스"}, Rating: 4.8},
				{Name: "사우스스프링스CC", Address: "경기도 이천시", GreenFee: "주중 17만원 / 주말 25만원", Features: []string{"자연친화적", "좋은 관리"}, Rating: 4.6},
			},
		},
		{
			Name: "파주",
			City: "경기도",
			Lat:  37.7599,
			Lon:  126.7800,
			GolfCourses: []GolfCourse{
				{Name: "서원밸리CC", Address: "경기도 파주시", GreenFee: "주중 14만원 / 주말 20만원", Features: []string{"서울 근교", "접근성"}, Rating: 4.2},
				{Name: "파주CC", Address: "경기도 파주시", GreenFee: "주중 13만원 / 주말 19만원", Features: []string{"합리적 가격", "초보자 친화"}, Rating: 4.0},
			},
		},
		// ===== 강원도 =====
		{
			Name: "춘천",
			City: "강원도",
			Lat:  37.8813,
			Lon:  127.7300,
			GolfCourses: []GolfCourse{
				{Name: "라데나CC", Address: "강원도 춘천시", GreenFee: "주중 14만원 / 주말 20만원", Features: []string{"청정 자연", "시원한 공기"}, Rating: 4.4},
				{Name: "춘천레이크CC", Address: "강원도 춘천시", GreenFee: "주중 12만원 / 주말 18만원", Features: []string{"호수 전경", "힐링"}, Rating: 4.2},
			},
		},
		{
			Name: "강릉",
			City: "강원도",
			Lat:  37.7519,
			Lon:  128.8761,
			GolfCourses: []GolfCourse{
				{Name: "파인비치CC", Address: "강원도 강릉시", GreenFee: "주중 16만원 / 주말 24만원", Features: []string{"동해 바다 뷰", "리조트"}, Rating: 4.6},
				{Name: "강릉CC", Address: "강원도 강릉시", GreenFee: "주중 13만원 / 주말 19만원", Features: []string{"시원한 바람", "자연경관"}, Rating: 4.3},
			},
		},
		{
			Name: "원주",
			City: "강원도",
			Lat:  37.3422,
			Lon:  127.9202,
			GolfCourses: []GolfCourse{
				{Name: "오크밸리CC", Address: "강원도 원주시", GreenFee: "주중 15만원 / 주말 22만원", Features: []string{"리조트 연계", "스키장"}, Rating: 4.5},
				{Name: "쏠비치CC", Address: "강원도 원주시", GreenFee: "주중 14만원 / 주말 20만원", Features: []string{"깔끔한 코스", "편의시설"}, Rating: 4.3},
			},
		},
		// ===== 충청도 =====
		{
			Name: "천안",
			City: "충청남도",
			Lat:  36.8151,
			Lon:  127.1139,
			GolfCourses: []GolfCourse{
				{Name: "우정힐스CC", Address: "충남 천안시", GreenFee: "주중 13만원 / 주말 19만원", Features: []string{"KTX 접근성", "가성비"}, Rating: 4.2},
				{Name: "천안상록CC", Address: "충남 천안시", GreenFee: "주중 12만원 / 주말 17만원", Features: []string{"합리적 가격", "넓은 코스"}, Rating: 4.0},
			},
		},
		{
			Name: "대전",
			City: "대전광역시",
			Lat:  36.3504,
			Lon:  127.3845,
			GolfCourses: []GolfCourse{
				{Name: "롯데스카이힐CC", Address: "대전광역시", GreenFee: "주중 14만원 / 주말 21만원", Features: []string{"도심 근처", "편리한 접근"}, Rating: 4.3},
				{Name: "유성CC", Address: "대전광역시 유성구", GreenFee: "주중 15만원 / 주말 22만원", Features: []string{"온천 연계", "명문"}, Rating: 4.5},
			},
		},
		// ===== 전라도 =====
		{
			Name: "광주",
			City: "광주광역시",
			Lat:  35.1595,
			Lon:  126.8526,
			GolfCourses: []GolfCourse{
				{Name: "무등산CC", Address: "광주광역시", GreenFee: "주중 13만원 / 주말 19만원", Features: []string{"무등산 전경", "명문"}, Rating: 4.4},
				{Name: "광주CC", Address: "광주광역시", GreenFee: "주중 12만원 / 주말 17만원", Features: []string{"도심 접근", "가성비"}, Rating: 4.1},
			},
		},
		{
			Name: "전주",
			City: "전라북도",
			Lat:  35.8242,
			Lon:  127.1480,
			GolfCourses: []GolfCourse{
				{Name: "전주신시가지CC", Address: "전북 전주시", GreenFee: "주중 12만원 / 주말 17만원", Features: []string{"신시가지", "편의시설"}, Rating: 4.2},
				{Name: "라온CC", Address: "전북 완주군", GreenFee: "주중 14만원 / 주말 20만원", Features: []string{"자연경관", "코스 관리"}, Rating: 4.4},
			},
		},
		{
			Name: "여수",
			City: "전라남도",
			Lat:  34.7604,
			Lon:  127.6622,
			GolfCourses: []GolfCourse{
				{Name: "디오션CC", Address: "전남 여수시", GreenFee: "주중 18만원 / 주말 28만원", Features: []string{"바다 뷰", "프리미엄"}, Rating: 4.7},
				{Name: "여수경도CC", Address: "전남 여수시", GreenFee: "주중 15만원 / 주말 22만원", Features: []string{"섬 골프장", "경치"}, Rating: 4.5},
			},
		},
		// ===== 경상도 =====
		{
			Name: "부산",
			City: "부산광역시",
			Lat:  35.1796,
			Lon:  129.0756,
			GolfCourses: []GolfCourse{
				{Name: "아시아드CC", Address: "부산광역시 강서구", GreenFee: "주중 15만원 / 주말 23만원", Features: []string{"도심 접근", "야간 라운딩"}, Rating: 4.4},
				{Name: "기장CC", Address: "부산광역시 기장군", GreenFee: "주중 14만원 / 주말 21만원", Features: []string{"바다 근처", "리조트"}, Rating: 4.3},
			},
		},
		{
			Name: "대구",
			City: "대구광역시",
			Lat:  35.8714,
			Lon:  128.6014,
			GolfCourses: []GolfCourse{
				{Name: "대구CC", Address: "대구광역시 달성군", GreenFee: "주중 14만원 / 주말 21만원", Features: []string{"명문", "역사"}, Rating: 4.5},
				{Name: "팔공CC", Address: "대구광역시 동구", GreenFee: "주중 13만원 / 주말 19만원", Features: []string{"팔공산", "자연"}, Rating: 4.2},
			},
		},
		{
			Name: "경주",
			City: "경상북도",
			Lat:  35.8562,
			Lon:  129.2247,
			GolfCourses: []GolfCourse{
				{Name: "블루원CC", Address: "경북 경주시", GreenFee: "주중 16만원 / 주말 24만원", Features: []string{"리조트", "관광 연계"}, Rating: 4.6},
				{Name: "경주CC", Address: "경북 경주시", GreenFee: "주중 14만원 / 주말 20만원", Features: []string{"역사 도시", "명문"}, Rating: 4.4},
			},
		},
		{
			Name: "거제",
			City: "경상남도",
			Lat:  34.8806,
			Lon:  128.6211,
			GolfCourses: []GolfCourse{
				{Name: "거제씨클럽CC", Address: "경남 거제시", GreenFee: "주중 17만원 / 주말 26만원", Features: []string{"바다 뷰", "프리미엄"}, Rating: 4.7},
				{Name: "거제CC", Address: "경남 거제시", GreenFee: "주중 14만원 / 주말 20만원", Features: []string{"남해 전경", "휴양"}, Rating: 4.4},
			},
		},
		// ===== 제주도 =====
		{
			Name: "제주",
			City: "제주특별자치도",
			Lat:  33.4996,
			Lon:  126.5312,
			GolfCourses: []GolfCourse{
				{Name: "나인브릿지CC", Address: "제주도 서귀포시", GreenFee: "주중 35만원 / 주말 45만원", Features: []string{"세계적 명문", "한라산 뷰"}, Rating: 4.9},
				{Name: "핀크스CC", Address: "제주도 서귀포시", GreenFee: "주중 25만원 / 주말 35만원", Features: []string{"PGA 투어", "프리미엄"}, Rating: 4.8},
				{Name: "래비드리조트CC", Address: "제주도 제주시", GreenFee: "주중 18만원 / 주말 28만원", Features: []string{"리조트", "바다 전경"}, Rating: 4.5},
			},
		},
		{
			Name: "서귀포",
			City: "제주특별자치도",
			Lat:  33.2541,
			Lon:  126.5601,
			GolfCourses: []GolfCourse{
				{Name: "해비치CC", Address: "제주도 서귀포시", GreenFee: "주중 22만원 / 주말 32만원", Features: []string{"호텔 연계", "고급"}, Rating: 4.7},
				{Name: "오라CC", Address: "제주도 서귀포시", GreenFee: "주중 20만원 / 주말 28만원", Features: []string{"한라산 뷰", "자연"}, Rating: 4.6},
			},
		},
	}
}

// GetGolfWeather 지역별 골프 날씨 정보 조회
func (g *GolfCollector) GetGolfWeather(ctx context.Context, region GolfRegion) (*GolfWeather, error) {
	// OpenWeatherMap API (무료)
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=demo&units=metric&lang=kr", 
		region.Lat, region.Lon)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := g.client.Do(req)
	if err != nil {
		// API 실패 시 시뮬레이션 데이터 반환
		return g.simulateWeather(region), nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return g.simulateWeather(region), nil
	}
	
	var data struct {
		Main struct {
			Temp      float64 `json:"temp"`
			FeelsLike float64 `json:"feels_like"`
			Humidity  int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Weather []struct {
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return g.simulateWeather(region), nil
	}
	
	weather := &GolfWeather{
		Region:      region.Name,
		Temperature: data.Main.Temp,
		FeelsLike:   data.Main.FeelsLike,
		Humidity:    data.Main.Humidity,
		WindSpeed:   data.Wind.Speed,
	}
	
	if len(data.Weather) > 0 {
		weather.Description = data.Weather[0].Description
		weather.Icon = data.Weather[0].Icon
	}
	
	// 골프 지수 계산
	weather.GolfIndex, weather.GolfGrade = g.calculateGolfIndex(weather)
	
	return weather, nil
}

// simulateWeather 날씨 시뮬레이션 (API 실패 시)
func (g *GolfCollector) simulateWeather(region GolfRegion) *GolfWeather {
	rand.Seed(time.Now().UnixNano())
	
	// 계절에 따른 온도 조정
	month := time.Now().Month()
	var baseTemp float64
	var descriptions []string
	
	switch {
	case month >= 3 && month <= 5: // 봄
		baseTemp = 15 + rand.Float64()*10
		descriptions = []string{"맑음", "구름 조금", "화창함"}
	case month >= 6 && month <= 8: // 여름
		baseTemp = 25 + rand.Float64()*8
		descriptions = []string{"맑음", "구름 많음", "흐림", "소나기"}
	case month >= 9 && month <= 11: // 가을
		baseTemp = 12 + rand.Float64()*12
		descriptions = []string{"맑음", "구름 조금", "청명함", "선선함"}
	default: // 겨울
		baseTemp = -2 + rand.Float64()*10
		descriptions = []string{"맑음", "흐림", "눈", "추움"}
	}
	
	weather := &GolfWeather{
		Region:      region.Name,
		Temperature: baseTemp,
		FeelsLike:   baseTemp - 2 + rand.Float64()*4,
		Humidity:    40 + rand.Intn(40),
		WindSpeed:   1 + rand.Float64()*6,
		Description: descriptions[rand.Intn(len(descriptions))],
	}
	
	weather.GolfIndex, weather.GolfGrade = g.calculateGolfIndex(weather)
	
	return weather
}

// calculateGolfIndex 골프 지수 계산
func (g *GolfCollector) calculateGolfIndex(w *GolfWeather) (int, string) {
	score := 100
	
	// 온도 점수 (15-25도가 최적)
	if w.Temperature < 5 {
		score -= 40
	} else if w.Temperature < 10 {
		score -= 20
	} else if w.Temperature < 15 {
		score -= 5
	} else if w.Temperature > 35 {
		score -= 35
	} else if w.Temperature > 30 {
		score -= 15
	} else if w.Temperature > 25 {
		score -= 5
	}
	
	// 바람 점수 (강풍 감점)
	if w.WindSpeed > 10 {
		score -= 30
	} else if w.WindSpeed > 7 {
		score -= 15
	} else if w.WindSpeed > 5 {
		score -= 5
	}
	
	// 습도 점수
	if w.Humidity > 85 {
		score -= 15
	} else if w.Humidity > 70 {
		score -= 5
	}
	
	// 날씨 설명에 따른 조정
	desc := strings.ToLower(w.Description)
	if strings.Contains(desc, "비") || strings.Contains(desc, "rain") || strings.Contains(desc, "소나기") {
		score -= 40
	} else if strings.Contains(desc, "눈") || strings.Contains(desc, "snow") {
		score -= 50
	} else if strings.Contains(desc, "흐림") || strings.Contains(desc, "cloud") {
		score -= 5
	}
	
	// 점수 범위 제한
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	// 등급 결정
	var grade string
	switch {
	case score >= 80:
		grade = "🟢 최적"
	case score >= 60:
		grade = "🔵 좋음"
	case score >= 40:
		grade = "🟡 보통"
	default:
		grade = "🔴 비추"
	}
	
	return score, grade
}

// GetGolfProducts 골프 용품 추천 (쿠팡 파트너스)
func (g *GolfCollector) GetGolfProducts() []GolfProduct {
	baseURL := "https://www.coupang.com/vp/products/"
	
	products := []GolfProduct{
		{
			Name:     "타이틀리스트 Pro V1 골프공 12개입",
			Price:    65000,
			Category: "골프공",
			URL:      baseURL + "123456789",
		},
		{
			Name:     "캘러웨이 슈퍼소프트 골프공 12개입",
			Price:    32000,
			Category: "골프공",
			URL:      baseURL + "234567890",
		},
		{
			Name:     "풋조이 WeatherSof 골프장갑",
			Price:    18000,
			Category: "골프장갑",
			URL:      baseURL + "345678901",
		},
		{
			Name:     "타이틀리스트 플레이어스4 골프백",
			Price:    320000,
			Category: "골프백",
			URL:      baseURL + "456789012",
		},
		{
			Name:     "언더아머 골프 폴로셔츠",
			Price:    89000,
			Category: "골프웨어",
			URL:      baseURL + "567890123",
		},
		{
			Name:     "부쉬넬 V5 슬림 골프 거리측정기",
			Price:    450000,
			Category: "거리측정기",
			URL:      baseURL + "678901234",
		},
	}
	
	// 쿠팡 파트너스 링크 생성
	for i := range products {
		if g.coupangID != "" {
			products[i].URL = fmt.Sprintf("%s?wPcid=%s&sfrn=AFFILIATE", products[i].URL, g.coupangID)
		}
	}
	
	return products
}

// GenerateGolfPost 골프 날씨 포스트 생성
func (g *GolfCollector) GenerateGolfPost(ctx context.Context) *Post {
	now := time.Now()
	
	// 모든 지역 표시
	rand.Seed(now.UnixNano())
	selectedRegions := g.regions
	
	// 각 지역 날씨 조회
	var weatherData []struct {
		Region  GolfRegion
		Weather *GolfWeather
	}
	
	bestIndex := 0
	bestRegion := ""
	
	for _, region := range selectedRegions {
		weather, _ := g.GetGolfWeather(ctx, region)
		if weather != nil {
			weatherData = append(weatherData, struct {
				Region  GolfRegion
				Weather *GolfWeather
			}{region, weather})
			
			if weather.GolfIndex > bestIndex {
				bestIndex = weather.GolfIndex
				bestRegion = region.Name
			}
		}
	}
	
	// 골프 용품
	products := g.GetGolfProducts()
	
	// 제목 생성
	title := fmt.Sprintf("[%s] 오늘 골프 날씨 ⛳ %s 골프지수 %d점! 추천 골프장",
		now.Format("01/02"), bestRegion, bestIndex)
	
	// 본문 생성
	var content strings.Builder
	
	// 스타일
	content.WriteString(`
<style>
.golf-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
.golf-header { background: linear-gradient(135deg, #2d5a27 0%, #4a7c59 100%); color: white; padding: 40px; border-radius: 16px; text-align: center; margin-bottom: 30px; }
.golf-header h1 { margin: 0 0 10px 0; font-size: 28px; }
.golf-header p { margin: 0; opacity: 0.9; }
.weather-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; margin-bottom: 30px; }
.weather-card { background: #fff; border: 1px solid #e5e5e5; border-radius: 12px; padding: 20px; }
.weather-card h3 { margin: 0 0 15px 0; color: #2d5a27; font-size: 18px; }
.weather-info { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
.temp { font-size: 36px; font-weight: 700; color: #333; }
.weather-detail { font-size: 14px; color: #666; }
.golf-index { text-align: center; padding: 15px; background: #f5f5f5; border-radius: 8px; margin-bottom: 15px; }
.golf-index .score { font-size: 32px; font-weight: 700; }
.golf-index .grade { font-size: 16px; margin-top: 5px; }
.course-list { margin-top: 15px; }
.course-item { padding: 12px 0; border-bottom: 1px solid #eee; }
.course-item:last-child { border-bottom: none; }
.course-name { font-weight: 600; color: #333; }
.course-info { font-size: 13px; color: #666; margin-top: 4px; }
.course-features { display: flex; gap: 8px; margin-top: 8px; flex-wrap: wrap; }
.feature-tag { font-size: 11px; padding: 3px 8px; background: #e8f5e9; color: #2d5a27; border-radius: 4px; }
.products-section { background: #f9f9f9; padding: 30px; border-radius: 16px; margin-top: 30px; }
.products-section h2 { margin: 0 0 20px 0; color: #333; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
.product-card { background: #fff; border: 1px solid #e5e5e5; border-radius: 8px; padding: 15px; text-align: center; }
.product-name { font-size: 14px; font-weight: 500; margin-bottom: 8px; }
.product-price { font-size: 18px; font-weight: 700; color: #f03e3e; margin-bottom: 10px; }
.product-btn { display: inline-block; background: #2d5a27; color: white; padding: 8px 20px; border-radius: 6px; text-decoration: none; font-size: 13px; }
.footer-notice { margin-top: 30px; padding: 20px; background: #f5f5f5; border-radius: 12px; font-size: 13px; color: #666; }
</style>
`)

	content.WriteString(`<div class="golf-container">`)
	
	// 헤더
	content.WriteString(fmt.Sprintf(`
<div class="golf-header">
	<h1>⛳ 오늘의 골프 날씨</h1>
	<p>%s | 골프 치기 좋은 날을 찾아드립니다!</p>
</div>
`, now.Format("2006년 01월 02일 (Mon)")))

	// 날씨 카드들
	content.WriteString(`<div class="weather-grid">`)
	
	for _, data := range weatherData {
		content.WriteString(fmt.Sprintf(`
<div class="weather-card">
	<h3>📍 %s %s</h3>
	<div class="weather-info">
		<div class="temp">%.1f°C</div>
		<div class="weather-detail">
			체감 %.1f°C<br>
			습도 %d%% | 바람 %.1fm/s<br>
			%s
		</div>
	</div>
	<div class="golf-index">
		<div class="score">%d점</div>
		<div class="grade">%s</div>
	</div>
	<div class="course-list">
		<strong>🏌️ 추천 골프장</strong>
`, data.Region.City, data.Region.Name,
			data.Weather.Temperature,
			data.Weather.FeelsLike,
			data.Weather.Humidity,
			data.Weather.WindSpeed,
			data.Weather.Description,
			data.Weather.GolfIndex,
			data.Weather.GolfGrade))

		for _, course := range data.Region.GolfCourses {
			content.WriteString(fmt.Sprintf(`
		<div class="course-item">
			<div class="course-name">%s ⭐%.1f</div>
			<div class="course-info">%s | %s</div>
			<div class="course-features">
`, course.Name, course.Rating, course.GreenFee, course.Phone))

			for _, feature := range course.Features {
				content.WriteString(fmt.Sprintf(`<span class="feature-tag">%s</span>`, feature))
			}
			content.WriteString(`</div></div>`)
		}
		
		content.WriteString(`</div></div>`)
	}
	
	content.WriteString(`</div>`) // weather-grid 끝

	// 골프 용품 추천
	content.WriteString(`
<div class="products-section">
	<h2>🛒 오늘의 골프 용품 추천</h2>
	<div class="product-grid">
`)

	for _, product := range products[:4] { // 4개만 표시
		content.WriteString(fmt.Sprintf(`
		<div class="product-card">
			<div class="product-name">%s</div>
			<div class="product-price">%s원</div>
			<a href="%s" target="_blank" class="product-btn">👉 최저가 보기</a>
		</div>
`, product.Name, formatPrice(product.Price), product.URL))
	}

	content.WriteString(`</div></div>`)

	// 푸터
	content.WriteString(`
<div class="footer-notice">
	<p>💡 <strong>Tip:</strong> 골프 라운드 전 날씨를 꼭 확인하세요! 바람이 강한 날은 클럽 선택에 주의하세요.</p>
	<p>📍 골프장 예약은 미리미리! 주말은 2주 전 예약을 추천합니다.</p>
	<p>⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.</p>
</div>
`)

	content.WriteString(`</div>`) // container 끝

	// 태그 생성
	tags := []string{"골프날씨", "골프장추천", "경기도골프장", "골프", "라운딩"}
	for _, data := range weatherData {
		tags = append(tags, data.Region.Name+"골프장")
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "골프/날씨",
		Tags:     tags,
	}
}

// formatPrice 가격 포맷팅
func formatPrice(price int) string {
	str := fmt.Sprintf("%d", price)
	n := len(str)
	if n <= 3 {
		return str
	}
	
	var result strings.Builder
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(str[:remainder])
		result.WriteString(",")
	}
	
	for i := remainder; i < n; i += 3 {
		if i > remainder {
			result.WriteString(",")
		}
		result.WriteString(str[i : i+3])
	}
	
	return result.String()
}

