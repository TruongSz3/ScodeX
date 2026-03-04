package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sz3/scodex/internal/api"
)

const defaultBaseURL = "http://127.0.0.1:7777"

type Client struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		trimmed := strings.TrimRight(baseURL, "/")
		if trimmed != "" {
			c.baseURL = trimmed
		}
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithAuthToken(authToken string) Option {
	return func(c *Client) {
		trimmed := strings.TrimSpace(authToken)
		if trimmed != "" {
			c.authToken = trimmed
		}
	}
}

func New(opts ...Option) *Client {
	authToken := strings.TrimSpace(os.Getenv("AGENT_AUTH_TOKEN"))
	if authToken == "" {
		authToken = api.DefaultAuthToken
	}

	client := &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		authToken: authToken,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type HealthResponse struct {
	Status string `json:"status"`
}

type InitializeRequest struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientName      string `json:"clientName"`
}

type InitializeResponse struct {
	SessionID string `json:"sessionId"`
}

type InitializedRequest struct {
	SessionID string `json:"sessionId"`
}

type InitializedResponse struct {
	Status string `json:"status"`
}

func (c *Client) CheckHealth(ctx context.Context) (HealthResponse, error) {
	var out HealthResponse
	err := c.getJSON(ctx, "/health", &out)
	if err != nil {
		return HealthResponse{}, err
	}
	return out, nil
}

func (c *Client) Initialize(ctx context.Context, req InitializeRequest) (InitializeResponse, error) {
	var out InitializeResponse
	err := c.postJSON(ctx, "/runtime/initialize", req, &out)
	if err != nil {
		return InitializeResponse{}, err
	}
	return out, nil
}

func (c *Client) Initialized(ctx context.Context, req InitializedRequest) (InitializedResponse, error) {
	var out InitializedResponse
	err := c.postJSON(ctx, "/runtime/initialized", req, &out)
	if err != nil {
		return InitializedResponse{}, err
	}
	return out, nil
}

func (c *Client) PingHandshake(ctx context.Context) error {
	initResp, err := c.Initialize(ctx, InitializeRequest{
		ProtocolVersion: "v1",
		ClientName:      "agent-cli",
	})
	if err != nil {
		return err
	}

	_, err = c.Initialized(ctx, InitializedRequest{SessionID: initResp.SessionID})
	return err
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errorFromResponse(response)
	}

	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) postJSON(ctx context.Context, path string, body any, out any) error {
	encodedBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode request body: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+path,
		bytes.NewReader(encodedBody),
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set(api.HeaderAuthToken, c.authToken)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errorFromResponse(response)
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(response.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func errorFromResponse(response *http.Response) error {
	bodyBytes, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	bodyText := strings.TrimSpace(string(bodyBytes))
	if bodyText == "" {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}
	return fmt.Errorf("unexpected status code: %d: %s", response.StatusCode, bodyText)
}
