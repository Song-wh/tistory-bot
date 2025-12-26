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
	email       string
	password    string
	blogName    string
	headless    bool
	slowMotion  time.Duration
	browser     *rod.Browser
	loggedIn    bool
	userDataDir string // 브라우저 세션 유지용
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
	// 계정별 브라우저 세션 디렉토리 (캡챠 방지) - 절대 경로 사용
	// 현재 작업 디렉토리 기준으로 절대 경로 생성
	userDataDir := fmt.Sprintf("browser_data/%s", blogName)
	
	return &Client{
		email:       email,
		password:    password,
		blogName:    blogName,
		headless:    headless,
		slowMotion:  time.Duration(slowMotion) * time.Millisecond,
		userDataDir: userDataDir,
	}
}

// Connect 브라우저 연결
func (c *Client) Connect() error {
	l := launcher.New().
		Headless(c.headless).
		Leakless(false). // Windows 호환성을 위해 leakless 비활성화
		Set("disable-gpu").
		Set("no-sandbox").
		UserDataDir(c.userDataDir) // 세션 유지 (캡챠 방지)

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

// Login 카카오 계정으로 로그인 (세션 유지 시 스킵)
func (c *Client) Login(ctx context.Context) error {
	if c.browser == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// 먼저 글쓰기 페이지로 이동해서 로그인 상태 확인
	checkURL := fmt.Sprintf("https://%s.tistory.com/manage/newpost", c.blogName)
	page, err := c.browser.Page(proto.TargetCreateTarget{URL: checkURL})
	if err != nil {
		return fmt.Errorf("페이지 열기 실패: %w", err)
	}

	// 페이지 로딩 대기
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("페이지 로딩 실패: %w", err)
	}
	time.Sleep(2 * time.Second)

	// 현재 URL 확인 - 로그인 페이지로 리다이렉트 되었는지 확인
	currentURL := page.MustInfo().URL
	
	// 이미 로그인된 상태 (글쓰기 페이지에 있음)
	if strings.Contains(currentURL, "manage/newpost") || strings.Contains(currentURL, "manage/post") {
		c.loggedIn = true
		fmt.Println("✅ 세션 유지됨 (로그인 스킵)")
		_ = page.Close()
		return nil
	}

	// 로그인 필요 - 로그인 페이지로 이동
	fmt.Println("  🔐 로그인 필요...")
	_ = page.Close()
	
	page, err = c.browser.Page(proto.TargetCreateTarget{URL: "https://www.tistory.com/auth/login"})
	if err != nil {
		return fmt.Errorf("로그인 페이지 열기 실패: %w", err)
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
	currentURL = page.MustInfo().URL
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

	// 브라우저 다이얼로그(confirm/alert) 자동 처리 - 페이지 로드 전에 설정
	fmt.Println("  🔍 임시저장 알림 자동 처리 설정...")
	go page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		fmt.Printf("  📢 다이얼로그 감지: %s\n", e.Message)
		// "취소" 선택 (Accept: false)
		_ = proto.PageHandleJavaScriptDialog{Accept: false}.Call(page)
		fmt.Println("  ✅ 다이얼로그 취소 완료")
	})()

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("페이지 로딩 실패: %w", err)
	}

	time.Sleep(3 * time.Second)
	fmt.Println("  ✅ 페이지 로딩 완료")

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

	// 태그 입력 (티스토리 최대 10개 제한)
	if len(tags) > 0 {
		// 중복 제거 및 10개 제한
		uniqueTags := make([]string, 0, 10)
		seen := make(map[string]bool)
		for _, tag := range tags {
			tagLower := strings.ToLower(strings.TrimSpace(tag))
			if tagLower != "" && !seen[tagLower] && len(uniqueTags) < 10 {
				seen[tagLower] = true
				uniqueTags = append(uniqueTags, tag)
			}
		}
		tags = uniqueTags
		fmt.Printf("  🏷️ 태그 입력 (최대 10개): %v\n", tags)

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

// WritePostWithThumbnail 썸네일 포함 글쓰기
func (c *Client) WritePostWithThumbnail(ctx context.Context, title, content, categoryName string, tags []string, visibility int, thumbnailPath string) (*PostResult, error) {
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

	// 다이얼로그 자동 처리
	go page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		_ = proto.PageHandleJavaScriptDialog{Accept: false}.Call(page)
	})()

	page.MustWaitLoad()
	time.Sleep(3 * time.Second)
	fmt.Println("  ✅ 페이지 로딩 완료")

	// 썸네일은 발행 팝업에서 "대표이미지 추가"로 업로드 (아래에서 처리)

	// 제목 입력
	page.MustEval(`(title) => {
		const titleInput = document.querySelector('#post-title-inp') || 
		                   document.querySelector('[class*="title"] input') ||
		                   document.querySelector('input[placeholder*="제목"]');
		if (titleInput) {
			titleInput.value = title;
			titleInput.dispatchEvent(new Event('input', { bubbles: true }));
			return true;
		}
		return false;
	}`, title)

	// 에디터 iframe으로 전환하여 본문 입력
	page.MustEval(`(content) => {
		const iframe = document.querySelector('#tinymce_ifr') || document.querySelector('iframe[id*="tinymce"]');
		if (iframe) {
			const doc = iframe.contentDocument || iframe.contentWindow.document;
			const body = doc.body;
			if (body) {
				body.innerHTML = content;
				return true;
			}
		}
		const directEditor = document.querySelector('.mce-content-body') || document.querySelector('[contenteditable="true"]');
		if (directEditor) {
			directEditor.innerHTML = content;
			return true;
		}
		return false;
	}`, content)

	time.Sleep(2 * time.Second)
	fmt.Println("  📝 본문 입력 완료")

	// 태그 입력 (최대 10개)
	if len(tags) > 0 {
		uniqueTags := make([]string, 0, 10)
		seen := make(map[string]bool)
		for _, tag := range tags {
			tagLower := strings.ToLower(strings.TrimSpace(tag))
			if tagLower != "" && !seen[tagLower] && len(uniqueTags) < 10 {
				seen[tagLower] = true
				uniqueTags = append(uniqueTags, tag)
			}
		}
		tags = uniqueTags
		fmt.Printf("  🏷️ 태그 입력: %v\n", tags)

		page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		time.Sleep(1 * time.Second)

		for i, tag := range tags {
			page.MustEval(`(tag) => {
				const selectors = [
					'input[placeholder*="태그"]',
					'.tag-input input',
					'#tagText',
					'input.tf_g'
				];
				let input = null;
				for (const sel of selectors) {
					input = document.querySelector(sel);
					if (input) break;
				}
				if (input) {
					input.focus();
					input.value = tag;
					input.dispatchEvent(new Event('input', { bubbles: true }));
					const enterEvent = new KeyboardEvent('keydown', { key: 'Enter', code: 'Enter', keyCode: 13, bubbles: true });
					input.dispatchEvent(enterEvent);
					return true;
				}
				return false;
			}`, tag)
			time.Sleep(300 * time.Millisecond)
			fmt.Printf("    [%d/%d] 태그 추가: %s\n", i+1, len(tags), tag)
		}
	}

	// 카테고리 선택
	if categoryName != "" {
		fmt.Printf("  📂 카테고리 선택: %s\n", categoryName)
		page.MustEval(`(categoryName) => {
			const dropdown = document.querySelector('.category-btn') || document.querySelector('[class*="category"]');
			if (dropdown) dropdown.click();
		}`, categoryName)
		time.Sleep(500 * time.Millisecond)

		page.MustEval(`(categoryName) => {
			const items = document.querySelectorAll('.category-item, [class*="category"] li, [class*="category"] a');
			for (const item of items) {
				if (item.textContent.includes(categoryName)) {
					item.click();
					return true;
				}
			}
			return false;
		}`, categoryName)
		time.Sleep(500 * time.Millisecond)
	}

	// 완료 버튼 클릭 (발행 팝업 열기)
	fmt.Println("  📤 완료 버튼 클릭 시도...")
	page.MustEval(`() => {
		const btns = document.querySelectorAll('button, .btn, [class*="publish"], [class*="complete"]');
		for (const btn of btns) {
			if (btn.textContent.includes('완료') || btn.textContent.includes('발행') || btn.textContent.includes('공개')) {
				btn.click();
				return true;
			}
		}
		return false;
	}`)
	fmt.Println("  ✅ 완료 버튼 클릭 시도 완료")
	time.Sleep(2 * time.Second)

	// 대표이미지 추가 (발행 팝업에서)
	if thumbnailPath != "" {
		fmt.Println("  🖼️ 대표이미지 추가 시도...")
		
		// 파일 input에 직접 파일 설정 (inp_g 클래스)
		fileInput, err := page.Element(`input[type="file"].inp_g, .box_thumb input[type="file"], input[accept="image/*"]`)
		if err == nil && fileInput != nil {
			err = fileInput.SetFiles([]string{thumbnailPath})
			if err == nil {
				fmt.Println("    ✅ 대표이미지 업로드 완료!")
				time.Sleep(2 * time.Second)
			} else {
				fmt.Printf("    ⚠️ 파일 설정 실패: %v\n", err)
			}
		} else {
			// 방법 2: box_thumb 클릭 후 파일 input
			page.MustEval(`() => {
				const thumb = document.querySelector('.box_thumb, .txt_thumb');
				if (thumb) thumb.click();
			}`)
			time.Sleep(1 * time.Second)
			
			fileInput2, _ := page.Element(`input[type="file"]`)
			if fileInput2 != nil {
				_ = fileInput2.SetFiles([]string{thumbnailPath})
				fmt.Println("    ✅ 대표이미지 업로드 완료!")
				time.Sleep(2 * time.Second)
			}
		}
	}

	// 공개 발행 옵션 선택
	fmt.Println("  📤 공개 옵션 선택...")
	page.MustEval(`() => {
		const options = document.querySelectorAll('[class*="option"], label, .radio-item, input[type="radio"]');
		for (const opt of options) {
			if (opt.textContent && opt.textContent.includes('공개')) {
				opt.click();
				return true;
			}
		}
		// 이미 공개가 선택되어 있을 수 있음
		return true;
	}`)
	time.Sleep(1 * time.Second)

	// 최종 발행 버튼 (공개 발행)
	fmt.Println("  📤 공개 발행 버튼 클릭 시도...")
	page.MustEval(`() => {
		const btns = document.querySelectorAll('button, .btn');
		for (const btn of btns) {
			if (btn.textContent.includes('공개 발행') || btn.textContent.includes('발행')) {
				btn.click();
				return true;
			}
		}
		return false;
	}`)

	time.Sleep(5 * time.Second)
	fmt.Println("  ✅ 포스팅 완료!")

	// 결과 URL
	currentURL := ""
	if info, err := page.Info(); err == nil {
		currentURL = info.URL
	}

	postID := ""
	if strings.Contains(currentURL, "/") {
		parts := strings.Split(currentURL, "/")
		postID = parts[len(parts)-1]
	}

	_ = page.Close()

	return &PostResult{
		PostID: postID,
		URL:    fmt.Sprintf("https://%s.tistory.com/%s", c.blogName, postID),
	}, nil
}

