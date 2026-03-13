package projects

import (
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"k8s.io/utils/ptr"

	v1alpha1 "github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
)

const errProtectedEnvironmentNotFound = "404 Not found"

// ProtectedEnvironmentClient defines GitLab Protected Environments service operations.
type ProtectedEnvironmentClient interface {
	GetProtectedEnvironment(pid interface{}, name string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error)
	ProtectRepositoryEnvironments(pid interface{}, opt *gitlab.ProtectRepositoryEnvironmentsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error)
	UpdateProtectedEnvironments(pid interface{}, name string, opt *gitlab.UpdateProtectedEnvironmentsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedEnvironment, *gitlab.Response, error)
	UnprotectEnvironment(pid interface{}, name string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewProtectedEnvironmentClient returns a new GitLab Protected Environments client.
func NewProtectedEnvironmentClient(cfg common.Config) ProtectedEnvironmentClient {
	git := common.NewClient(cfg)
	return git.ProtectedEnvironments
}

// IsErrorProtectedEnvironmentNotFound helper function to test for errProtectedEnvironmentNotFound error.
func IsErrorProtectedEnvironmentNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errProtectedEnvironmentNotFound)
}

// LateInitializeProtectedEnvironment fills empty fields in spec with values from GitLab.
func LateInitializeProtectedEnvironment(in *v1alpha1.ProtectedEnvironmentParameters, pe *gitlab.ProtectedEnvironment) {
	if pe == nil || in == nil {
		return
	}
	if in.RequiredApprovalCount == nil && pe.RequiredApprovalCount != 0 {
		in.RequiredApprovalCount = ptr.To(pe.RequiredApprovalCount)
	}
}

// GenerateProtectedEnvironmentObservation builds status.atProvider from GitLab object.
func GenerateProtectedEnvironmentObservation(pe *gitlab.ProtectedEnvironment) v1alpha1.ProtectedEnvironmentObservation {
	if pe == nil {
		return v1alpha1.ProtectedEnvironmentObservation{}
	}

	out := v1alpha1.ProtectedEnvironmentObservation{
		Name:                  ptr.To(pe.Name),
		RequiredApprovalCount: ptr.To(pe.RequiredApprovalCount),
	}

	if len(pe.DeployAccessLevels) > 0 {
		dls := make([]v1alpha1.EnvironmentAccessLevelObservation, 0, len(pe.DeployAccessLevels))
		for _, dl := range pe.DeployAccessLevels {
			accessLevel := int(dl.AccessLevel)
			dls = append(dls, v1alpha1.EnvironmentAccessLevelObservation{
				ID:                     ptr.To(dl.ID),
				AccessLevel:            ptr.To(accessLevel),
				AccessLevelDescription: ptr.To(dl.AccessLevelDescription),
				UserID:                 ptr.To(dl.UserID),
				GroupID:                ptr.To(dl.GroupID),
				GroupInheritanceType:   ptr.To(dl.GroupInheritanceType),
			})
		}
		out.DeployAccessLevels = dls
	}

	if len(pe.ApprovalRules) > 0 {
		ars := make([]v1alpha1.EnvironmentApprovalRuleObservation, 0, len(pe.ApprovalRules))
		for _, ar := range pe.ApprovalRules {
			accessLevel := int(ar.AccessLevel)
			ars = append(ars, v1alpha1.EnvironmentApprovalRuleObservation{
				ID:                     ptr.To(ar.ID),
				UserID:                 ptr.To(ar.UserID),
				GroupID:                ptr.To(ar.GroupID),
				AccessLevel:            ptr.To(accessLevel),
				AccessLevelDescription: ptr.To(ar.AccessLevelDescription),
				RequiredApprovals:      ptr.To(ar.RequiredApprovalCount),
				GroupInheritanceType:   ptr.To(ar.GroupInheritanceType),
			})
		}
		out.ApprovalRules = ars
	}

	return out
}

