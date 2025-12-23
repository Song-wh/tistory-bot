package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Song-wh/tistory-bot/internal/collector"
	"github.com/Song-wh/tistory-bot/internal/config"
	"github.com/Song-wh/tistory-bot/internal/tistory"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var cfgFile string
var accountName string // 특정 계정만 실행할 때 사용

var rootCmd = &cobra.Command{
	Use:   "tistory-bot",
	Short: "티스토리 자동 포스팅 봇 (다중 계정 지원)",
	Long: `티스토리에 자동으로 글을 포스팅합니다.

✨ 다중 계정 지원
  --account [name] 옵션으로 특정 계정만 실행 가능
  생략하면 모든 활성화된 계정에 포스팅

카테고리:
  • 주식/코인 정보
  • 핫딜/할인 정보  
  • IT/테크 뉴스
  • 영화/드라마 정보
  • 트렌드/실검
  • 쿠팡 특가 💰

⚠️  브라우저 자동화 방식으로 동작합니다`,
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

		accounts := getTargetAccounts(cfg)
		if len(accounts) == 0 {
			fmt.Println("❌ 활성화된 계정이 없습니다.")
			os.Exit(1)
		}

		for _, acc := range accounts {
			fmt.Printf("\n🔑 [%s] 로그인 테스트 중...\n", acc.Name)

			client := tistory.NewClient(
				acc.Tistory.Email,
				acc.Tistory.Password,
				acc.Tistory.BlogName,
				false,
				500,
			)

			ctx := context.Background()
			if err := client.TestLogin(ctx); err != nil {
				fmt.Printf("❌ [%s] 로그인 실패: %v\n", acc.Name, err)
				continue
			}

			fmt.Printf("✅ [%s] 로그인 성공!\n", acc.Name)
		}
	},
}

// post 명령어 - 글 작성
var postCmd = &cobra.Command{
	Use:   "post [category]",
	Short: "글 작성 (모든 계정 또는 특정 계정)",
	Long: `지정한 카테고리의 글을 자동으로 작성합니다.

--account [name] 옵션으로 특정 계정만 포스팅 가능

카테고리:
  crypto       - 코인 시세 정보
  deals        - 핫딜/할인 정보
  tech         - IT/테크 뉴스
  movie        - 영화/드라마 정보
  trend        - 트렌드/실검
  lotto        - 로또 당첨번호
  lotto-predict - 로또 예측번호 (AI 분석)
  fortune      - 오늘의 운세
  sports       - 스포츠 뉴스
  coupang      - 쿠팡 특가/파트너스 💰
  golf         - 내일 골프 날씨 예보 ⛳
  golf-tips    - 골프 레슨 팁 + 용품 추천 🏌️
  error        - 에러/장애 해결 아카이브 🔴`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		accounts := getTargetAccounts(cfg)
		if len(accounts) == 0 {
			fmt.Println("❌ 활성화된 계정이 없습니다.")
			os.Exit(1)
		}

		category := args[0]
		ctx := context.Background()

		fmt.Printf("📝 카테고리: %s | 대상 계정: %d개\n", category, len(accounts))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for _, acc := range accounts {
			fmt.Printf("\n🔄 [%s] 포스팅 시작...\n", acc.Name)

			// 쿠팡 카테고리인데 쿠팡 설정이 없으면 건너뛰기
			if category == "coupang" && !acc.HasCoupang() {
				fmt.Printf("  ⏭️ [%s] 쿠팡 파트너스 설정 없음, 건너뜀\n", acc.Name)
				continue
			}

			// 카테고리 매핑 확인
			post := generatePost(ctx, cfg, &acc, category)
			if post == nil {
				continue
			}

			categoryName := acc.GetCategoryName(post.Category)
			if categoryName == "" {
				fmt.Printf("  ℹ️ [%s] 카테고리 '%s' 미설정, 기본 카테고리 사용\n", acc.Name, post.Category)
				// 빈 문자열 = 카테고리 선택 안 함 (기본 카테고리에 게시)
			}

			// 티스토리 클라이언트 생성
			client := tistory.NewClient(
				acc.Tistory.Email,
				acc.Tistory.Password,
				acc.Tistory.BlogName,
				cfg.Browser.Headless,
				cfg.Browser.SlowMotion,
			)
			defer client.Close()

			fmt.Printf("  📝 제목: %s\n", post.Title)

			result, err := client.WritePost(ctx, post.Title, post.Content, categoryName, post.Tags, 3)
			if err != nil {
				fmt.Printf("  ❌ [%s] 포스팅 실패: %v\n", acc.Name, err)
				continue
			}

			fmt.Printf("  ✅ [%s] 포스팅 완료! URL: %s\n", acc.Name, result.URL)
		}

		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✅ 모든 계정 포스팅 완료!")
	},
}

