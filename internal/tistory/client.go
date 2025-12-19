package tistory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Client 티스토리 브라우저 자동화 클라이언트
type Client struct {
	email      string
	password   string
	blogName   string
	headless   bool
	slowMotion time.Duration
	browser    *rod.Browser
	loggedIn   bool
}

// Category 카테고리 정보
type Category struct {
	ID   string
	Name string
}

// PostResult 포스팅 결과
type PostResult struct {
	PostID string
	URL    string
}

// NewClient 새 클라이언트 생성
func NewClient(email, password, blogName string, headless bool, slowMotion int) *Client {
	return &Client{
		email:      email,
		password:   password,
		blogName:   blogName,
		headless:   headless,
		slowMotion: time.Duration(slowMotion) * time.Millisecond,
	}
}

// Connect 브라우저 연결
func (c *Client) Connect() error {
	l := launcher.New().
		Headless(c.headless).
		Leakless(false). // Windows 호환성을 위해 leakless 비활성화
		Set("disable-gpu").
		Set("no-sandbox")

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("브라우저 실행 실패: %w", err)
	}

	c.browser = rod.New().ControlURL(url)
	if c.slowMotion > 0 {
		c.browser = c.browser.SlowMotion(c.slowMotion)
	}

	if err := c.browser.Connect(); err != nil {
		return fmt.Errorf("브라우저 연결 실패: %w", err)
	}

	return nil
}

// Close 브라우저 종료
func (c *Client) Close() {
	if c.browser != nil {
		c.browser.MustClose()
	}
}

// Login 카카오 계정으로 로그인
func (c *Client) Login(ctx context.Context) error {
	if c.browser == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	page, err := c.browser.Page(proto.TargetCreateTarget{URL: "https://www.tistory.com/auth/login"})
	if err != nil {
		return fmt.Errorf("페이지 열기 실패: %w", err)
	}

	// 페이지 로딩 대기
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("페이지 로딩 실패: %w", err)
	}

	// 카카오 로그인 버튼 클릭
	kakaoBtn, err := page.Timeout(10 * time.Second).Element("a.link_kakao_id")
	if err != nil {
		return fmt.Errorf("카카오 로그인 버튼을 찾을 수 없습니다: %w", err)
	}
	if err := kakaoBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("카카오 로그인 버튼 클릭 실패: %w", err)
	}

	// 카카오 로그인 페이지 대기
	time.Sleep(2 * time.Second)

	// 이메일 입력
	emailInput, err := page.Timeout(10 * time.Second).Element("input[name='loginId']")
	if err != nil {
		return fmt.Errorf("이메일 입력란을 찾을 수 없습니다: %w", err)
	}
	if err := emailInput.Input(c.email); err != nil {
		return fmt.Errorf("이메일 입력 실패: %w", err)
	}

	// 비밀번호 입력
	pwdInput, err := page.Element("input[name='password']")
	if err != nil {
		return fmt.Errorf("비밀번호 입력란을 찾을 수 없습니다: %w", err)
	}
	if err := pwdInput.Input(c.password); err != nil {
		return fmt.Errorf("비밀번호 입력 실패: %w", err)
	}

	// 로그인 버튼 클릭
	loginBtn, err := page.Element("button[type='submit']")
	if err != nil {
		return fmt.Errorf("로그인 버튼을 찾을 수 없습니다: %w", err)
	}
	if err := loginBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("로그인 버튼 클릭 실패: %w", err)
	}

	// 로그인 완료 대기
	time.Sleep(3 * time.Second)

	// 로그인 성공 확인 (티스토리 메인 페이지로 리다이렉트)
	currentURL := page.MustInfo().URL
	if strings.Contains(currentURL, "tistory.com") && !strings.Contains(currentURL, "auth/login") {
		c.loggedIn = true
		fmt.Println("✅ 로그인 성공!")
		return nil
	}

	return fmt.Errorf("로그인 실패: 현재 URL = %s", currentURL)
}

