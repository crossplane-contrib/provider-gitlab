package projects

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

var (
	testHarborURL                        = "https://demo.goharbor.io"
	testHarborProjectName                = "myproject"
	testHarborUsername                   = "robot$gitlab"
	testHarborPassword                   = "supersecret"
	testHarborUseInheritedSettings       = false
	testHarborProjectID            int64 = 123
	testHarborIntegrationID        int64 = 999
)

func TestGenerateSetHarborServiceOptions(t *testing.T) {
	type args struct {
		parameters *projectsv1alpha1.IntegrationHarborParameters
		password   string
	}
	cases := map[string]struct {
		args args
		want *gitlab.SetHarborServiceOptions
	}{
		"AllFieldsSet": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationHarborParameters{
					ProjectID:            &testHarborProjectID,
					URL:                  testHarborURL,
					ProjectName:          testHarborProjectName,
					Username:             testHarborUsername,
					UseInheritedSettings: &testHarborUseInheritedSettings,
				},
				password: testHarborPassword,
			},
			want: &gitlab.SetHarborServiceOptions{
				URL:                  &testHarborURL,
				ProjectName:          &testHarborProjectName,
				Username:             &testHarborUsername,
				Password:             &testHarborPassword,
				UseInheritedSettings: &testHarborUseInheritedSettings,
			},
		},
		"EmptyPasswordOmitted": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationHarborParameters{
					URL:         testHarborURL,
					ProjectName: testHarborProjectName,
					Username:    testHarborUsername,
				},
				password: "",
			},
			want: &gitlab.SetHarborServiceOptions{
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
			want: &gitlab.SetHarborServiceOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetHarborServiceOptions(tc.args.parameters, tc.args.password)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetHarborServiceOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIntegrationHarborObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.HarborService
		want    projectsv1alpha1.IntegrationHarborObservation
	}{
		"FullObservation": {
			service: &gitlab.HarborService{
				Service: gitlab.Service{
					ID:    testHarborIntegrationID,
					Title: "Harbor",
					Slug:  "harbor",
				},
				Properties: &gitlab.HarborServiceProperties{
					URL:                  testHarborURL,
					ProjectName:          testHarborProjectName,
					Username:             testHarborUsername,
					Password:             testHarborPassword,
					UseInheritedSettings: testHarborUseInheritedSettings,
				},
			},
			want: projectsv1alpha1.IntegrationHarborObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To(testHarborIntegrationID),
					Title:                          ptr.To("Harbor"),
					Slug:                           ptr.To("harbor"),
					Active:                         ptr.To(false),
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
				URL:                  testHarborURL,
				ProjectName:          testHarborProjectName,
				Username:             testHarborUsername,
				UseInheritedSettings: testHarborUseInheritedSettings,
			},
		},
		"NilService": {
			service: nil,
			want:    projectsv1alpha1.IntegrationHarborObservation{},
		},
		"NilProperties": {
			service: &gitlab.HarborService{
				Service: gitlab.Service{ID: testHarborIntegrationID},
			},
			want: projectsv1alpha1.IntegrationHarborObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationHarborObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationHarborObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationHarborUpToDate(t *testing.T) {
	useInheritedTrue := true
	useInheritedFalse := false

	cases := map[string]struct {
		spec        *projectsv1alpha1.IntegrationHarborParameters
		observation *gitlab.HarborService
		want        bool
	}{
		"UpToDate": {
			spec: &projectsv1alpha1.IntegrationHarborParameters{
				URL:                  testHarborURL,
				ProjectName:          testHarborProjectName,
				Username:             testHarborUsername,
				UseInheritedSettings: &useInheritedFalse,
			},
			observation: &gitlab.HarborService{
				Properties: &gitlab.HarborServiceProperties{
					URL:                  testHarborURL,
					ProjectName:          testHarborProjectName,
					Username:             testHarborUsername,
					UseInheritedSettings: false,
				},
			},
			want: true,
		},
		"OutOfDateURL": {
			spec: &projectsv1alpha1.IntegrationHarborParameters{
				URL:         "https://other.harbor",
				ProjectName: testHarborProjectName,
				Username:    testHarborUsername,
			},
			observation: &gitlab.HarborService{
				Properties: &gitlab.HarborServiceProperties{
					URL:         testHarborURL,
					ProjectName: testHarborProjectName,
					Username:    testHarborUsername,
				},
			},
			want: false,
		},
		"OutOfDateProjectName": {
			spec: &projectsv1alpha1.IntegrationHarborParameters{
				URL:         testHarborURL,
				ProjectName: "other-project",
				Username:    testHarborUsername,
			},
			observation: &gitlab.HarborService{
				Properties: &gitlab.HarborServiceProperties{
					URL:         testHarborURL,
					ProjectName: testHarborProjectName,
					Username:    testHarborUsername,
				},
			},
			want: false,
		},
		"OutOfDateUseInheritedSettings": {
			spec: &projectsv1alpha1.IntegrationHarborParameters{
				URL:                  testHarborURL,
				ProjectName:          testHarborProjectName,
				Username:             testHarborUsername,
				UseInheritedSettings: &useInheritedTrue,
			},
			observation: &gitlab.HarborService{
				Properties: &gitlab.HarborServiceProperties{
					URL:                  testHarborURL,
					ProjectName:          testHarborProjectName,
					Username:             testHarborUsername,
					UseInheritedSettings: false,
				},
			},
			want: false,
		},
		"NilObservation": {
			spec:        &projectsv1alpha1.IntegrationHarborParameters{},
			observation: nil,
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

func TestLateInitializeIntegrationHarbor(t *testing.T) {
	useInheritedTrue := true

	cases := map[string]struct {
		spec *projectsv1alpha1.IntegrationHarborParameters
		svc  *gitlab.HarborService
		want *projectsv1alpha1.IntegrationHarborParameters
	}{
		"FillsUseInheritedSettings": {
			spec: &projectsv1alpha1.IntegrationHarborParameters{},
			svc: &gitlab.HarborService{
				Properties: &gitlab.HarborServiceProperties{UseInheritedSettings: true},
			},
			want: &projectsv1alpha1.IntegrationHarborParameters{
				UseInheritedSettings: ptr.To(true),
			},
		},
		"KeepsExistingUseInheritedSettings": {
			spec: &projectsv1alpha1.IntegrationHarborParameters{
				UseInheritedSettings: &useInheritedTrue,
			},
			svc: &gitlab.HarborService{
				Properties: &gitlab.HarborServiceProperties{UseInheritedSettings: false},
			},
			want: &projectsv1alpha1.IntegrationHarborParameters{
				UseInheritedSettings: &useInheritedTrue,
			},
		},
		"NilSpecIsNoop": {
			spec: nil,
			svc:  &gitlab.HarborService{Properties: &gitlab.HarborServiceProperties{}},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeIntegrationHarbor(tc.spec, tc.svc)
			if diff := cmp.Diff(tc.want, tc.spec); diff != "" {
				t.Errorf("LateInitializeIntegrationHarbor(): -want, +got:\n%s", diff)
			}
		})
	}
}
