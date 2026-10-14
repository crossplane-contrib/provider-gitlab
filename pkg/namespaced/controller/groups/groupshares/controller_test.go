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

package groupshares

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	v2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
)

var (
	errBoom        = errors.New("boom")
	sharedGroupID  = int64(1)
	invitedGroupID = int64(2)
	expiresAt      = "2027-01-31"
	expiresAtTime  = gitlab.ISOTime(time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC))
	notFound       = &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}
	serverError    = &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}
)

type shareModifier func(*v1alpha1.GroupShare)

func withSpec(p v1alpha1.GroupShareParameters) shareModifier {
	return func(cr *v1alpha1.GroupShare) { cr.Spec.ForProvider = p }
}

func withStatus(o v1alpha1.GroupShareObservation) shareModifier {
	return func(cr *v1alpha1.GroupShare) { cr.Status.AtProvider = o }
}

func withConditions(c ...v2.Condition) shareModifier {
	return func(cr *v1alpha1.GroupShare) { cr.Status.SetConditions(c...) }
}

func groupShare(m ...shareModifier) *v1alpha1.GroupShare {
	cr := &v1alpha1.GroupShare{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func params(access int64, expires *string) v1alpha1.GroupShareParameters {
	return v1alpha1.GroupShareParameters{
		SharedGroupID:    new(sharedGroupID),
		InvitedGroupID:   new(invitedGroupID),
		GroupAccessLevel: access,
		ExpiresAt:        expires,
	}
}

// sharedGroup is the shared group as GitLab returns it, with the given shares.
func sharedGroup(shares ...gitlab.SharedWithGroup) func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
	return func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
		return &gitlab.Group{ID: sharedGroupID, SharedWithGroups: shares}, &gitlab.Response{}, nil
	}
}

func share(access int64, expires *gitlab.ISOTime) gitlab.SharedWithGroup {
	return gitlab.SharedWithGroup{GroupID: invitedGroupID, GroupFullPath: "org/developers", GroupAccessLevel: access, ExpiresAt: expires}
}

