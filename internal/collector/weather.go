package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WeatherCollector 날씨 정보 수집기
type WeatherCollector struct {
	client *http.Client
}

// Weather 날씨 정보
type Weather struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temp"`
	TempMin     float64 `json:"temp_min"`
	TempMax     float64 `json:"temp_max"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

func NewWeatherCollector() *WeatherCollector {
	return &WeatherCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetWeather 주요 도시 날씨 정보 조회 (wttr.in 무료 API 사용)
func (w *WeatherCollector) GetWeather(ctx context.Context) ([]Weather, error) {
	cities := []string{"Seoul", "Busan", "Incheon", "Daegu", "Daejeon", "Gwangju", "Jeju"}
	var weathers []Weather

	for _, city := range cities {
		url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "TistoryBot/1.0")

		resp, err := w.client.Do(req)
		if err != nil {
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// 파싱
		if current, ok := result["current_condition"].([]interface{}); ok && len(current) > 0 {
			if cc, ok := current[0].(map[string]interface{}); ok {
				temp := parseFloat(cc["temp_C"])
				humidity := parseInt(cc["humidity"])
				desc := ""
				if weatherDesc, ok := cc["weatherDesc"].([]interface{}); ok && len(weatherDesc) > 0 {
					if d, ok := weatherDesc[0].(map[string]interface{}); ok {
						desc = fmt.Sprintf("%v", d["value"])
					}
				}

				weathers = append(weathers, Weather{
					City:        getCityKorean(city),
					Temperature: temp,
					Humidity:    humidity,
					Description: desc,
				})
			}
		}
	}

	return weathers, nil
}

// GenerateWeatherPost 날씨 포스트 생성
func (w *WeatherCollector) GenerateWeatherPost(weathers []Weather) *Post {
	now := time.Now()
	title := fmt.Sprintf("🌤️ 오늘의 날씨 [%s] 전국 주요 도시", now.Format("01/02"))

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<h2>🌤️ 오늘의 날씨</h2>
<p>업데이트: %s</p>

<div style="background: linear-gradient(135deg, #74b9ff 0%%, #0984e3 100%%); padding: 20px; border-radius: 15px; color: white; margin: 20px 0;">
<h3 style="color: white; margin-bottom: 20px;">📍 전국 주요 도시 날씨</h3>
`, now.Format("2006년 01월 02일 15:04")))

	for _, weather := range weathers {
		emoji := getWeatherEmoji(weather.Description)
		content.WriteString(fmt.Sprintf(`
<div style="background: rgba(255,255,255,0.2); padding: 15px; border-radius: 10px; margin-bottom: 10px; display: flex; justify-content: space-between; align-items: center;">
<span style="font-size: 1.2em;">%s %s</span>
<span style="font-size: 1.5em; font-weight: bold;">%s %.0f°C</span>
<span>습도 %d%%</span>
</div>
`, emoji, weather.City, getWeatherEmoji(weather.Description), weather.Temperature, weather.Humidity))
	}

	content.WriteString(`</div>

<h3>👔 오늘의 옷차림 추천</h3>
`)

	// 서울 기온 기준 옷차림 추천
	if len(weathers) > 0 {
		temp := weathers[0].Temperature
		content.WriteString(getClothingRecommendation(temp))
	}

	content.WriteString(`
<h3>☔ 우산 체크</h3>
<p>외출 전 기상청 레이더 영상을 확인하세요!</p>

<p style="color: #888; font-size: 0.9em; margin-top: 30px;">
※ 날씨 정보는 참고용이며, 정확한 정보는 기상청에서 확인하세요.
</p>
`)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "날씨/생활",
		Tags:     []string{"오늘날씨", "전국날씨", "날씨", "기온", "옷차림추천", now.Format("01월02일날씨")},
	}
}

func getCityKorean(city string) string {
	cities := map[string]string{
		"Seoul":   "서울",
		"Busan":   "부산",
		"Incheon": "인천",
		"Daegu":   "대구",
		"Daejeon": "대전",
		"Gwangju": "광주",
		"Jeju":    "제주",
	}
	if k, ok := cities[city]; ok {
		return k
	}
	return city
}

func getWeatherEmoji(desc string) string {
	desc = strings.ToLower(desc)
	switch {
	case strings.Contains(desc, "rain") || strings.Contains(desc, "비"):
		return "🌧️"
	case strings.Contains(desc, "snow") || strings.Contains(desc, "눈"):
		return "❄️"
	case strings.Contains(desc, "cloud") || strings.Contains(desc, "구름"):
		return "☁️"
	case strings.Contains(desc, "sun") || strings.Contains(desc, "맑"):
		return "☀️"
	case strings.Contains(desc, "fog") || strings.Contains(desc, "안개"):
		return "🌫️"
	default:
		return "🌤️"
	}
}

func getClothingRecommendation(temp float64) string {
	switch {
	case temp >= 28:
		return `<div style="background: #ff7675; padding: 15px; border-radius: 10px; color: white;">
<p><strong>🔥 무더위 (28°C 이상)</strong></p>
<p>민소매, 반팔, 반바지, 원피스</p>
</div>`
	case temp >= 23:
		return `<div style="background: #fdcb6e; padding: 15px; border-radius: 10px;">
<p><strong>☀️ 더움 (23~27°C)</strong></p>
<p>반팔, 얇은 셔츠, 면바지</p>
</div>`
	case temp >= 17:
		return `<div style="background: #74b9ff; padding: 15px; border-radius: 10px; color: white;">
<p><strong>🌤️ 따뜻함 (17~22°C)</strong></p>
<p>얇은 가디건, 긴팔, 면바지</p>
</div>`
	case temp >= 12:
		return `<div style="background: #a29bfe; padding: 15px; border-radius: 10px; color: white;">
<p><strong>🍂 선선함 (12~16°C)</strong></p>
<p>자켓, 가디건, 니트</p>
</div>`
	case temp >= 6:
		return `<div style="background: #636e72; padding: 15px; border-radius: 10px; color: white;">
<p><strong>🧥 쌀쌀함 (6~11°C)</strong></p>
<p>코트, 점퍼, 니트, 스타킹</p>
</div>`
	default:
		return `<div style="background: #2d3436; padding: 15px; border-radius: 10px; color: white;">
<p><strong>❄️ 추움 (5°C 이하)</strong></p>
<p>패딩, 두꺼운 코트, 목도리, 장갑</p>
</div>`
	}
}

func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	}
	return 0
}

func parseInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	}
	return 0
}
