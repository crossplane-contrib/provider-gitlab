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

package groups

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	groupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

// GitLab returns notify_only_broken_pipelines, branches_to_be_notified,
// labels_to_be_notified and labels_to_be_notified_behavior under properties,
// and null for them at the top level, where gitlab.GroupMattermostIntegration
// reads them. notify_only_default_branch is not returned at all.
// These tests go through the real client with responses of that shape, as
// returned by GitLab 19.4, so that these fields are compared with what GitLab
// actually stores.
const mattermostResponse = `{
	"id": 1000,
	"title": "Mattermost notifications",
	"slug": "mattermost",
	"active": true,
	"inherited": false,
	"push_events": true,
	"pipeline_events": true,
	"notify_only_broken_pipelines": null,
	"branches_to_be_notified": null,
	"labels_to_be_notified": null,
	"labels_to_be_notified_behavior": null,
	"notify_only_default_branch": null,
	"properties": {
		"username": "gitlab-bot",
		"channel": "gitlab-notifications",
		"notify_only_broken_pipelines": %s,
		"branches_to_be_notified": "all",
		"labels_to_be_notified": "foo",
		"labels_to_be_notified_behavior": "match_all",
		"pipeline_channel": "gitlab-pipelines",
		"vulnerability_channel": "vulnerability"
	}
}`

// getMattermost serves body for the group Mattermost integration and returns
// what the client decodes, along with the escaped path that was requested.
func getMattermost(t *testing.T, gid any, body string) (*GroupMattermostIntegration, string) {
	t.Helper()

	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewMattermostClient(common.Config{Token: "token", BaseURL: server.URL})
	got, _, err := client.GetGroupMattermostIntegration(gid)
	if err != nil {
		t.Fatalf("GetGroupMattermostIntegration(): %v", err)
	}
	return got, path
}

func TestGetGroupMattermostIntegrationWire(t *testing.T) {
	cases := map[string]struct {
		gid      any
		broken   string
		wantPath string
	}{
		"IntegerID": {
			gid:      int64(20),
			broken:   `true`,
			wantPath: "/api/v4/groups/20/integrations/mattermost",
		},
		"GroupPath": {
			gid:      "parent/child",
			broken:   `true`,
			wantPath: "/api/v4/groups/parent%2Fchild/integrations/mattermost",
		},
		"StringBoolean": {
			gid:      20,
			broken:   `"true"`,
			wantPath: "/api/v4/groups/20/integrations/mattermost",
		},
	}

	spec := &groupsv1alpha1.IntegrationMattermostParameters{
		Username:                   ptr.To("gitlab-bot"),
		Channel:                    ptr.To("gitlab-notifications"),
		NotifyOnlyBrokenPipelines:  ptr.To(true),
		BranchesToBeNotified:       ptr.To("all"),
		LabelsToBeNotified:         ptr.To("foo"),
		LabelsToBeNotifiedBehavior: ptr.To("match_all"),
		PushEvents:                 ptr.To(true),
		PipelineEvents:             ptr.To(true),
		PipelineChannel:            ptr.To("gitlab-pipelines"),
		VulnerabilityChannel:       ptr.To("vulnerability"),
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, path := getMattermost(t, tc.gid, fmt.Sprintf(mattermostResponse, tc.broken))
			if diff := cmp.Diff(tc.wantPath, path); diff != "" {
				t.Errorf("request path: -want, +got:\n%s", diff)
			}

			if !IsIntegrationMattermostUpToDate(spec, got) {
				t.Errorf("IsIntegrationMattermostUpToDate(): got false, want true")
			}

			obs := GenerateIntegrationMattermostObservation(got)
			want := groupsv1alpha1.IntegrationMattermostObservation{
				CommonIntegrationObservation: obs.CommonIntegrationObservation,
				Username:                     "gitlab-bot",
				Channel:                      "gitlab-notifications",
				NotifyOnlyBrokenPipelines:    true,
				BranchesToBeNotified:         "all",
				LabelsToBeNotified:           "foo",
				LabelsToBeNotifiedBehavior:   "match_all",
				PipelineChannel:              "gitlab-pipelines",
				VulnerabilityChannel:         "vulnerability",
			}
			if diff := cmp.Diff(want, obs); diff != "" {
				t.Errorf("GenerateIntegrationMattermostObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetGroupMattermostIntegrationWireDiffers(t *testing.T) {
	got, _ := getMattermost(t, 20, fmt.Sprintf(mattermostResponse, `false`))

	differs := map[string]*groupsv1alpha1.IntegrationMattermostParameters{
		"NotifyOnlyBrokenPipelines":  {NotifyOnlyBrokenPipelines: ptr.To(true)},
		"BranchesToBeNotified":       {BranchesToBeNotified: ptr.To("default")},
		"LabelsToBeNotified":         {LabelsToBeNotified: ptr.To("bar")},
		"LabelsToBeNotifiedBehavior": {LabelsToBeNotifiedBehavior: ptr.To("match_any")},
	}
	for name, spec := range differs {
		t.Run(name, func(t *testing.T) {
			if IsIntegrationMattermostUpToDate(spec, got) {
				t.Errorf("IsIntegrationMattermostUpToDate(): got true, want false")
			}
		})
	}
}

func TestGetGroupMattermostIntegrationInvalidID(t *testing.T) {
	client := NewMattermostClient(common.Config{Token: "token", BaseURL: "http://127.0.0.1:0"})
	if _, _, err := client.GetGroupMattermostIntegration(1.5); err == nil {
		t.Errorf("GetGroupMattermostIntegration(): got nil error, want an invalid ID error")
	}
}
