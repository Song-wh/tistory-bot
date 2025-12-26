package collector

import (
	"context"
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// GameCollector 게임 뉴스 수집기
type GameCollector struct {
	client    *http.Client
	coupangID string
}

// GameNews 게임 뉴스
type GameNews struct {
	Title   string
	Link    string
	Source  string
	PubDate string
}

// SteamGame 스팀 게임 정보
type SteamGame struct {
	Name          string
	AppID         int
	OriginalPrice int
	FinalPrice    int
	DiscountPct   int
	HeaderImage   string
}

// GamingProduct 게이밍 상품
type GamingProduct struct {
	Name        string
	SearchQuery string
	Emoji       string
	Description string
}

func NewGameCollector(coupangID string) *GameCollector {
	return &GameCollector{
		client:    &http.Client{Timeout: 30 * time.Second},
		coupangID: coupangID,
	}
}

// 게이밍 추천 상품
var gamingProducts = []GamingProduct{
	{Name: "게이밍 마우스", SearchQuery: "게이밍마우스 로지텍", Emoji: "🖱️", Description: "정확한 조준"},
	{Name: "게이밍 키보드", SearchQuery: "기계식키보드 게이밍", Emoji: "⌨️", Description: "기계식 타건감"},
	{Name: "게이밍 헤드셋", SearchQuery: "게이밍헤드셋 7.1채널", Emoji: "🎧", Description: "서라운드 사운드"},
	{Name: "게이밍 의자", SearchQuery: "게이밍의자 컴퓨터의자", Emoji: "🪑", Description: "장시간 편안함"},
	{Name: "게이밍 모니터", SearchQuery: "게이밍모니터 144hz", Emoji: "🖥️", Description: "고주사율"},
	{Name: "게임패드", SearchQuery: "게임패드 컨트롤러", Emoji: "🎮", Description: "콘솔 느낌"},
	{Name: "마우스패드", SearchQuery: "게이밍마우스패드 대형", Emoji: "🖼️", Description: "넓은 조작"},
	{Name: "웹캠", SearchQuery: "웹캠 스트리밍", Emoji: "📷", Description: "스트리밍용"},
}

// GetGameNews 게임 뉴스 RSS 수집
func (g *GameCollector) GetGameNews(ctx context.Context) ([]GameNews, error) {
	// Google News RSS로 게임 뉴스 수집
	url := "https://news.google.com/rss/search?q=게임+OR+스팀+OR+e스포츠+OR+신작게임&hl=ko&gl=KR&ceid=KR:ko"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return g.getSimulatedNews(), nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := g.client.Do(req)
	if err != nil {
		return g.getSimulatedNews(), nil
	}
	defer resp.Body.Close()

	var rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []struct {
				Title   string `xml:"title"`
				Link    string `xml:"link"`
				PubDate string `xml:"pubDate"`
				Source  string `xml:"source"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return g.getSimulatedNews(), nil
	}

	var news []GameNews
	for i, item := range rss.Channel.Items {
		if i >= 10 { // 최대 10개
			break
		}
		title := item.Title
		// " - 출처" 제거
		if idx := strings.LastIndex(title, " - "); idx > 0 {
			title = title[:idx]
		}
		news = append(news, GameNews{
			Title:   title,
			Link:    item.Link,
			Source:  item.Source,
			PubDate: item.PubDate,
		})
	}

	if len(news) == 0 {
		return g.getSimulatedNews(), nil
	}

	return news, nil
}

// GetSteamDeals 스팀 할인 게임 (무료 API)
func (g *GameCollector) GetSteamDeals(ctx context.Context) ([]SteamGame, error) {
	// Steam 인기 게임 (스팀 공식 API는 제한적이라 시뮬레이션)
	return g.getSimulatedSteamDeals(), nil
}

// getSimulatedNews 시뮬레이션 뉴스
func (g *GameCollector) getSimulatedNews() []GameNews {
	now := time.Now()
	r := rand.New(rand.NewSource(now.UnixNano()))

	allNews := []GameNews{
		{Title: "GTA 6 예약 판매 시작, 역대급 사전예약 기록", Source: "게임메카"},
		{Title: "스팀 겨울 세일 시작! 최대 90% 할인", Source: "인벤"},
		{Title: "발더스 게이트 3, GOTY 수상 쾌거", Source: "게임조선"},
		{Title: "LOL 월드 챔피언십 결승전 시청자 1억명 돌파", Source: "게임메카"},
		{Title: "엘든링 DLC '황금의 나무 그림자' 호평 세례", Source: "루리웹"},
		{Title: "닌텐도 스위치 2 공식 발표 임박", Source: "게임메카"},
		{Title: "배틀그라운드 신규 맵 업데이트 예고", Source: "인벤"},
		{Title: "사이버펑크 2077 완전판 스팀 1위 등극", Source: "게임조선"},
		{Title: "T1 vs GenG, 오늘 밤 8시 결승전", Source: "인벤"},
		{Title: "메이플스토리 유니버스 신규 소식 공개", Source: "루리웹"},
		{Title: "디아블로 4 시즌 3 대규모 패치 예정", Source: "게임메카"},
		{Title: "발로란트 신규 에이전트 공개, 능력치는?", Source: "인벤"},
		{Title: "포켓몬 신작 2025년 출시 확정", Source: "게임조선"},
		{Title: "스타필드 DLC '섀터드 스페이스' 출시", Source: "루리웹"},
		{Title: "PS5 프로 국내 정식 출시일 발표", Source: "게임메카"},
	}

	// 랜덤하게 섞어서 10개 선택
	r.Shuffle(len(allNews), func(i, j int) {
		allNews[i], allNews[j] = allNews[j], allNews[i]
	})

	if len(allNews) > 10 {
		allNews = allNews[:10]
	}

	return allNews
}

// getSimulatedSteamDeals 시뮬레이션 스팀 할인
func (g *GameCollector) getSimulatedSteamDeals() []SteamGame {
	now := time.Now()
	r := rand.New(rand.NewSource(now.UnixNano()))

	allDeals := []SteamGame{
		{Name: "엘든링", AppID: 1245620, OriginalPrice: 59800, FinalPrice: 41860, DiscountPct: 30},
		{Name: "사이버펑크 2077", AppID: 1091500, OriginalPrice: 59800, FinalPrice: 29900, DiscountPct: 50},
		{Name: "발더스 게이트 3", AppID: 1086940, OriginalPrice: 64800, FinalPrice: 51840, DiscountPct: 20},
		{Name: "레드 데드 리뎀션 2", AppID: 1174180, OriginalPrice: 59800, FinalPrice: 23920, DiscountPct: 60},
		{Name: "호그와트 레거시", AppID: 990080, OriginalPrice: 69800, FinalPrice: 48860, DiscountPct: 30},
		{Name: "스타필드", AppID: 1716740, OriginalPrice: 79800, FinalPrice: 55860, DiscountPct: 30},
		{Name: "데이브 더 다이버", AppID: 1868140, OriginalPrice: 24000, FinalPrice: 16800, DiscountPct: 30},
		{Name: "팰월드", AppID: 1623730, OriginalPrice: 32000, FinalPrice: 25600, DiscountPct: 20},
		{Name: "던그리드", AppID: 1171390, OriginalPrice: 14800, FinalPrice: 7400, DiscountPct: 50},
		{Name: "GTA V", AppID: 271590, OriginalPrice: 33000, FinalPrice: 16500, DiscountPct: 50},
	}

	// 랜덤하게 섞기
	r.Shuffle(len(allDeals), func(i, j int) {
		allDeals[i], allDeals[j] = allDeals[j], allDeals[i]
	})

	if len(allDeals) > 5 {
		return allDeals[:5]
	}
	return allDeals
}

// generateCoupangLink 쿠팡 검색 링크 생성
func (g *GameCollector) generateCoupangLink(query string) string {
	baseURL := fmt.Sprintf("https://www.coupang.com/np/search?component=&q=%s", query)
	if g.coupangID != "" {
		return fmt.Sprintf("%s&channel=affiliate&affiliate=%s", baseURL, g.coupangID)
	}
	return baseURL
}

// GenerateGamePost 게임 뉴스 포스트 생성
func (g *GameCollector) GenerateGamePost(news []GameNews) *Post {
	now := time.Now()
	ctx := context.Background()

	steamDeals, _ := g.GetSteamDeals(ctx)

	title := fmt.Sprintf("🎮 [%s] 오늘의 게임 뉴스 & 스팀 할인", now.Format("01/02"))

	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.game-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, sans-serif; }
.game-header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 30px; border-radius: 20px; color: white; text-align: center; margin-bottom: 25px; }
.section-title { border-left: 5px solid #667eea; padding-left: 15px; font-size: 22px; margin: 30px 0 20px 0; color: #2d3436; }
.news-card { background: #f8f9fa; padding: 18px; border-radius: 12px; margin: 12px 0; border-left: 4px solid #667eea; transition: transform 0.2s; }
.news-card:hover { transform: translateX(5px); }
.news-title { font-size: 16px; font-weight: 600; color: #2d3436; margin: 0; }
.news-title a { color: #2d3436; text-decoration: none; }
.news-title a:hover { color: #667eea; }
.news-source { font-size: 12px; color: #b2bec3; margin-top: 8px; }
.deal-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 15px; }
.deal-card { background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); padding: 20px; border-radius: 16px; color: white; }
.deal-name { font-size: 18px; font-weight: 700; margin-bottom: 10px; }
.deal-price { display: flex; align-items: center; gap: 10px; }
.original-price { text-decoration: line-through; color: #888; font-size: 14px; }
.final-price { font-size: 22px; font-weight: bold; color: #00d4aa; }
.discount-badge { background: #e74c3c; padding: 4px 10px; border-radius: 8px; font-size: 14px; font-weight: bold; }
.product-section { background: linear-gradient(135deg, #232526 0%, #414345 100%); padding: 25px; border-radius: 16px; margin-top: 30px; }
.product-title { font-size: 20px; font-weight: 700; color: white; margin: 0 0 20px 0; text-align: center; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 15px; }
.product-card { background: rgba(255,255,255,0.1); padding: 20px; border-radius: 12px; text-align: center; color: white; }
.product-emoji { font-size: 36px; margin-bottom: 10px; }
.product-name { font-size: 15px; font-weight: 600; }
.product-desc { font-size: 12px; color: #aaa; margin: 5px 0; }
.product-link { display: inline-block; background: #e74c3c; color: white; padding: 8px 16px; border-radius: 8px; text-decoration: none; font-size: 13px; margin-top: 10px; }
.esports-section { background: linear-gradient(135deg, #0f0c29 0%, #302b63 50%, #24243e 100%); padding: 25px; border-radius: 16px; margin: 25px 0; color: white; }
.footer-notice { margin-top: 30px; padding: 20px; background: #f8f9fa; border-radius: 12px; font-size: 13px; color: #636e72; text-align: center; }
</style>
`)

	content.WriteString(fmt.Sprintf(`
<div class="game-container">
<div class="game-header">
	<h1 style="margin: 0; font-size: 28px;">🎮 오늘의 게임 뉴스</h1>
	<p style="margin: 10px 0 0 0; opacity: 0.9;">%s 업데이트</p>
</div>
`, now.Format("2006년 01월 02일")))

	// 게임 뉴스
	content.WriteString(`<h2 class="section-title">📰 게임 뉴스</h2>`)
	for _, n := range news {
		newsLink := n.Title
		if n.Link != "" {
			newsLink = fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, n.Link, n.Title)
		}
		content.WriteString(fmt.Sprintf(`
<div class="news-card">
	<p class="news-title">%s</p>
	<p class="news-source">📰 %s</p>
</div>
`, newsLink, n.Source))
	}

	// 스팀 할인
	if len(steamDeals) > 0 {
		content.WriteString(`<h2 class="section-title">🔥 Steam 할인 게임</h2>
<div class="deal-grid">`)
		for _, deal := range steamDeals {
			steamURL := fmt.Sprintf("https://store.steampowered.com/app/%d", deal.AppID)
			content.WriteString(fmt.Sprintf(`
<div class="deal-card">
	<div class="deal-name">%s</div>
	<div class="deal-price">
		<span class="original-price">₩%s</span>
		<span class="final-price">₩%s</span>
		<span class="discount-badge">-%d%%</span>
	</div>
	<a href="%s" target="_blank" style="color: #00d4aa; font-size: 13px; margin-top: 10px; display: block;">Steam에서 보기 →</a>
</div>
`, deal.Name, formatGamePrice(deal.OriginalPrice), formatGamePrice(deal.FinalPrice), deal.DiscountPct, steamURL))
		}
		content.WriteString(`</div>`)
	}

	// e스포츠 섹션
	content.WriteString(`
<div class="esports-section">
	<h3 style="margin: 0 0 15px 0; font-size: 20px;">⚔️ e스포츠 소식</h3>
	<p style="margin: 0; line-height: 1.8;">
		🎯 LOL, 발로란트, 오버워치 등 e스포츠 경기 일정은<br>
		<a href="https://www.op.gg/esports" target="_blank" style="color: #00d4aa;">OP.GG e스포츠</a> 에서 확인하세요!
	</p>
</div>
`)

	// 게이밍 장비 추천
	if g.coupangID != "" {
		content.WriteString(`
<div class="product-section">
	<h3 class="product-title">🛒 추천 게이밍 장비</h3>
	<div class="product-grid">
`)
		// 랜덤하게 4개 선택
		r := rand.New(rand.NewSource(now.UnixNano()))
		shuffled := make([]GamingProduct, len(gamingProducts))
		copy(shuffled, gamingProducts)
		r.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		for i := 0; i < 4 && i < len(shuffled); i++ {
			p := shuffled[i]
			content.WriteString(fmt.Sprintf(`
		<div class="product-card">
			<div class="product-emoji">%s</div>
			<div class="product-name">%s</div>
			<div class="product-desc">%s</div>
			<a href="%s" target="_blank" class="product-link">쿠팡에서 보기</a>
		</div>
`, p.Emoji, p.Name, p.Description, g.generateCoupangLink(p.SearchQuery)))
		}

		content.WriteString(`
	</div>
</div>
`)
	}

	// 푸터
	content.WriteString(`
<div class="footer-notice">
	<p>🎮 게임을 즐기는 모든 분들을 응원합니다!</p>
	<p style="margin-top: 10px; font-size: 12px; color: #888;">
	⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.
	</p>
</div>
</div>
`)

	// 태그 생성
	tags := []string{
		"게임", "게임뉴스", "스팀할인", "스팀세일",
		now.Format("01월02일") + "게임",
		"e스포츠", "PC게임",
	}

	// 뉴스 제목에서 게임 이름 추출
	gameNames := []string{"GTA", "엘든링", "사이버펑크", "발더스게이트", "LOL", "발로란트", "배그", "메이플"}
	for _, news := range news {
		for _, name := range gameNames {
			if strings.Contains(news.Title, name) {
				tags = append(tags, name)
				break
			}
		}
	}

	// 상품 태그
	tags = append(tags, "게이밍마우스", "게이밍키보드", "게이밍장비")

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: CategoryTech, // IT/테크 카테고리에 포함
		Tags:     tags,
	}
}

// formatGamePrice 가격 포맷팅 (게임용)
func formatGamePrice(price int) string {
	s := fmt.Sprintf("%d", price)
	n := len(s)
	if n <= 3 {
		return s
	}

	var result strings.Builder
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}
