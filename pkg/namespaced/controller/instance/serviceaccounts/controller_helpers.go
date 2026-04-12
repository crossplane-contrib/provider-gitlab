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
	"sync"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

// addServiceAccountToGroups is a helper function that adds the service account to the specified groups with the specified access level. This is used to satisfy the BaselinePermissions field by adding the service account to all groups that have at least the specified access level.
func (e *external) addServiceAccountToGroups(ctx context.Context, serviceAccountID int64, groupIDs []int64, accessLevel int) error {
	var addToGroupsWaitGroup sync.WaitGroup
	var requestErr error
	var requestErrMu sync.Mutex

	setRequestErr := func(err error) {
		if err == nil {
			return
		}
		requestErrMu.Lock()
		if requestErr == nil {
			requestErr = err
		}
		requestErrMu.Unlock()
	}

	for _, groupID := range groupIDs {
		addToGroupsWaitGroup.Add(1)
		go func(groupID int64) {
			defer addToGroupsWaitGroup.Done()

			_, _, err := e.groupMemberClient.AddGroupMember(
				groupID,
				&gitlab.AddGroupMemberOptions{
					UserID:      ptr.To(serviceAccountID),
					AccessLevel: ptr.To(gitlab.AccessLevelValue(accessLevel)),
				},
				gitlab.WithContext(ctx),
			)
			if err != nil {
				setRequestErr(errors.Wrapf(err, "cannot add service account to Gitlab group with ID %d", groupID))
				return
			}
		}(groupID)
	}

	addToGroupsWaitGroup.Wait()
	return requestErr
}

// updateServiceAccountGroupPermissions is a helper function that updates the service account's permissions for the specified groups to the specified access level. This is used to satisfy the BaselinePermissions field by updating the service account's permissions for all groups that have at least the specified access level.
func (e *external) updateServiceAccountGroupPermissions(ctx context.Context, serviceAccountID int64, groupIDs []int64, accessLevel int) error {
	var updateGroupsWaitGroup sync.WaitGroup
	var requestErr error
	var requestErrMu sync.Mutex

	setRequestErr := func(err error) {
		if err == nil {
			return
		}
		requestErrMu.Lock()
		if requestErr == nil {
			requestErr = err
		}
		requestErrMu.Unlock()
	}

	for _, groupID := range groupIDs {
		updateGroupsWaitGroup.Add(1)
		go func(groupID int64) {
			defer updateGroupsWaitGroup.Done()

			_, _, err := e.groupMemberClient.EditGroupMember(
				groupID,
				serviceAccountID,
				&gitlab.EditGroupMemberOptions{
					AccessLevel: ptr.To(gitlab.AccessLevelValue(accessLevel)),
				},
				gitlab.WithContext(ctx),
			)
			if err != nil {
				setRequestErr(errors.Wrapf(err, "cannot update service account permissions for Gitlab group with ID %d", groupID))
				return
			}
		}(groupID)
	}

	updateGroupsWaitGroup.Wait()
	return requestErr
}

// fetchTopLevelGroupsMissingPermissions is a helper function that returns a list of top level group IDs that the service account is missing permissions for, based on the specified minimum permission level. This is used to determine which groups the service account needs to be added to in order to satisfy the BaselinePermissions field.
//
//nolint:gocyclo
func (e *external) fetchTopLevelGroupsMissingPermissions(minPermission int, serviceAccountID int64) (notInGroups []int64, wrongPermsGroups []int64, err error) {
	var (
		permissionsCheckWaitGroup  sync.WaitGroup
		groupsMissingPermissionsMu sync.Mutex
		requestErr                 error
		requestErrMu               sync.Mutex
	)

	setRequestErr := func(err error) {
		if err == nil {
			return
		}
		requestErrMu.Lock()
		if requestErr == nil {
			requestErr = err
		}
		requestErrMu.Unlock()
	}

	for page := int64(1); ; {
		groups, resp, err := e.fetchTopLevelGroupsPage(page)
		if err != nil {
			return nil, nil, err
		}

		for _, group := range groups {
			permissionsCheckWaitGroup.Add(1)
			go func(groupID int64) {
				defer permissionsCheckWaitGroup.Done()

				isMissingMembership, hasInsufficientPermissions, err := e.getGroupPermissionStatus(groupID, serviceAccountID, minPermission)
				if err != nil {
					setRequestErr(err)
					return
				}

				groupsMissingPermissionsMu.Lock()
				if isMissingMembership {
					notInGroups = append(notInGroups, groupID)
				}
				if hasInsufficientPermissions {
					wrongPermsGroups = append(wrongPermsGroups, groupID)
				}
				groupsMissingPermissionsMu.Unlock()
			}(group.ID)
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	permissionsCheckWaitGroup.Wait()
	if requestErr != nil {
		return nil, nil, requestErr
	}

	return notInGroups, wrongPermsGroups, nil
}

func (e *external) fetchTopLevelGroupsPage(page int64) ([]*gitlab.Group, *gitlab.Response, error) {
	groups, resp, err := e.groupsClient.ListGroups(&gitlab.ListGroupsOptions{
		TopLevelOnly: ptr.To(true),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    page,
		},
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot list Gitlab groups")
	}
	if resp == nil {
		return nil, nil, errors.New("no response received when listing Gitlab groups")
	}

	return groups, resp, nil
}

func (e *external) getGroupPermissionStatus(groupID int64, serviceAccountID int64, minPermission int) (isMissingMembership bool, hasInsufficientPermissions bool, err error) {
	groupMember, resp, err := e.groupMemberClient.GetGroupMember(groupID, serviceAccountID)
	if err != nil {
		if clients.IsResponseNotFound(resp) {
			return true, false, nil
		}
		return false, false, errors.Wrap(err, "cannot check Gitlab group member")
	}

	if groupMember == nil {
		return false, true, nil
	}

	if int(groupMember.AccessLevel) < minPermission {
		return false, true, nil
	}

	return false, false, nil
}

// getAccessLevelValue is a helper function that returns the integer value of the given access level string. This is used to convert the BaselinePermissions field from a string to an integer that can be used with the Gitlab API.
func getAccessLevelValue(accessLevel string) int {
	accessLevelsMapping := map[string]int{
		accessLevelNoAccess:        accessLevelNoAccessValue,
		accessLevelMinimalAccess:   accessLevelMinimalAccessValue,
		accessLevelGuest:           accessLevelGuestValue,
		accessLevelPlanner:         accessLevelPlannerValue,
		accessLevelReporter:        accessLevelReporterValue,
		accessLevelSecurityManager: accessLevelSecurityManagerValue,
		accessLevelDeveloper:       accessLevelDeveloperValue,
		accessLevelMaintainer:      accessLevelMaintainerValue,
		accessLevelOwner:           accessLevelOwnerValue,
	}

	value, ok := accessLevelsMapping[accessLevel]
	if !ok {
		return accessLevelNoAccessValue
	}
	return value
}
