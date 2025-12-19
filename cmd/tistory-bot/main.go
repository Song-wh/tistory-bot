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
  • 트렌드/실검`,
}

// auth 명령어 - 티스토리 인증
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "티스토리 API 인증",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		authURL := tistory.GetAuthURL(cfg.Tistory.ClientID, cfg.Tistory.RedirectURI)
		fmt.Println("🔑 브라우저에서 다음 URL을 열어 인증하세요:")
		fmt.Println(authURL)
		fmt.Println("\n인증 후 리다이렉트된 URL의 code 파라미터를 복사하세요.")
	},
}

// token 명령어 - 토큰 발급
var tokenCmd = &cobra.Command{
	Use:   "token [code]",
	Short: "액세스 토큰 발급",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		code := args[0]
		token, err := tistory.GetAccessToken(
			cfg.Tistory.ClientID,
			cfg.Tistory.ClientSecret,
			cfg.Tistory.RedirectURI,
			code,
		)
		if err != nil {
			fmt.Printf("토큰 발급 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ 액세스 토큰 발급 성공!")
		fmt.Printf("토큰: %s\n", token)
		fmt.Println("\n이 토큰을 config.yaml의 access_token에 저장하세요.")
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

		category := args[0]
		ctx := context.Background()

		client := tistory.NewClient(cfg.Tistory.AccessToken, cfg.Tistory.BlogName)

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

		// 카테고리 ID 찾기
		categoryID := cfg.Categories[post.Category]
		if categoryID == "" {
			fmt.Printf("⚠️ 카테고리 '%s'의 ID가 설정되지 않았습니다.\n", post.Category)
			fmt.Println("config.yaml에서 카테고리 ID를 설정하세요.")
			os.Exit(1)
		}

		fmt.Printf("📝 포스팅: %s\n", post.Title)

		result, err := client.WritePost(ctx, post.Title, post.Content, categoryID, post.Tags, 3)
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

		client := tistory.NewClient(cfg.Tistory.AccessToken, cfg.Tistory.BlogName)
		ctx := context.Background()

		categories, err := client.GetCategories(ctx)
		if err != nil {
			fmt.Printf("카테고리 조회 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("📂 블로그 카테고리:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, cat := range categories {
			fmt.Printf("  [%s] %s (글 %s개)\n", cat.ID, cat.Name, cat.Entries)
		}
		fmt.Println("\nconfig.yaml의 categories에 ID를 설정하세요.")
	},
}

// run 명령어 - 전체 자동 실행
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "모든 카테고리 자동 포스팅",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 티스토리 자동 포스팅 시작!")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		categories := []string{"crypto", "tech", "movie", "trend"}

		for _, cat := range categories {
			fmt.Printf("\n📝 [%s] 포스팅 중...\n", cat)
			// 각 카테고리별 포스팅 로직 실행
		}

		fmt.Println("\n✅ 모든 포스팅 완료!")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "설정 파일 경로")
	
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(tokenCmd)
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

