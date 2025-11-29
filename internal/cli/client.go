package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gomonov/otus-go-project/internal/domain"
)

type CreateSubnetRequest struct {
	CIDR string `json:"cidr"`
}

type DeleteSubnetRequest struct {
	CIDR string `json:"cidr"`
}

type SubnetsListResponse struct {
	Subnets []SubnetResponse `json:"subnets"`
	Count   int              `json:"count"`
}

type SubnetResponse struct {
	ListType domain.ListType `json:"listType"`
	CIDR     string          `json:"cidr"`
}

type ResetBucketsRequest struct {
	Login string `json:"login,omitempty"`
	IP    string `json:"ip,omitempty"`
}

type ResetBucketsResponse struct {
	Reset bool `json:"reset"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) makeRequest(method, path string, body interface{}) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, errorResp.Error)
		}
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) AddToBlacklist(cidr string) error {
	req := CreateSubnetRequest{CIDR: cidr}
	_, err := c.makeRequest("POST", "/blacklist", req)
	return err
}

func (c *Client) RemoveFromBlacklist(cidr string) error {
	req := DeleteSubnetRequest{CIDR: cidr}
	_, err := c.makeRequest("DELETE", "/blacklist", req)
	return err
}

func (c *Client) GetBlacklist() (*SubnetsListResponse, error) {
	respBody, err := c.makeRequest("GET", "/blacklist", nil)
	if err != nil {
		return nil, err
	}

	var response SubnetsListResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *Client) AddToWhitelist(cidr string) error {
	req := CreateSubnetRequest{CIDR: cidr}
	_, err := c.makeRequest("POST", "/whitelist", req)
	return err
}

func (c *Client) RemoveFromWhitelist(cidr string) error {
	req := DeleteSubnetRequest{CIDR: cidr}
	_, err := c.makeRequest("DELETE", "/whitelist", req)
	return err
}

func (c *Client) GetWhitelist() (*SubnetsListResponse, error) {
	respBody, err := c.makeRequest("GET", "/whitelist", nil)
	if err != nil {
		return nil, err
	}

	var response SubnetsListResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *Client) ResetBuckets(login, ip string) (*ResetBucketsResponse, error) {
	req := ResetBucketsRequest{
		Login: login,
		IP:    ip,
	}

	respBody, err := c.makeRequest("POST", "/reset", req)
	if err != nil {
		return nil, err
	}

	var response ResetBucketsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
