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

package clients

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

func TestIsResponseNotFound(t *testing.T) {
	cases := map[string]struct {
		status      int
		contentType string
		body        string
		want        bool
	}{
		"GitLabNotFound": {
			status: http.StatusNotFound, contentType: "application/json", body: `{"message":"404 User Not Found"}`, want: true,
		},
		"GitLabNotFoundWithCharset": {
			status: http.StatusNotFound, contentType: "application/json; charset=utf-8", body: `{"message":"404 Not found"}`, want: true,
		},
		"ProxyPlainTextNotFound": {
			status: http.StatusNotFound, contentType: "text/plain; charset=utf-8", body: "404 page not found\n", want: false,
		},
		"ProxyHTMLNotFound": {
			status: http.StatusNotFound, contentType: "text/html", body: "<html><body>Not Found</body></html>", want: false,
		},
		"NotFoundWithoutContentType": {
			status: http.StatusNotFound, contentType: "", body: "", want: false,
		},
		"GitLabServerError": {
			status: http.StatusInternalServerError, contentType: "application/json", body: `{"message":"500 Internal Server Error"}`, want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.contentType == "" {
					// A nil value stops net/http from sniffing a Content-Type.
					w.Header()["Content-Type"] = nil
				} else {
					w.Header().Set("Content-Type", tc.contentType)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			git, err := gitlab.NewClient("token", gitlab.WithBaseURL(srv.URL), gitlab.WithoutRetries())
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			_, res, err := git.Users.GetUser(1, &gitlab.GetUserOptions{})
			if err == nil {
				t.Fatalf("GetUser: want an error for HTTP %d, got nil", tc.status)
			}

			if got := IsResponseNotFound(res); got != tc.want {
				t.Errorf("IsResponseNotFound(HTTP %d %q) = %v, want %v", tc.status, tc.contentType, got, tc.want)
			}
		})
	}
}

func TestIsResponseNotFoundWithoutHTTPResponse(t *testing.T) {
	if IsResponseNotFound(nil) {
		t.Error("IsResponseNotFound(nil) = true, want false")
	}
	if IsResponseNotFound(&gitlab.Response{}) {
		t.Error("IsResponseNotFound(&gitlab.Response{}) = true, want false")
	}
}

func TestAreStringSlicesEqual(t *testing.T) {
	cases := map[string]struct {
		a    []string
		b    []string
		want bool
	}{
		"BothNil": {
			a: nil, b: nil, want: true,
		},
		"BothEmpty": {
			a: []string{}, b: []string{}, want: true,
		},
		"NilAndEmpty": {
			a: nil, b: []string{}, want: true,
		},
		"Equal": {
			a: []string{"a", "b"}, b: []string{"a", "b"}, want: true,
		},
		"EqualDifferentOrder": {
			a: []string{"b", "a"}, b: []string{"a", "b"}, want: true,
		},
		"DifferentLength": {
			a: []string{"a", "b"}, b: []string{"a"}, want: false,
		},
		"DifferentElements": {
			a: []string{"a", "b"}, b: []string{"a", "c"}, want: false,
		},
		"DuplicatesInA": {
			a: []string{"a", "a"}, b: []string{"a", "b"}, want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := AreStringSlicesEqual(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("AreStringSlicesEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
