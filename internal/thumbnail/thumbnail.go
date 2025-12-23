package thumbnail

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fogleman/gg"
)

// Generator 썸네일 생성기
type Generator struct {
	Width     int
	Height    int
	OutputDir string
}

// CategoryStyle 카테고리별 스타일
type CategoryStyle struct {
	GradientStart color.Color
	GradientEnd   color.Color
	Emoji         string
	SubText       string
}

// 카테고리별 스타일 정의 (이모지 대신 텍스트 아이콘 사용 - 폰트 호환성)
var categoryStyles = map[string]CategoryStyle{
	"crypto": {
		GradientStart: color.RGBA{255, 175, 0, 255},   // 골드
		GradientEnd:   color.RGBA{255, 100, 0, 255},   // 오렌지
		Emoji:         "BTC",
		SubText:       "암호화폐 시세",
	},
	"tech": {
		GradientStart: color.RGBA{0, 150, 255, 255},   // 블루
		GradientEnd:   color.RGBA{100, 50, 200, 255},  // 퍼플
		Emoji:         "TECH",
		SubText:       "IT/테크 뉴스",
	},
	"movie": {
		GradientStart: color.RGBA{220, 20, 60, 255},   // 크림슨
		GradientEnd:   color.RGBA{139, 0, 139, 255},   // 다크마젠타
		Emoji:         "MOVIE",
		SubText:       "영화/드라마",
	},
	"trend": {
		GradientStart: color.RGBA{255, 65, 108, 255},  // 핑크
		GradientEnd:   color.RGBA{255, 75, 43, 255},   // 레드오렌지
		Emoji:         "HOT",
		SubText:       "실시간 트렌드",
	},
	"lotto": {
		GradientStart: color.RGBA{50, 205, 50, 255},   // 라임그린
		GradientEnd:   color.RGBA{34, 139, 34, 255},   // 포레스트그린
		Emoji:         "LOTTO",
		SubText:       "로또 당첨번호",
	},
	"lotto-predict": {
		GradientStart: color.RGBA{138, 43, 226, 255},  // 블루바이올렛
		GradientEnd:   color.RGBA{75, 0, 130, 255},    // 인디고
		Emoji:         "AI",
		SubText:       "로또 예측",
	},
	"fortune": {
		GradientStart: color.RGBA{255, 215, 0, 255},   // 골드
		GradientEnd:   color.RGBA{255, 140, 0, 255},   // 다크오렌지
		Emoji:         "FORTUNE",
		SubText:       "오늘의 운세",
	},
	"sports": {
		GradientStart: color.RGBA{0, 184, 148, 255},   // 그린
		GradientEnd:   color.RGBA{0, 206, 201, 255},   // 시안
		Emoji:         "SPORTS",
		SubText:       "스포츠 뉴스",
	},
	"golf": {
		GradientStart: color.RGBA{46, 125, 50, 255},   // 그린
		GradientEnd:   color.RGBA{76, 175, 80, 255},   // 라이트그린
		Emoji:         "GOLF",
		SubText:       "골프 날씨",
	},
	"golf-tips": {
		GradientStart: color.RGBA{27, 94, 32, 255},    // 다크그린
		GradientEnd:   color.RGBA{56, 142, 60, 255},   // 그린
		Emoji:         "LESSON",
		SubText:       "골프 레슨",
	},
	"coupang": {
		GradientStart: color.RGBA{230, 57, 70, 255},   // 쿠팡 레드
		GradientEnd:   color.RGBA{168, 50, 62, 255},   // 다크레드
		Emoji:         "DEAL",
		SubText:       "오늘의 특가",
	},
	"error": {
		GradientStart: color.RGBA{45, 52, 54, 255},    // 다크그레이
		GradientEnd:   color.RGBA{99, 110, 114, 255},  // 그레이
		Emoji:         "DEBUG",
		SubText:       "에러 해결",
	},
}

// NewGenerator 썸네일 생성기 생성
func NewGenerator(outputDir string) *Generator {
	// 디렉토리 생성
	os.MkdirAll(outputDir, 0755)
	
	return &Generator{
		Width:     1200,
		Height:    630, // OG 이미지 권장 크기
		OutputDir: outputDir,
	}
}

