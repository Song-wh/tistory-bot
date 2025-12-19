package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Song-wh/tistory-bot/internal/collector"
	"github.com/Song-wh/tistory-bot/internal/config"
	"github.com/Song-wh/tistory-bot/internal/tistory"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tistory-bot",
	Short: "티스토리 자동 포스팅 봇",
	Long: `티스토리에 자동으로 글을 포스팅합니다.

카테고리:
  • 주식/코인 정보
  • 핫딜/할인 정보
  • IT/테크 뉴스
  • 영화/드라마 정보
  • 트렌드/실검

⚠️  브라우저 자동화 방식으로 동작합니다 (API 키 필요 없음)`,
}

// login 명령어 - 로그인 테스트
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "티스토리 로그인 테스트",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		if cfg.Tistory.Email == "" || cfg.Tistory.Password == "" {
			fmt.Println("❌ config.yaml에 email과 password를 설정하세요.")
			os.Exit(1)
		}

		fmt.Println("🔑 로그인 테스트 중...")
		fmt.Println("  (브라우저가 실행됩니다)")

		client := tistory.NewClient(
			cfg.Tistory.Email,
			cfg.Tistory.Password,
			cfg.Tistory.BlogName,
			false, // headless=false로 브라우저 표시
			500,   // 느린 동작으로 확인 가능
		)

		ctx := context.Background()
		if err := client.TestLogin(ctx); err != nil {
			fmt.Printf("❌ 로그인 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ 로그인 성공!")
	},
}

// post 명령어 - 글 작성
var postCmd = &cobra.Command{
	Use:   "post [category]",
	Short: "글 작성",
	Long: `지정한 카테고리의 글을 자동으로 작성합니다.

카테고리:
  crypto  - 코인 시세 정보
  deals   - 핫딜/할인 정보
  tech    - IT/테크 뉴스
  movie   - 영화/드라마 정보
  trend   - 트렌드/실검`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		if cfg.Tistory.Email == "" || cfg.Tistory.Password == "" {
			fmt.Println("❌ config.yaml에 email과 password를 설정하세요.")
			os.Exit(1)
		}

		category := args[0]
		ctx := context.Background()

		client := tistory.NewClient(
			cfg.Tistory.Email,
			cfg.Tistory.Password,
			cfg.Tistory.BlogName,
			cfg.Browser.Headless,
			cfg.Browser.SlowMotion,
		)
		defer client.Close()

		var post *collector.Post

		switch category {
		case "crypto":
			fmt.Println("🪙 코인 시세 수집 중...")
			c := collector.NewStockCollector()
			cryptos, err := c.GetTopCryptos(ctx, 10)
			if err != nil {
				fmt.Printf("수집 실패: %v\n", err)
				os.Exit(1)
			}
			post = c.GenerateCryptoPost(cryptos)

		case "tech":
			fmt.Println("💻 IT/테크 뉴스 수집 중...")
			c := collector.NewTechCollector()
			news, err := c.GetTechNews(ctx, 10)
			if err != nil {
				fmt.Printf("수집 실패: %v\n", err)
				os.Exit(1)
			}
			post = c.GenerateTechPost(news)

		case "movie":
			fmt.Println("🎬 영화 정보 수집 중...")
			c := collector.NewMovieCollector(cfg.TMDB.APIKey)
			movies, err := c.GetNowPlaying(ctx, 10)
			if err != nil {
				fmt.Printf("수집 실패: %v\n", err)
				os.Exit(1)
			}
			post = c.GenerateMoviePost(movies, "now_playing")

		case "trend":
			fmt.Println("🔥 트렌드 수집 중...")
			c := collector.NewTrendCollector()
			trends, err := c.GetGoogleTrends(ctx, 10)
			if err != nil {
				fmt.Printf("수집 실패: %v\n", err)
				os.Exit(1)
			}
			post = c.GenerateTrendPost(trends)

		default:
			fmt.Printf("알 수 없는 카테고리: %s\n", category)
			os.Exit(1)
		}

		// 카테고리 이름 찾기
		categoryName := cfg.Categories[post.Category]
		if categoryName == "" {
			fmt.Printf("⚠️ 카테고리 '%s'가 설정되지 않았습니다.\n", post.Category)
			fmt.Println("config.yaml에서 카테고리 이름을 설정하세요.")
			os.Exit(1)
		}

		fmt.Printf("📝 포스팅: %s\n", post.Title)
		fmt.Println("  (브라우저에서 작업 중...)")

		result, err := client.WritePost(ctx, post.Title, post.Content, categoryName, post.Tags, 3)
		if err != nil {
			fmt.Printf("포스팅 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 포스팅 완료!\n")
		fmt.Printf("URL: %s\n", result.URL)
	},
}