// GenerateProtectRepositoryEnvironmentsOptions produces *gitlab.ProtectRepositoryEnvironmentsOptions from spec.
func GenerateProtectRepositoryEnvironmentsOptions(p *v1alpha1.ProtectedEnvironmentParameters) *gitlab.ProtectRepositoryEnvironmentsOptions {
	opt := &gitlab.ProtectRepositoryEnvironmentsOptions{}
	if p == nil {
		return opt
	}

	opt.Name = p.Name
	addDeployAccessLevels(p, opt)
	addApprovalRules(p, opt)

	// Apply unified approval count only when rules are explicitly managed and empty.
	if p.ApprovalRules != nil && len(*p.ApprovalRules) == 0 && p.RequiredApprovalCount != nil {
		opt.RequiredApprovalCount = p.RequiredApprovalCount
	}

	return opt
}

func isEmptySubject(accessLevel *int, userID *int64, groupID *int64) bool {
	return accessLevel == nil && userID == nil && groupID == nil
}

func accessLevelPtr(in *int) *gitlab.AccessLevelValue {
	if in == nil {
		return nil
	}
	v := gitlab.AccessLevelValue(*in)
	return &v
}

// addDeployAccessLevels fills DeployAccessLevels for create.
func addDeployAccessLevels(p *v1alpha1.ProtectedEnvironmentParameters, opt *gitlab.ProtectRepositoryEnvironmentsOptions) {
	if p == nil || opt == nil || p.DeployAccessLevels == nil || len(*p.DeployAccessLevels) == 0 {
		return
	}

	items := make([]*gitlab.EnvironmentAccessOptions, 0, len(*p.DeployAccessLevels))
	for _, s := range *p.DeployAccessLevels {
		if isEmptySubject(s.AccessLevel, s.UserID, s.GroupID) {
			continue
		}
		items = append(items, &gitlab.EnvironmentAccessOptions{
			AccessLevel:          accessLevelPtr(s.AccessLevel),
			UserID:               s.UserID,
			GroupID:              s.GroupID,
			GroupInheritanceType: s.GroupInheritanceType,
		})
	}

	if len(items) > 0 {
		opt.DeployAccessLevels = &items
	}
}

// addApprovalRules fills ApprovalRules for create.
func addApprovalRules(p *v1alpha1.ProtectedEnvironmentParameters, opt *gitlab.ProtectRepositoryEnvironmentsOptions) {
	if p == nil || opt == nil || p.ApprovalRules == nil || len(*p.ApprovalRules) == 0 {
		return
	}

	items := make([]*gitlab.EnvironmentApprovalRuleOptions, 0, len(*p.ApprovalRules))
	for _, s := range *p.ApprovalRules {
		if isEmptySubject(s.AccessLevel, s.UserID, s.GroupID) {
			continue
		}
		items = append(items, &gitlab.EnvironmentApprovalRuleOptions{
			AccessLevel:           accessLevelPtr(s.AccessLevel),
			UserID:                s.UserID,
			GroupID:               s.GroupID,
			RequiredApprovalCount: s.RequiredApprovals,
			GroupInheritanceType:  s.GroupInheritanceType,
		})
	}

	if len(items) > 0 {
		opt.ApprovalRules = &items
	}
}

func approvalCountUpToDate(p *v1alpha1.ProtectedEnvironmentParameters, pe *gitlab.ProtectedEnvironment) bool {
	if p == nil || pe == nil {
		return false
	}
	if p.ApprovalRules == nil || len(*p.ApprovalRules) != 0 || p.RequiredApprovalCount == nil {
		return true
	}
	return *p.RequiredApprovalCount == pe.RequiredApprovalCount
}

// IsProtectedEnvironmentUpToDate checks whether desired spec equals GitLab state.
func IsProtectedEnvironmentUpToDate(p *v1alpha1.ProtectedEnvironmentParameters, pe *gitlab.ProtectedEnvironment) bool {
	if p == nil || pe == nil {
		return false
	}

	if !approvalCountUpToDate(p, pe) {
		return false
	}

	if p.DeployAccessLevels != nil && !matchDeployAccessLevels(p.DeployAccessLevels, pe.DeployAccessLevels) {
		return false
	}

	if p.ApprovalRules != nil && !matchApprovalRules(p.ApprovalRules, pe.ApprovalRules) {
		return false
	}

	return true
}

