package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HTTPTransport provides HTTP transport for MCP servers
type HTTPTransport struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewHTTP creates a new HTTP transport
func NewHTTP(baseURL string) *HTTPTransport {
	return &HTTPTransport{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// Execute sends a request to the MCP server and returns the response
func (t *HTTPTransport) Execute(method string, params interface{}) (map[string]interface{}, error) {
	// For HTTP, we translate the method to a URL pattern
	var endpoint string
	var httpMethod string
	var reqBody io.Reader

	if method == "tools/list" {
		endpoint = fmt.Sprintf("%s/v1/tools", t.BaseURL)
		httpMethod = "GET"
	} else if method == "resources/list" {
		endpoint = fmt.Sprintf("%s/v1/resources", t.BaseURL)
		httpMethod = "GET"
	} else if method == "tools/call" {
		if toolParams, ok := params.(map[string]interface{}); ok {
			toolName, _ := toolParams["name"].(string)
			endpoint = fmt.Sprintf("%s/v1/tools/%s", t.BaseURL, url.PathEscape(toolName))
			httpMethod = "POST"
			
			if arguments, ok := toolParams["arguments"].(map[string]interface{}); ok && len(arguments) > 0 {
				jsonBody, err := json.Marshal(arguments)
				if err != nil {
					return nil, fmt.Errorf("error marshaling JSON: %w", err)
				}
				reqBody = bytes.NewBuffer(jsonBody)
			}
		} else {
			return nil, fmt.Errorf("invalid params for tools/call")
		}
	} else {
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	req, err := http.NewRequest(httpMethod, endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned non-success status: %d - %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	return result, nil
} 