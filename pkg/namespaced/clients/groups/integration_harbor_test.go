package groups

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	groupsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

var (
	testHarborURL                        = "https://demo.goharbor.io"
	testHarborProjectName                = "myproject"
	testHarborUsername                   = "robot$gitlab"
	testHarborPassword                   = "supersecret"
	testHarborUseInheritedSettings       = false
	testHarborGroupID              int64 = 321
	testHarborIntegrationID        int64 = 999
)

func TestGenerateSetUpHarborOptions(t *testing.T) {
	type args struct {
		parameters *groupsv1alpha1.IntegrationHarborParameters
		password   string
	}
	cases := map[string]struct {
		args args
		want *gitlab.SetUpHarborOptions
	}{
		"AllFieldsSet": {
			args: args{
				parameters: &groupsv1alpha1.IntegrationHarborParameters{
					GroupID:              &testHarborGroupID,
					URL:                  testHarborURL,
					ProjectName:          testHarborProjectName,
					Username:             testHarborUsername,
					UseInheritedSettings: &testHarborUseInheritedSettings,
				},
				password: testHarborPassword,
			},
			want: &gitlab.SetUpHarborOptions{
				URL:                  &testHarborURL,
				ProjectName:          &testHarborProjectName,
				Username:             &testHarborUsername,
				Password:             &testHarborPassword,
				UseInheritedSettings: &testHarborUseInheritedSettings,
			},
		},
		"EmptyPasswordOmitted": {
			args: args{
				parameters: &groupsv1alpha1.IntegrationHarborParameters{
					URL:         testHarborURL,
					ProjectName: testHarborProjectName,
					Username:    testHarborUsername,
				},
				password: "",
			},
			want: &gitlab.SetUpHarborOptions{
				URL:         &testHarborURL,
				ProjectName: &testHarborProjectName,
				Username:    &testHarborUsername,
			},
		},
		"NilInput": {
			args: args{
				parameters: nil,
				password:   testHarborPassword,
			},
			want: &gitlab.SetUpHarborOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetUpHarborOptions(tc.args.parameters, tc.args.password)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetUpHarborOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIntegrationHarborObservation(t *testing.T) {
	cases := map[string]struct {
		integration *gitlab.Integration
		want        groupsv1alpha1.IntegrationHarborObservation
	}{
		"ActiveIntegration": {
			integration: &gitlab.Integration{
				ID:     testHarborIntegrationID,
				Title:  "Harbor",
				Slug:   "harbor",
				Active: true,
			},
			want: groupsv1alpha1.IntegrationHarborObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To(testHarborIntegrationID),
					Title:                          ptr.To("Harbor"),
					Slug:                           ptr.To("harbor"),
					Active:                         ptr.To(true),
					AlertEvents:                    ptr.To(false),
					CommitEvents:                   ptr.To(false),
					ConfidentialIssuesEvents:       ptr.To(false),
					ConfidentialNoteEvents:         ptr.To(false),
					DeploymentEvents:               ptr.To(false),
					GroupConfidentialMentionEvents: ptr.To(false),
					GroupMentionEvents:             ptr.To(false),
					IncidentEvents:                 ptr.To(false),
					IssuesEvents:                   ptr.To(false),
					JobEvents:                      ptr.To(false),
					MergeRequestsEvents:            ptr.To(false),
					NoteEvents:                     ptr.To(false),
					PipelineEvents:                 ptr.To(false),
					PushEvents:                     ptr.To(false),
					TagPushEvents:                  ptr.To(false),
					VulnerabilityEvents:            ptr.To(false),
					WikiPageEvents:                 ptr.To(false),
					CommentOnEventEnabled:          ptr.To(false),
					Inherited:                      ptr.To(false),
				},
			},
		},
		"NilIntegration": {
			integration: nil,
			want:        groupsv1alpha1.IntegrationHarborObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationHarborObservation(tc.integration)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationHarborObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationHarborUpToDate(t *testing.T) {
	cases := map[string]struct {
		spec        *groupsv1alpha1.IntegrationHarborParameters
		observation *gitlab.Integration
		want        bool
	}{
		"ActiveIntegration": {
			spec:        &groupsv1alpha1.IntegrationHarborParameters{URL: testHarborURL},
			observation: &gitlab.Integration{Active: true},
			want:        true,
		},
		"InactiveIntegration": {
			spec:        &groupsv1alpha1.IntegrationHarborParameters{URL: testHarborURL},
			observation: &gitlab.Integration{Active: false},
			want:        false,
		},
		"NilObservation": {
			spec:        &groupsv1alpha1.IntegrationHarborParameters{},
			observation: nil,
			want:        false,
		},
		"NilSpec": {
			spec:        nil,
			observation: &gitlab.Integration{Active: true},
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationHarborUpToDate(tc.spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationHarborUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}