// Generate 썸네일 생성
func (g *Generator) Generate(category, title string) (string, error) {
	dc := gg.NewContext(g.Width, g.Height)

	// 스타일 가져오기
	style, ok := categoryStyles[category]
	if !ok {
		style = CategoryStyle{
			GradientStart: color.RGBA{100, 100, 100, 255},
			GradientEnd:   color.RGBA{50, 50, 50, 255},
			Emoji:         "📝",
			SubText:       "블로그",
		}
	}

	// 그라데이션 배경
	gradient := gg.NewLinearGradient(0, 0, float64(g.Width), float64(g.Height))
	r1, g1, b1, _ := style.GradientStart.RGBA()
	r2, g2, b2, _ := style.GradientEnd.RGBA()
	gradient.AddColorStop(0, color.RGBA{uint8(r1 >> 8), uint8(g1 >> 8), uint8(b1 >> 8), 255})
	gradient.AddColorStop(1, color.RGBA{uint8(r2 >> 8), uint8(g2 >> 8), uint8(b2 >> 8), 255})
	dc.SetFillStyle(gradient)
	dc.DrawRectangle(0, 0, float64(g.Width), float64(g.Height))
	dc.Fill()

	// 패턴 오버레이 (약간의 텍스처)
	dc.SetColor(color.RGBA{255, 255, 255, 15})
	for i := 0; i < g.Width; i += 30 {
		dc.DrawLine(float64(i), 0, float64(i+100), float64(g.Height))
		dc.SetLineWidth(1)
		dc.Stroke()
	}

	// 반투명 박스 (텍스트 가독성)
	dc.SetColor(color.RGBA{0, 0, 0, 80})
	dc.DrawRoundedRectangle(60, 150, float64(g.Width-120), float64(g.Height-200), 20)
	dc.Fill()

	// 카테고리 아이콘 텍스트 (큰 사이즈, 스타일리시)
	dc.SetColor(color.White)
	if err := g.loadFont(dc, 60); err == nil {
		// 배지 스타일 배경
		textWidth, _ := dc.MeasureString(style.Emoji)
		dc.SetColor(color.RGBA{255, 255, 255, 40})
		dc.DrawRoundedRectangle(float64(g.Width)/2-textWidth/2-20, 60, textWidth+40, 80, 15)
		dc.Fill()
		
		// 텍스트
		dc.SetColor(color.White)
		dc.DrawStringAnchored(style.Emoji, float64(g.Width)/2, 100, 0.5, 0.5)
	}

	// 제목 텍스트
	if err := g.loadFont(dc, 48); err == nil {
		// 제목이 너무 길면 자르기
		displayTitle := truncateText(title, 25)
		dc.SetColor(color.White)
		dc.DrawStringAnchored(displayTitle, float64(g.Width)/2, float64(g.Height)/2-20, 0.5, 0.5)
	}

	// 서브 텍스트
	if err := g.loadFont(dc, 28); err == nil {
		dc.SetColor(color.RGBA{255, 255, 255, 200})
		dc.DrawStringAnchored(style.SubText, float64(g.Width)/2, float64(g.Height)/2+50, 0.5, 0.5)
	}

	// 날짜
	if err := g.loadFont(dc, 20); err == nil {
		dateStr := time.Now().Format("2006.01.02")
		dc.SetColor(color.RGBA{255, 255, 255, 150})
		dc.DrawStringAnchored(dateStr, float64(g.Width)/2, float64(g.Height)-80, 0.5, 0.5)
	}

	// 브랜드 로고/텍스트
	if err := g.loadFont(dc, 18); err == nil {
		dc.SetColor(color.RGBA{255, 255, 255, 120})
		dc.DrawStringAnchored("🔗 song-circle.tistory.com", float64(g.Width)/2, float64(g.Height)-50, 0.5, 0.5)
	}

	// 파일 저장
	filename := fmt.Sprintf("%s_%d.png", category, time.Now().UnixNano())
	filepath := filepath.Join(g.OutputDir, filename)
	
	if err := dc.SavePNG(filepath); err != nil {
		return "", err
	}

	return filepath, nil
}

// loadFont 폰트 로드 (시스템 폰트 사용)
func (g *Generator) loadFont(dc *gg.Context, size float64) error {
	// Windows 한글 폰트 경로들
	fontPaths := []string{
		"C:/Windows/Fonts/malgun.ttf",      // 맑은 고딕
		"C:/Windows/Fonts/NanumGothic.ttf", // 나눔고딕
		"C:/Windows/Fonts/gulim.ttc",       // 굴림
		"C:/Windows/Fonts/arial.ttf",       // Arial
		"/usr/share/fonts/truetype/nanum/NanumGothic.ttf", // Linux
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			if err := dc.LoadFontFace(path, size); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("폰트를 찾을 수 없음")
}

// truncateText 텍스트 자르기 (한글 고려)
func truncateText(text string, maxLen int) string {
	// 대괄호 내용 제거 (날짜 등)
	if idx := strings.Index(text, "]"); idx > 0 && idx < 20 {
		text = strings.TrimSpace(text[idx+1:])
	}

	runes := []rune(text)
	if utf8.RuneCountInString(text) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// GenerateForPost 포스트용 썸네일 생성 (카테고리 자동 감지)
func (g *Generator) GenerateForPost(category, title string) (string, error) {
	return g.Generate(category, title)
}

// Cleanup 오래된 썸네일 삭제 (1일 이상)
func (g *Generator) Cleanup() error {
	entries, err := os.ReadDir(g.OutputDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(g.OutputDir, entry.Name()))
		}
	}
	return nil
}

