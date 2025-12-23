package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ErrorArchiveCollector 에러/장애 해결 아카이브 수집기
type ErrorArchiveCollector struct {
	client *http.Client
}

// ErrorEntry 에러 정보
type ErrorEntry struct {
	Title       string   `json:"title"`        // 에러 제목
	ErrorMsg    string   `json:"error_msg"`    // 에러 메시지
	Language    string   `json:"language"`     // 프로그래밍 언어
	Tags        []string `json:"tags"`         // 태그
	Cause       string   `json:"cause"`        // 원인
	Solution    string   `json:"solution"`     // 해결책
	CodeExample string   `json:"code_example"` // 코드 예시
	Source      string   `json:"source"`       // 출처 (SO/GitHub)
	SourceURL   string   `json:"source_url"`   // 원본 URL
	Views       int      `json:"views"`        // 조회수
	Score       int      `json:"score"`        // 점수/스타
}

// NewErrorArchiveCollector 생성자
func NewErrorArchiveCollector() *ErrorArchiveCollector {
	return &ErrorArchiveCollector{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetStackOverflowErrors Stack Overflow에서 인기 에러 수집
func (c *ErrorArchiveCollector) GetStackOverflowErrors(ctx context.Context, tag string, limit int) ([]ErrorEntry, error) {
	// Stack Overflow API - 인기 질문
	url := fmt.Sprintf(
		"https://api.stackexchange.com/2.3/questions?order=desc&sort=votes&tagged=%s;error&site=stackoverflow&pagesize=%d&filter=withbody",
		tag, limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// API 실패 시 시뮬레이션 데이터
		return c.getSimulatedSOErrors(tag, limit), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.getSimulatedSOErrors(tag, limit), nil
	}

	var result struct {
		Items []struct {
			Title       string   `json:"title"`
			Body        string   `json:"body"`
			Tags        []string `json:"tags"`
			Score       int      `json:"score"`
			ViewCount   int      `json:"view_count"`
			Link        string   `json:"link"`
			IsAnswered  bool     `json:"is_answered"`
			AnswerCount int      `json:"answer_count"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.getSimulatedSOErrors(tag, limit), nil
	}

	var errors []ErrorEntry
	for _, item := range result.Items {
		if !item.IsAnswered || item.AnswerCount == 0 {
			continue
		}

		errorMsg := c.extractErrorMessage(item.Title, item.Body)
		lang := c.detectLanguage(item.Tags)

		errors = append(errors, ErrorEntry{
			Title:     item.Title,
			ErrorMsg:  errorMsg,
			Language:  lang,
			Tags:      item.Tags,
			Source:    "Stack Overflow",
			SourceURL: item.Link,
			Views:     item.ViewCount,
			Score:     item.Score,
		})
	}

	return errors, nil
}

// GetGitHubIssues GitHub에서 인기 이슈 수집
func (c *ErrorArchiveCollector) GetGitHubIssues(ctx context.Context, language string, limit int) ([]ErrorEntry, error) {
	// GitHub Search API - 에러 관련 이슈
	url := fmt.Sprintf(
		"https://api.github.com/search/issues?q=label:bug+language:%s+state:closed&sort=reactions&order=desc&per_page=%d",
		language, limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.getSimulatedGitHubErrors(language, limit), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.getSimulatedGitHubErrors(language, limit), nil
	}

	var result struct {
		Items []struct {
			Title     string `json:"title"`
			Body      string `json:"body"`
			HTMLURL   string `json:"html_url"`
			Reactions struct {
				TotalCount int `json:"total_count"`
			} `json:"reactions"`
			Labels []struct {
				Name string `json:"name"`
			} `json:"labels"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.getSimulatedGitHubErrors(language, limit), nil
	}

	var errors []ErrorEntry
	for _, item := range result.Items {
		errorMsg := c.extractErrorMessage(item.Title, item.Body)

		var tags []string
		for _, label := range item.Labels {
			tags = append(tags, label.Name)
		}

		errors = append(errors, ErrorEntry{
			Title:     item.Title,
			ErrorMsg:  errorMsg,
			Language:  language,
			Tags:      tags,
			Source:    "GitHub",
			SourceURL: item.HTMLURL,
			Score:     item.Reactions.TotalCount,
		})
	}

	return errors, nil
}

// extractErrorMessage 에러 메시지 추출
func (c *ErrorArchiveCollector) extractErrorMessage(title, body string) string {
	// 일반적인 에러 패턴
	patterns := []string{
		`(?i)(error|exception|panic|fatal|failed):\s*(.+)`,
		`(?i)cannot\s+(.+)`,
		`(?i)unable\s+to\s+(.+)`,
		`(?i)undefined\s+(.+)`,
		`(?i)null\s+pointer`,
		`(?i)type\s+error`,
		`(?i)syntax\s+error`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(body); len(matches) > 0 {
			return strings.TrimSpace(matches[0])
		}
		if matches := re.FindStringSubmatch(title); len(matches) > 0 {
			return strings.TrimSpace(matches[0])
		}
	}

	// 패턴 못 찾으면 제목 사용
	return title
}

// detectLanguage 언어 감지
func (c *ErrorArchiveCollector) detectLanguage(tags []string) string {
	langMap := map[string]string{
		"javascript": "JavaScript",
		"js":         "JavaScript",
		"typescript": "TypeScript",
		"ts":         "TypeScript",
		"python":     "Python",
		"java":       "Java",
		"go":         "Go",
		"golang":     "Go",
		"rust":       "Rust",
		"c++":        "C++",
		"cpp":        "C++",
		"c#":         "C#",
		"csharp":     "C#",
		"php":        "PHP",
		"ruby":       "Ruby",
		"swift":      "Swift",
		"kotlin":     "Kotlin",
		"react":      "React",
		"vue":        "Vue.js",
		"angular":    "Angular",
		"node.js":    "Node.js",
		"nodejs":     "Node.js",
	}

	for _, tag := range tags {
		if lang, ok := langMap[strings.ToLower(tag)]; ok {
			return lang
		}
	}

	return "General"
}

// getSimulatedSOErrors Stack Overflow 시뮬레이션 데이터
func (c *ErrorArchiveCollector) getSimulatedSOErrors(tag string, limit int) []ErrorEntry {
	allErrors := []ErrorEntry{
		// JavaScript
		{
			Title:    "Cannot read properties of null (reading 'toLowerCase')",
			ErrorMsg: "TypeError: Cannot read properties of null (reading 'toLowerCase')",
			Language: "JavaScript",
			Tags:     []string{"javascript", "null", "typeerror"},
			Cause:    "변수가 null인 상태에서 메서드를 호출할 때 발생합니다. 주로 DOM 요소가 없거나 API 응답이 null일 때 발생합니다.",
			Solution: "Optional Chaining(?.)을 사용하거나 null 체크를 먼저 수행하세요.",
			CodeExample: `// ❌ 에러 발생
const text = data.name.toLowerCase();

// ✅ 해결 방법 1: Optional Chaining
const text = data?.name?.toLowerCase();

// ✅ 해결 방법 2: Null 체크
const text = data && data.name ? data.name.toLowerCase() : '';`,
			Source: "Stack Overflow",
			Views:  150000,
			Score:  320,
		},
		{
			Title:    "Uncaught ReferenceError: X is not defined",
			ErrorMsg: "ReferenceError: X is not defined",
			Language: "JavaScript",
			Tags:     []string{"javascript", "reference-error"},
			Cause:    "변수나 함수가 선언되지 않은 상태에서 사용하려고 할 때 발생합니다. 스코프 문제나 오타가 원인일 수 있습니다.",
			Solution: "변수 선언 여부 확인, import 누락 확인, 스코프 확인이 필요합니다.",
			CodeExample: `// ❌ 에러 발생
console.log(myVariable); // 선언 안 됨

// ✅ 해결 방법
let myVariable = "Hello";
console.log(myVariable);`,
			Source: "Stack Overflow",
			Views:  200000,
			Score:  450,
		},
		{
			Title:    "SyntaxError: Unexpected token < in JSON",
			ErrorMsg: "SyntaxError: Unexpected token '<' in JSON at position 0",
			Language: "JavaScript",
			Tags:     []string{"javascript", "json", "api"},
			Cause:    "JSON 파싱 시 HTML이 반환될 때 발생합니다. 주로 API 엔드포인트가 잘못되었거나 서버가 HTML 에러 페이지를 반환할 때 발생합니다.",
			Solution: "API URL 확인, Content-Type 헤더 확인, 서버 응답 로깅이 필요합니다.",
			CodeExample: `// ❌ 에러 발생 원인
// 서버가 HTML을 반환: "<html>..."
const data = JSON.parse(response);

// ✅ 해결 방법
fetch('/api/data')
  .then(res => {
    if (!res.ok) throw new Error('API 오류');
    return res.json();
  })
  .catch(err => console.error('파싱 실패:', err));`,
			Source: "Stack Overflow",
			Views:  180000,
			Score:  380,
		},
		// React
		{
			Title:    "React: Each child in a list should have a unique key prop",
			ErrorMsg: "Warning: Each child in a list should have a unique \"key\" prop",
			Language: "React",
			Tags:     []string{"react", "javascript", "key-prop"},
			Cause:    "리스트 렌더링 시 각 요소에 고유한 key가 없을 때 발생합니다. React가 효율적으로 DOM을 업데이트하기 위해 필요합니다.",
			Solution: "map() 사용 시 각 요소에 고유한 key prop을 추가하세요. index보다 고유 ID 사용을 권장합니다.",
			CodeExample: `// ❌ 에러 발생
{items.map(item => <li>{item.name}</li>)}

// ⚠️ index 사용 (비권장)
{items.map((item, index) => <li key={index}>{item.name}</li>)}

// ✅ 고유 ID 사용 (권장)
{items.map(item => <li key={item.id}>{item.name}</li>)}`,
			Source: "Stack Overflow",
			Views:  250000,
			Score:  520,
		},
		{
			Title:    "React Hooks: Too many re-renders",
			ErrorMsg: "Error: Too many re-renders. React limits the number of renders to prevent an infinite loop.",
			Language: "React",
			Tags:     []string{"react", "hooks", "infinite-loop"},
			Cause:    "컴포넌트 내에서 setState를 조건 없이 호출하거나, useEffect 의존성 배열이 잘못 설정된 경우 발생합니다.",
			Solution: "이벤트 핸들러 내에서 setState 호출, useEffect 의존성 배열 확인이 필요합니다.",
			CodeExample: `// ❌ 에러 발생 - 렌더링마다 setState 호출
function Component() {
  const [count, setCount] = useState(0);
  setCount(count + 1); // 무한 루프!
  return <div>{count}</div>;
}

// ✅ 해결 방법 - 이벤트 핸들러 사용
function Component() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(c => c + 1)}>{count}</button>;
}`,
			Source: "Stack Overflow",
			Views:  180000,
			Score:  420,
		},
		// Go
		{
			Title:    "panic: runtime error: invalid memory address or nil pointer dereference",
			ErrorMsg: "panic: runtime error: invalid memory address or nil pointer dereference",
			Language: "Go",
			Tags:     []string{"go", "panic", "nil-pointer"},
			Cause:    "nil 포인터를 역참조하려고 할 때 발생합니다. 초기화되지 않은 포인터나 nil을 반환하는 함수 결과를 사용할 때 발생합니다.",
			Solution: "nil 체크 후 사용, 포인터 초기화 확인, 에러 반환값 확인이 필요합니다.",
			CodeExample: `// ❌ 에러 발생
var user *User // nil 상태
fmt.Println(user.Name) // panic!

// ✅ 해결 방법 1: nil 체크
if user != nil {
    fmt.Println(user.Name)
}

// ✅ 해결 방법 2: 초기화
user := &User{Name: "John"}
fmt.Println(user.Name)`,
			Source: "Stack Overflow",
			Views:  120000,
			Score:  280,
		},
		{
			Title:    "Go: cannot use X (type Y) as type Z in argument",
			ErrorMsg: "cannot use X (type Y) as type Z in argument to function",
			Language: "Go",
			Tags:     []string{"go", "type-error"},
			Cause:    "함수 인자 타입이 맞지 않을 때 발생합니다. Go는 암시적 타입 변환을 지원하지 않습니다.",
			Solution: "명시적 타입 변환을 수행하거나 인터페이스를 사용하세요.",
			CodeExample: `// ❌ 에러 발생
var num int = 10
var num64 int64 = num // 타입 불일치!

// ✅ 해결 방법: 명시적 변환
var num int = 10
var num64 int64 = int64(num)`,
			Source: "Stack Overflow",
			Views:  95000,
			Score:  210,
		},
		// Python
		{
			Title:    "Python: IndentationError: unexpected indent",
			ErrorMsg: "IndentationError: unexpected indent",
			Language: "Python",
			Tags:     []string{"python", "indentation"},
			Cause:    "들여쓰기가 잘못되었을 때 발생합니다. 탭과 스페이스 혼용, 불필요한 들여쓰기가 원인입니다.",
			Solution: "일관된 들여쓰기 사용 (스페이스 4칸 권장), IDE 설정 확인이 필요합니다.",
			CodeExample: `# ❌ 에러 발생
def hello():
    print("Hello")
        print("World")  # 불필요한 들여쓰기!

# ✅ 해결 방법
def hello():
    print("Hello")
    print("World")`,
			Source: "Stack Overflow",
			Views:  300000,
			Score:  580,
		},
		{
			Title:    "Python: ModuleNotFoundError: No module named 'X'",
			ErrorMsg: "ModuleNotFoundError: No module named 'X'",
			Language: "Python",
			Tags:     []string{"python", "import", "module"},
			Cause:    "모듈이 설치되지 않았거나 가상환경이 활성화되지 않았을 때 발생합니다.",
			Solution: "pip install로 모듈 설치, 가상환경 활성화 확인이 필요합니다.",
			CodeExample: `# ❌ 에러 발생
import pandas  # 설치 안 됨

# ✅ 해결 방법
# 터미널에서:
pip install pandas

# 가상환경 사용 시:
python -m venv venv
source venv/bin/activate  # Linux/Mac
venv\Scripts\activate     # Windows
pip install pandas`,
			Source: "Stack Overflow",
			Views:  400000,
			Score:  650,
		},
		{
			Title:    "Python: TypeError: 'NoneType' object is not subscriptable",
			ErrorMsg: "TypeError: 'NoneType' object is not subscriptable",
			Language: "Python",
			Tags:     []string{"python", "none", "typeerror"},
			Cause:    "None 값에 인덱싱([])을 시도할 때 발생합니다. 함수가 None을 반환하는 경우 발생합니다.",
			Solution: "None 체크 후 인덱싱, 함수 반환값 확인이 필요합니다.",
			CodeExample: `# ❌ 에러 발생
result = some_function()  # None 반환
print(result[0])  # TypeError!

# ✅ 해결 방법
result = some_function()
if result is not None:
    print(result[0])
else:
    print("결과 없음")`,
			Source: "Stack Overflow",
			Views:  180000,
			Score:  320,
		},
		// TypeScript
		{
			Title:    "TypeScript: Object is possibly 'undefined'",
			ErrorMsg: "TS2532: Object is possibly 'undefined'",
			Language: "TypeScript",
			Tags:     []string{"typescript", "undefined"},
			Cause:    "Optional 프로퍼티나 nullable 타입을 안전하게 처리하지 않았을 때 발생합니다.",
			Solution: "Optional Chaining, Non-null assertion, 또는 타입 가드를 사용하세요.",
			CodeExample: `// ❌ 에러 발생
interface User {
  name?: string;
}
const user: User = {};
console.log(user.name.toUpperCase()); // Object is possibly 'undefined'

// ✅ 해결 방법 1: Optional Chaining
console.log(user.name?.toUpperCase());

// ✅ 해결 방법 2: Non-null assertion (확실할 때만!)
console.log(user.name!.toUpperCase());

// ✅ 해결 방법 3: 타입 가드
if (user.name) {
  console.log(user.name.toUpperCase());
}`,
			Source: "Stack Overflow",
			Views:  220000,
			Score:  480,
		},
		// Node.js
		{
			Title:    "Node.js: EADDRINUSE: address already in use",
			ErrorMsg: "Error: listen EADDRINUSE: address already in use :::3000",
			Language: "Node.js",
			Tags:     []string{"nodejs", "port", "server"},
			Cause:    "다른 프로세스가 이미 해당 포트를 사용 중일 때 발생합니다.",
			Solution: "기존 프로세스 종료, 다른 포트 사용, 또는 포트 확인 후 kill이 필요합니다.",
			CodeExample: `// 포트 사용 프로세스 확인 및 종료
// Windows:
netstat -ano | findstr :3000
taskkill /PID <PID> /F

// Linux/Mac:
lsof -i :3000
kill -9 <PID>

// 코드에서 다른 포트 사용
const PORT = process.env.PORT || 3001;
app.listen(PORT, () => console.log('Server running on port ' + PORT));`,
			Source: "Stack Overflow",
			Views:  280000,
			Score:  520,
		},
	}

	// 태그 필터링
	var filtered []ErrorEntry
	for _, e := range allErrors {
		for _, t := range e.Tags {
			if strings.Contains(strings.ToLower(t), strings.ToLower(tag)) {
				filtered = append(filtered, e)
				break
			}
		}
	}

	if len(filtered) == 0 {
		filtered = allErrors
	}

	// 셔플
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// getSimulatedGitHubErrors GitHub 시뮬레이션 데이터
func (c *ErrorArchiveCollector) getSimulatedGitHubErrors(language string, limit int) []ErrorEntry {
	// 시뮬레이션 데이터 - SO와 동일한 구조 재활용
	return c.getSimulatedSOErrors(language, limit)
}

// GenerateErrorPost 에러 해결 포스트 생성
func (c *ErrorArchiveCollector) GenerateErrorPost(ctx context.Context) *Post {
	now := time.Now()

	// 언어 목록 (순환)
	languages := []string{"javascript", "python", "go", "typescript", "react"}
	lang := languages[now.Day()%len(languages)]

	// 에러 수집
	errors := c.getSimulatedSOErrors(lang, 3)
	if len(errors) == 0 {
		return nil
	}

	// 메인 에러 선택
	mainError := errors[0]

	// SEO 최적화 제목
	title := fmt.Sprintf("[%s] %s - 원인과 해결방법 완벽 정리",
		mainError.Language, c.truncateTitle(mainError.ErrorMsg, 50))

	// 본문 생성
	var content strings.Builder

	// 스타일
	content.WriteString(`
<style>
.error-container { max-width: 900px; margin: 0 auto; font-family: 'Fira Code', 'Consolas', monospace; }
.error-header { background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); color: #fff; padding: 40px; border-radius: 16px; margin-bottom: 30px; }
.error-header h1 { margin: 0 0 15px 0; font-size: 24px; color: #e94560; }
.error-header .lang-badge { display: inline-block; background: #e94560; padding: 4px 12px; border-radius: 4px; font-size: 12px; margin-bottom: 15px; }
.error-msg { background: #0f0f23; padding: 20px; border-radius: 8px; font-family: monospace; color: #ff6b6b; font-size: 14px; overflow-x: auto; }
.section { background: #fff; border: 1px solid #e5e5e5; border-radius: 12px; padding: 25px; margin-bottom: 20px; }
.section h2 { margin: 0 0 15px 0; color: #1a1a2e; font-size: 18px; display: flex; align-items: center; gap: 10px; }
.section h2::before { content: ''; width: 4px; height: 20px; background: #e94560; border-radius: 2px; }
.cause-box { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px 20px; border-radius: 0 8px 8px 0; }
.solution-box { background: #d4edda; border-left: 4px solid #28a745; padding: 15px 20px; border-radius: 0 8px 8px 0; }
.code-block { background: #1e1e1e; color: #d4d4d4; padding: 20px; border-radius: 8px; overflow-x: auto; font-size: 13px; line-height: 1.6; }
.code-block .comment { color: #6a9955; }
.code-block .error { color: #f14c4c; }
.code-block .success { color: #4ec9b0; }
.more-errors { background: #f8f9fa; padding: 25px; border-radius: 12px; margin-top: 30px; }
.more-errors h3 { margin: 0 0 15px 0; }
.error-item { padding: 15px; background: #fff; border-radius: 8px; margin-bottom: 10px; border-left: 3px solid #e94560; }
.error-item .title { font-weight: 600; color: #333; margin-bottom: 5px; }
.error-item .meta { font-size: 12px; color: #666; }
.tags { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 15px; }
.tag { font-size: 11px; padding: 3px 10px; background: #e9ecef; color: #495057; border-radius: 4px; }
.footer-note { margin-top: 30px; padding: 20px; background: #f5f5f5; border-radius: 12px; font-size: 13px; color: #666; }
.source-link { color: #e94560; text-decoration: none; }
</style>
`)

	content.WriteString(`<div class="error-container">`)

	// 헤더
	content.WriteString(fmt.Sprintf(`
<div class="error-header">
	<span class="lang-badge">%s</span>
	<h1>🔴 %s</h1>
	<div class="error-msg">%s</div>
</div>
`, mainError.Language, mainError.Title, mainError.ErrorMsg))

	// 원인 섹션
	content.WriteString(fmt.Sprintf(`
<div class="section">
	<h2>❓ 왜 이 에러가 발생하나요?</h2>
	<div class="cause-box">
		<p>%s</p>
	</div>
</div>
`, mainError.Cause))

	// 해결책 섹션
	content.WriteString(fmt.Sprintf(`
<div class="section">
	<h2>✅ 해결 방법</h2>
	<div class="solution-box">
		<p>%s</p>
	</div>
</div>
`, mainError.Solution))

	// 코드 예시
	content.WriteString(fmt.Sprintf(`
<div class="section">
	<h2>💻 코드 예시</h2>
	<pre class="code-block">%s</pre>
</div>
`, c.formatCodeBlock(mainError.CodeExample)))

	// 관련 에러 더보기
	if len(errors) > 1 {
		content.WriteString(`
<div class="more-errors">
	<h3>📚 관련 에러 더보기</h3>
`)
		for _, e := range errors[1:] {
			content.WriteString(fmt.Sprintf(`
	<div class="error-item">
		<div class="title">%s</div>
		<div class="meta">🏷️ %s | 👀 조회수 %s</div>
	</div>
`, e.Title, e.Language, formatViews(e.Views)))
		}
		content.WriteString(`</div>`)
	}

	// 태그
	content.WriteString(`<div class="tags">`)
	for _, tag := range mainError.Tags {
		content.WriteString(fmt.Sprintf(`<span class="tag">#%s</span>`, tag))
	}
	content.WriteString(`</div>`)

	// 푸터
	content.WriteString(fmt.Sprintf(`
<div class="footer-note">
	<p>📅 작성일: %s</p>
	<p>💡 이 글이 도움이 되셨다면 공유해주세요!</p>
	<p>🔍 더 많은 에러 해결법은 블로그를 구독해주세요.</p>
</div>
`, now.Format("2006년 01월 02일")))

	content.WriteString(`</div>`)

	// 공격적인 태그 전략
	langLower := strings.ToLower(mainError.Language)
	tags := []string{
		// 기본 태그
		mainError.Language + "에러", mainError.Language + "해결", mainError.Language + "오류",
		"프로그래밍에러", "개발에러해결", "코딩에러",
		// 언어별 태그
		langLower + "error", langLower + "버그", langLower + "디버깅",
		// 인기 키워드
		"에러해결", "오류해결", "버그수정", "디버깅", "트러블슈팅",
		"개발자팁", "코딩팁", "프로그래밍팁",
		// 플랫폼 태그
		"StackOverflow", "GitHub", "개발블로그",
		// 검색 키워드
		"에러메시지", "에러코드", "에러원인", "에러해결방법",
	}
	tags = append(tags, mainError.Tags...)

	return &Post{
		Title:    title,
		Content:  content.String(),
		Category: "에러/해결",
		Tags:     tags,
	}
}

// truncateTitle 제목 자르기
func (c *ErrorArchiveCollector) truncateTitle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// formatCodeBlock 코드 블록 포맷팅
func (c *ErrorArchiveCollector) formatCodeBlock(code string) string {
	// 주석 강조
	code = strings.ReplaceAll(code, "// ❌", `<span class="error">// ❌</span>`)
	code = strings.ReplaceAll(code, "// ✅", `<span class="success">// ✅</span>`)
	code = strings.ReplaceAll(code, "// ⚠️", `<span class="comment">// ⚠️</span>`)
	code = strings.ReplaceAll(code, "# ❌", `<span class="error"># ❌</span>`)
	code = strings.ReplaceAll(code, "# ✅", `<span class="success"># ✅</span>`)
	return code
}

// formatViews 조회수 포맷팅
func formatViews(views int) string {
	if views >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(views)/1000000)
	}
	if views >= 1000 {
		return fmt.Sprintf("%.1fK", float64(views)/1000)
	}
	return fmt.Sprintf("%d", views)
}