// accounts 명령어 - 계정 목록
var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "등록된 계정 목록 조회",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("📋 등록된 계정 목록")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for i, acc := range cfg.Accounts {
			status := "🔴 비활성"
			if acc.Enabled {
				status = "🟢 활성"
			}

			fmt.Printf("\n%d. %s %s\n", i+1, acc.Name, status)
			fmt.Printf("   📧 티스토리: %s (%s.tistory.com)\n", acc.Tistory.Email, acc.Tistory.BlogName)

			if acc.HasCoupang() {
				fmt.Printf("   🛒 쿠팡: %s\n", acc.Coupang.PartnerID)
			} else {
				fmt.Printf("   🛒 쿠팡: ❌ 미설정\n")
			}

			if acc.HasNaver() {
				fmt.Printf("   🌐 네이버: ✅ 설정됨\n")
			}

			fmt.Printf("   📂 카테고리: %d개\n", len(acc.Categories))

			if acc.Schedule.Enabled {
				fmt.Printf("   ⏰ 스케줄: %d개 작업\n", len(acc.Schedule.Jobs))
			}
		}
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

		accounts := getTargetAccounts(cfg)
		if len(accounts) == 0 {
			fmt.Println("❌ 활성화된 계정이 없습니다.")
			os.Exit(1)
		}

		for _, acc := range accounts {
			fmt.Printf("\n📂 [%s] 카테고리 조회 중...\n", acc.Name)

			client := tistory.NewClient(
				acc.Tistory.Email,
				acc.Tistory.Password,
				acc.Tistory.BlogName,
				cfg.Browser.Headless,
				cfg.Browser.SlowMotion,
			)
			defer client.Close()

			ctx := context.Background()
			categories, err := client.GetCategories(ctx)
			if err != nil {
				fmt.Printf("❌ [%s] 카테고리 조회 실패: %v\n", acc.Name, err)
				continue
			}

			fmt.Printf("\n📂 [%s] 블로그 카테고리:\n", acc.Name)
			for _, cat := range categories {
				fmt.Printf("  • %s\n", cat.Name)
			}
		}
	},
}

// run 명령어 - 전체 자동 실행
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "모든 카테고리 자동 포스팅 (모든 계정)",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		accounts := getTargetAccounts(cfg)
		if len(accounts) == 0 {
			fmt.Println("❌ 활성화된 계정이 없습니다.")
			os.Exit(1)
		}

		fmt.Println("🚀 티스토리 자동 포스팅 시작!")
		fmt.Printf("📋 대상 계정: %d개\n", len(accounts))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		ctx := context.Background()
		categories := []string{"crypto", "tech", "movie", "trend", "lotto", "lotto-predict", "weather", "fortune", "sports", "coupang"}

		for _, acc := range accounts {
			fmt.Printf("\n\n📌 [%s] 포스팅 시작\n", acc.Name)
			fmt.Println("────────────────────────────")

			client := tistory.NewClient(
				acc.Tistory.Email,
				acc.Tistory.Password,
				acc.Tistory.BlogName,
				cfg.Browser.Headless,
				cfg.Browser.SlowMotion,
			)

			for _, cat := range categories {
				fmt.Printf("\n  📝 [%s] 카테고리...\n", cat)

				// 쿠팡인데 설정 없으면 건너뛰기
				if cat == "coupang" && !acc.HasCoupang() {
					fmt.Printf("    ⏭️ 쿠팡 설정 없음, 건너뜀\n")
					continue
				}

				post := generatePost(ctx, cfg, &acc, cat)
				if post == nil {
					continue
				}

				categoryName := acc.GetCategoryName(post.Category)
				if categoryName == "" {
					fmt.Printf("    ℹ️ 카테고리 '%s' 미설정, 기본 카테고리 사용\n", post.Category)
				}

				_, err := client.WritePost(ctx, post.Title, post.Content, categoryName, post.Tags, 3)
				if err != nil {
					fmt.Printf("    ❌ 포스팅 실패: %v\n", err)
					continue
				}

				fmt.Printf("    ✅ 완료: %s\n", post.Title)
			}

			client.Close()
		}

		fmt.Println("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✅ 모든 계정 포스팅 완료!")
	},
}

