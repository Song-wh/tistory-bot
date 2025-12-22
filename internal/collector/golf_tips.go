package collector

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GolfTipsCollector 골프 레슨 팁 + 용품 추천 수집기
type GolfTipsCollector struct {
	coupangID string
	tips      []GolfTip
	products  []GolfEquipment
}

// GolfTip 골프 레슨 팁
type GolfTip struct {
	Category    string   // 카테고리 (드라이버, 아이언, 퍼팅, 어프로치 등)
	Title       string   // 제목
	Description string   // 설명
	Steps       []string // 단계별 설명
	ProTip      string   // 프로 팁
	CommonError string   // 흔한 실수
	ImageURL    string   // 이미지 (선택)
}

// GolfEquipment 골프 용품
type GolfEquipment struct {
	Name        string
	Category    string // 클럽, 공, 장갑, 의류, 악세서리
	Price       int
	Description string
	Features    []string
	ProductID   string
	Rating      float64
}

// NewGolfTipsCollector 생성자
func NewGolfTipsCollector(coupangID string) *GolfTipsCollector {
	return &GolfTipsCollector{
		coupangID: coupangID,
		tips:      getGolfTips(),
		products:  getGolfEquipments(),
	}
}

// getGolfTips 골프 레슨 팁 데이터
func getGolfTips() []GolfTip {
	return []GolfTip{
		// 드라이버
		{
			Category:    "드라이버",
			Title:       "드라이버 비거리 늘리는 3가지 핵심 포인트",
			Description: "비거리를 늘리기 위한 가장 중요한 3가지 요소를 알려드립니다.",
			Steps: []string{
				"1️⃣ 어드레스에서 공을 왼발 안쪽에 위치시키세요",
				"2️⃣ 백스윙 시 어깨 회전을 90도 이상 충분히 하세요",
				"3️⃣ 다운스윙에서 하체 리드를 먼저 시작하세요",
			},
			ProTip:      "프로들은 임팩트 순간 왼쪽 벽을 만들어 파워를 극대화합니다",
			CommonError: "상체로 먼저 치려고 하면 슬라이스가 납니다",
		},
		{
			Category:    "드라이버",
			Title:       "슬라이스 교정하는 확실한 방법",
			Description: "아마추어 골퍼의 80%가 겪는 슬라이스, 이렇게 고치세요!",
			Steps: []string{
				"1️⃣ 그립을 스트롱 그립으로 잡으세요 (왼손 너클 3개 보이게)",
				"2️⃣ 백스윙 탑에서 클럽 페이스가 하늘을 향하게 하세요",
				"3️⃣ 다운스윙에서 인사이드-아웃 경로로 스윙하세요",
			},
			ProTip:      "연습 때 발 사이에 공을 놓고 치면 인아웃 경로가 자연스럽게 됩니다",
			CommonError: "아웃사이드-인 경로가 슬라이스의 주범입니다",
		},
		{
			Category:    "드라이버",
			Title:       "티샷 정확도 높이는 루틴 만들기",
			Description: "일관된 티샷을 위한 프리샷 루틴을 배워보세요.",
			Steps: []string{
				"1️⃣ 공 뒤에서 목표를 정확히 설정하세요",
				"2️⃣ 두 번의 연습 스윙으로 리듬을 찾으세요",
				"3️⃣ 어드레스 후 3초 안에 스윙을 시작하세요",
			},
			ProTip:      "매번 같은 루틴을 지키면 긴장 상황에서도 일관된 샷이 가능합니다",
			CommonError: "어드레스에서 너무 오래 서 있으면 몸이 경직됩니다",
		},
		// 아이언
		{
			Category:    "아이언",
			Title:       "아이언 다운블로우 마스터하기",
			Description: "프로처럼 공을 찍어치는 다운블로우 비법!",
			Steps: []string{
				"1️⃣ 공 위치를 스탠스 중앙보다 약간 오른쪽에 두세요",
				"2️⃣ 손이 항상 클럽헤드보다 앞서가게 하세요",
				"3️⃣ 임팩트 후에도 손목 각도를 유지하세요",
			},
			ProTip:      "디봇이 공 앞쪽에 생겨야 정확한 다운블로우입니다",
			CommonError: "공을 띄우려고 손목을 풀면 토핑이 납니다",
		},
		{
			Category:    "아이언",
			Title:       "거리 컨트롤의 핵심, 하프스윙 연습법",
			Description: "100야드 이내 거리 컨트롤을 완벽하게!",
			Steps: []string{
				"1️⃣ 9시-3시 스윙으로 기본 거리를 파악하세요",
				"2️⃣ 10시-2시 스윙으로 풀스윙 대비 80% 거리",
				"3️⃣ 그립을 짧게 잡으면 10% 거리 감소",
			},
			ProTip:      "각 클럽별 하프스윙 거리를 메모해두세요",
			CommonError: "거리 조절을 스윙 속도로 하면 일관성이 떨어집니다",
		},
		// 퍼팅
		{
			Category:    "퍼팅",
			Title:       "3퍼트 없애는 거리감 연습법",
			Description: "롱퍼팅 거리감을 잡아 3퍼트를 없애세요!",
			Steps: []string{
				"1️⃣ 10m, 15m, 20m 거리별 스트로크 크기를 정하세요",
				"2️⃣ 공을 보지 말고 홀을 보며 연습 스트로크하세요",
				"3️⃣ 거리에 집중하고 방향은 70%만 맞추세요",
			},
			ProTip:      "롱퍼팅은 홀에 넣기보다 1m 반경 안에 붙이는 게 목표입니다",
			CommonError: "방향에 집중하면 거리감을 잃습니다",
		},
		{
			Category:    "퍼팅",
			Title:       "숏퍼팅 자신감 키우기",
			Description: "1m 퍼팅을 100% 성공하는 방법!",
			Steps: []string{
				"1️⃣ 공을 스탠스 중앙 왼쪽 눈 아래에 위치시키세요",
				"2️⃣ 어깨로만 스트로크하고 손목은 고정하세요",
				"3️⃣ 홀 뒤쪽 가장자리를 노리고 치세요",
			},
			ProTip:      "매일 1m 퍼팅 연속 20개 성공 챌린지를 하세요",
			CommonError: "머리를 빨리 들면 밀거나 당깁니다",
		},
		// 어프로치
		{
			Category:    "어프로치",
			Title:       "50야드 어프로치 완벽 정복",
			Description: "애매한 50야드 거리, 이렇게 공략하세요!",
			Steps: []string{
				"1️⃣ 56도 웨지로 3/4 스윙을 기본으로 하세요",
				"2️⃣ 공 위치는 스탠스 중앙에 두세요",
				"3️⃣ 피니시를 허리 높이에서 멈추세요",
			},
			ProTip:      "클럽을 1인치 짧게 잡으면 컨트롤이 좋아집니다",
			CommonError: "풀스윙하고 속도를 줄이면 미스샷이 납니다",
		},
		{
			Category:    "어프로치",
			Title:       "벙커샷 두려움 극복하기",
			Description: "벙커가 더 이상 무섭지 않아요!",
			Steps: []string{
				"1️⃣ 페이스를 열고 스탠스도 열어주세요",
				"2️⃣ 공 5cm 뒤 모래를 목표로 치세요",
				"3️⃣ 모래를 홀 방향으로 던진다고 생각하세요",
			},
			ProTip:      "벙커에서는 가속하면서 쳐야 합니다, 감속하면 안 돼요",
			CommonError: "공을 직접 맞추려 하면 토핑이 납니다",
		},
		// 멘탈
		{
			Category:    "멘탈",
			Title:       "라운드 중 멘탈 관리법",
			Description: "나쁜 샷 후에도 평정심을 유지하는 방법!",
			Steps: []string{
				"1️⃣ 나쁜 샷 후 심호흡 3번을 하세요",
				"2️⃣ 다음 샷에만 집중하고 이전 샷은 잊으세요",
				"3️⃣ 18홀 전체로 생각하고 한 홀에 연연하지 마세요",
			},
			ProTip:      "프로들도 미스샷을 합니다, 회복력이 중요합니다",
			CommonError: "화를 내면 다음 샷도 망칩니다",
		},
		{
			Category:    "멘탈",
			Title:       "첫 티샷 긴장 극복하기",
			Description: "첫 홀 긴장감을 이겨내는 방법!",
			Steps: []string{
				"1️⃣ 라운드 전 충분한 퍼팅 연습으로 워밍업하세요",
				"2️⃣ 첫 티샷은 드라이버 대신 안전한 클럽을 고려하세요",
				"3️⃣ 결과보다 스윙 리듬에 집중하세요",
			},
			ProTip:      "프로들도 첫 홀은 안전하게 플레이합니다",
			CommonError: "첫 홀부터 무리하면 전체 라운드가 망가집니다",
		},
		// 코스 전략
		{
			Category:    "코스전략",
			Title:       "스코어 줄이는 코스 매니지먼트",
			Description: "무리한 샷 대신 현명한 선택으로 스코어를 줄이세요!",
			Steps: []string{
				"1️⃣ OB가 있는 쪽 반대로 조준하세요",
				"2️⃣ 핀이 어려운 위치면 그린 중앙을 노리세요",
				"3️⃣ 파5에서 무리하게 투온 시도하지 마세요",
			},
			ProTip:      "보기 없는 골프가 80대 비결입니다",
			CommonError: "영웅 샷을 시도하면 더블 보기가 됩니다",
		},
	}
}

