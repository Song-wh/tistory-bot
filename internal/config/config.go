package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 설정
type Config struct {
	Tistory    TistoryConfig     `yaml:"tistory"`
	Browser    BrowserConfig     `yaml:"browser"`
	TMDB       TMDBConfig        `yaml:"tmdb"`
	Naver      NaverConfig       `yaml:"naver"`
	Coupang    CoupangConfig     `yaml:"coupang"`
	Categories map[string]string `yaml:"categories"`
	Schedule   ScheduleConfig    `yaml:"schedule"`
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
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
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

	return &cfg, nil
}

