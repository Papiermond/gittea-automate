package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	token  string
	client *http.Client
}

type CreateRepoRequest struct {
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	AutoInit bool   `json:"auto_init"`
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		client: &http.Client{},
	}
}

func (c *Client) RepoExists(username, repo string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", username, repo)
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
	url := "https://api.github.com/user/repos"
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