// getGolfEquipments 골프 용품 데이터
func getGolfEquipments() []GolfEquipment {
	return []GolfEquipment{
		// 골프공
		{Name: "타이틀리스트 Pro V1 12개입", Category: "골프공", Price: 65000, Description: "투어 선수들이 가장 많이 사용하는 프리미엄 골프공", Features: []string{"뛰어난 스핀", "일관된 비행", "부드러운 타감"}, Rating: 4.9, ProductID: "123456"},
		{Name: "캘러웨이 크롬소프트 12개입", Category: "골프공", Price: 55000, Description: "부드러운 타감과 긴 비거리의 조화", Features: []string{"하이퍼 엘라스틱 코어", "낮은 스핀", "긴 비거리"}, Rating: 4.7, ProductID: "123457"},
		{Name: "브리지스톤 투어B X 12개입", Category: "골프공", Price: 58000, Description: "타이거 우즈가 선택한 골프공", Features: []string{"정확한 컨트롤", "우수한 내구성", "안정적 비행"}, Rating: 4.8, ProductID: "123458"},
		{Name: "스릭슨 Z-STAR 12개입", Category: "골프공", Price: 52000, Description: "가성비 최고의 투어 볼", Features: []string{"3피스 구조", "스핀 컨트롤", "합리적 가격"}, Rating: 4.6, ProductID: "123459"},
		// 골프장갑
		{Name: "풋조이 WeatherSof 장갑", Category: "장갑", Price: 18000, Description: "세계 1위 판매 골프장갑", Features: []string{"내구성", "그립감", "통기성"}, Rating: 4.8, ProductID: "234567"},
		{Name: "타이틀리스트 플레이어스 장갑", Category: "장갑", Price: 22000, Description: "프리미엄 카브레타 가죽", Features: []string{"고급 가죽", "뛰어난 핏", "부드러운 감촉"}, Rating: 4.7, ProductID: "234568"},
		{Name: "캘러웨이 투어 오센틱 장갑", Category: "장갑", Price: 25000, Description: "투어 선수용 퍼포먼스 장갑", Features: []string{"최고급 가죽", "퍼포먼스 핏", "땀 흡수"}, Rating: 4.6, ProductID: "234569"},
		// 거리측정기
		{Name: "부쉬넬 V5 슬림 거리측정기", Category: "거리측정기", Price: 450000, Description: "가장 인기 있는 프리미엄 거리측정기", Features: []string{"핀시커 기능", "슬로프 모드", "빠른 측정"}, Rating: 4.9, ProductID: "345678"},
		{Name: "가민 어프로치 Z82", Category: "거리측정기", Price: 580000, Description: "GPS + 레이저 하이브리드", Features: []string{"코스맵 내장", "풀컬러 디스플레이", "터치스크린"}, Rating: 4.8, ProductID: "345679"},
		{Name: "보이스캐디 T9", Category: "거리측정기", Price: 350000, Description: "가성비 좋은 거리측정기", Features: []string{"슬로프 기능", "컴팩트", "진동 알림"}, Rating: 4.5, ProductID: "345680"},
		// 골프웨어
		{Name: "나이키 드라이핏 폴로셔츠", Category: "의류", Price: 79000, Description: "시원하고 쾌적한 골프 폴로", Features: []string{"드라이핏 기술", "스트레치", "UV 차단"}, Rating: 4.6, ProductID: "456789"},
		{Name: "아디다스 골프 바지", Category: "의류", Price: 89000, Description: "편안한 스트레치 골프 팬츠", Features: []string{"4웨이 스트레치", "발수 가공", "슬림핏"}, Rating: 4.5, ProductID: "456790"},
		{Name: "언더아머 골프 벨트", Category: "악세서리", Price: 45000, Description: "스트레치 브레이드 벨트", Features: []string{"탄성 소재", "조절 가능", "가벼움"}, Rating: 4.4, ProductID: "456791"},
		// 연습용품
		{Name: "퍼팅 연습 매트 3m", Category: "연습용품", Price: 35000, Description: "실내 퍼팅 연습 필수템", Features: []string{"실제 그린과 유사", "자동 리턴", "휴대 가능"}, Rating: 4.5, ProductID: "567890"},
		{Name: "스윙 연습기", Category: "연습용품", Price: 45000, Description: "올바른 스윙 궤도 연습", Features: []string{"궤도 교정", "실내외 사용", "접이식"}, Rating: 4.3, ProductID: "567891"},
		{Name: "얼라이먼트 스틱 2개입", Category: "연습용품", Price: 15000, Description: "정렬 연습 필수 아이템", Features: []string{"다용도", "가벼움", "튼튼함"}, Rating: 4.6, ProductID: "567892"},
	}
}

