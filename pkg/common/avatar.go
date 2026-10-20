/*
Copyright 2026 The Crossplane Authors.

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

package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

const errUnsupportedAvatar = "unsupported avatar image type %q, use PNG, JPEG, GIF, BMP, WEBP or ICO"

// avatarExtensions maps the detected content type of an avatar image to the
// file extension GitLab requires to accept the upload.
var avatarExtensions = map[string]string{
	"image/png":    "png",
	"image/jpeg":   "jpg",
	"image/gif":    "gif",
	"image/bmp":    "bmp",
	"image/webp":   "webp",
	"image/x-icon": "ico",
}

// avatarFilename returns the name "<sha256>.<ext>" an avatar image is uploaded
// under. GitLab only returns the avatar URL, and that URL keeps the uploaded
// filename, so the name is what ties the URL back to the image.
func avatarFilename(image []byte) (string, error) {
	contentType := http.DetectContentType(image)
	ext, ok := avatarExtensions[contentType]
	if !ok {
		return "", errors.Errorf(errUnsupportedAvatar, contentType)
	}
	sum := sha256.Sum256(image)
	return hex.EncodeToString(sum[:]) + "." + ext, nil
}

// IsAvatarUpToDate reports whether avatarURL shows image. An empty image means
// the avatar is not managed, which is always up to date.
func IsAvatarUpToDate(image []byte, avatarURL string) bool {
	if len(image) == 0 {
		return true
	}
	name, err := avatarFilename(image)
	if err != nil {
		return false
	}
	u, err := url.Parse(avatarURL)
	if err != nil {
		return false
	}
	return path.Base(u.Path) == name
}

// AvatarUpload returns the upload that makes avatarURL show image, or nil when
// avatarURL already shows it, matching the nil-means-unchanged contract of
// gitlab.ModifyUserOptions.Avatar.
func AvatarUpload(image []byte, avatarURL string) (*gitlab.UserAvatar, error) {
	if IsAvatarUpToDate(image, avatarURL) {
		return nil, nil
	}
	name, err := avatarFilename(image)
	if err != nil {
		return nil, err
	}
	return &gitlab.UserAvatar{Filename: name, Image: bytes.NewReader(image)}, nil
}
