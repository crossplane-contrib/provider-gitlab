package projects

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	v1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
)

func TestFilterRuleSpecs_DropsEmptySubjects(t *testing.T) {

	in := []v1alpha1.EnvironmentApprovalRuleParameters{

		{}, // empty

		{AccessLevel: ptr.To(40)},

		{UserID: ptr.To(int64(1))},

		{GroupID: ptr.To(int64(2))},
	}

	out := filterRuleSpecs(in)

	if len(out) != 3 {

		t.Fatalf("expected 3 entries, got %d: %#v", len(out), out)

	}

}

func TestFilterDeploySpecs_DropsEmptySubjects(t *testing.T) {

	in := []v1alpha1.EnvironmentAccessLevelParameters{

		{}, // empty

		{AccessLevel: ptr.To(20)},

		{UserID: ptr.To(int64(1))},

		{GroupID: ptr.To(int64(2))},
	}

	out := filterDeploySpecs(in)

	if len(out) != 3 {

		t.Fatalf("expected 3 entries, got %d: %#v", len(out), out)

	}

}

func TestMatchApprovalRules_OrderInsensitive(t *testing.T) {

	spec := []v1alpha1.EnvironmentApprovalRuleParameters{

		{AccessLevel: ptr.To(40), RequiredApprovals: ptr.To(int64(1))},

		{AccessLevel: ptr.To(30), RequiredApprovals: ptr.To(int64(2))},
	}

	got := []*gitlab.EnvironmentApprovalRule{

		{ID: 2, AccessLevel: 30, RequiredApprovalCount: 2, GroupInheritanceType: 0},

		{ID: 1, AccessLevel: 40, RequiredApprovalCount: 1, GroupInheritanceType: 0},
	}

	if !matchApprovalRules(&spec, got) {

		t.Fatalf("expected match=true")

	}

}

func TestMatchDeployAccessLevels_OrderInsensitive(t *testing.T) {

	spec := []v1alpha1.EnvironmentAccessLevelParameters{

		{AccessLevel: ptr.To(20)},

		{AccessLevel: ptr.To(40)},
	}

	got := []*gitlab.EnvironmentAccessDescription{

		{ID: 2, AccessLevel: gitlab.AccessLevelValue(40), GroupInheritanceType: 0},

		{ID: 1, AccessLevel: gitlab.AccessLevelValue(20), GroupInheritanceType: 0},
	}

	if !matchDeployAccessLevels(&spec, got) {

		t.Fatalf("expected match=true")

	}

}

func TestBuildApprovalRulesDelta_UnmanagedNilSpec(t *testing.T) {

	got := []*gitlab.EnvironmentApprovalRule{

		{ID: 1, AccessLevel: 40, RequiredApprovalCount: 1},
	}

	out := buildApprovalRulesDelta(nil, got)

	if out != nil {

		t.Fatalf("expected nil, got %#v", out)

	}

}

func TestBuildApprovalRulesDelta_EmptySpecDeletesAll_WithSubject(t *testing.T) {

	spec := []v1alpha1.EnvironmentApprovalRuleParameters{}

	got := []*gitlab.EnvironmentApprovalRule{

		{ID: 10, AccessLevel: 40, UserID: 0, GroupID: 0, GroupInheritanceType: 0},

		{ID: 11, AccessLevel: 30, UserID: 0, GroupID: 0, GroupInheritanceType: 0},
	}

	out := buildApprovalRulesDelta(&spec, got)

	if len(out) != 2 {

		t.Fatalf("expected 2 deltas, got %d: %#v", len(out), out)

	}

	for _, d := range out {

		if d.ID == nil {

			t.Fatalf("expected ID set, got %#v", d)

		}

		if d.Destroy == nil || *d.Destroy != true {

			t.Fatalf("expected destroy=true, got %#v", d)

		}

		// GitLab требует subject даже на destroy: access_level | user_id | group_id

		if d.AccessLevel == nil && d.UserID == nil && d.GroupID == nil {

			t.Fatalf("expected subject for destroy entry, got %#v", d)

		}

	}

}

func TestBuildApprovalRulesDelta_UpdateRequiredApprovalsByIdentity(t *testing.T) {

	spec := []v1alpha1.EnvironmentApprovalRuleParameters{

		{AccessLevel: ptr.To(40), RequiredApprovals: ptr.To(int64(3))},
	}

	got := []*gitlab.EnvironmentApprovalRule{

		{ID: 10, AccessLevel: 40, RequiredApprovalCount: 1, GroupInheritanceType: 0},
	}

	out := buildApprovalRulesDelta(&spec, got)

	if len(out) != 1 {

		t.Fatalf("expected 1 delta, got %d: %#v", len(out), out)

	}

	d := out[0]

	if d.ID == nil || *d.ID != 10 {

		t.Fatalf("expected update by ID=10, got %#v", d)

	}

	if d.RequiredApprovalCount == nil || *d.RequiredApprovalCount != 3 {

		t.Fatalf("expected required approvals=3, got %#v", d)

	}

}

