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

package groups

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/groups/v1alpha1"
)

func TestGenerateAddEditOptionsAndObservation(t *testing.T) {
	name := "n"
	img := "img"
	link := "link"
	p := &v1alpha1.BadgeParameters{
		Name:     &name,
		ImageURL: &img,
		LinkURL:  &link,
	}

	add := GenerateAddGroupBadgeOptions(p)
	if add == nil || *add.Name != name || *add.ImageURL != img || *add.LinkURL != link {
		t.Fatalf("GenerateAddGroupBadgeOptions did not produce expected values")
	}

	edit := GenerateEditGroupBadgeOptions(p)
	if edit == nil || *edit.Name != name || *edit.ImageURL != img || *edit.LinkURL != link {
		t.Fatalf("GenerateEditGroupBadgeOptions did not produce expected values")
	}

	// observation
	b := &gitlab.GroupBadge{ID: 5, LinkURL: link, RenderedLinkURL: link, ImageURL: img, RenderedImageURL: img, Name: name}
	obs := GenerateBadgeObservation(b)
	want := v1alpha1.BadgeObservation{ID: 5, LinkURL: link, RenderedLinkURL: link, ImageURL: img, RenderedImageURL: img, Name: name}
	if diff := cmp.Diff(want, obs); diff != "" {
		t.Fatalf("GenerateBadgeObservation() mismatch (-want +got):\n%s", diff)
	}
}

func TestIsErrorGroupBadgeNotFound(t *testing.T) {
	if !IsErrorGroupBadgeNotFound(errors.New(errGroupNotFound)) {
		t.Fatalf("expected IsErrorGroupBadgeNotFound to return true for %s", errGroupNotFound)
	}

	// nil error should return false
	if IsErrorGroupBadgeNotFound(nil) {
		t.Fatalf("expected IsErrorGroupBadgeNotFound to return false for nil error")
	}

	// other errors should return false
	if IsErrorGroupBadgeNotFound(errors.New("some other error")) {
		t.Fatalf("expected IsErrorGroupBadgeNotFound to return false for some other error")
	}
}