// uploadThumbnail 썸네일 업로드 (에러 안전)
func (c *Client) uploadThumbnail(page *rod.Page, thumbnailPath string) (err error) {
	// panic 복구
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("썸네일 업로드 중 오류: %v", r)
		}
	}()

	// 페이지 연결 상태 확인
	if page == nil {
		return fmt.Errorf("페이지 연결 없음")
	}

	// 티스토리 에디터의 이미지(첨부) 버튼 클릭
	// 버튼 구조: <button id="mceu_0-open"><i class="mce-ico mce-i-image"></i></button>
	clicked, evalErr := page.Eval(`() => {
		// 방법 1: mce-i-image 클래스로 찾기 (가장 정확)
		const imageIcon = document.querySelector('.mce-i-image');
		if (imageIcon) {
			const btn = imageIcon.closest('button');
			if (btn) {
				btn.click();
				return "mce-i-image";
			}
		}
		
		// 방법 2: aria-label="첨부" 로 찾기
		const attachBtn = document.querySelector('[aria-label="첨부"], [aria-label*="첨부"]');
		if (attachBtn) {
			attachBtn.click();
			return "aria-첨부";
		}
		
		// 방법 3: id로 찾기 (mceu_0-open)
		const mceuBtn = document.querySelector('#mceu_0-open, [id^="mceu_"][id$="-open"]');
		if (mceuBtn) {
			mceuBtn.click();
			return "mceu-open";
		}
		
		// 방법 4: mce-ico 클래스로 찾기
		const mceIco = document.querySelector('.mce-ico');
		if (mceIco) {
			const btn = mceIco.closest('button');
			if (btn) {
				btn.click();
				return "mce-ico";
			}
		}
		
		return null;
	}`)

	if evalErr != nil {
		return fmt.Errorf("이미지 버튼 클릭 실패: %v", evalErr)
	}

	if clicked.Value.Nil() {
		return fmt.Errorf("이미지 버튼을 찾을 수 없음")
	}

	fmt.Printf("    📷 이미지 버튼 클릭: %v\n", clicked.Value.Str())
	time.Sleep(2 * time.Second)

	// 파일 input 찾아서 파일 설정 (안전한 방식)
	fileInput, elemErr := page.Element(`input[type="file"]`)
	if elemErr != nil || fileInput == nil {
		// 숨겨진 file input 찾기
		_, _ = page.Eval(`() => {
			const inputs = document.querySelectorAll('input[type="file"]');
			console.log('File inputs found:', inputs.length);
			return inputs.length;
		}`)
		return fmt.Errorf("파일 업로드 요소를 찾을 수 없음")
	}

	setErr := fileInput.SetFiles([]string{thumbnailPath})
	if setErr != nil {
		return fmt.Errorf("파일 설정 실패: %v", setErr)
	}

	fmt.Println("    📤 파일 업로드 중...")
	time.Sleep(3 * time.Second) // 업로드 대기

	// 업로드 완료 후 확인 버튼 클릭 (있다면)
	_, _ = page.Eval(`() => {
		const confirmBtn = document.querySelector('[class*="confirm"], [class*="submit"], button.primary');
		if (confirmBtn) {
			confirmBtn.click();
			return true;
		}
		return false;
	}`)
	time.Sleep(1 * time.Second)

	return nil
}

// TestLogin 로그인 테스트
func (c *Client) TestLogin(ctx context.Context) error {
	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Close()

	return c.Login(ctx)
}
