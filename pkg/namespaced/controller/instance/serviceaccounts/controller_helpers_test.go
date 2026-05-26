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

package serviceaccounts

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	xerrors "github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	groupsfake "github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/groups/fake"
)

func TestGetAccessLevelValue(t *testing.T) {
	tests := map[string]struct {
		accessLevel string
		want        int
	}{
		"NoAccess":                  {accessLevel: accessLevelNoAccess, want: accessLevelNoAccessValue},
		"MinimalAccess":             {accessLevel: accessLevelMinimalAccess, want: accessLevelMinimalAccessValue},
		"Guest":                     {accessLevel: accessLevelGuest, want: accessLevelGuestValue},
		"Planner":                   {accessLevel: accessLevelPlanner, want: accessLevelPlannerValue},
		"Reporter":                  {accessLevel: accessLevelReporter, want: accessLevelReporterValue},
		"SecurityManager":           {accessLevel: accessLevelSecurityManager, want: accessLevelSecurityManagerValue},
		"Developer":                 {accessLevel: accessLevelDeveloper, want: accessLevelDeveloperValue},
		"Maintainer":                {accessLevel: accessLevelMaintainer, want: accessLevelMaintainerValue},
		"Owner":                     {accessLevel: accessLevelOwner, want: accessLevelOwnerValue},
		"UnknownDefaultsToNoAccess": {accessLevel: "unknown", want: accessLevelNoAccessValue},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := getAccessLevelValue(tc.accessLevel)
			if got != tc.want {
				t.Fatalf("getAccessLevelValue(%q) = %d, want %d", tc.accessLevel, got, tc.want)
			}
		})
	}
}

func TestFetchTopLevelGroupsPage(t *testing.T) {
	client := &groupsfake.MockClient{
		MockListGroups: func(opt *gitlab.ListGroupsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
			if opt.TopLevelOnly == nil || !*opt.TopLevelOnly {
				t.Fatalf("fetchTopLevelGroupsPage() called ListGroups without TopLevelOnly")
			}
			if opt.PerPage != 100 {
				t.Fatalf("fetchTopLevelGroupsPage() called ListGroups with PerPage = %d, want 100", opt.PerPage)
			}
			if opt.Page != 2 {
				t.Fatalf("fetchTopLevelGroupsPage() called ListGroups with Page = %d, want 2", opt.Page)
			}
			return []*gitlab.Group{{ID: 123}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}, NextPage: 3}, nil
		},
	}

	e := &external{groupsClient: client}

	got, resp, err := e.fetchTopLevelGroupsPage(2)
	if err != nil {
		t.Fatalf("fetchTopLevelGroupsPage() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("fetchTopLevelGroupsPage() returned nil response")
	}
	if resp.NextPage != 3 {
		t.Fatalf("fetchTopLevelGroupsPage() NextPage = %d, want 3", resp.NextPage)
	}
	if diff := cmp.Diff([]int64{123}, []int64{got[0].ID}); diff != "" {
		t.Fatalf("fetchTopLevelGroupsPage() unexpected groups (-want +got):\n%s", diff)
	}
}

func TestGetGroupPermissionStatus(t *testing.T) {
	tests := map[string]struct {
		memberClient *groupsfake.MockClient
		wantMissing  bool
		wantWrong    bool
		wantErr      error
	}{
		"NotFound": {
			memberClient: &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
			}},
			wantMissing: true,
		},
		"NilMember": {
			memberClient: &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
			}},
			wantWrong: true,
		},
		"LowAccess": {
			memberClient: &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
				return &gitlab.GroupMember{AccessLevel: gitlab.AccessLevelValue(accessLevelReporterValue)}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
			}},
			wantWrong: true,
		},
		"EnoughAccess": {
			memberClient: &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
				return &gitlab.GroupMember{AccessLevel: gitlab.AccessLevelValue(accessLevelMaintainerValue)}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
			}},
		},
		"UnexpectedError": {
			memberClient: &groupsfake.MockClient{MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errors.New("boom")
			}},
			wantErr: xerrors.New("cannot check Gitlab group member: boom"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := &external{groupMemberClient: tc.memberClient}
			missing, wrong, err := e.getGroupPermissionStatus(42, 99, accessLevelDeveloperValue)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("getGroupPermissionStatus() unexpected error: %v", err)
				}
			} else if err == nil || err.Error() != tc.wantErr.Error() {
				t.Fatalf("getGroupPermissionStatus() error = %v, want %v", err, tc.wantErr)
			}
			if missing != tc.wantMissing {
				t.Fatalf("getGroupPermissionStatus() missing = %v, want %v", missing, tc.wantMissing)
			}
			if wrong != tc.wantWrong {
				t.Fatalf("getGroupPermissionStatus() wrong = %v, want %v", wrong, tc.wantWrong)
			}
		})
	}
}

