package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	url    string
	token  string
	client *http.Client
}

type CreateRepoRequest struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility"` // "private" or "public"
}

func NewClient(baseURL, token string) *Client {
	// Default to gitlab.com if no URL provided
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	return &Client{
		url:    baseURL,
		token:  token,
		client: &http.Client{},
	}
}

func (c *Client) RepoExists(username, repo string) (bool, error) {
	// GitLab uses namespace/project format
	projectPath := url.PathEscape(fmt.Sprintf("%s/%s", username, repo))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s", c.url, projectPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	}
	if resp.StatusCode == 200 {
		return true, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
}

func (c *Client) CreateRepo(req CreateRepoRequest) error {
	apiURL := fmt.Sprintf("%s/api/v4/projects", c.url)

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("PRIVATE-TOKEN", c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create repo (status %d): %s", resp.StatusCode, respBody)
	}

	return nil
}
