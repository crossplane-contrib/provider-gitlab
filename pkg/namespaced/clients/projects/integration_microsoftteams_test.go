package projects

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/apis/common/v1alpha1"
	projectsv1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestGenerateSetMicrosoftTeamsServiceOptions(t *testing.T) {
	type args struct {
		parameters *projectsv1alpha1.IntegrationMicrosoftTeamsParameters
		webhook    string
	}
	cases := map[string]struct {
		args args
		want *gitlab.SetMicrosoftTeamsServiceOptions
	}{
		"AllFieldsSet": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
					ProjectID:                 ptr.To[int64](123),
					NotifyOnlyBrokenPipelines: ptr.To(false),
					BranchesToBeNotified:      ptr.To("default_and_protected"),
					PushEvents:                ptr.To(false),
					IssuesEvents:              ptr.To(true),
					ConfidentialIssuesEvents:  ptr.To(false),
					MergeRequestsEvents:       ptr.To(true),
					TagPushEvents:             ptr.To(false),
					NoteEvents:                ptr.To(true),
					ConfidentialNoteEvents:    ptr.To(false),
					PipelineEvents:            ptr.To(true),
					WikiPageEvents:            ptr.To(false),
				},
				webhook: "https://outlook.office.com/webhook/abc",
			},
			want: &gitlab.SetMicrosoftTeamsServiceOptions{
				WebHook:                   ptr.To("https://outlook.office.com/webhook/abc"),
				NotifyOnlyBrokenPipelines: ptr.To(false),
				BranchesToBeNotified:      ptr.To("default_and_protected"),
				PushEvents:                ptr.To(false),
				IssuesEvents:              ptr.To(true),
				ConfidentialIssuesEvents:  ptr.To(false),
				MergeRequestsEvents:       ptr.To(true),
				TagPushEvents:             ptr.To(false),
				NoteEvents:                ptr.To(true),
				ConfidentialNoteEvents:    ptr.To(false),
				PipelineEvents:            ptr.To(true),
				WikiPageEvents:            ptr.To(false),
			},
		},
		"MinimalSpec": {
			args: args{
				parameters: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{},
				webhook:    "https://outlook.office.com/webhook/abc",
			},
			want: &gitlab.SetMicrosoftTeamsServiceOptions{
				WebHook: ptr.To("https://outlook.office.com/webhook/abc"),
			},
		},
		"NilInput": {
			args: args{webhook: "https://outlook.office.com/webhook/abc"},
			want: &gitlab.SetMicrosoftTeamsServiceOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetMicrosoftTeamsServiceOptions(tc.args.parameters, tc.args.webhook)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetMicrosoftTeamsServiceOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIntegrationMicrosoftTeamsObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.MicrosoftTeamsService
		want    projectsv1alpha1.IntegrationMicrosoftTeamsObservation
	}{
		"FullObservation": {
			service: &gitlab.MicrosoftTeamsService{
				Service: gitlab.Service{
					ID:             8,
					Title:          "Microsoft Teams notifications",
					Slug:           "microsoft-teams",
					Active:         true,
					PipelineEvents: true,
				},
				Properties: &gitlab.MicrosoftTeamsServiceProperties{
					NotifyOnlyBrokenPipelines: true,
					BranchesToBeNotified:      "protected",
				},
			},
			want: projectsv1alpha1.IntegrationMicrosoftTeamsObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To[int64](8),
					Title:                          ptr.To("Microsoft Teams notifications"),
					Slug:                           ptr.To("microsoft-teams"),
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
					PipelineEvents:                 ptr.To(true),
					PushEvents:                     ptr.To(false),
					TagPushEvents:                  ptr.To(false),
					VulnerabilityEvents:            ptr.To(false),
					WikiPageEvents:                 ptr.To(false),
					CommentOnEventEnabled:          ptr.To(false),
					Inherited:                      ptr.To(false),
				},
				NotifyOnlyBrokenPipelines: true,
				BranchesToBeNotified:      "protected",
			},
		},
		"NilService": {
			want: projectsv1alpha1.IntegrationMicrosoftTeamsObservation{},
		},
		"NilProperties": {
			service: &gitlab.MicrosoftTeamsService{Service: gitlab.Service{ID: 8}},
			want:    projectsv1alpha1.IntegrationMicrosoftTeamsObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationMicrosoftTeamsObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationMicrosoftTeamsObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsIntegrationMicrosoftTeamsUpToDate(t *testing.T) {
	observed := func() *gitlab.MicrosoftTeamsService {
		return &gitlab.MicrosoftTeamsService{
			Service: gitlab.Service{IssuesEvents: true, NoteEvents: false},
			Properties: &gitlab.MicrosoftTeamsServiceProperties{
				NotifyOnlyBrokenPipelines: true,
				BranchesToBeNotified:      "default",
				IssuesEvents:              false,
			},
		}
	}

	cases := map[string]struct {
		spec        *projectsv1alpha1.IntegrationMicrosoftTeamsParameters
		observation *gitlab.MicrosoftTeamsService
		want        bool
	}{
		"UpToDate": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(true),
				BranchesToBeNotified:      ptr.To("default"),
				IssuesEvents:              ptr.To(true),
				NoteEvents:                ptr.To(false),
			},
			observation: observed(),
			want:        true,
		},
		"OutOfDateNotifyOnlyBrokenPipelines": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(false),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateBranchesToBeNotified": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				BranchesToBeNotified: ptr.To("all"),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDateTopLevelIssuesEvents": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				IssuesEvents: ptr.To(false),
			},
			observation: observed(),
			want:        false,
		},
		"OutOfDatePushEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{PushEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateConfidentialIssuesEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{ConfidentialIssuesEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateMergeRequestsEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{MergeRequestsEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateTagPushEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{TagPushEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateNoteEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{NoteEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateConfidentialNoteEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{ConfidentialNoteEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDatePipelineEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{PipelineEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"OutOfDateWikiPageEvents": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{WikiPageEvents: ptr.To(true)},
			observation: observed(),
			want:        false,
		},
		"NilObservation": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{},
			want: false,
		},
		"NilProperties": {
			spec:        &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{},
			observation: &gitlab.MicrosoftTeamsService{},
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationMicrosoftTeamsUpToDate(tc.spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationMicrosoftTeamsUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLateInitializeIntegrationMicrosoftTeams(t *testing.T) {
	cases := map[string]struct {
		spec *projectsv1alpha1.IntegrationMicrosoftTeamsParameters
		svc  *gitlab.MicrosoftTeamsService
		want *projectsv1alpha1.IntegrationMicrosoftTeamsParameters
	}{
		"FillsNilFields": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{},
			svc: &gitlab.MicrosoftTeamsService{
				Service: gitlab.Service{PushEvents: true, ConfidentialNoteEvents: true},
				Properties: &gitlab.MicrosoftTeamsServiceProperties{
					NotifyOnlyBrokenPipelines: true,
					BranchesToBeNotified:      "default",
				},
			},
			want: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(true),
				BranchesToBeNotified:      ptr.To("default"),
				PushEvents:                ptr.To(true),
				IssuesEvents:              ptr.To(false),
				ConfidentialIssuesEvents:  ptr.To(false),
				MergeRequestsEvents:       ptr.To(false),
				TagPushEvents:             ptr.To(false),
				NoteEvents:                ptr.To(false),
				ConfidentialNoteEvents:    ptr.To(true),
				PipelineEvents:            ptr.To(false),
				WikiPageEvents:            ptr.To(false),
			},
		},
		"KeepsExistingValues": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(false),
				BranchesToBeNotified:      ptr.To("all"),
				PushEvents:                ptr.To(false),
				IssuesEvents:              ptr.To(true),
				ConfidentialIssuesEvents:  ptr.To(true),
				MergeRequestsEvents:       ptr.To(true),
				TagPushEvents:             ptr.To(true),
				NoteEvents:                ptr.To(true),
				ConfidentialNoteEvents:    ptr.To(false),
				PipelineEvents:            ptr.To(true),
				WikiPageEvents:            ptr.To(true),
			},
			svc: &gitlab.MicrosoftTeamsService{
				Service: gitlab.Service{PushEvents: true, ConfidentialNoteEvents: true},
				Properties: &gitlab.MicrosoftTeamsServiceProperties{
					NotifyOnlyBrokenPipelines: true,
				},
			},
			want: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(false),
				BranchesToBeNotified:      ptr.To("all"),
				PushEvents:                ptr.To(false),
				IssuesEvents:              ptr.To(true),
				ConfidentialIssuesEvents:  ptr.To(true),
				MergeRequestsEvents:       ptr.To(true),
				TagPushEvents:             ptr.To(true),
				NoteEvents:                ptr.To(true),
				ConfidentialNoteEvents:    ptr.To(false),
				PipelineEvents:            ptr.To(true),
				WikiPageEvents:            ptr.To(true),
			},
		},
		"EmptyBranchesNotInitialized": {
			spec: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(true),
				PushEvents:                ptr.To(true),
				IssuesEvents:              ptr.To(true),
				ConfidentialIssuesEvents:  ptr.To(true),
				MergeRequestsEvents:       ptr.To(true),
				TagPushEvents:             ptr.To(true),
				NoteEvents:                ptr.To(true),
				ConfidentialNoteEvents:    ptr.To(true),
				PipelineEvents:            ptr.To(true),
				WikiPageEvents:            ptr.To(true),
			},
			svc: &gitlab.MicrosoftTeamsService{
				Properties: &gitlab.MicrosoftTeamsServiceProperties{},
			},
			want: &projectsv1alpha1.IntegrationMicrosoftTeamsParameters{
				NotifyOnlyBrokenPipelines: ptr.To(true),
				PushEvents:                ptr.To(true),
				IssuesEvents:              ptr.To(true),
				ConfidentialIssuesEvents:  ptr.To(true),
				MergeRequestsEvents:       ptr.To(true),
				TagPushEvents:             ptr.To(true),
				NoteEvents:                ptr.To(true),
				ConfidentialNoteEvents:    ptr.To(true),
				PipelineEvents:            ptr.To(true),
				WikiPageEvents:            ptr.To(true),
			},
		},
		"NilSpecIsNoop": {
			svc:  &gitlab.MicrosoftTeamsService{Properties: &gitlab.MicrosoftTeamsServiceProperties{}},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeIntegrationMicrosoftTeams(tc.spec, tc.svc)
			if diff := cmp.Diff(tc.want, tc.spec); diff != "" {
				t.Errorf("LateInitializeIntegrationMicrosoftTeams(): -want, +got:\n%s", diff)
			}
		})
	}
}