// schedule 명령어 - 자동 스케줄 실행
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "자동 스케줄러 실행 (모든 계정)",
	Long: `설정된 스케줄에 따라 자동으로 포스팅합니다.
모든 활성화된 계정에 대해 각각의 스케줄을 실행합니다.
프로그램을 종료하려면 Ctrl+C를 누르세요.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("설정 로드 실패: %v\n", err)
			os.Exit(1)
		}

		accounts := cfg.GetEnabledAccounts()
		if len(accounts) == 0 {
			fmt.Println("❌ 활성화된 계정이 없습니다.")
			os.Exit(1)
		}

		fmt.Println("🚀 티스토리 자동 포스팅 스케줄러 시작!")
		fmt.Printf("📋 활성 계정: %d개\n", len(accounts))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		c := cron.New()

		for _, acc := range accounts {
			if !acc.Schedule.Enabled || len(acc.Schedule.Jobs) == 0 {
				fmt.Printf("\n⏭️ [%s] 스케줄 비활성화\n", acc.Name)
				continue
			}

			fmt.Printf("\n📅 [%s] 스케줄 등록:\n", acc.Name)

			for _, job := range acc.Schedule.Jobs {
				category := job.Category
				cronExpr := job.Cron
				accCopy := acc // 클로저용 복사

				fmt.Printf("  • %s: %s\n", category, cronExpr)

				c.AddFunc(cronExpr, func() {
					// 랜덤 딜레이 (0~45분) - 자동화 티 안 나게
					rand.Seed(time.Now().UnixNano())
					delay := time.Duration(rand.Intn(45)) * time.Minute
					fmt.Printf("\n⏰ [%s] 스케줄 트리거: %s (%.0f분 후 실행)\n", accCopy.Name, category, delay.Minutes())
					time.Sleep(delay)
					fmt.Printf("▶️ [%s] 포스팅 시작: %s\n", accCopy.Name, category)
					runPostForAccount(cfg, &accCopy, category)
				})
			}
		}

		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⏳ 스케줄 대기 중... (종료: Ctrl+C)")

		c.Start()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n🛑 스케줄러 종료...")
		c.Stop()
	},
}

// getTargetAccounts 대상 계정 목록 반환
func getTargetAccounts(cfg *config.Config) []config.AccountConfig {
	accounts := cfg.GetEnabledAccounts()

	// --account 옵션이 있으면 해당 계정만 반환
	if accountName != "" {
		for _, acc := range accounts {
			if acc.Name == accountName {
				return []config.AccountConfig{acc}
			}
		}
		fmt.Printf("⚠️ 계정 '%s'를 찾을 수 없습니다.\n", accountName)
		return nil
	}

	return accounts
}

// generatePost 카테고리에 맞는 포스트 생성
func generatePost(ctx context.Context, cfg *config.Config, acc *config.AccountConfig, category string) *collector.Post {
	var post *collector.Post

	switch category {
	case "crypto":
		c := collector.NewStockCollector()
		cryptos, err := c.GetTopCryptos(ctx, 10)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateCryptoPost(cryptos)

	case "tech":
		c := collector.NewTechCollector()
		news, err := c.GetTechNews(ctx, 10)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateTechPost(news)

	case "movie":
		c := collector.NewMovieCollector(cfg.TMDB.APIKey, acc.Coupang.PartnerID)
		movies, err := c.GetNowPlaying(ctx, 10)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateMoviePost(movies, "now_playing")

	case "trend":
		c := collector.NewTrendCollector()
		trends, err := c.GetGoogleTrends(ctx, 10)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateTrendPost(trends)

	case "lotto":
		c := collector.NewLottoCollector()
		result, err := c.GetLatestLotto(ctx)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateLottoPost(result)

	case "lotto-predict":
		c := collector.NewLottoCollector()
		results, err := c.GetRecentResults(ctx, 20)
		if err != nil {
			fmt.Printf("    ❌ 분석 실패: %v\n", err)
			return nil
		}
		hotNumbers, coldNumbers := c.AnalyzeNumbers(results)
		predictions := c.GeneratePredictions(hotNumbers, coldNumbers, acc.Name)
		nextRound := results[0].DrawNo + 1
		post = c.GeneratePredictionPost(nextRound, predictions, hotNumbers, coldNumbers)

	case "weather":
		c := collector.NewWeatherCollector()
		weathers, err := c.GetWeather(ctx)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateWeatherPost(weathers)

	case "fortune":
		c := collector.NewFortuneCollector(acc.Coupang.PartnerID)
		fortunes := c.GetTodayFortune()
		post = c.GenerateFortunePost(fortunes)

	case "sports":
		c := collector.NewSportsCollector(acc.Coupang.PartnerID)
		news, err := c.GetSportsNews(ctx)
		if err != nil {
			fmt.Printf("    ❌ 수집 실패: %v\n", err)
			return nil
		}
		post = c.GenerateSportsPost(news)

	case "coupang":
		if !acc.HasCoupang() {
			fmt.Printf("    ⏭️ 쿠팡 설정 없음, 건너뜀\n")
			return nil
		}
		c := collector.NewCoupangCollector(acc.Coupang.PartnerID)
		products, err := c.GetGoldboxProducts(ctx, 10)
		if err != nil {
			fmt.Printf("    ❌ 크롤링 실패: %v\n", err)
			return nil
		}
		if len(products) == 0 {
			fmt.Printf("    ❌ 상품 없음, 건너뜀\n")
			return nil
		}
		post = c.GenerateCoupangPost(products)

	case "golf":
		coupangID := ""
		if acc.HasCoupang() {
			coupangID = acc.Coupang.PartnerID
		}
		c := collector.NewGolfCollector(coupangID)
		post = c.GenerateGolfPost(ctx)

	case "golf-tips":
		coupangID := ""
		if acc.HasCoupang() {
			coupangID = acc.Coupang.PartnerID
		}
		c := collector.NewGolfTipsCollector(coupangID)
		post = c.GenerateGolfTipsPost(ctx)

	case "error":
		c := collector.NewErrorArchiveCollector()
		post = c.GenerateErrorPost(ctx)

	default:
		fmt.Printf("    ❌ 알 수 없는 카테고리: %s\n", category)
		return nil
	}

	return post
}

// runPostForAccount 특정 계정에 포스팅
func runPostForAccount(cfg *config.Config, acc *config.AccountConfig, category string) {
	ctx := context.Background()

	// 쿠팡인데 설정 없으면 건너뛰기
	if category == "coupang" && !acc.HasCoupang() {
		fmt.Printf("  ⏭️ [%s] 쿠팡 설정 없음, 건너뜀\n", acc.Name)
		return
	}

	post := generatePost(ctx, cfg, acc, category)
	if post == nil {
		return
	}

	categoryName := acc.GetCategoryName(post.Category)
	if categoryName == "" {
		fmt.Printf("  ℹ️ [%s] 카테고리 '%s' 미설정, 기본 카테고리 사용\n", acc.Name, post.Category)
	}

	client := tistory.NewClient(
		acc.Tistory.Email,
		acc.Tistory.Password,
		acc.Tistory.BlogName,
		cfg.Browser.Headless,
		cfg.Browser.SlowMotion,
	)
	defer client.Close()

	_, err := client.WritePost(ctx, post.Title, post.Content, categoryName, post.Tags, 3)
	if err != nil {
		fmt.Printf("  ❌ [%s] 포스팅 실패: %v\n", acc.Name, err)
		return
	}

	fmt.Printf("  ✅ [%s] 포스팅 완료: %s\n", acc.Name, post.Title)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "설정 파일 경로")
	rootCmd.PersistentFlags().StringVar(&accountName, "account", "", "특정 계정만 실행 (생략시 전체)")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(postCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(categoriesCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(scheduleCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
