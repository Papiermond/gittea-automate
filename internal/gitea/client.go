package gitea

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

type CreateRepoRequest struct {
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	AutoInit bool   `json:"auto_init"`
}

type PushMirrorRequest struct {
	RemoteAddress  string `json:"remote_address"`
	RemotePassword string `json:"remote_password"`
	RemoteUsername string `json:"remote_username"`
	SyncOnCommit   bool   `json:"sync_on_commit"`
	Interval       string `json:"interval"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}
}

func (c *Client) RepoExists(username, repo string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s", c.baseURL, username, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "token "+c.token)

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
	url := fmt.Sprintf("%s/api/v1/user/repos", c.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "token "+c.token)
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

func (c *Client) AddPushMirror(username, repo string, req PushMirrorRequest) error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/push_mirrors", c.baseURL, username, repo)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "token "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		// Mirror might already exist, check response
		if bytes.Contains(respBody, []byte("already exists")) {
			return nil // Not an error
		}
		return fmt.Errorf("failed to add mirror (status %d): %s", resp.StatusCode, respBody)
	}

	return nil
}
