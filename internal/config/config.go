package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 설정
type Config struct {
	// 다중 계정 지원
	Accounts []AccountConfig `yaml:"accounts"`

	// 전역 설정 (하위 호환성)
	Tistory      TistoryConfig      `yaml:"tistory"`
	Browser      BrowserConfig      `yaml:"browser"`
	TMDB         TMDBConfig         `yaml:"tmdb"`
	Naver        NaverConfig        `yaml:"naver"`
	Coupang      CoupangConfig      `yaml:"coupang"`
	FootballData *FootballDataConfig `yaml:"football_data"` // 스포츠 API (선택)
	Thumbnail    *ThumbnailConfig   `yaml:"thumbnail"`     // 썸네일 설정 (선택)
	Categories   map[string]string  `yaml:"categories"`
	Schedule     ScheduleConfig     `yaml:"schedule"`
}

// FootballDataConfig 스포츠 API 설정
type FootballDataConfig struct {
	APIKey string `yaml:"api_key"`
}

// ThumbnailConfig 썸네일 설정
type ThumbnailConfig struct {
	Enabled   bool   `yaml:"enabled"`
	OutputDir string `yaml:"output_dir"`
}

// AccountConfig 개별 계정 설정
type AccountConfig struct {
	Name       string            `yaml:"name"`       // 계정 식별자
	Enabled    bool              `yaml:"enabled"`    // 활성화 여부
	Tistory    TistoryConfig     `yaml:"tistory"`    // 티스토리 설정
	Coupang    CoupangConfig     `yaml:"coupang"`    // 쿠팡 파트너스 설정
	Naver      NaverConfig       `yaml:"naver"`      // 네이버 API 설정
	Categories map[string]string `yaml:"categories"` // 카테고리 매핑
	Schedule   ScheduleConfig    `yaml:"schedule"`   // 스케줄 설정
}

// TistoryConfig 티스토리 설정 (브라우저 자동화용)
type TistoryConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
	BlogName string `yaml:"blog_name"`
}

// BrowserConfig 브라우저 설정
type BrowserConfig struct {
	Headless   bool `yaml:"headless"`
	SlowMotion int  `yaml:"slow_motion"`
}

// TMDBConfig TMDB API 설정
type TMDBConfig struct {
	APIKey string `yaml:"api_key"`
}

// NaverConfig 네이버 API 설정
type NaverConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// CoupangConfig 쿠팡파트너스 설정
type CoupangConfig struct {
	PartnerID string `yaml:"partner_id"` // 파트너스 ID (예: AF3262952)
	AccessKey string `yaml:"access_key"` // API용 (선택)
	SecretKey string `yaml:"secret_key"` // API용 (선택)
}

// ScheduleConfig 스케줄 설정
type ScheduleConfig struct {
	Enabled bool          `yaml:"enabled"`
	Jobs    []ScheduleJob `yaml:"jobs"`
}

// ScheduleJob 개별 스케줄 작업
type ScheduleJob struct {
	Category string `yaml:"category"`
	Cron     string `yaml:"cron"`
}

// Load 설정 파일 로드
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 기본값 설정
	if cfg.Browser.Headless == false && cfg.Browser.SlowMotion == 0 {
		cfg.Browser.Headless = true
	}

	// 하위 호환성: accounts가 없으면 기존 설정으로 단일 계정 생성
	if len(cfg.Accounts) == 0 && cfg.Tistory.Email != "" {
		cfg.Accounts = []AccountConfig{
			{
				Name:       cfg.Tistory.BlogName,
				Enabled:    true,
				Tistory:    cfg.Tistory,
				Coupang:    cfg.Coupang,
				Naver:      cfg.Naver,
				Categories: cfg.Categories,
				Schedule:   cfg.Schedule,
			},
		}
	}

	// enabled 기본값 true
	for i := range cfg.Accounts {
		if cfg.Accounts[i].Name != "" && !cfg.Accounts[i].Enabled {
			// enabled가 명시적으로 false가 아니면 true로
			// YAML에서 enabled 필드가 없으면 false가 되므로, 명시적으로 체크
		}
	}

	return &cfg, nil
}

// GetEnabledAccounts 활성화된 계정들만 반환
func (c *Config) GetEnabledAccounts() []AccountConfig {
	var accounts []AccountConfig
	for _, acc := range c.Accounts {
		if acc.Enabled && acc.Tistory.Email != "" {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

// HasCoupang 쿠팡 파트너스 설정이 있는지 확인
func (a *AccountConfig) HasCoupang() bool {
	return a.Coupang.PartnerID != ""
}

// HasNaver 네이버 API 설정이 있는지 확인
func (a *AccountConfig) HasNaver() bool {
	return a.Naver.ClientID != "" && a.Naver.ClientSecret != ""
}

// GetCategoryName 카테고리 이름 반환 (없으면 빈 문자열)
func (a *AccountConfig) GetCategoryName(category string) string {
	if name, ok := a.Categories[category]; ok {
		return name
	}
	return ""
}

// GetCategoryNameOrDefault 카테고리 이름 반환 (없으면 기본값)
func (a *AccountConfig) GetCategoryNameOrDefault(category string) string {
	if name, ok := a.Categories[category]; ok {
		return name
	}
	// 기본 카테고리 (티스토리 기본값)
	return "" // 빈 문자열 = 카테고리 선택 안 함 (기본 카테고리)
}
