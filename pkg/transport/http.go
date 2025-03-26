package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HTTP implements the Transport interface using HTTP calls.
type HTTP struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTP creates a new HTTP transport with the given base URL.
func NewHTTP(baseURL string) *HTTP {
	return &HTTP{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Execute implements the Transport interface by sending HTTP requests
// to the MCP server and parsing the responses.
func (t *HTTP) Execute(method string, params any) (map[string]any, error) {
	var endpoint string
	var httpMethod string
	var reqBody io.Reader

	switch method {
	case "tools/list":
		endpoint = fmt.Sprintf("%s/v1/tools", t.baseURL)
		httpMethod = http.MethodGet
	case "resources/list":
		endpoint = fmt.Sprintf("%s/v1/resources", t.baseURL)
		httpMethod = http.MethodGet
	case "prompts/list":
		endpoint = fmt.Sprintf("%s/v1/prompts", t.baseURL)
		httpMethod = http.MethodGet
	case "tools/call":
		toolParams, ok := params.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid params for tools/call: expected map[string]any")
		}

		toolName, ok := toolParams["name"].(string)
		if !ok {
			return nil, fmt.Errorf("tool name is required for tools/call")
		}

		endpoint = fmt.Sprintf("%s/v1/tools/%s", t.baseURL, url.PathEscape(toolName))
		httpMethod = http.MethodPost

		if arguments, ok := toolParams["arguments"].(map[string]any); ok && len(arguments) > 0 {
			jsonBody, err := json.Marshal(arguments)
			if err != nil {
				return nil, fmt.Errorf("error marshaling tool arguments: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		}
	case "resources/read":
		resParams, ok := params.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid params for resources/read: expected map[string]any")
		}

		uri, ok := resParams["uri"].(string)
		if !ok {
			return nil, fmt.Errorf("uri is required for resources/read")
		}

		endpoint = fmt.Sprintf("%s/v1/resources/%s", t.baseURL, url.PathEscape(uri))
		httpMethod = http.MethodGet
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, httpMethod, endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	return result, nil
}
