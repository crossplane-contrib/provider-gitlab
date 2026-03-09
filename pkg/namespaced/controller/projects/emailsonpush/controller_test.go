package emailsonpush

import (
	"context"
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

var (
	errBoom   = errors.New("boom")
	projectID = int64(123)
)

type mockClient struct {
	get func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error)
	set func(pid interface{}, opt *gitlab.SetEmailsOnPushServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error)
	del func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

func (m *mockClient) GetEmailsOnPushService(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
	if m.get != nil {
		return m.get(pid, options...)
	}
	return nil, nil, nil
}

func (m *mockClient) SetEmailsOnPushService(pid interface{}, opt *gitlab.SetEmailsOnPushServiceOptions, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
	if m.set != nil {
		return m.set(pid, opt, options...)
	}
	return nil, nil, nil
}

func (m *mockClient) DeleteEmailsOnPushService(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	if m.del != nil {
		return m.del(pid, options...)
	}
	return nil, nil
}

func emailOnPush() *v1alpha1.EmailsOnPush {
	return &v1alpha1.EmailsOnPush{}
}

type args struct {
	client *mockClient
	cr     resource.Managed
}

func TestObserve(t *testing.T) {
	type want struct {
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InvalidType": {
			args: args{
				cr: &v1alpha1.Project{},
			},
			want: want{
				err: errors.New(errNotEmailsOnPush),
			},
		},
		"MissingProjectID": {
			args: args{
				cr: emailOnPush(),
			},
			want: want{
				err: errors.New(errProjectIDMissing),
			},
		},
		"GetError": {
			args: args{
				client: &mockClient{
					get: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
						return nil, nil, errBoom
					},
				},
				cr: func() resource.Managed {
					cr := emailOnPush()
					cr.Spec.ForProvider.ProjectID = &projectID
					return cr
				}(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"ObserveSuccess": {
			args: args{
				client: &mockClient{
					get: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
						return &gitlab.EmailsOnPushService{
							Properties: &gitlab.EmailsOnPushServiceProperties{},
						}, &gitlab.Response{}, nil
					},
				},
				cr: func() resource.Managed {
					cr := emailOnPush()
					cr.Spec.ForProvider.ProjectID = &projectID
					return cr
				}(),
			},
			want: want{
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"Observe404": {
			args: args{
				client: &mockClient{
					get: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.EmailsOnPushService, *gitlab.Response, error) {
						return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
					},
				},
				cr: func() resource.Managed {
					cr := emailOnPush()
					cr.Spec.ForProvider.ProjectID = &projectID
					return cr
				}(),
			},
			want: want{
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.client,
			}

			res, err := e.Observe(context.Background(), tc.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Fatalf("Observe error mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.result.ResourceExists, res.ResourceExists); diff != "" {
				t.Fatalf("Observe result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