func TestBuildApprovalRulesDelta_ReplaceSubject_DeleteOldCreateNew(t *testing.T) {

	spec := []v1alpha1.EnvironmentApprovalRuleParameters{

		{UserID: ptr.To(int64(123)), RequiredApprovals: ptr.To(int64(1))},
	}

	got := []*gitlab.EnvironmentApprovalRule{

		{ID: 10, AccessLevel: 40, GroupInheritanceType: 0},
	}

	out := buildApprovalRulesDelta(&spec, got)

	if len(out) != 2 {

		t.Fatalf("expected 2 deltas (create+destroy), got %d: %#v", len(out), out)

	}

	var sawCreate, sawDestroy bool

	for _, d := range out {

		if d.Destroy != nil && *d.Destroy {

			sawDestroy = true

			if d.ID == nil || *d.ID != 10 {

				t.Fatalf("expected destroy of ID=10, got %#v", d)

			}

			if d.AccessLevel == nil && d.UserID == nil && d.GroupID == nil {

				t.Fatalf("expected subject for destroy entry, got %#v", d)

			}

			continue

		}

		sawCreate = true

		if d.ID != nil {

			t.Fatalf("expected create without ID, got %#v", d)

		}

		if d.UserID == nil || *d.UserID != 123 {

			t.Fatalf("expected create userID=123, got %#v", d)

		}

	}

	if !sawCreate || !sawDestroy {

		t.Fatalf("expected both create and destroy, got %#v", out)

	}

}

func TestBuildDeployAccessLevelsDelta_UnmanagedNilSpec(t *testing.T) {

	got := []*gitlab.EnvironmentAccessDescription{

		{ID: 1, AccessLevel: gitlab.AccessLevelValue(20)},
	}

	out := buildDeployAccessLevelsDelta(nil, got)

	if out != nil {

		t.Fatalf("expected nil, got %#v", out)

	}

}

func TestBuildDeployAccessLevelsDelta_EmptySpecDeletesAll(t *testing.T) {

	spec := []v1alpha1.EnvironmentAccessLevelParameters{}

	got := []*gitlab.EnvironmentAccessDescription{

		{ID: 1, AccessLevel: gitlab.AccessLevelValue(20)},

		{ID: 2, AccessLevel: gitlab.AccessLevelValue(40)},
	}

	out := buildDeployAccessLevelsDelta(&spec, got)

	if len(out) != 2 {

		t.Fatalf("expected 2 deltas, got %d: %#v", len(out), out)

	}

	for _, d := range out {

		if d.ID == nil {

			t.Fatalf("expected ID set, got %#v", d)

		}

		if d.Destroy == nil || *d.Destroy != true {

			t.Fatalf("expected destroy=true, got %#v", d)

		}

	}

}

func TestGenerateUpdateProtectedEnvironmentsOptions_NoChangesReturnsNil(t *testing.T) {

	als := []v1alpha1.EnvironmentAccessLevelParameters{{AccessLevel: ptr.To(20)}}

	ars := []v1alpha1.EnvironmentApprovalRuleParameters{{AccessLevel: ptr.To(40), RequiredApprovals: ptr.To(int64(1))}}

	p := &v1alpha1.ProtectedEnvironmentParameters{

		DeployAccessLevels: &als,

		ApprovalRules: &ars,
	}

	pe := &gitlab.ProtectedEnvironment{

		DeployAccessLevels: []*gitlab.EnvironmentAccessDescription{

			{ID: 1, AccessLevel: gitlab.AccessLevelValue(20), GroupInheritanceType: 0},
		},

		ApprovalRules: []*gitlab.EnvironmentApprovalRule{

			{ID: 2, AccessLevel: 40, RequiredApprovalCount: 1, GroupInheritanceType: 0},
		},

		RequiredApprovalCount: 0,
	}

	u := GenerateUpdateProtectedEnvironmentsOptions(p, pe)

	if u != nil {

		t.Fatalf("expected nil update options, got %#v", u)

	}

}

func TestGenerateUpdateProtectedEnvironmentsOptions_ApprovalCountOnlyWhenRulesExplicitEmpty(t *testing.T) {

	emptyRules := []v1alpha1.EnvironmentApprovalRuleParameters{}

	p := &v1alpha1.ProtectedEnvironmentParameters{

		ApprovalRules: &emptyRules,

		RequiredApprovalCount: ptr.To(int64(2)),
	}

	pe := &gitlab.ProtectedEnvironment{RequiredApprovalCount: 1}

	u := GenerateUpdateProtectedEnvironmentsOptions(p, pe)

	if u == nil || u.RequiredApprovalCount == nil || *u.RequiredApprovalCount != 2 {

		t.Fatalf("expected requiredApprovalCount update to 2, got %#v", u)

	}

}

func TestSameAccessSubject_PrefersUserOverGroupOverAccessLevel(t *testing.T) {

	if !sameAccessSubject(ptr.To(40), ptr.To(int64(1)), ptr.To(int64(2)), 40, 1, 2) {

		t.Fatalf("expected match by user")

	}

	if !sameAccessSubject(ptr.To(40), nil, ptr.To(int64(2)), 40, 99, 2) {

		t.Fatalf("expected match by group")

	}

	if !sameAccessSubject(ptr.To(40), nil, nil, 40, 99, 98) {

		t.Fatalf("expected match by access level")

	}

	if sameAccessSubject(nil, nil, nil, 40, 0, 0) {

		t.Fatalf("expected false for empty subject")

	}

}

func TestFilterRuleSpecs_StableOutput(t *testing.T) {

	in := []v1alpha1.EnvironmentApprovalRuleParameters{

		{UserID: ptr.To(int64(1))},

		{},

		{AccessLevel: ptr.To(40)},
	}

	out := filterRuleSpecs(in)

	want := []v1alpha1.EnvironmentApprovalRuleParameters{

		{UserID: ptr.To(int64(1))},

		{AccessLevel: ptr.To(40)},
	}

	if diff := cmp.Diff(want, out); diff != "" {

		t.Fatalf("unexpected diff: %s", diff)

	}

}
