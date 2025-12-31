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

package users

import (
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const (
	errFetchFailed = "can not fetch userID by userName"
	errPullUserID  = "cant determine user by userName. Amount of users received: %v"
)

// UserClient defines Gitlab User service operations
type UserClient interface {
	ListUsers(opt *gitlab.ListUsersOptions, options ...gitlab.RequestOptionFunc) ([]*gitlab.User, *gitlab.Response, error)
}

// NewUserClient returns a new Gitlab User service
func NewUserClient(cfg common.Config) UserClient {
	git := common.NewClient(cfg)
	return git.Users
}

// GetUserID gets Gitlab userID by Gitlab username
func GetUserID(git UserClient, username string) (*int64, error) {
	userOptions := gitlab.ListUsersOptions{Username: &username}
	userArr, _, err := git.ListUsers(&userOptions)
	if err != nil {
		return nil, errors.Wrap(err, errFetchFailed)
	}
	if len(userArr) == 0 || len(userArr) > 1 {
		return nil, errors.Errorf(errPullUserID, len(userArr))
	}

	pulledUserID := userArr[0].ID

	return &pulledUserID, nil
}
