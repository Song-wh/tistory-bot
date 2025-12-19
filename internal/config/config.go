package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 설정
type Config struct {
	Tistory    TistoryConfig         `yaml:"tistory"`
	TMDB       TMDBConfig            `yaml:"tmdb"`
	Naver      NaverConfig           `yaml:"naver"`
	Coupang    CoupangConfig         `yaml:"coupang"`
	Categories map[string]string     `yaml:"categories"`
	Schedule   ScheduleConfig        `yaml:"schedule"`
}

// TistoryConfig 티스토리 설정
type TistoryConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
	AccessToken  string `yaml:"access_token"`
	BlogName     string `yaml:"blog_name"`
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
	Enabled  bool     `yaml:"enabled"`
	Cron     string   `yaml:"cron"`
	Categories []string `yaml:"categories"`
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

	return &cfg, nil
}