func TestFetchTopLevelGroupsMissingPermissions(t *testing.T) {
	var (
		pagesMu sync.Mutex
		pages   []int64
	)

	client := &groupsfake.MockClient{
		MockListGroups: func(opt *gitlab.ListGroupsOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
			pagesMu.Lock()
			pages = append(pages, opt.Page)
			pagesMu.Unlock()

			switch opt.Page {
			case 1:
				return []*gitlab.Group{{ID: 1}, {ID: 2}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}, NextPage: 2}, nil
			case 2:
				return []*gitlab.Group{{ID: 3}}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}, NextPage: 0}, nil
			default:
				t.Fatalf("unexpected page %d", opt.Page)
				return nil, nil, nil
			}
		},
		MockGetMember: func(gid interface{}, user int64, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
			switch gid.(int64) {
			case 1:
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
			case 2:
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
			case 3:
				return &gitlab.GroupMember{AccessLevel: gitlab.AccessLevelValue(accessLevelReporterValue)}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
			default:
				t.Fatalf("unexpected group id %v", gid)
				return nil, nil, nil
			}
		},
	}

	e := &external{groupsClient: client, groupMemberClient: client}

	notIn, wrong, err := e.fetchTopLevelGroupsMissingPermissions(accessLevelDeveloperValue, 99)
	if err != nil {
		t.Fatalf("fetchTopLevelGroupsMissingPermissions() unexpected error: %v", err)
	}
	if diff := cmp.Diff([]int64{1}, notIn, cmpopts.SortSlices(func(a, b int64) bool { return a < b })); diff != "" {
		t.Fatalf("fetchTopLevelGroupsMissingPermissions() notInGroups mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]int64{2, 3}, wrong, cmpopts.SortSlices(func(a, b int64) bool { return a < b })); diff != "" {
		t.Fatalf("fetchTopLevelGroupsMissingPermissions() wrongPermsGroups mismatch (-want +got):\n%s", diff)
	}

	pagesMu.Lock()
	defer pagesMu.Unlock()
	if diff := cmp.Diff([]int64{1, 2}, pages); diff != "" {
		t.Fatalf("fetchTopLevelGroupsMissingPermissions() page sequence mismatch (-want +got):\n%s", diff)
	}
}

func TestAddServiceAccountToGroups(t *testing.T) {
	var (
		callsMu sync.Mutex
		calls   = map[int64]*gitlab.AddGroupMemberOptions{}
	)

	client := &groupsfake.MockClient{
		MockAddMember: func(gid interface{}, opt *gitlab.AddGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
			groupID := gid.(int64)
			callsMu.Lock()
			calls[groupID] = opt
			callsMu.Unlock()

			if groupID == 12 {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errors.New("boom")
			}

			return &gitlab.GroupMember{}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusCreated}}, nil
		},
	}

	e := &external{groupMemberClient: client}
	err := e.addServiceAccountToGroups(context.Background(), 99, []int64{11, 12}, accessLevelMaintainerValue)
	if err == nil {
		t.Fatal("addServiceAccountToGroups() expected an error")
	}

	callsMu.Lock()
	defer callsMu.Unlock()
	if len(calls) != 2 {
		t.Fatalf("addServiceAccountToGroups() called AddGroupMember %d times, want 2", len(calls))
	}
	for groupID, call := range calls {
		if call == nil {
			t.Fatalf("addServiceAccountToGroups() nil call for group %d", groupID)
		}
		if call.UserID == nil || *call.UserID != 99 {
			t.Fatalf("addServiceAccountToGroups() UserID for group %d = %v, want 99", groupID, call.UserID)
		}
		if call.AccessLevel == nil || int(*call.AccessLevel) != accessLevelMaintainerValue {
			t.Fatalf("addServiceAccountToGroups() AccessLevel for group %d = %v, want %d", groupID, call.AccessLevel, accessLevelMaintainerValue)
		}
	}
}

func TestUpdateServiceAccountGroupPermissions(t *testing.T) {
	var (
		callsMu sync.Mutex
		calls   = map[int64]*gitlab.EditGroupMemberOptions{}
	)

	client := &groupsfake.MockClient{
		MockEditMember: func(gid interface{}, user int64, opt *gitlab.EditGroupMemberOptions, options ...gitlab.RequestOptionFunc) (*gitlab.GroupMember, *gitlab.Response, error) {
			groupID := gid.(int64)
			callsMu.Lock()
			calls[groupID] = opt
			callsMu.Unlock()

			if groupID == 12 {
				return nil, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errors.New("boom")
			}

			return &gitlab.GroupMember{}, &gitlab.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
		},
	}

	e := &external{groupMemberClient: client}
	err := e.updateServiceAccountGroupPermissions(context.Background(), 99, []int64{11, 12}, accessLevelMaintainerValue)
	if err == nil {
		t.Fatal("updateServiceAccountGroupPermissions() expected an error")
	}

	callsMu.Lock()
	defer callsMu.Unlock()
	if len(calls) != 2 {
		t.Fatalf("updateServiceAccountGroupPermissions() called EditGroupMember %d times, want 2", len(calls))
	}
	for groupID, call := range calls {
		if call == nil {
			t.Fatalf("updateServiceAccountGroupPermissions() nil call for group %d", groupID)
		}
		if call.AccessLevel == nil || int(*call.AccessLevel) != accessLevelMaintainerValue {
			t.Fatalf("updateServiceAccountGroupPermissions() AccessLevel for group %d = %v, want %d", groupID, call.AccessLevel, accessLevelMaintainerValue)
		}
	}
}
