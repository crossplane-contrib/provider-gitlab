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
	"io"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

var (
	testPNG  = []byte("\x89PNG\r\n\x1a\n")
	testJPEG = []byte("\xff\xd8\xff")
)

const (
	testPNGName       = "4c4b6a3be1314ab86138bef4314dde022e600960d8689a2c8f8631802d20dab6.png"
	testJPEGName      = "6e568e1f67fba258184c78181539e5e8fdee447e49bb706fc0ea34fbf12336a5.jpg"
	testPNGAvatarURL  = "https://gitlab.example.com/uploads/-/system/user/avatar/7/" + testPNGName + "?v=1791979682"
	testGravatarURL   = "https://secure.gravatar.com/avatar/3c9a0d1d?s=80&d=identicon"
	testUnsupportedCT = "text/plain; charset=utf-8"
)

func TestIsAvatarUpToDate(t *testing.T) {
	cases := map[string]struct {
		image     []byte
		avatarURL string
		want      bool
	}{
		"Unmanaged":         {image: nil, avatarURL: testGravatarURL, want: true},
		"SameImage":         {image: testPNG, avatarURL: testPNGAvatarURL, want: true},
		"DefaultAvatar":     {image: testPNG, avatarURL: testGravatarURL, want: false},
		"OtherImage":        {image: testJPEG, avatarURL: testPNGAvatarURL, want: false},
		"NoAvatarURL":       {image: testPNG, avatarURL: "", want: false},
		"UnsupportedFormat": {image: []byte("not an image"), avatarURL: testPNGAvatarURL, want: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsAvatarUpToDate(tc.image, tc.avatarURL); got != tc.want {
				t.Errorf("IsAvatarUpToDate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAvatarUpload(t *testing.T) {
	type want struct {
		filename string
		image    []byte
		err      error
	}

	cases := map[string]struct {
		image     []byte
		avatarURL string
		want      *want
	}{
		"Unmanaged":   {image: nil, avatarURL: testGravatarURL, want: nil},
		"UpToDate":    {image: testPNG, avatarURL: testPNGAvatarURL, want: nil},
		"ChangedPNG":  {image: testPNG, avatarURL: testGravatarURL, want: &want{filename: testPNGName, image: testPNG}},
		"ChangedJPEG": {image: testJPEG, avatarURL: testPNGAvatarURL, want: &want{filename: testJPEGName, image: testJPEG}},
		"UnsupportedFormat": {
			image:     []byte("not an image"),
			avatarURL: testGravatarURL,
			want:      &want{err: errors.Errorf(errUnsupportedAvatar, testUnsupportedCT)},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			upload, err := AvatarUpload(tc.image, tc.avatarURL)
			var got *want
			if upload != nil || err != nil {
				got = &want{err: err}
			}
			if upload != nil {
				image, readErr := io.ReadAll(upload.Image)
				if readErr != nil {
					t.Fatalf("reading upload: %v", readErr)
				}
				got.filename = upload.Filename
				got.image = image
			}
			if diff := cmp.Diff(tc.want, got, test.EquateErrors(), cmp.AllowUnexported(want{})); diff != "" {
				t.Errorf("AvatarUpload(): -want, +got:\n%s", diff)
			}
		})
	}
}