func filterDeploySpecs(spec []v1alpha1.EnvironmentAccessLevelParameters) []v1alpha1.EnvironmentAccessLevelParameters {
	out := make([]v1alpha1.EnvironmentAccessLevelParameters, 0, len(spec))
	for _, s := range spec {
		if isEmptySubject(s.AccessLevel, s.UserID, s.GroupID) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func filterRuleSpecs(spec []v1alpha1.EnvironmentApprovalRuleParameters) []v1alpha1.EnvironmentApprovalRuleParameters {
	out := make([]v1alpha1.EnvironmentApprovalRuleParameters, 0, len(spec))
	for _, s := range spec {
		if isEmptySubject(s.AccessLevel, s.UserID, s.GroupID) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func sameAccessSubject(
	accessLevelSpec *int, userIDSpec *int64, groupIDSpec *int64,
	accessLevelGit gitlab.AccessLevelValue, userIDGit int64, groupIDGit int64,
) bool {
	switch {
	case userIDSpec != nil:
		return *userIDSpec == userIDGit
	case groupIDSpec != nil:
		return *groupIDSpec == groupIDGit
	case accessLevelSpec != nil:
		return *accessLevelSpec == int(accessLevelGit)
	default:
		return false
	}
}

func deployAccessMatchesSpec(s v1alpha1.EnvironmentAccessLevelParameters, g *gitlab.EnvironmentAccessDescription) bool {
	return sameAccessSubject(s.AccessLevel, s.UserID, s.GroupID, g.AccessLevel, g.UserID, g.GroupID) &&
		clients.IsInt64EqualToInt64Ptr(s.GroupInheritanceType, g.GroupInheritanceType)
}

func approvalRuleEqualsSpec(s v1alpha1.EnvironmentApprovalRuleParameters, g *gitlab.EnvironmentApprovalRule) bool {
	return sameAccessSubject(s.AccessLevel, s.UserID, s.GroupID, g.AccessLevel, g.UserID, g.GroupID) &&
		clients.IsInt64EqualToInt64Ptr(s.GroupInheritanceType, g.GroupInheritanceType) &&
		clients.IsInt64EqualToInt64Ptr(s.RequiredApprovals, g.RequiredApprovalCount)
}

func approvalRuleIdentityMatchesSpec(s v1alpha1.EnvironmentApprovalRuleParameters, g *gitlab.EnvironmentApprovalRule) bool {
	return sameAccessSubject(s.AccessLevel, s.UserID, s.GroupID, g.AccessLevel, g.UserID, g.GroupID) &&
		clients.IsInt64EqualToInt64Ptr(s.GroupInheritanceType, g.GroupInheritanceType)
}

func matchDeployAccessLevels(spec *[]v1alpha1.EnvironmentAccessLevelParameters, got []*gitlab.EnvironmentAccessDescription) bool {
	desired := filterDeploySpecs(*spec)
	if len(desired) != len(got) {
		return false
	}

	used := make([]bool, len(got))
	for _, s := range desired {
		found := false
		for i, g := range got {
			if used[i] {
				continue
			}
			if deployAccessMatchesSpec(s, g) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func matchApprovalRules(spec *[]v1alpha1.EnvironmentApprovalRuleParameters, got []*gitlab.EnvironmentApprovalRule) bool {
	desired := filterRuleSpecs(*spec)
	if len(desired) != len(got) {
		return false
	}

	used := make([]bool, len(got))
	for _, s := range desired {
		found := false
		for i, g := range got {
			if used[i] {
				continue
			}
			if approvalRuleEqualsSpec(s, g) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GenerateUpdateProtectedEnvironmentsOptions builds the delta; returns nil if there are no changes.
func GenerateUpdateProtectedEnvironmentsOptions(
	p *v1alpha1.ProtectedEnvironmentParameters,
	pe *gitlab.ProtectedEnvironment,
) *gitlab.UpdateProtectedEnvironmentsOptions {
	if p == nil || pe == nil {
		return nil
	}

	var changed bool
	u := &gitlab.UpdateProtectedEnvironmentsOptions{}

	deltasDeploy := buildDeployAccessLevelsDelta(p.DeployAccessLevels, pe.DeployAccessLevels)
	if len(deltasDeploy) > 0 {
		u.DeployAccessLevels = &deltasDeploy
		changed = true
	}

	deltasRules := buildApprovalRulesDelta(p.ApprovalRules, pe.ApprovalRules)
	if len(deltasRules) > 0 {
		u.ApprovalRules = &deltasRules
		changed = true
	}

	// Unified approvals: only when rules are explicitly managed and empty.
	if p.ApprovalRules != nil && len(*p.ApprovalRules) == 0 && p.RequiredApprovalCount != nil {
		if *p.RequiredApprovalCount != pe.RequiredApprovalCount {
			u.RequiredApprovalCount = p.RequiredApprovalCount
			changed = true
		}
	}

	if !changed {
		return nil
	}
	return u
}

func buildDeployAccessLevelsDelta(
	spec *[]v1alpha1.EnvironmentAccessLevelParameters,
	got []*gitlab.EnvironmentAccessDescription,
) []*gitlab.UpdateEnvironmentAccessOptions {
	// nil => unmanaged => no ops
	if spec == nil {
		return nil
	}

	desired := filterDeploySpecs(*spec)
	out := make([]*gitlab.UpdateEnvironmentAccessOptions, 0)
	used := make([]bool, len(got))

	for _, s := range desired {
		matched := -1
		for i, g := range got {
			if used[i] {
				continue
			}
			if deployAccessMatchesSpec(s, g) {
				matched = i
				break
			}
		}

		if matched >= 0 {
			used[matched] = true
			continue
		}

		out = append(out, &gitlab.UpdateEnvironmentAccessOptions{
			AccessLevel:          accessLevelPtr(s.AccessLevel),
			UserID:               s.UserID,
			GroupID:              s.GroupID,
			GroupInheritanceType: s.GroupInheritanceType,
		})
	}

	for i, g := range got {
		if used[i] {
			continue
		}
		out = append(out, &gitlab.UpdateEnvironmentAccessOptions{
			ID:      ptr.To(g.ID),
			Destroy: ptr.To(true),
		})
	}

	return out
}

//nolint:gocyclo
func buildApprovalRulesDelta(
	spec *[]v1alpha1.EnvironmentApprovalRuleParameters,
	got []*gitlab.EnvironmentApprovalRule,
) []*gitlab.UpdateEnvironmentApprovalRuleOptions {
	// nil => unmanaged => no ops
	if spec == nil {
		return nil
	}

	desired := filterRuleSpecs(*spec)
	out := make([]*gitlab.UpdateEnvironmentApprovalRuleOptions, 0)
	used := make([]bool, len(got))

	for _, s := range desired {
		matched := -1
		for i, g := range got {
			if used[i] {
				continue
			}
			// Match by identity (subject + inheritance), so RequiredApprovals changes become updates.
			if approvalRuleIdentityMatchesSpec(s, g) {
				matched = i
				break
			}
		}

		if matched >= 0 {
			cur := got[matched]

			// If required approvals are equal, this is a no-op.
			if clients.IsInt64EqualToInt64Ptr(s.RequiredApprovals, cur.RequiredApprovalCount) {
				used[matched] = true
				continue
			}

			out = append(out, &gitlab.UpdateEnvironmentApprovalRuleOptions{
				ID:                    ptr.To(cur.ID),
				AccessLevel:           accessLevelPtr(s.AccessLevel),
				UserID:                s.UserID,
				GroupID:               s.GroupID,
				RequiredApprovalCount: s.RequiredApprovals,
				GroupInheritanceType:  s.GroupInheritanceType,
			})
			used[matched] = true
			continue
		}

		out = append(out, &gitlab.UpdateEnvironmentApprovalRuleOptions{
			AccessLevel:           accessLevelPtr(s.AccessLevel),
			UserID:                s.UserID,
			GroupID:               s.GroupID,
			RequiredApprovalCount: s.RequiredApprovals,
			GroupInheritanceType:  s.GroupInheritanceType,
		})
	}

	for i, g := range got {
		if used[i] {
			continue
		}

		u := &gitlab.UpdateEnvironmentApprovalRuleOptions{
			ID:      ptr.To(g.ID),
			Destroy: ptr.To(true),
		}

		if g.GroupInheritanceType != 0 {
			u.GroupInheritanceType = ptr.To(g.GroupInheritanceType)
		}

		switch {
		case g.UserID != 0:
			u.UserID = ptr.To(g.UserID)
		case g.GroupID != 0:
			u.GroupID = ptr.To(g.GroupID)
		default:
			u.AccessLevel = ptr.To(g.AccessLevel)
		}

		out = append(out, u)
	}

	return out
}
