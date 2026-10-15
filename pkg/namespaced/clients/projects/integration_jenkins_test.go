package projects

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestGenerateSetJenkinsServiceOptions(t *testing.T) {
	type args struct {
		parameters *projectsv1alpha1.IntegrationJenkinsParameters
		password   *string
	}
	cases := map[string]struct {
		args args
		want *gitlab.SetJenkinsCIServiceOptions
	}{
		"AllFieldsSet": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationJenkinsParameters{
					ProjectID:             ptr.To[int64](123),
					URL:                   "https://jenkins.example.com",
					ProjectName:           "my-job",
					EnableSSLVerification: ptr.To(false),
					Username:              ptr.To("bot"),
					PushEvents:            ptr.To(true),
					MergeRequestsEvents:   ptr.To(false),
					TagPushEvents:         ptr.To(true),
				},
				password: ptr.To("supersecret"),
			},
			want: &gitlab.SetJenkinsCIServiceOptions{
				URL:                   ptr.To("https://jenkins.example.com"),
				ProjectName:           ptr.To("my-job"),
				EnableSSLVerification: ptr.To(false),
				Username:              ptr.To("bot"),
				Password:              ptr.To("supersecret"),
				PushEvents:            ptr.To(true),
				MergeRequestsEvents:   ptr.To(false),
				TagPushEvents:         ptr.To(true),
			},
		},
		"NilPasswordOmitted": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationJenkinsParameters{
					URL:         "https://jenkins.example.com",
					ProjectName: "my-job",
				},
			},
			want: &gitlab.SetJenkinsCIServiceOptions{
				URL:         ptr.To("https://jenkins.example.com"),
				ProjectName: ptr.To("my-job"),
			},
		},
		"NilInput": {
			args: args{password: ptr.To("supersecret")},
			want: &gitlab.SetJenkinsCIServiceOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetJenkinsServiceOptions(tc.args.parameters, tc.args.password)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetJenkinsServiceOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIntegrationJenkinsObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.JenkinsCIService
		want    projectsv1alpha1.IntegrationJenkinsObservation
	}{
		"FullObservation": {
			service: &gitlab.JenkinsCIService{
				Service: gitlab.Service{
					ID:         7,
					Title:      "Jenkins",
					Slug:       "jenkins",
					Active:     true,
					PushEvents: true,
				},
				Properties: &gitlab.JenkinsCIServiceProperties{
					URL:                   "https://jenkins.example.com",
					EnableSSLVerification: true,
					ProjectName:           "my-job",
					Username:              "bot",
				},
			},
			want: projectsv1alpha1.IntegrationJenkinsObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To[int64](7),
					Title:                          ptr.To("Jenkins"),
					Slug:                           ptr.To("jenkins"),
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
					PushEvents:                     ptr.To(true),
					TagPushEvents:                  ptr.To(false),
					VulnerabilityEvents:            ptr.To(false),
					WikiPageEvents:                 ptr.To(false),
					CommentOnEventEnabled:          ptr.To(false),
					Inherited:                      ptr.To(false),
				},
				URL:                   "https://jenkins.example.com",
				ProjectName:           "my-job",
				EnableSSLVerification: true,
				Username:              "bot",
			},
		},
		"NilService": {
			want: projectsv1alpha1.IntegrationJenkinsObservation{},
		},
		"NilProperties": {
			service: &gitlab.JenkinsCIService{Service: gitlab.Service{ID: 7}},
			want:    projectsv1alpha1.IntegrationJenkinsObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationJenkinsObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationJenkinsObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationJenkinsUpToDate(t *testing.T) {
	observed := func() *gitlab.JenkinsCIService {
		return &gitlab.JenkinsCIService{
			Service: gitlab.Service{PushEvents: true},
			Properties: &gitlab.JenkinsCIServiceProperties{
				URL:                   "https://jenkins.example.com",
				EnableSSLVerification: true,
				ProjectName:           "my-job",
				Username:              "bot",
			},
		}
	}

	cases := map[string]struct {
		spec        *projectsv1alpha1.IntegrationJenkinsParameters
		observation *gitlab.JenkinsCIService
		want        bool
	}{
		"UpToDate": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:                   "https://jenkins.example.com",
				ProjectName:           "my-job",
				EnableSSLVerification: ptr.To(true),
				Username:              ptr.To("bot"),
				PushEvents:            ptr.To(true),
			},
			observation: observed(),
			want:        true,
		},
		"OutOfDateURL": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:         "https://other.example.com",
				ProjectName: "my-job",
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateProjectName": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:         "https://jenkins.example.com",
				ProjectName: "other-job",
			},
			observation: observed(),
			want:        false,
		},
		"UnsetEnableSSLVerificationIgnoresNullRemote": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:         "https://jenkins.example.com",
				ProjectName: "my-job",
			},
			observation: decodeJenkins(t, `{"properties":{"jenkins_url":"https://jenkins.example.com","project_name":"my-job","enable_ssl_verification":null}}`),
			want:        true,
		},
		"OutOfDateEnableSSLVerification": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:                   "https://jenkins.example.com",
				ProjectName:           "my-job",
				EnableSSLVerification: ptr.To(false),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDatePushEvents": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:         "https://jenkins.example.com",
				ProjectName: "my-job",
				PushEvents:  ptr.To(false),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateUsername": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:         "https://jenkins.example.com",
				ProjectName: "my-job",
				Username:    ptr.To("other"),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateMergeRequestsEvents": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:                 "https://jenkins.example.com",
				ProjectName:         "my-job",
				MergeRequestsEvents: ptr.To(true),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateTagPushEvents": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				URL:           "https://jenkins.example.com",
				ProjectName:   "my-job",
				TagPushEvents: ptr.To(true),
			},
			observation: observed(),
			want:        false,
		},
		"NilObservation": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{},
			want: false,
		},
		"NilProperties": {
			spec:        &projectsv1alpha1.IntegrationJenkinsParameters{},
			observation: &gitlab.JenkinsCIService{},
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationJenkinsUpToDate(tc.spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationJenkinsUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLateInitializeIntegrationJenkins(t *testing.T) {
	cases := map[string]struct {
		spec *projectsv1alpha1.IntegrationJenkinsParameters
		svc  *gitlab.JenkinsCIService
		want *projectsv1alpha1.IntegrationJenkinsParameters
	}{
		"FillsNilFields": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{},
			svc: &gitlab.JenkinsCIService{
				Service: gitlab.Service{PushEvents: true, MergeRequestsEvents: false, TagPushEvents: true},
				Properties: &gitlab.JenkinsCIServiceProperties{
					EnableSSLVerification: true,
					Username:              "bot",
				},
			},
			want: &projectsv1alpha1.IntegrationJenkinsParameters{
				Username:            ptr.To("bot"),
				PushEvents:          ptr.To(true),
				MergeRequestsEvents: ptr.To(false),
				TagPushEvents:       ptr.To(true),
			},
		},
		"EmptyUsernameNotInitialized": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{},
			svc: &gitlab.JenkinsCIService{
				Properties: &gitlab.JenkinsCIServiceProperties{},
			},
			want: &projectsv1alpha1.IntegrationJenkinsParameters{
				PushEvents:          ptr.To(false),
				MergeRequestsEvents: ptr.To(false),
				TagPushEvents:       ptr.To(false),
			},
		},
		"NullEnableSSLVerificationNotInitialized": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{},
			svc:  decodeJenkins(t, `{"properties":{"enable_ssl_verification":null}}`),
			want: &projectsv1alpha1.IntegrationJenkinsParameters{
				PushEvents:          ptr.To(false),
				MergeRequestsEvents: ptr.To(false),
				TagPushEvents:       ptr.To(false),
			},
		},
		"KeepsExistingValues": {
			spec: &projectsv1alpha1.IntegrationJenkinsParameters{
				EnableSSLVerification: ptr.To(false),
				Username:              ptr.To("mine"),
				PushEvents:            ptr.To(false),
				MergeRequestsEvents:   ptr.To(true),
				TagPushEvents:         ptr.To(false),
			},
			svc: &gitlab.JenkinsCIService{
				Service: gitlab.Service{PushEvents: true, MergeRequestsEvents: false, TagPushEvents: true},
				Properties: &gitlab.JenkinsCIServiceProperties{
					EnableSSLVerification: true,
					Username:              "bot",
				},
			},
			want: &projectsv1alpha1.IntegrationJenkinsParameters{
				EnableSSLVerification: ptr.To(false),
				Username:              ptr.To("mine"),
				PushEvents:            ptr.To(false),
				MergeRequestsEvents:   ptr.To(true),
				TagPushEvents:         ptr.To(false),
			},
		},
		"NilSpecIsNoop": {
			svc:  &gitlab.JenkinsCIService{Properties: &gitlab.JenkinsCIServiceProperties{}},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeIntegrationJenkins(tc.spec, tc.svc)
			if diff := cmp.Diff(tc.want, tc.spec); diff != "" {
				t.Errorf("LateInitializeIntegrationJenkins(): -want, +got:\n%s", diff)
			}
		})
	}
}

func decodeJenkins(t *testing.T, raw string) *gitlab.JenkinsCIService {
	t.Helper()
	svc := &gitlab.JenkinsCIService{}
	if err := json.Unmarshal([]byte(raw), svc); err != nil {
		t.Fatalf("json.Unmarshal(): %v", err)
	}
	return svc
}

func TestIntegrationJenkinsObservationJSONKeepsFalse(t *testing.T) {
	obs := GenerateIntegrationJenkinsObservation(&gitlab.JenkinsCIService{
		Properties: &gitlab.JenkinsCIServiceProperties{EnableSSLVerification: false},
	})
	raw, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("json.Marshal(): %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal(): %v", err)
	}
	if diff := cmp.Diff("false", string(got["enableSslVerification"])); diff != "" {
		t.Errorf("enableSslVerification: -want, +got:\n%s", diff)
	}
}
