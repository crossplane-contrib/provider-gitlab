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
	"context"
	"testing"
	"time"

	v1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/crossplane-contrib/provider-gitlab/apis/shared/projects/v1alpha1"
)

var (
	url                      = "https://my-project.example.com"
	confidentialNoteEvents   = true
	pushEvents               = true
	pushEventsBranchFilter   = "foo"
	issuesEvents             = true
	confidentialIssuesEvents = true
	mergeRequestsEvents      = true
	tagPushEvents            = true
	noteEvents               = true
	jobEvents                = true
	pipelineEvents           = true
	wikiPageEvents           = true
	enableSSLVerification    = true

	tokenValue = "84B9C651-9025-47D2-9124-DD951BD268E8"
	tokenRef   = &v1.SecretKeySelector{
		Key: "token", SecretReference: v1.SecretReference{Name: "test", Namespace: "test"},
	}
)

func TestGenerateHookObservation(t *testing.T) {
	id := 0
	createdAt := time.Now()

	type args struct {
		ph *gitlab.ProjectHook
	}

	cases := map[string]struct {
		args args
		want v1alpha1.HookObservation
	}{
		"Full": {
			args: args{
				ph: &gitlab.ProjectHook{
					ID:        id,
					CreatedAt: &createdAt,
				},
			},
			want: v1alpha1.HookObservation{
				ID:        id,
				CreatedAt: &metav1.Time{Time: createdAt},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateHookObservation(tc.args.ph)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestLateInitializeHook(t *testing.T) {
	cases := map[string]struct {
		parameters  *v1alpha1.HookParameters
		projecthook *gitlab.ProjectHook
		want        *v1alpha1.HookParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.HookParameters{},
			projecthook: &gitlab.ProjectHook{
				ConfidentialNoteEvents:   confidentialNoteEvents,
				PushEvents:               pushEvents,
				PushEventsBranchFilter:   pushEventsBranchFilter,
				IssuesEvents:             issuesEvents,
				ConfidentialIssuesEvents: confidentialIssuesEvents,
				MergeRequestsEvents:      mergeRequestsEvents,
				TagPushEvents:            tagPushEvents,
				NoteEvents:               noteEvents,
				JobEvents:                jobEvents,
				PipelineEvents:           pipelineEvents,
				WikiPageEvents:           wikiPageEvents,
				EnableSSLVerification:    enableSSLVerification,
			},
			want: &v1alpha1.HookParameters{
				ConfidentialNoteEvents:   &confidentialNoteEvents,
				PushEvents:               &pushEvents,
				PushEventsBranchFilter:   &pushEventsBranchFilter,
				IssuesEvents:             &issuesEvents,
				ConfidentialIssuesEvents: &confidentialIssuesEvents,
				MergeRequestsEvents:      &mergeRequestsEvents,
				TagPushEvents:            &tagPushEvents,
				NoteEvents:               &noteEvents,
				JobEvents:                &jobEvents,
				PipelineEvents:           &pipelineEvents,
				WikiPageEvents:           &wikiPageEvents,
				EnableSSLVerification:    &enableSSLVerification,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeHook(tc.parameters, tc.projecthook)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestGenerateCreateHookOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.HookParameters
		secret     *corev1.Secret
	}
	type want struct {
		addProjectHookOptions *gitlab.AddProjectHookOptions
		err                   error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.HookParameters{
					URL:                      &url,
					ConfidentialNoteEvents:   &confidentialNoteEvents,
					PushEvents:               &pushEvents,
					PushEventsBranchFilter:   &pushEventsBranchFilter,
					IssuesEvents:             &issuesEvents,
					ConfidentialIssuesEvents: &confidentialIssuesEvents,
					MergeRequestsEvents:      &mergeRequestsEvents,
					TagPushEvents:            &tagPushEvents,
					NoteEvents:               &noteEvents,
					JobEvents:                &jobEvents,
					PipelineEvents:           &pipelineEvents,
					WikiPageEvents:           &wikiPageEvents,
					EnableSSLVerification:    &enableSSLVerification,
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "test"},
					Data: map[string][]byte{
						"token": []byte(tokenValue),
					},
				},
			},
			want: want{
				err: nil,
				addProjectHookOptions: &gitlab.AddProjectHookOptions{
					URL:                      &url,
					ConfidentialNoteEvents:   &confidentialNoteEvents,
					PushEvents:               &pushEvents,
					PushEventsBranchFilter:   &pushEventsBranchFilter,
					IssuesEvents:             &issuesEvents,
					ConfidentialIssuesEvents: &confidentialIssuesEvents,
					MergeRequestsEvents:      &mergeRequestsEvents,
					TagPushEvents:            &tagPushEvents,
					NoteEvents:               &noteEvents,
					JobEvents:                &jobEvents,
					PipelineEvents:           &pipelineEvents,
					WikiPageEvents:           &wikiPageEvents,
					EnableSSLVerification:    &enableSSLVerification,
					Token:                    &tokenValue,
				},
			},
		},
		"SomeFields": {
			args: args{
				parameters: &v1alpha1.HookParameters{
					PushEvents:             &pushEvents,
					PushEventsBranchFilter: &pushEventsBranchFilter,
					IssuesEvents:           &issuesEvents,
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "test"},
					Data: map[string][]byte{
						"token": []byte(tokenValue),
					},
				},
			},
			want: want{
				err: nil,
				addProjectHookOptions: &gitlab.AddProjectHookOptions{
					PushEvents:             &pushEvents,
					PushEventsBranchFilter: &pushEventsBranchFilter,
					IssuesEvents:           &issuesEvents,
					Token:                  &tokenValue,
				},
			},
		},
		"FailNoSecret": {
			args: args{
				parameters: &v1alpha1.HookParameters{
					PushEvents:             &pushEvents,
					PushEventsBranchFilter: &pushEventsBranchFilter,
					IssuesEvents:           &issuesEvents,
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: "other", Name: "other"},
					Data: map[string][]byte{
						"token": []byte(tokenValue),
					},
				},
			},
			want: want{
				err:                   errors.New(`Cannot get referenced Secret: secrets "test" not found`),
				addProjectHookOptions: nil,
			},
		},
		"FailWrongKey": {
			args: args{
				parameters: &v1alpha1.HookParameters{
					PushEvents:             &pushEvents,
					PushEventsBranchFilter: &pushEventsBranchFilter,
					IssuesEvents:           &issuesEvents,
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "test"},
					Data: map[string][]byte{
						"wrongKey": []byte(tokenValue),
					},
				},
			},
			want: want{
				err:                   errors.New("Could not find key token in the referenced secret"),
				addProjectHookOptions: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			client := fake.NewClientBuilder().WithObjects(tc.args.secret).Build()

			got, err := GenerateCreateHookOptions(tc.args.parameters, tokenRef, client, context.Background())
			if err != nil && tc.want.err != nil {
				if diff := cmp.Diff(tc.want.err.Error(), err.Error(), test.EquateErrors()); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.addProjectHookOptions, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestGenerateEditHookOptions(t *testing.T) {
	type args struct {
		parameters *v1alpha1.HookParameters
	}
	cases := map[string]struct {
		args args
		want *gitlab.EditProjectHookOptions
	}{
		"AllFields": {
			args: args{
				parameters: &v1alpha1.HookParameters{
					URL:                      &url,
					ConfidentialNoteEvents:   &confidentialNoteEvents,
					PushEvents:               &pushEvents,
					PushEventsBranchFilter:   &pushEventsBranchFilter,
					IssuesEvents:             &issuesEvents,
					ConfidentialIssuesEvents: &confidentialIssuesEvents,
					MergeRequestsEvents:      &mergeRequestsEvents,
					TagPushEvents:            &tagPushEvents,
					NoteEvents:               &noteEvents,
					JobEvents:                &jobEvents,
					PipelineEvents:           &pipelineEvents,
					WikiPageEvents:           &wikiPageEvents,
					EnableSSLVerification:    &enableSSLVerification,
				},
			},
			want: &gitlab.EditProjectHookOptions{
				URL:                      &url,
				ConfidentialNoteEvents:   &confidentialNoteEvents,
				PushEvents:               &pushEvents,
				PushEventsBranchFilter:   &pushEventsBranchFilter,
				IssuesEvents:             &issuesEvents,
				ConfidentialIssuesEvents: &confidentialIssuesEvents,
				MergeRequestsEvents:      &mergeRequestsEvents,
				TagPushEvents:            &tagPushEvents,
				NoteEvents:               &noteEvents,
				JobEvents:                &jobEvents,
				PipelineEvents:           &pipelineEvents,
				WikiPageEvents:           &wikiPageEvents,
				EnableSSLVerification:    &enableSSLVerification,
				Token:                    &tokenValue,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "test"},
				Data: map[string][]byte{
					"token": []byte(tokenValue),
				},
			}
			client := fake.NewClientBuilder().WithObjects(secret).Build()
			got, _ := GenerateEditHookOptions(tc.args.parameters, tokenRef, client, context.Background())

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestIsHookUpToDate(t *testing.T) {
	type args struct {
		projecthook *gitlab.ProjectHook
		p           *v1alpha1.HookParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: &v1alpha1.HookParameters{
					URL:                      &url,
					ConfidentialNoteEvents:   &confidentialNoteEvents,
					PushEvents:               &pushEvents,
					PushEventsBranchFilter:   &pushEventsBranchFilter,
					IssuesEvents:             &issuesEvents,
					ConfidentialIssuesEvents: &confidentialIssuesEvents,
					MergeRequestsEvents:      &mergeRequestsEvents,
					TagPushEvents:            &tagPushEvents,
					NoteEvents:               &noteEvents,
					JobEvents:                &jobEvents,
					PipelineEvents:           &pipelineEvents,
					WikiPageEvents:           &wikiPageEvents,
					EnableSSLVerification:    &enableSSLVerification,
				},
				projecthook: &gitlab.ProjectHook{
					URL:                      url,
					ConfidentialNoteEvents:   confidentialNoteEvents,
					PushEvents:               pushEvents,
					PushEventsBranchFilter:   pushEventsBranchFilter,
					IssuesEvents:             issuesEvents,
					ConfidentialIssuesEvents: confidentialIssuesEvents,
					MergeRequestsEvents:      mergeRequestsEvents,
					TagPushEvents:            tagPushEvents,
					NoteEvents:               noteEvents,
					JobEvents:                jobEvents,
					PipelineEvents:           pipelineEvents,
					WikiPageEvents:           wikiPageEvents,
					EnableSSLVerification:    enableSSLVerification,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: &v1alpha1.HookParameters{
					URL:                      &url,
					ConfidentialNoteEvents:   &confidentialNoteEvents,
					PushEvents:               &pushEvents,
					PushEventsBranchFilter:   &pushEventsBranchFilter,
					IssuesEvents:             &issuesEvents,
					ConfidentialIssuesEvents: &confidentialIssuesEvents,
					MergeRequestsEvents:      &mergeRequestsEvents,
					TagPushEvents:            &tagPushEvents,
					NoteEvents:               &noteEvents,
					JobEvents:                &jobEvents,
					PipelineEvents:           &pipelineEvents,
					WikiPageEvents:           &wikiPageEvents,
					EnableSSLVerification:    &enableSSLVerification,
				},
				projecthook: &gitlab.ProjectHook{
					URL:                      "http://some.other.url",
					ConfidentialNoteEvents:   false,
					PushEvents:               false,
					PushEventsBranchFilter:   "bar",
					IssuesEvents:             false,
					ConfidentialIssuesEvents: false,
					MergeRequestsEvents:      false,
					TagPushEvents:            false,
					NoteEvents:               false,
					JobEvents:                false,
					PipelineEvents:           false,
					WikiPageEvents:           false,
					EnableSSLVerification:    false,
				},
			},
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsHookUpToDate(tc.args.p, tc.args.projecthook)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
