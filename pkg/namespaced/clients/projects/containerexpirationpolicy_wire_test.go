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

package projects

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go/v3"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

// GitLab only declares name_regex as the deletion pattern. Earlier client-go
// versions overwrote ContainerExpirationPolicyAttributes.NameRegex with
// NameRegexDelete before encoding the request, so a pattern written to
// NameRegex was dropped; since v3.16.2 NameRegexDelete is only a fallback for
// an unset NameRegex. Either way the pattern has to leave as name_regex.
//
// GenerateCreateProjectOptions and GenerateEditProjectOptions are checked
// elsewhere against the options struct, which is the state before client-go
// touches those fields. These tests go through the real client and assert the
// encoded request body instead, so the mapping stays correct even if client-go
// changes which field it forwards.

// captureRequestBody runs call against a server that records the request body
// and returns container_expiration_policy_attributes from it.
func captureRequestBody(t *testing.T, call func(*gitlab.Client)) map[string]any {
	t.Helper()

	var raw []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		raw = body
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"id":1}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := gitlab.NewClient("token", gitlab.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	call(client)

	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal request body %q: %v", string(raw), err)
	}
	attributes, ok := decoded["container_expiration_policy_attributes"].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return attributes
}

func containerExpirationPolicyParameters() *v1alpha1.ProjectParameters {
	pattern := "delete-me-.*"
	return &v1alpha1.ProjectParameters{
		ContainerExpirationPolicyAttributes: &v1alpha1.ContainerExpirationPolicyAttributes{
			NameRegex: &pattern,
		},
	}
}

func TestCreateProjectSendsNameRegex(t *testing.T) {
	options := GenerateCreateProjectOptions("example", containerExpirationPolicyParameters())

	attributes := captureRequestBody(t, func(client *gitlab.Client) {
		_, _, _ = client.Projects.CreateProject(options)
	})

	if got := attributes["name_regex"]; got != "delete-me-.*" {
		t.Errorf("name_regex in request body: got %v, want %q (body attributes: %v)",
			got, "delete-me-.*", attributes)
	}
}

func TestEditProjectSendsNameRegex(t *testing.T) {
	options := GenerateEditProjectOptions("example", containerExpirationPolicyParameters())

	attributes := captureRequestBody(t, func(client *gitlab.Client) {
		_, _, _ = client.Projects.EditProject(1, options)
	})

	if got := attributes["name_regex"]; got != "delete-me-.*" {
		t.Errorf("name_regex in request body: got %v, want %q (body attributes: %v)",
			got, "delete-me-.*", attributes)
	}
}