func TestConnect(t *testing.T) {
	cases := map[string]struct {
		kube client.Client
		cr   resource.Managed
		err  error
	}{
		"InvalidInput": {
			err: errors.New(errNotGroupShare),
		},
		"ProviderConfigRefNotGivenError": {
			kube: &test.MockClient{MockGet: test.NewMockGetFn(nil)},
			cr:   groupShare(),
			err:  errors.New("providerConfigRef is not given"),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube}
			_, err := c.Connect(context.Background(), tc.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Connect(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr  resource.Managed
		obs managed.ExternalObservation
		err error
	}
	cases := map[string]struct {
		reason string
		client *fake.MockClient
		cr     resource.Managed
		want   want
	}{
		"InvalidInput": {
			reason: "Observe must reject resources that are not a GroupShare.",
			client: &fake.MockClient{},
			want:   want{err: errors.New(errNotGroupShare)},
		},
		"MissingGroupIDs": {
			reason: "Both groups must be known before GitLab can be asked about the share.",
			client: &fake.MockClient{},
			cr:     groupShare(withSpec(v1alpha1.GroupShareParameters{SharedGroupID: new(sharedGroupID), GroupAccessLevel: 30})),
			want: want{
				cr:  groupShare(withSpec(v1alpha1.GroupShareParameters{SharedGroupID: new(sharedGroupID), GroupAccessLevel: 30})),
				err: errors.New(errMissingGroupIDs),
			},
		},
		"SharedGroupNotFound": {
			reason: "A share of a group that does not exist does not exist either.",
			client: &fake.MockClient{MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
				return nil, notFound, errBoom
			}},
			cr: groupShare(withSpec(params(30, nil))),
			want: want{
				cr:  groupShare(withSpec(params(30, nil))),
				obs: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"GetGroupFailed": {
			reason: "Errors other than 404 must be returned.",
			client: &fake.MockClient{MockGetGroup: func(pid interface{}, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
				return nil, serverError, errBoom
			}},
			cr: groupShare(withSpec(params(30, nil))),
			want: want{
				cr:  groupShare(withSpec(params(30, nil))),
				err: errors.Wrap(errBoom, errGetGroup),
			},
		},
		"ShareNotFound": {
			reason: "The share does not exist if the invited group is not among the shared group's shares.",
			client: &fake.MockClient{MockGetGroup: sharedGroup(gitlab.SharedWithGroup{GroupID: 99, GroupAccessLevel: 30})},
			cr:     groupShare(withSpec(params(30, nil))),
			want: want{
				cr:  groupShare(withSpec(params(30, nil))),
				obs: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"UpToDate": {
			reason: "A share with the desired access level and expiry is up to date.",
			client: &fake.MockClient{MockGetGroup: sharedGroup(share(30, &expiresAtTime))},
			cr:     groupShare(withSpec(params(30, new(expiresAt)))),
			want: want{
				cr: groupShare(
					withSpec(params(30, new(expiresAt))),
					withStatus(v1alpha1.GroupShareObservation{InvitedGroupFullPath: "org/developers", GroupAccessLevel: 30, ExpiresAt: new(expiresAt)}),
					withConditions(v2.Available()),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
			},
		},
		"AccessLevelDiffers": {
			reason: "A different access level must be applied.",
			client: &fake.MockClient{MockGetGroup: sharedGroup(share(20, nil))},
			cr:     groupShare(withSpec(params(30, nil))),
			want: want{
				cr: groupShare(
					withSpec(params(30, nil)),
					withStatus(v1alpha1.GroupShareObservation{InvitedGroupFullPath: "org/developers", GroupAccessLevel: 20}),
					withConditions(v2.Available()),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false},
			},
		},
		"ExpiresAtDiffers": {
			reason: "A different expiry date must be applied.",
			client: &fake.MockClient{MockGetGroup: sharedGroup(share(30, &expiresAtTime))},
			cr:     groupShare(withSpec(params(30, new("2027-02-28")))),
			want: want{
				cr: groupShare(
					withSpec(params(30, new("2027-02-28"))),
					withStatus(v1alpha1.GroupShareObservation{InvitedGroupFullPath: "org/developers", GroupAccessLevel: 30, ExpiresAt: new(expiresAt)}),
					withConditions(v2.Available()),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false},
			},
		},
		"LateInitializeExpiresAt": {
			reason: "An expiry set in GitLab is copied into the spec when the spec has none.",
			client: &fake.MockClient{MockGetGroup: sharedGroup(share(30, &expiresAtTime))},
			cr:     groupShare(withSpec(params(30, nil))),
			want: want{
				cr: groupShare(
					withSpec(params(30, new(expiresAt))),
					withStatus(v1alpha1.GroupShareObservation{InvitedGroupFullPath: "org/developers", GroupAccessLevel: 30, ExpiresAt: new(expiresAt)}),
					withConditions(v2.Available()),
				),
				obs: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ResourceLateInitialized: true},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			obs, err := e.Observe(context.Background(), tc.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nObserve(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("%s\nObserve(...): -want, +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.cr, test.EquateConditions()); diff != "" {
				t.Errorf("%s\nObserve(...): -want resource, +got resource:\n%s", tc.reason, diff)
			}
		})
	}
}

// equateISOTime compares GitLab dates by their instant.
var equateISOTime = cmp.Comparer(func(a, b gitlab.ISOTime) bool { return time.Time(a).Equal(time.Time(b)) })

// call records a request the controller sent to GitLab.
type call struct {
	Method       string
	Group        interface{}
	Opt          *gitlab.ShareGroupWithGroupOptions
	InvitedGroup int64
}

func recordingClient(calls *[]call, shareErr, unshareErr error, unshareRes *gitlab.Response) *fake.MockClient {
	return &fake.MockClient{
		MockShareGroupWithGroup: func(gid interface{}, opt *gitlab.ShareGroupWithGroupOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Group, *gitlab.Response, error) {
			*calls = append(*calls, call{Method: "share", Group: gid, Opt: opt})
			return &gitlab.Group{}, &gitlab.Response{}, shareErr
		},
		MockUnshareGroupFromGroup: func(gid interface{}, groupID int64, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error) {
			*calls = append(*calls, call{Method: "unshare", Group: gid, InvitedGroup: groupID})
			return unshareRes, unshareErr
		},
	}
}

func TestCreate(t *testing.T) {
	cases := map[string]struct {
		reason   string
		cr       resource.Managed
		shareErr error
		calls    []call
		err      error
	}{
		"InvalidInput": {
			reason: "Create must reject resources that are not a GroupShare.",
			err:    errors.New(errNotGroupShare),
		},
		"MissingGroupIDs": {
			reason: "Both groups must be known before sharing.",
			cr:     groupShare(withSpec(v1alpha1.GroupShareParameters{InvitedGroupID: new(invitedGroupID), GroupAccessLevel: 30})),
			err:    errors.New(errMissingGroupIDs),
		},
		"Shared": {
			reason: "The shared group is shared with the invited group at the desired level and expiry.",
			cr:     groupShare(withSpec(params(30, new(expiresAt)))),
			calls: []call{{Method: "share", Group: sharedGroupID, Opt: &gitlab.ShareGroupWithGroupOptions{
				GroupID:     new(invitedGroupID),
				GroupAccess: new(gitlab.AccessLevelValue(30)),
				ExpiresAt:   new(expiresAtTime),
			}}},
		},
		"SharedWithoutExpiry": {
			reason: "Without an expiry date, none is sent.",
			cr:     groupShare(withSpec(params(40, nil))),
			calls: []call{{Method: "share", Group: sharedGroupID, Opt: &gitlab.ShareGroupWithGroupOptions{
				GroupID:     new(invitedGroupID),
				GroupAccess: new(gitlab.AccessLevelValue(40)),
			}}},
		},
		"InvalidExpiresAt": {
			reason: "An expiry date that is not a date is an error, not an unlimited share.",
			cr:     groupShare(withSpec(params(30, new("tomorrow")))),
			err:    errors.Wrap(errors.New(`parsing time "tomorrow" as "2006-01-02": cannot parse "tomorrow" as "2006"`), errExpiresAt),
		},
		"ShareFailed": {
			reason: "Errors from GitLab must be returned.",
			cr:     groupShare(withSpec(params(30, nil))),
			calls: []call{{Method: "share", Group: sharedGroupID, Opt: &gitlab.ShareGroupWithGroupOptions{
				GroupID:     new(invitedGroupID),
				GroupAccess: new(gitlab.AccessLevelValue(30)),
			}}},
			shareErr: errBoom,
			err:      errors.Wrap(errBoom, errShare),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var calls []call
			e := &external{client: recordingClient(&calls, tc.shareErr, nil, &gitlab.Response{})}
			_, err := e.Create(context.Background(), tc.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nCreate(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.calls, calls, equateISOTime); diff != "" {
				t.Errorf("%s\nCreate(...): -want calls, +got calls:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := map[string]struct {
		reason     string
		shareErr   error
		unshareErr error
		calls      []call
		err        error
	}{
		"Reshared": {
			reason: "GitLab cannot change a share, so it is removed and created again with the desired settings.",
			calls: []call{
				{Method: "unshare", Group: sharedGroupID, InvitedGroup: invitedGroupID},
				{Method: "share", Group: sharedGroupID, Opt: &gitlab.ShareGroupWithGroupOptions{
					GroupID:     new(invitedGroupID),
					GroupAccess: new(gitlab.AccessLevelValue(40)),
				}},
			},
		},
		"ShareFailedAfterUnshare": {
			reason: "If the new share cannot be created, the error is returned so that a later reconcile creates it.",
			calls: []call{
				{Method: "unshare", Group: sharedGroupID, InvitedGroup: invitedGroupID},
				{Method: "share", Group: sharedGroupID, Opt: &gitlab.ShareGroupWithGroupOptions{
					GroupID:     new(invitedGroupID),
					GroupAccess: new(gitlab.AccessLevelValue(40)),
				}},
			},
			shareErr: errBoom,
			err:      errors.Wrap(errBoom, errShare),
		},
		"UnshareFailed": {
			reason: "If the old share cannot be removed, no new one is created.",
			calls: []call{
				{Method: "unshare", Group: sharedGroupID, InvitedGroup: invitedGroupID},
			},
			unshareErr: errBoom,
			err:        errors.Wrap(errBoom, errUnshare),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var calls []call
			e := &external{client: recordingClient(&calls, tc.shareErr, tc.unshareErr, serverError)}
			_, err := e.Update(context.Background(), groupShare(withSpec(params(40, nil))))
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nUpdate(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.calls, calls, equateISOTime); diff != "" {
				t.Errorf("%s\nUpdate(...): -want calls, +got calls:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	unshareCall := []call{{Method: "unshare", Group: sharedGroupID, InvitedGroup: invitedGroupID}}
	cases := map[string]struct {
		reason     string
		cr         resource.Managed
		unshareRes *gitlab.Response
		unshareErr error
		calls      []call
		err        error
	}{
		"InvalidInput": {
			reason: "Delete must reject resources that are not a GroupShare.",
			err:    errors.New(errNotGroupShare),
		},
		"Unshared": {
			reason:     "Deleting the resource removes the share.",
			cr:         groupShare(withSpec(params(30, nil))),
			unshareRes: &gitlab.Response{},
			calls:      unshareCall,
		},
		"AlreadyGone": {
			reason:     "A share that no longer exists is deleted.",
			cr:         groupShare(withSpec(params(30, nil))),
			unshareRes: notFound,
			unshareErr: errBoom,
			calls:      unshareCall,
		},
		"UnshareFailed": {
			reason:     "Errors other than 404 must be returned.",
			cr:         groupShare(withSpec(params(30, nil))),
			unshareRes: serverError,
			unshareErr: errBoom,
			err:        errors.Wrap(errBoom, errUnshare),
			calls:      unshareCall,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var calls []call
			e := &external{client: recordingClient(&calls, nil, tc.unshareErr, tc.unshareRes)}
			_, err := e.Delete(context.Background(), tc.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nDelete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.calls, calls); diff != "" {
				t.Errorf("%s\nDelete(...): -want calls, +got calls:\n%s", tc.reason, diff)
			}
		})
	}
}