// GenerateGolfTipsPost 골프 레슨 팁 포스트 생성
func (g *GolfTipsCollector) GenerateGolfTipsPost(ctx context.Context) *Post {
	now := time.Now()
	rand.Seed(now.UnixNano())

	// 오늘의 팁 선택 (랜덤 3개, 카테고리 다르게)
	categories := []string{"드라이버", "아이언", "퍼팅", "어프로치", "멘탈", "코스전략"}
	rand.Shuffle(len(categories), func(i, j int) {
		categories[i], categories[j] = categories[j], categories[i]
	})
	selectedCategories := categories[:3]

	var selectedTips []GolfTip
	for _, cat := range selectedCategories {
		for _, tip := range g.tips {
			if tip.Category == cat {
				selectedTips = append(selectedTips, tip)
				break
			}
		}
	}

	// 관련 용품 선택 (4개)
	rand.Shuffle(len(g.products), func(i, j int) {
		g.products[i], g.products[j] = g.products[j], g.products[i]
	})
	selectedProducts := g.products[:4]

	// 제목 생성
	mainTip := selectedTips[0]
	title := fmt.Sprintf("[골프레슨] %s | 오늘의 골프 팁 ⛳", mainTip.Title)

	// 본문 생성
	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.golf-tips-container { max-width: 900px; margin: 0 auto; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
.tips-header { background: linear-gradient(135deg, #1a472a 0%, #2d5a27 100%); color: white; padding: 40px; border-radius: 16px; text-align: center; margin-bottom: 30px; }
.tips-header h1 { margin: 0 0 10px 0; font-size: 26px; }
.tips-header p { margin: 0; opacity: 0.9; font-size: 14px; }
.tip-card { background: #fff; border: 1px solid #e5e5e5; border-radius: 12px; padding: 25px; margin-bottom: 25px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
.tip-category { display: inline-block; background: #2d5a27; color: white; padding: 4px 12px; border-radius: 20px; font-size: 12px; margin-bottom: 15px; }
.tip-title { font-size: 20px; font-weight: 700; color: #1a472a; margin-bottom: 10px; }
.tip-desc { color: #555; margin-bottom: 20px; line-height: 1.6; }
.tip-steps { background: #f8faf8; padding: 20px; border-radius: 8px; margin-bottom: 15px; }
.tip-steps li { padding: 8px 0; border-bottom: 1px dashed #ddd; }
.tip-steps li:last-child { border-bottom: none; }
.pro-tip { background: #fff3cd; padding: 15px; border-radius: 8px; margin-bottom: 10px; }
.pro-tip::before { content: '💡 Pro Tip: '; font-weight: 700; }
.common-error { background: #f8d7da; padding: 15px; border-radius: 8px; }
.common-error::before { content: '⚠️ 주의: '; font-weight: 700; }
.products-section { background: #f5f5f5; padding: 30px; border-radius: 16px; margin-top: 30px; }
.products-section h2 { margin: 0 0 20px 0; color: #1a472a; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
.product-card { background: white; border-radius: 10px; padding: 15px; text-align: center; }
.product-name { font-size: 14px; font-weight: 600; margin-bottom: 5px; color: #333; }
.product-price { font-size: 18px; font-weight: 700; color: #e53935; margin-bottom: 8px; }
.product-rating { font-size: 12px; color: #ffc107; margin-bottom: 10px; }
.product-btn { display: inline-block; background: #2d5a27; color: white; padding: 10px 20px; border-radius: 6px; text-decoration: none; font-size: 13px; }
.footer-note { margin-top: 30px; padding: 20px; background: #f9f9f9; border-radius: 12px; font-size: 13px; color: #666; }
</style>
`)

	content.WriteString(`<div class="golf-tips-container">`)

	// 헤더
	content.WriteString(fmt.Sprintf(`
<div class="tips-header">
	<h1>⛳ 오늘의 골프 레슨</h1>
	<p>%s | 스코어를 줄이는 실전 팁!</p>
</div>
`, now.Format("2006년 01월 02일")))

	// 팁 카드들
	for _, tip := range selectedTips {
		content.WriteString(fmt.Sprintf(`
<div class="tip-card">
	<span class="tip-category">%s</span>
	<h3 class="tip-title">%s</h3>
	<p class="tip-desc">%s</p>
	<div class="tip-steps">
		<ul style="list-style: none; padding: 0; margin: 0;">
`, tip.Category, tip.Title, tip.Description))

		for _, step := range tip.Steps {
			content.WriteString(fmt.Sprintf(`<li>%s</li>`, step))
		}

		content.WriteString(`</ul></div>`)
		content.WriteString(fmt.Sprintf(`<div class="pro-tip">%s</div>`, tip.ProTip))
		content.WriteString(fmt.Sprintf(`<div class="common-error">%s</div>`, tip.CommonError))
		content.WriteString(`</div>`)
	}

	// 추천 용품
	content.WriteString(`
<div class="products-section">
	<h2>🛒 오늘의 추천 골프용품</h2>
	<div class="product-grid">
`)

	for _, product := range selectedProducts {
		url := g.generatePartnerLink(product.ProductID)
		stars := strings.Repeat("⭐", int(product.Rating))

		content.WriteString(fmt.Sprintf(`
		<div class="product-card">
			<div class="product-name">%s</div>
			<div class="product-price">%s원</div>
			<div class="product-rating">%s %.1f</div>
			<a href="%s" target="_blank" class="product-btn">👉 최저가 보기</a>
		</div>
`, product.Name, formatPrice(product.Price), stars, product.Rating, url))
	}

	content.WriteString(`</div></div>`)

	// 푸터
	content.WriteString(`
<div class="footer-note">
	<p>📌 오늘 배운 팁을 연습장에서 꼭 연습해보세요!</p>
	<p>🏌️ 좋은 장비도 중요하지만, 꾸준한 연습이 실력 향상의 핵심입니다.</p>
	<p>⚠️ 본 포스팅은 쿠팡 파트너스 활동의 일환으로, 이에 따른 일정액의 수수료를 제공받습니다.</p>
</div>
`)

	content.WriteString(`</div>`)

	// 태그
	tags := []string{"골프레슨", "골프팁", "골프스윙", "골프연습", "골프입문", "골프용품추천"}
	for _, tip := range selectedTips {
		tags = append(tags, "골프"+tip.Category)
	}

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "골프/날씨",
		Tags:     tags,
	}
}

// generatePartnerLink 쿠팡 파트너스 링크 생성
func (g *GolfTipsCollector) generatePartnerLink(productID string) string {
	baseURL := fmt.Sprintf("https://www.coupang.com/vp/products/%s", productID)
	if g.coupangID != "" {
		return fmt.Sprintf("%s?wPcid=%s&sfrn=AFFILIATE", baseURL, g.coupangID)
	}
	return baseURL
}
