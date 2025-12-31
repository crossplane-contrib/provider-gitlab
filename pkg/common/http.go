/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"
)

const (
	httpRequestTimeout         = 30 * time.Second
	maxResponseSize            = 1024 * 1024 // 1MB max response size
	errFetchFromEndpoint       = "cannot fetch from endpoint"
	errInvalidEndpointResponse = "invalid response from endpoint"
)

// AuthParameters holds authentication details for HTTP requests
type AuthParameters struct {
	Username *string
	Password *string
	Token    *string
}

// RequestParameters holds parameters for making HTTP requests
type RequestParameters struct {
	Auth        *AuthParameters
	EndpointURL string
	Timeout     *time.Duration
	MaxSize     *int
	Retries     *int
}

// FetchFromEndpoint fetches content from the provided endpoint URL
// using the provided authentication details. It supports:
// - Unauthenticated requests (Auth nil or all auth params nil)
// - Basic authentication (username + password)
// - Bearer token authentication (token)
// Token authentication takes precedence over basic auth if both are provided.
func FetchFromEndpoint(ctx context.Context, params RequestParameters) (string, error) {
	if params.EndpointURL == "" {
		return "", errors.New("endpoint URL cannot be empty")
	}

	// Determine timeout
	timeout := httpRequestTimeout
	if params.Timeout != nil {
		timeout = *params.Timeout
	}

	// Determine max size
	maxSize := int64(maxResponseSize)
	if params.MaxSize != nil {
		maxSize = int64(*params.MaxSize)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.EndpointURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create HTTP request")
	}

	// Apply authentication
	applyAuthentication(req, params.Auth)

	// Use cleanhttp for a safer default HTTP client
	client := cleanhttp.DefaultClient()
	client.Timeout = timeout

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, errFetchFromEndpoint)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check HTTP status code
	if resp.StatusCode/100 != 2 {
		return "", errors.Errorf("endpoint returned non-success status code: %d", resp.StatusCode)
	}

	// Read response body with size limit
	limitedReader := io.LimitReader(resp.Body, maxSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	// Convert to string and trim whitespace
	result := strings.TrimSpace(string(body))
	if result == "" {
		return "", errors.New(errInvalidEndpointResponse)
	}

	return result, nil
}

// applyAuthentication applies the appropriate authentication method to the HTTP request.
// Priority: Bearer token > Basic Auth > None

func applyAuthentication(req *http.Request, auth *AuthParameters) {
	if auth == nil {
		return
	}
	// Token authentication takes precedence
	if auth.Token != nil && *auth.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *auth.Token))
		return
	}

	// Basic authentication if username is provided
	if auth.Username != nil && *auth.Username != "" {
		// Password can be empty for basic auth
		pwd := ""
		if auth.Password != nil {
			pwd = *auth.Password
		}
		// Encode credentials
		creds := fmt.Sprintf("%s:%s", *auth.Username, pwd)
		encodedCreds := base64.StdEncoding.EncodeToString([]byte(creds))
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encodedCreds))
		return
	}

	// No authentication - allow unauthenticated requests
}