// GetCategories 카테고리 목록 가져오기
func (c *Client) GetCategories(ctx context.Context) ([]Category, error) {
	if !c.loggedIn {
		if err := c.Login(ctx); err != nil {
			return nil, err
		}
	}

	// 글쓰기 페이지로 이동
	editorURL := fmt.Sprintf("https://%s.tistory.com/manage/newpost", c.blogName)
	page, err := c.browser.Page(proto.TargetCreateTarget{URL: editorURL})
	if err != nil {
		return nil, fmt.Errorf("에디터 페이지 열기 실패: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("페이지 로딩 실패: %w", err)
	}

	time.Sleep(2 * time.Second)

	// 카테고리 선택 영역 찾기
	categorySelect, err := page.Timeout(10 * time.Second).Element("#category")
	if err != nil {
		return nil, fmt.Errorf("카테고리 선택 영역을 찾을 수 없습니다: %w", err)
	}

	// 모든 옵션 가져오기
	options, err := categorySelect.Elements("option")
	if err != nil {
		return nil, fmt.Errorf("카테고리 옵션을 찾을 수 없습니다: %w", err)
	}

	var categories []Category
	for _, opt := range options {
		value, _ := opt.Attribute("value")
		text, _ := opt.Text()

		if value != nil && *value != "" {
			categories = append(categories, Category{
				ID:   *value,
				Name: strings.TrimSpace(text),
			})
		}
	}

	page.MustClose()
	return categories, nil
}

// WritePost 글 작성
func (c *Client) WritePost(ctx context.Context, title, content, categoryName string, tags []string, visibility int) (*PostResult, error) {
	if !c.loggedIn {
		if err := c.Login(ctx); err != nil {
			return nil, err
		}
	}

	// 글쓰기 페이지로 이동
	editorURL := fmt.Sprintf("https://%s.tistory.com/manage/newpost", c.blogName)
	page, err := c.browser.Page(proto.TargetCreateTarget{URL: editorURL})
	if err != nil {
		return nil, fmt.Errorf("에디터 페이지 열기 실패: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("페이지 로딩 실패: %w", err)
	}

	time.Sleep(3 * time.Second)

	// 임시저장 알림창 처리 (있으면 닫기)
	page.MustEval(`() => {
		// 모든 버튼에서 "사용 안함", "취소", "닫기" 텍스트 찾기
		const buttons = document.querySelectorAll('button');
		for (const btn of buttons) {
			const text = (btn.textContent || '').trim();
			if (text.includes('사용 안함') || text.includes('사용안함') || 
			    text === '취소' || text === '닫기' || text === '아니오') {
				btn.click();
				console.log('Alert dismissed:', text);
				return true;
			}
		}
		return false;
	}`)

	time.Sleep(1 * time.Second)

	// 제목 입력
	titleInput, err := page.Timeout(10 * time.Second).Element("#post-title-inp")
	if err != nil {
		return nil, fmt.Errorf("제목 입력란을 찾을 수 없습니다: %w", err)
	}
	if err := titleInput.Input(title); err != nil {
		return nil, fmt.Errorf("제목 입력 실패: %w", err)
	}

	// 본문 입력 (TinyMCE 에디터)
	time.Sleep(2 * time.Second)
	fmt.Println("  📝 본문 입력 중...")

	// TinyMCE에 직접 내용 삽입 + 저장 트리거
	page.MustEval(`(content) => {
		// TinyMCE 에디터에 접근
		if (typeof tinymce !== 'undefined' && tinymce.activeEditor) {
			const editor = tinymce.activeEditor;
			editor.setContent(content);
			// 변경 이벤트 트리거
			editor.fire('change');
			editor.fire('input');
			// 저장 (내부 상태 업데이트)
			editor.save();
			console.log('TinyMCE content set successfully');
			return true;
		}
		// iframe 방식
		const iframe = document.querySelector('iframe');
		if (iframe && iframe.contentDocument) {
			const body = iframe.contentDocument.body;
			if (body) {
				body.innerHTML = content;
				// 변경 이벤트 트리거
				const event = new Event('input', { bubbles: true });
				body.dispatchEvent(event);
				console.log('iframe content set successfully');
				return true;
			}
		}
		return false;
	}`, content)

	time.Sleep(3 * time.Second)
	fmt.Println("  📝 본문 입력 완료")

	// 태그 입력
	if len(tags) > 0 {
		fmt.Printf("  🏷️ 태그 입력: %v\n", tags)

		// 페이지 하단으로 스크롤
		page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		time.Sleep(1 * time.Second)

		// JavaScript로 태그 입력 처리
		for i, tag := range tags {
			result := page.MustEval(`(tag) => {
				// 태그 입력란 찾기
				const selectors = [
					'input[placeholder*="태그"]',
					'input[placeholder*="Tag"]',
					'input[placeholder*="tag"]',
					'.tag-input input',
					'#tagText',
					'input.tf_g'
				];
				
				let input = null;
				for (const sel of selectors) {
					input = document.querySelector(sel);
					if (input) break;
				}
				
				if (!input) {
					// 모든 input 순회
					const inputs = document.querySelectorAll('input[type="text"], input:not([type])');
					for (const inp of inputs) {
						if (inp.placeholder && (inp.placeholder.includes('태그') || inp.placeholder.includes('tag'))) {
							input = inp;
							break;
						}
					}
				}
				
				if (!input) {
					return { success: false, error: '태그 입력란을 찾을 수 없음' };
				}
				
				// 스크롤 및 포커스
				input.scrollIntoView({ behavior: 'smooth', block: 'center' });
				input.focus();
				input.click();
				
				// 값 설정
				const nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
				nativeInputValueSetter.call(input, tag);
				
				// React/Vue 등 프레임워크용 이벤트
				input.dispatchEvent(new Event('input', { bubbles: true, cancelable: true }));
				input.dispatchEvent(new Event('change', { bubbles: true, cancelable: true }));
				
				// Enter 키 이벤트 시퀀스
				const enterOptions = {
					key: 'Enter',
					code: 'Enter', 
					keyCode: 13,
					which: 13,
					charCode: 13,
					bubbles: true,
					cancelable: true
				};
				
				input.dispatchEvent(new KeyboardEvent('keydown', enterOptions));
				input.dispatchEvent(new KeyboardEvent('keypress', enterOptions));
				input.dispatchEvent(new KeyboardEvent('keyup', enterOptions));
				
				// 폼 submit 이벤트도 시도
				const form = input.closest('form');
				if (form) {
					// submit 방지하면서 이벤트만 발생
					const submitEvent = new Event('submit', { bubbles: true, cancelable: true });
					form.dispatchEvent(submitEvent);
				}
				
				return { success: true, placeholder: input.placeholder };
			}`, tag)

			resultMap := result.Map()
			if resultMap["success"].Bool() {
				fmt.Printf("    [%d/%d] 태그 추가: %s\n", i+1, len(tags), tag)
			} else {
				errMsg := resultMap["error"].String()
				fmt.Printf("    ⚠️ 태그 실패: %s - %s\n", tag, errMsg)
			}
			time.Sleep(800 * time.Millisecond)
		}

		fmt.Println("    태그 입력 완료")
		time.Sleep(1 * time.Second)
	}

	// 카테고리 선택 (제목 위의 드롭다운)
	if categoryName != "" {
		fmt.Printf("  📂 카테고리 선택: %s\n", categoryName)

		// 1. 카테고리 드롭다운 클릭
		clicked := page.MustEval(`() => {
			const elements = document.querySelectorAll('*');
			for (const el of elements) {
				const text = (el.textContent || '').trim();
				if (text === '카테고리' && el.closest('button, [role="button"], .dropdown')) {
					el.click();
					return true;
				}
			}
			// 드롭다운 버튼 직접 찾기
			const dropdown = document.querySelector('[class*="category"]');
			if (dropdown) {
				dropdown.click();
				return true;
			}
			return false;
		}`).Bool()

		if clicked {
			fmt.Println("    드롭다운 클릭됨")
		} else {
			fmt.Println("    ⚠️ 드롭다운을 찾을 수 없음")
		}

		time.Sleep(1 * time.Second)

		// 2. 카테고리 옵션 선택
		selected := page.MustEval(`(name) => {
			const options = document.querySelectorAll('li, [role="option"], [role="menuitem"], .category-item');
			for (const opt of options) {
				const text = (opt.textContent || '').trim();
				if (text === name || text.includes(name)) {
					opt.click();
					return true;
				}
			}
			return false;
		}`, categoryName).Bool()

		if selected {
			fmt.Printf("    카테고리 '%s' 선택됨\n", categoryName)
		} else {
			fmt.Printf("    ⚠️ 카테고리 '%s'를 찾을 수 없음\n", categoryName)
		}

		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Second)

	// 완료 버튼 클릭 (키보드 단축키 사용)
	fmt.Println("  📤 완료 버튼 클릭 시도...")

	// 방법: JavaScript로 직접 버튼 클릭
	page.MustEval(`() => {
		// 완료 버튼 찾기 (여러 방법 시도)
		let btn = document.querySelector('button.btn-publish');
		if (!btn) {
			btn = document.querySelector('.btn_submit');
		}
		if (!btn) {
			// 모든 버튼에서 찾기
			const buttons = document.querySelectorAll('button');
			for (const b of buttons) {
				if (b.textContent.trim() === '완료' || b.innerText.trim() === '완료') {
					btn = b;
					break;
				}
			}
		}
		if (btn) {
			btn.click();
			console.log('완료 버튼 클릭됨');
			return true;
		}
		console.log('완료 버튼을 찾을 수 없음');
		return false;
	}`)
	fmt.Println("  ✅ 완료 버튼 클릭 시도 완료")

	// 발행 다이얼로그 대기
	time.Sleep(3 * time.Second)

	// "공개" 옵션 선택
	fmt.Println("  📤 공개 옵션 선택...")
	page.MustEval(`() => {
		// 공개 라디오 버튼 찾아서 클릭
		const labels = document.querySelectorAll('label');
		for (const label of labels) {
			if (label.textContent.trim() === '공개') {
				label.click();
				return true;
			}
		}
		// input radio로 시도
		const radios = document.querySelectorAll('input[type="radio"]');
		for (const radio of radios) {
			const label = radio.nextElementSibling || radio.parentElement;
			if (label && label.textContent && label.textContent.includes('공개') && !label.textContent.includes('비공개')) {
				radio.click();
				return true;
			}
		}
		return false;
	}`)

	time.Sleep(1 * time.Second)

	// "공개 발행" 버튼 클릭
	fmt.Println("  📤 공개 발행 버튼 클릭 시도...")
	page.MustEval(`() => {
		const buttons = document.querySelectorAll('button');
		for (const b of buttons) {
			const text = b.textContent || b.innerText || '';
			// "공개 발행" 또는 "발행" 또는 "저장" 버튼 클릭
			if (text.includes('공개 발행') || text.includes('발행') || (text.includes('저장') && !text.includes('임시'))) {
				b.click();
				console.log('발행 버튼 클릭됨:', text);
				return true;
			}
		}
		return false;
	}`)
	fmt.Println("  ✅ 발행 버튼 클릭 완료")

	// 발행 완료 대기
	time.Sleep(5 * time.Second)

	// 발행 완료 후 URL 가져오기
	time.Sleep(2 * time.Second)

	currentURL := ""
	if info, err := page.Info(); err == nil {
		currentURL = info.URL
	}

	// 포스트 ID 추출 시도
	postID := ""
	if strings.Contains(currentURL, "/") {
		parts := strings.Split(currentURL, "/")
		postID = parts[len(parts)-1]
	}

	// 페이지 닫기 (에러 무시)
	_ = page.Close()

	return &PostResult{
		PostID: postID,
		URL:    fmt.Sprintf("https://%s.tistory.com/%s", c.blogName, postID),
	}, nil
}

// TestLogin 로그인 테스트
func (c *Client) TestLogin(ctx context.Context) error {
	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Close()

	return c.Login(ctx)
}
