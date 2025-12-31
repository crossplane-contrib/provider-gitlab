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

package badges

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	meta "github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups"
)

var (
	unexpecedItem resource.Managed
	errBoom       = errors.New("boom")
	groupID       = int64(1234)
)

// mockBadgeClient implements groups.BadgeClient for tests
type mockBadgeClient struct {
	GetFn    func(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error)
	AddFn    func(gid any, opt *gitlab.AddGroupBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error)
	EditFn   func(gid any, badge int64, opt *gitlab.EditGroupBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error)
	DeleteFn func(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

func (m *mockBadgeClient) ListGroupBadges(gid any, opt *gitlab.ListGroupBadgesOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.GroupBadge, *gitlab.Response, error) {
	return nil, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
}

func (m *mockBadgeClient) GetGroupBadge(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
	if m.GetFn == nil {
		return nil, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
	}
	return m.GetFn(gid, badge, options...)
}
func (m *mockBadgeClient) AddGroupBadge(gid any, opt *gitlab.AddGroupBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
	if m.AddFn == nil {
		return nil, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
	}
	return m.AddFn(gid, opt, options...)
}

func (m *mockBadgeClient) EditGroupBadge(gid any, badge int64, opt *gitlab.EditGroupBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
	if m.EditFn == nil {
		return nil, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
	}
	return m.EditFn(gid, badge, opt, options...)
}

func (m *mockBadgeClient) DeleteGroupBadge(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
	if m.DeleteFn == nil {
		return &gitlab.Response{Response: &http.Response{StatusCode: 204}}, nil
	}
	return m.DeleteFn(gid, badge, options...)
}

func (m *mockBadgeClient) PreviewGroupBadge(gid any, opt *gitlab.GroupBadgePreviewOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
	return nil, nil, nil
}

// helper to create a Badge CR with modifiers
type badgeModifier func(*v1alpha1.Badge)

func withGroupID() badgeModifier {
	return func(r *v1alpha1.Badge) { r.Spec.ForProvider.GroupID = &groupID }
}

func withSpec(s v1alpha1.BadgeParameters) badgeModifier {
	return func(r *v1alpha1.Badge) { r.Spec.ForProvider = s }
}

func withConditions(c ...xpv1.Condition) badgeModifier {
	return func(r *v1alpha1.Badge) { r.Status.ConditionedStatus.Conditions = c }
}

func badge(m ...badgeModifier) *v1alpha1.Badge {
	cr := &v1alpha1.Badge{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type args struct {
	badge groups.BadgeClient
	kube  client.Client
	cr    resource.Managed
}

func TestConnect(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalClient
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{cr: unexpecedItem},
			want: want{cr: unexpecedItem, err: errors.New(errNotBadge)},
		},
		"ProviderConfigRefNotGivenError": {
			args: args{cr: badge()},
			want: want{cr: badge(), err: errors.New("providerConfigRef is not given")},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newGitlabClientFn: nil}
			o, err := c.Connect(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{cr: unexpecedItem},
			want: want{cr: unexpecedItem, err: errors.New(errNotBadge)},
		},
		"MissingGroupID": {
			args: args{cr: badge()},
			want: want{cr: badge(), err: errors.New(errGroupIDMissing)},
		},
		"EmptyExternalName": {
			args: args{cr: func() resource.Managed {
				r := badge(withGroupID())
				r.SetName("test")
				r.SetGenerateName("n")
				return r
			}()},
			want: want{cr: func() resource.Managed {
				r := badge(withGroupID())
				r.SetName("test")
				r.SetGenerateName("n")
				return r
			}(), result: managed.ExternalObservation{ResourceExists: false}},
		},
		"InvalidExternalName": {
			args: args{cr: func() resource.Managed {
				r := badge(withGroupID())
				meta.SetExternalName(r, "invalid")
				return r
			}()},
			want: want{cr: func() resource.Managed {
				r := badge(withGroupID())
				meta.SetExternalName(r, "invalid")
				return r
			}(), err: errors.New(errNotBadge)},
		},
		"NotFound": {
			args: args{
				badge: &mockBadgeClient{GetFn: func(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
					return nil, &gitlab.Response{Response: &http.Response{StatusCode: 404}}, errBoom
				}},
				cr: func() resource.Managed { cr := badge(withGroupID()); meta.SetExternalName(cr, "999"); return cr }(),
			},
			want: want{cr: func() resource.Managed { cr := badge(withGroupID()); meta.SetExternalName(cr, "999"); return cr }(), result: managed.ExternalObservation{ResourceExists: false}},
		},
		"FoundAndUpToDate": {
			args: args{
				badge: &mockBadgeClient{GetFn: func(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
					return &gitlab.GroupBadge{ID: 1, Name: "b"}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
				}},
				cr: func() resource.Managed { cr := badge(withGroupID()); meta.SetExternalName(cr, "1"); return cr }(),
			},
			want: want{cr: func() resource.Managed {
				cr := badge(withGroupID())
				meta.SetExternalName(cr, "1")
				// lateInit should set these values
				name := "b"
				empty := ""
				cr.Spec.ForProvider.Name = &name
				cr.Spec.ForProvider.ImageURL = empty
				cr.Spec.ForProvider.LinkURL = empty
				withConditions(xpv1.Available())(cr)
				cr.Status.AtProvider = v1alpha1.BadgeObservation{ID: 1, Name: "b", LinkURL: empty, RenderedLinkURL: empty, ImageURL: empty, RenderedImageURL: empty}
				return cr
			}(), result: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: true}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.badge}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreateUpdateDeleteDisconnect(t *testing.T) {
	// Add tests for Create, Update, Delete and Disconnect
	t.Run("CreateInvalidInput", func(t *testing.T) {
		e := &external{kube: nil, client: nil}
		_, err := e.Create(context.Background(), unexpecedItem)
		if diff := cmp.Diff(errors.New(errNotBadge), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("CreateMissingGroupID", func(t *testing.T) {
		cr := badge()
		e := &external{kube: nil, client: &mockBadgeClient{}}
		_, err := e.Create(context.Background(), cr)
		if diff := cmp.Diff(errors.New(errCreateFailed), errors.Cause(err), test.EquateErrors()); diff == "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("CreateWithExistingID", func(t *testing.T) {
		id := int64(5)
		cr := badge(withSpec(v1alpha1.BadgeParameters{ID: &id, GroupID: &groupID}))
		e := &external{kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)}, client: &mockBadgeClient{GetFn: func(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
			return &gitlab.GroupBadge{ID: 5}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
		}}}
		_, err := e.Create(context.Background(), cr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := meta.GetExternalName(cr); got != strconv.Itoa(5) {
			t.Fatalf("external name was not set, got: %s", got)
		}
	})

	t.Run("CreateSuccess", func(t *testing.T) {
		name := "b"
		cr := badge(withSpec(v1alpha1.BadgeParameters{Name: &name, GroupID: &groupID}))
		e := &external{kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)}, client: &mockBadgeClient{AddFn: func(gid any, opt *gitlab.AddGroupBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
			return &gitlab.GroupBadge{ID: 7}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
		}}}
		_, err := e.Create(context.Background(), cr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := meta.GetExternalName(cr); got != strconv.Itoa(7) {
			t.Fatalf("external name was not set, got: %s", got)
		}
	})

	t.Run("UpdateInvalidInput", func(t *testing.T) {
		e := &external{kube: nil, client: nil}
		_, err := e.Update(context.Background(), unexpecedItem)
		if diff := cmp.Diff(errors.New(errNotBadge), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("UpdateParseExternalNameError", func(t *testing.T) {
		cr := badge(withGroupID())
		meta.SetExternalName(cr, "not-int")
		e := &external{kube: nil, client: &mockBadgeClient{}}
		_, err := e.Update(context.Background(), cr)
		if diff := cmp.Diff(errors.New(errNotBadge), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("UpdateMissingGroupID", func(t *testing.T) {
		cr := badge()
		meta.SetExternalName(cr, "1")
		e := &external{kube: nil, client: &mockBadgeClient{}}
		_, err := e.Update(context.Background(), cr)
		if diff := cmp.Diff(errors.New(errGroupIDMissing), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("UpdateSuccess", func(t *testing.T) {
		cr := badge(withGroupID())
		meta.SetExternalName(cr, "1")
		e := &external{kube: nil, client: &mockBadgeClient{EditFn: func(gid any, badge int64, opt *gitlab.EditGroupBadgeOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupBadge, *gitlab.Response, error) {
			return &gitlab.GroupBadge{}, &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
		}}}
		_, err := e.Update(context.Background(), cr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("DeleteInvalidInput", func(t *testing.T) {
		e := &external{kube: nil, client: nil}
		_, err := e.Delete(context.Background(), unexpecedItem)
		if diff := cmp.Diff(errors.New(errNotBadge), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("DeleteParseExternalNameError", func(t *testing.T) {
		cr := badge(withGroupID())
		meta.SetExternalName(cr, "not-int")
		e := &external{kube: nil, client: &mockBadgeClient{}}
		_, err := e.Delete(context.Background(), cr)
		if diff := cmp.Diff(errors.New(errNotBadge), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("DeleteMissingGroupID", func(t *testing.T) {
		cr := badge()
		meta.SetExternalName(cr, "1")
		e := &external{kube: nil, client: &mockBadgeClient{}}
		_, err := e.Delete(context.Background(), cr)
		if diff := cmp.Diff(errors.New(errGroupIDMissing), err, test.EquateErrors()); diff != "" {
			t.Errorf("unexpected error: %s", diff)
		}
	})

	t.Run("DeleteSuccess", func(t *testing.T) {
		cr := badge(withGroupID())
		meta.SetExternalName(cr, "1")
		e := &external{kube: nil, client: &mockBadgeClient{DeleteFn: func(gid any, badge int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
			return &gitlab.Response{Response: &http.Response{StatusCode: 200}}, nil
		}}}
		_, err := e.Delete(context.Background(), cr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Disconnect", func(t *testing.T) {
		e := &external{}
		if err := e.Disconnect(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestLateInitializeBadge(t *testing.T) {
	p := &v1alpha1.BadgeParameters{}
	lateInitializeBadge(p, nil)
	if p.ID != nil || p.Name != nil {
		t.Fatalf("expected no changes for nil badge")
	}
	badge := &gitlab.GroupBadge{ID: 10, Name: "n", ImageURL: "img", LinkURL: "link"}
	lateInitializeBadge(p, badge)
	if p.Name == nil || *p.Name != "n" {
		t.Fatalf("expected Name to be set")
	}
}
