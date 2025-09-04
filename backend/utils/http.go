package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api-monitor/app/models"
)

// CheckEndpoint performs an HTTP check on the given endpoint
func CheckEndpoint(endpoint models.APIEndpoint) (int, int, string, string, error) {
	start := time.Now()

	// Create HTTP client with optional proxy
	client := &http.Client{
		Timeout: time.Duration(endpoint.TimeoutSeconds) * time.Second,
	}

	// Configure proxy if specified
	if endpoint.Proxy != nil && endpoint.Proxy.Host != "" {
		proxyURL := fmt.Sprintf("http://%s:%d", endpoint.Proxy.Host, endpoint.Proxy.Port)

		// Add authentication if provided
		if endpoint.Proxy.Username != "" && endpoint.Proxy.Password != "" {
			proxyURL = fmt.Sprintf("http://%s:%s@%s:%d",
				endpoint.Proxy.Username, endpoint.Proxy.Password,
				endpoint.Proxy.Host, endpoint.Proxy.Port)
		}

		proxy, err := url.Parse(proxyURL)
		if err != nil {
			log.Printf("Error parsing proxy URL for endpoint %s: %v", endpoint.Name, err)
		} else {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	var req *http.Request
	var err error

	if endpoint.Body != "" {
		req, err = http.NewRequest(endpoint.Method, endpoint.URL, strings.NewReader(endpoint.Body))
	} else {
		req, err = http.NewRequest(endpoint.Method, endpoint.URL, nil)
	}

	if err != nil {
		return 0, 0, "", "", fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	duration := time.Since(start)
	durationMs := int(duration.Milliseconds())

	if err != nil {
		return 0, durationMs, "", "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Collect response headers
	headersStr := ""
	if resp.Header != nil {
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[key] = values[0] // Take first value if multiple
			}
		}
		if headersBytes, err := json.Marshal(headers); err == nil {
			headersStr = string(headersBytes)
		}
	}

	// Truncate response body if too long
	if len(bodyStr) > 1000 {
		bodyStr = bodyStr[:1000] + "... (truncated)"
	}

	return resp.StatusCode, durationMs, bodyStr, headersStr, nil
}

// ValidateUTF8 cleans strings to ensure UTF-8 compatibility
func ValidateUTF8(s string) string {
	return strings.ToValidUTF8(s, "")
}
