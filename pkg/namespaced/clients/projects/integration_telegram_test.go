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

func TestGenerateSetTelegramServiceOptions(t *testing.T) {
	type args struct {
		parameters *projectsv1alpha1.IntegrationTelegramParameters
		token      string
	}
	cases := map[string]struct {
		args args
		want *gitlab.SetTelegramServiceOptions
	}{
		"AllFieldsSet": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationTelegramParameters{
					ProjectID:                 ptr.To[int64](123),
					Room:                      "@channel",
					NotifyOnlyBrokenPipelines: ptr.To(true),
					BranchesToBeNotified:      ptr.To("protected"),
					PushEvents:                ptr.To(true),
					IssuesEvents:              ptr.To(false),
					ConfidentialIssuesEvents:  ptr.To(true),
					MergeRequestsEvents:       ptr.To(false),
					TagPushEvents:             ptr.To(true),
					NoteEvents:                ptr.To(false),
					ConfidentialNoteEvents:    ptr.To(true),
					PipelineEvents:            ptr.To(false),
					WikiPageEvents:            ptr.To(true),
				},
				token: "123456:ABC",
			},
			want: &gitlab.SetTelegramServiceOptions{
				Token:                     ptr.To("123456:ABC"),
				Room:                      ptr.To("@channel"),
				NotifyOnlyBrokenPipelines: ptr.To(true),
				BranchesToBeNotified:      ptr.To("protected"),
				PushEvents:                ptr.To(true),
				IssuesEvents:              ptr.To(false),
				ConfidentialIssuesEvents:  ptr.To(true),
				MergeRequestsEvents:       ptr.To(false),
				TagPushEvents:             ptr.To(true),
				NoteEvents:                ptr.To(false),
				ConfidentialNoteEvents:    ptr.To(true),
				PipelineEvents:            ptr.To(false),
				WikiPageEvents:            ptr.To(true),
			},
		},
		"MinimalSpec": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel"},
				token:      "123456:ABC",
			},
			want: &gitlab.SetTelegramServiceOptions{
				Token: ptr.To("123456:ABC"),
				Room:  ptr.To("@channel"),
			},
		},
		"NilInput": {
			args: args{token: "123456:ABC"},
			want: &gitlab.SetTelegramServiceOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetTelegramServiceOptions(tc.args.parameters, tc.args.token)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetTelegramServiceOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIntegrationTelegramObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.TelegramService
		want    projectsv1alpha1.IntegrationTelegramObservation
	}{
		"FullObservation": {
			service: &gitlab.TelegramService{
				Service: gitlab.Service{
					ID:           9,
					Title:        "Telegram",
					Slug:         "telegram",
					Active:       true,
					NoteEvents:   true,
					IssuesEvents: true,
				},
				Properties: &gitlab.TelegramServiceProperties{
					Room:                      "@channel",
					NotifyOnlyBrokenPipelines: true,
					BranchesToBeNotified:      "all",
				},
			},
			want: projectsv1alpha1.IntegrationTelegramObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To[int64](9),
					Title:                          ptr.To("Telegram"),
					Slug:                           ptr.To("telegram"),
					Active:                         ptr.To(true),
					AlertEvents:                    ptr.To(false),
					CommitEvents:                   ptr.To(false),
					ConfidentialIssuesEvents:       ptr.To(false),
					ConfidentialNoteEvents:         ptr.To(false),
					DeploymentEvents:               ptr.To(false),
					GroupConfidentialMentionEvents: ptr.To(false),
					GroupMentionEvents:             ptr.To(false),
					IncidentEvents:                 ptr.To(false),
					IssuesEvents:                   ptr.To(true),
					JobEvents:                      ptr.To(false),
					MergeRequestsEvents:            ptr.To(false),
					NoteEvents:                     ptr.To(true),
					PipelineEvents:                 ptr.To(false),
					PushEvents:                     ptr.To(false),
					TagPushEvents:                  ptr.To(false),
					VulnerabilityEvents:            ptr.To(false),
					WikiPageEvents:                 ptr.To(false),
					CommentOnEventEnabled:          ptr.To(false),
					Inherited:                      ptr.To(false),
				},
				Room:                      "@channel",
				NotifyOnlyBrokenPipelines: true,
				BranchesToBeNotified:      "all",
			},
		},
		"NilService": {
			want: projectsv1alpha1.IntegrationTelegramObservation{},
		},
		"NilProperties": {
			service: &gitlab.TelegramService{Service: gitlab.Service{ID: 9}},
			want:    projectsv1alpha1.IntegrationTelegramObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationTelegramObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationTelegramObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationTelegramUpToDate(t *testing.T) {
	observed := func() *gitlab.TelegramService {
		return &gitlab.TelegramService{
			Service: gitlab.Service{PushEvents: true, PipelineEvents: true},
			Properties: &gitlab.TelegramServiceProperties{
				Room:                      "@channel",
				NotifyOnlyBrokenPipelines: true,
				BranchesToBeNotified:      "default",
			},
		}
	}

	cases := map[string]struct {
		spec        *projectsv1alpha1.IntegrationTelegramParameters
		observation *gitlab.TelegramService
		want        bool
	}{
		"UpToDate": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{
				Room:                      "@channel",
				NotifyOnlyBrokenPipelines: ptr.To(true),
				BranchesToBeNotified:      ptr.To("default"),
				PushEvents:                ptr.To(true),
				PipelineEvents:            ptr.To(true),
				IssuesEvents:              ptr.To(false),
			},
			observation: observed(),
			want:        true,
		},
		"OutOfDateRoom": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@other"},
			observation: observed(),
			want:        false,
		},
		"OutOfDateNotifyOnlyBrokenPipelines": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{
				Room:                      "@channel",
				NotifyOnlyBrokenPipelines: ptr.To(false),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateBranchesToBeNotified": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{
				Room:                 "@channel",
				BranchesToBeNotified: ptr.To("all"),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDatePipelineEvents": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{
				Room:           "@channel",
				PipelineEvents: ptr.To(false),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDatePushEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", PushEvents: ptr.To(false)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateIssuesEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", IssuesEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateConfidentialIssuesEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", ConfidentialIssuesEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateMergeRequestsEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", MergeRequestsEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateTagPushEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", TagPushEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateNoteEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", NoteEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateConfidentialNoteEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", ConfidentialNoteEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateWikiPageEvents": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel", WikiPageEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"NilObservation": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{},
			want: false,
		},
		"NilProperties": {
			spec:        &projectsv1alpha1.IntegrationTelegramParameters{},
			observation: &gitlab.TelegramService{},
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationTelegramUpToDate(tc.spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationTelegramUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLateInitializeIntegrationTelegram(t *testing.T) {
	cases := map[string]struct {
		spec *projectsv1alpha1.IntegrationTelegramParameters
		svc  *gitlab.TelegramService
		want *projectsv1alpha1.IntegrationTelegramParameters
	}{
		"FillsBranchesOnly": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel"},
			svc: &gitlab.TelegramService{
				Service: gitlab.Service{PushEvents: true, WikiPageEvents: true},
				Properties: &gitlab.TelegramServiceProperties{
					Room:                      "@channel",
					NotifyOnlyBrokenPipelines: true,
					BranchesToBeNotified:      "default",
				},
			},
			want: &projectsv1alpha1.IntegrationTelegramParameters{
				Room:                 "@channel",
				BranchesToBeNotified: ptr.To("default"),
			},
		},
		"NullBoolsNotInitialized": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{Room: "@channel"},
			svc: decodeTelegram(t, `{"push_events":null,"issues_events":null,"confidential_issues_events":null,"merge_requests_events":null,"tag_push_events":null,"note_events":null,"confidential_note_events":null,"pipeline_events":null,"wiki_page_events":null,`+
				`"properties":{"room":"@channel","notify_only_broken_pipelines":null,"branches_to_be_notified":"default"}}`),
			want: &projectsv1alpha1.IntegrationTelegramParameters{
				Room:                 "@channel",
				BranchesToBeNotified: ptr.To("default"),
			},
		},
		"KeepsExistingValues": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{
				NotifyOnlyBrokenPipelines: ptr.To(false),
				BranchesToBeNotified:      ptr.To("all"),
			},
			svc: &gitlab.TelegramService{
				Properties: &gitlab.TelegramServiceProperties{
					NotifyOnlyBrokenPipelines: true,
					BranchesToBeNotified:      "default",
				},
			},
			want: &projectsv1alpha1.IntegrationTelegramParameters{
				NotifyOnlyBrokenPipelines: ptr.To(false),
				BranchesToBeNotified:      ptr.To("all"),
			},
		},
		"EmptyBranchesNotInitialized": {
			spec: &projectsv1alpha1.IntegrationTelegramParameters{},
			svc: &gitlab.TelegramService{
				Properties: &gitlab.TelegramServiceProperties{},
			},
			want: &projectsv1alpha1.IntegrationTelegramParameters{},
		},
		"NilSpecIsNoop": {
			svc:  &gitlab.TelegramService{Properties: &gitlab.TelegramServiceProperties{}},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeIntegrationTelegram(tc.spec, tc.svc)
			if diff := cmp.Diff(tc.want, tc.spec); diff != "" {
				t.Errorf("LateInitializeIntegrationTelegram(): -want, +got:\n%s", diff)
			}
		})
	}
}

func decodeTelegram(t *testing.T, raw string) *gitlab.TelegramService {
	t.Helper()
	svc := &gitlab.TelegramService{}
	if err := json.Unmarshal([]byte(raw), svc); err != nil {
		t.Fatalf("json.Unmarshal(): %v", err)
	}
	return svc
}

func TestIntegrationTelegramObservationJSONKeepsFalse(t *testing.T) {
	obs := GenerateIntegrationTelegramObservation(&gitlab.TelegramService{
		Properties: &gitlab.TelegramServiceProperties{NotifyOnlyBrokenPipelines: false},
	})
	raw, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("json.Marshal(): %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal(): %v", err)
	}
	if diff := cmp.Diff("false", string(got["notifyOnlyBrokenPipelines"])); diff != "" {
		t.Errorf("notifyOnlyBrokenPipelines: -want, +got:\n%s", diff)
	}
}
