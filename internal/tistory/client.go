package tistory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client 티스토리 API 클라이언트
type Client struct {
	accessToken string
	blogName    string
	client      *http.Client
}

// Category 카테고리 정보
type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Parent   string `json:"parent"`
	Label    string `json:"label"`
	Entries  string `json:"entries"`
}

// PostResult 포스팅 결과
type PostResult struct {
	PostID string `json:"postId"`
	URL    string `json:"url"`
}

// NewClient 새 클라이언트 생성
func NewClient(accessToken, blogName string) *Client {
	return &Client{
		accessToken: accessToken,
		blogName:    blogName,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// GetCategories 카테고리 목록 가져오기
func (c *Client) GetCategories(ctx context.Context) ([]Category, error) {
	apiURL := fmt.Sprintf(
		"https://www.tistory.com/apis/category/list?access_token=%s&blogName=%s&output=json",
		c.accessToken, c.blogName,
	)

	resp, err := c.client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Tistory struct {
			Status string `json:"status"`
			Item   struct {
				Categories []Category `json:"category"`
			} `json:"item"`
		} `json:"tistory"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Tistory.Status != "200" {
		return nil, fmt.Errorf("API 에러: %s", result.Tistory.Status)
	}

	return result.Tistory.Item.Categories, nil
}

// WritePost 글 작성
func (c *Client) WritePost(ctx context.Context, title, content, categoryID string, tags []string, visibility int) (*PostResult, error) {
	apiURL := "https://www.tistory.com/apis/post/write"

	data := url.Values{}
	data.Set("access_token", c.accessToken)
	data.Set("blogName", c.blogName)
	data.Set("title", title)
	data.Set("content", content)
	data.Set("category", categoryID)
	data.Set("tag", strings.Join(tags, ","))
	data.Set("visibility", fmt.Sprintf("%d", visibility)) // 0: 비공개, 1: 보호, 3: 공개
	data.Set("output", "json")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Tistory struct {
			Status string `json:"status"`
			PostID string `json:"postId"`
			URL    string `json:"url"`
			Error  string `json:"error_message"`
		} `json:"tistory"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %s", string(body))
	}

	if result.Tistory.Status != "200" {
		return nil, fmt.Errorf("API 에러: %s - %s", result.Tistory.Status, result.Tistory.Error)
	}

	return &PostResult{
		PostID: result.Tistory.PostID,
		URL:    result.Tistory.URL,
	}, nil
}

// ModifyPost 글 수정
func (c *Client) ModifyPost(ctx context.Context, postID, title, content, categoryID string, tags []string) error {
	apiURL := "https://www.tistory.com/apis/post/modify"

	data := url.Values{}
	data.Set("access_token", c.accessToken)
	data.Set("blogName", c.blogName)
	data.Set("postId", postID)
	data.Set("title", title)
	data.Set("content", content)
	data.Set("category", categoryID)
	data.Set("tag", strings.Join(tags, ","))
	data.Set("output", "json")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Tistory struct {
			Status string `json:"status"`
		} `json:"tistory"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Tistory.Status != "200" {
		return fmt.Errorf("API 에러: %s", result.Tistory.Status)
	}

	return nil
}

// GetAuthURL 인증 URL 생성
func GetAuthURL(clientID, redirectURI string) string {
	return fmt.Sprintf(
		"https://www.tistory.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code",
		clientID, url.QueryEscape(redirectURI),
	)
}

// GetAccessToken 액세스 토큰 발급
func GetAccessToken(clientID, clientSecret, redirectURI, code string) (string, error) {
	apiURL := fmt.Sprintf(
		"https://www.tistory.com/oauth/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s&grant_type=authorization_code",
		clientID, clientSecret, url.QueryEscape(redirectURI), code,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	// access_token=xxx 형식으로 반환됨
	parts := strings.Split(string(body), "=")
	if len(parts) == 2 && parts[0] == "access_token" {
		return parts[1], nil
	}

	return "", fmt.Errorf("토큰 발급 실패: %s", string(body))
}