// categories 명령어 - 카테고리 목록
var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "블로그 카테고리 목록 조회",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		if cfg.Tistory.Email == "" || cfg.Tistory.Password == "" {
			fmt.Println("❌ config.yaml에 email과 password를 설정하세요.")
			os.Exit(1)
		}

		fmt.Println("📂 카테고리 조회 중...")
		fmt.Println("  (브라우저에서 작업 중...)")

		client := tistory.NewClient(
			cfg.Tistory.Email,
			cfg.Tistory.Password,
			cfg.Tistory.BlogName,
			cfg.Browser.Headless,
			cfg.Browser.SlowMotion,
		)
		defer client.Close()

		ctx := context.Background()

		categories, err := client.GetCategories(ctx)
		if err != nil {
			fmt.Printf("카테고리 조회 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n📂 블로그 카테고리:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, cat := range categories {
			fmt.Printf("  • %s\n", cat.Name)
		}
		fmt.Println("\nconfig.yaml의 categories에 이름을 그대로 입력하세요.")
	},
}

// run 명령어 - 전체 자동 실행
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "모든 카테고리 자동 포스팅",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🚀 티스토리 자동 포스팅 시작!")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		client := tistory.NewClient(
			cfg.Tistory.Email,
			cfg.Tistory.Password,
			cfg.Tistory.BlogName,
			cfg.Browser.Headless,
			cfg.Browser.SlowMotion,
		)
		defer client.Close()

		ctx := context.Background()

		categories := []string{"crypto", "tech", "movie", "trend"}

		for _, cat := range categories {
			fmt.Printf("\n📝 [%s] 포스팅 중...\n", cat)

			var post *collector.Post
			var err error

			switch cat {
			case "crypto":
				c := collector.NewStockCollector()
				cryptos, e := c.GetTopCryptos(ctx, 10)
				if e != nil {
					fmt.Printf("  ❌ 수집 실패: %v\n", e)
					continue
				}
				post = c.GenerateCryptoPost(cryptos)

			case "tech":
				c := collector.NewTechCollector()
				news, e := c.GetTechNews(ctx, 10)
				if e != nil {
					fmt.Printf("  ❌ 수집 실패: %v\n", e)
					continue
				}
				post = c.GenerateTechPost(news)

			case "movie":
				c := collector.NewMovieCollector(cfg.TMDB.APIKey)
				movies, e := c.GetNowPlaying(ctx, 10)
				if e != nil {
					fmt.Printf("  ❌ 수집 실패: %v\n", e)
					continue
				}
				post = c.GenerateMoviePost(movies, "now_playing")

			case "trend":
				c := collector.NewTrendCollector()
				trends, e := c.GetGoogleTrends(ctx, 10)
				if e != nil {
					fmt.Printf("  ❌ 수집 실패: %v\n", e)
					continue
				}
				post = c.GenerateTrendPost(trends)
			}

			categoryName := cfg.Categories[post.Category]
			if categoryName == "" {
				fmt.Printf("  ⚠️ 카테고리 '%s' 미설정, 건너뜀\n", post.Category)
				continue
			}

			_, err = client.WritePost(ctx, post.Title, post.Content, categoryName, post.Tags, 3)
			if err != nil {
				fmt.Printf("  ❌ 포스팅 실패: %v\n", err)
				continue
			}

			fmt.Printf("  ✅ 완료: %s\n", post.Title)
		}

		fmt.Println("\n✅ 모든 포스팅 완료!")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "설정 파일 경로")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(postCmd)
	rootCmd.AddCommand(categoriesCmd)
	rootCmd.AddCommand(runCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
