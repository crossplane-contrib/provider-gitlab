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

const (
	testEmailsOnPushRecipients        = "dev@example.com ops@example.com"
	testEmailsOnPushIntegrationID     = int64(999)
	testEmailsOnPushBranchesProtected = "protected"
)

func TestGenerateSetEmailsOnPushServiceOptions(t *testing.T) {
	cases := map[string]struct {
		parameters *projectsv1alpha1.IntegrationEmailsOnPushParameters
		want       *gitlab.SetEmailsOnPushServiceOptions
	}{
		"AllFieldsSet": {
			parameters: &projectsv1alpha1.IntegrationEmailsOnPushParameters{
				ProjectID:              ptr.To(int64(123)),
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           ptr.To(true),
				SendFromCommitterEmail: ptr.To(false),
				PushEvents:             ptr.To(false),
				TagPushEvents:          ptr.To(true),
				BranchesToBeNotified:   ptr.To(testEmailsOnPushBranchesProtected),
			},
			want: &gitlab.SetEmailsOnPushServiceOptions{
				Recipients:             ptr.To(testEmailsOnPushRecipients),
				DisableDiffs:           ptr.To(true),
				SendFromCommitterEmail: ptr.To(false),
				PushEvents:             ptr.To(false),
				TagPushEvents:          ptr.To(true),
				BranchesToBeNotified:   ptr.To(testEmailsOnPushBranchesProtected),
			},
		},
		"OnlyRecipients": {
			parameters: &projectsv1alpha1.IntegrationEmailsOnPushParameters{
				Recipients: testEmailsOnPushRecipients,
			},
			want: &gitlab.SetEmailsOnPushServiceOptions{
				Recipients: ptr.To(testEmailsOnPushRecipients),
			},
		},
		"NilInput": {
			parameters: nil,
			want:       &gitlab.SetEmailsOnPushServiceOptions{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateSetEmailsOnPushServiceOptions(tc.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateSetEmailsOnPushServiceOptions(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIntegrationEmailsOnPushObservation(t *testing.T) {
	cases := map[string]struct {
		service *gitlab.EmailsOnPushService
		want    projectsv1alpha1.IntegrationEmailsOnPushObservation
	}{
		"FullObservation": {
			service: &gitlab.EmailsOnPushService{
				Service: gitlab.Service{
					ID:            testEmailsOnPushIntegrationID,
					Title:         "Emails on push",
					Slug:          "emails-on-push",
					Active:        true,
					PushEvents:    true,
					TagPushEvents: true,
				},
				Properties: &gitlab.EmailsOnPushServiceProperties{
					Recipients:             testEmailsOnPushRecipients,
					DisableDiffs:           true,
					SendFromCommitterEmail: true,
					BranchesToBeNotified:   testEmailsOnPushBranchesProtected,
				},
			},
			want: projectsv1alpha1.IntegrationEmailsOnPushObservation{
				CommonIntegrationObservation: v1alpha1.CommonIntegrationObservation{
					ID:                             ptr.To(testEmailsOnPushIntegrationID),
					Title:                          ptr.To("Emails on push"),
					Slug:                           ptr.To("emails-on-push"),
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
					TagPushEvents:                  ptr.To(true),
					VulnerabilityEvents:            ptr.To(false),
					WikiPageEvents:                 ptr.To(false),
					CommentOnEventEnabled:          ptr.To(false),
					Inherited:                      ptr.To(false),
				},
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           true,
				SendFromCommitterEmail: true,
				BranchesToBeNotified:   testEmailsOnPushBranchesProtected,
			},
		},
		"NilService": {
			service: nil,
			want:    projectsv1alpha1.IntegrationEmailsOnPushObservation{},
		},
		"NilProperties": {
			service: &gitlab.EmailsOnPushService{
				Service: gitlab.Service{ID: testEmailsOnPushIntegrationID},
			},
			want: projectsv1alpha1.IntegrationEmailsOnPushObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateIntegrationEmailsOnPushObservation(tc.service)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateIntegrationEmailsOnPushObservation(): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIntegrationEmailsOnPushObservationJSONKeepsFalse(t *testing.T) {
	type falseFields struct {
		DisableDiffs           *bool `json:"disableDiffs"`
		SendFromCommitterEmail *bool `json:"sendFromCommitterEmail"`
	}

	observation := GenerateIntegrationEmailsOnPushObservation(&gitlab.EmailsOnPushService{
		Properties: &gitlab.EmailsOnPushServiceProperties{Recipients: testEmailsOnPushRecipients},
	})
	raw, err := json.Marshal(observation)
	if err != nil {
		t.Fatalf("marshal observation: %v", err)
	}

	got := falseFields{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal observation %q: %v", string(raw), err)
	}
	want := falseFields{DisableDiffs: ptr.To(false), SendFromCommitterEmail: ptr.To(false)}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("observation JSON %s: -want, +got:\n%s", string(raw), diff)
	}
}

func TestIsIntegrationEmailsOnPushUpToDate(t *testing.T) {
	observed := func() *gitlab.EmailsOnPushService {
		return &gitlab.EmailsOnPushService{
			Service: gitlab.Service{PushEvents: true, TagPushEvents: false},
			Properties: &gitlab.EmailsOnPushServiceProperties{
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           false,
				SendFromCommitterEmail: true,
				BranchesToBeNotified:   "all",
			},
		}
	}
	fullSpec := func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
		return &projectsv1alpha1.IntegrationEmailsOnPushParameters{
			Recipients:             testEmailsOnPushRecipients,
			DisableDiffs:           ptr.To(false),
			SendFromCommitterEmail: ptr.To(true),
			PushEvents:             ptr.To(true),
			TagPushEvents:          ptr.To(false),
			BranchesToBeNotified:   ptr.To("all"),
		}
	}

	cases := map[string]struct {
		spec        *projectsv1alpha1.IntegrationEmailsOnPushParameters
		observation *gitlab.EmailsOnPushService
		want        bool
	}{
		"UpToDate": {
			spec:        fullSpec(),
			observation: observed(),
			want:        true,
		},
		"UnsetOptionalFieldsAreIgnored": {
			spec:        &projectsv1alpha1.IntegrationEmailsOnPushParameters{Recipients: testEmailsOnPushRecipients},
			observation: observed(),
			want:        true,
		},
		"OutOfDateRecipients": {
			spec: func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
				s := fullSpec()
				s.Recipients = "other@example.com"
				return s
			}(),
			observation: observed(),
			want:        false,
		},
		"OutOfDateDisableDiffs": {
			spec: func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
				s := fullSpec()
				s.DisableDiffs = ptr.To(true)
				return s
			}(),
			observation: observed(),
			want:        false,
		},
		"OutOfDateSendFromCommitterEmail": {
			spec: func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
				s := fullSpec()
				s.SendFromCommitterEmail = ptr.To(false)
				return s
			}(),
			observation: observed(),
			want:        false,
		},
		"OutOfDatePushEvents": {
			spec: func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
				s := fullSpec()
				s.PushEvents = ptr.To(false)
				return s
			}(),
			observation: observed(),
			want:        false,
		},
		"OutOfDateTagPushEvents": {
			spec: func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
				s := fullSpec()
				s.TagPushEvents = ptr.To(true)
				return s
			}(),
			observation: observed(),
			want:        false,
		},
		"OutOfDateBranchesToBeNotified": {
			spec: func() *projectsv1alpha1.IntegrationEmailsOnPushParameters {
				s := fullSpec()
				s.BranchesToBeNotified = ptr.To(testEmailsOnPushBranchesProtected)
				return s
			}(),
			observation: observed(),
			want:        false,
		},
		"NilObservation": {
			spec:        fullSpec(),
			observation: nil,
			want:        false,
		},
		"NilProperties": {
			spec:        fullSpec(),
			observation: &gitlab.EmailsOnPushService{},
			want:        false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIntegrationEmailsOnPushUpToDate(tc.spec, tc.observation)
			if got != tc.want {
				t.Errorf("IsIntegrationEmailsOnPushUpToDate(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLateInitializeIntegrationEmailsOnPush(t *testing.T) {
	remote := &gitlab.EmailsOnPushService{
		Service: gitlab.Service{PushEvents: true, TagPushEvents: true},
		Properties: &gitlab.EmailsOnPushServiceProperties{
			Recipients:             "remote@example.com",
			DisableDiffs:           true,
			SendFromCommitterEmail: true,
			BranchesToBeNotified:   "all",
		},
	}

	cases := map[string]struct {
		spec *projectsv1alpha1.IntegrationEmailsOnPushParameters
		svc  *gitlab.EmailsOnPushService
		want *projectsv1alpha1.IntegrationEmailsOnPushParameters
	}{
		"FillsUnsetFields": {
			spec: &projectsv1alpha1.IntegrationEmailsOnPushParameters{Recipients: testEmailsOnPushRecipients},
			svc:  remote,
			want: &projectsv1alpha1.IntegrationEmailsOnPushParameters{
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           ptr.To(true),
				SendFromCommitterEmail: ptr.To(true),
				PushEvents:             ptr.To(true),
				TagPushEvents:          ptr.To(true),
				BranchesToBeNotified:   ptr.To("all"),
			},
		},
		"KeepsSetFields": {
			spec: &projectsv1alpha1.IntegrationEmailsOnPushParameters{
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           ptr.To(false),
				SendFromCommitterEmail: ptr.To(false),
				PushEvents:             ptr.To(false),
				TagPushEvents:          ptr.To(false),
				BranchesToBeNotified:   ptr.To(testEmailsOnPushBranchesProtected),
			},
			svc: remote,
			want: &projectsv1alpha1.IntegrationEmailsOnPushParameters{
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           ptr.To(false),
				SendFromCommitterEmail: ptr.To(false),
				PushEvents:             ptr.To(false),
				TagPushEvents:          ptr.To(false),
				BranchesToBeNotified:   ptr.To(testEmailsOnPushBranchesProtected),
			},
		},
		"EmptyRemoteBranchesStaysUnset": {
			spec: &projectsv1alpha1.IntegrationEmailsOnPushParameters{Recipients: testEmailsOnPushRecipients},
			svc: &gitlab.EmailsOnPushService{
				Properties: &gitlab.EmailsOnPushServiceProperties{},
			},
			want: &projectsv1alpha1.IntegrationEmailsOnPushParameters{
				Recipients:             testEmailsOnPushRecipients,
				DisableDiffs:           ptr.To(false),
				SendFromCommitterEmail: ptr.To(false),
				PushEvents:             ptr.To(false),
				TagPushEvents:          ptr.To(false),
			},
		},
		"NilSpecIsNoop": {
			spec: nil,
			svc:  remote,
			want: nil,
		},
		"NilPropertiesIsNoop": {
			spec: &projectsv1alpha1.IntegrationEmailsOnPushParameters{Recipients: testEmailsOnPushRecipients},
			svc:  &gitlab.EmailsOnPushService{},
			want: &projectsv1alpha1.IntegrationEmailsOnPushParameters{Recipients: testEmailsOnPushRecipients},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeIntegrationEmailsOnPush(tc.spec, tc.svc)
			if diff := cmp.Diff(tc.want, tc.spec); diff != "" {
				t.Errorf("LateInitializeIntegrationEmailsOnPush(): -want, +got:\n%s", diff)
			}
		})
	}
}
