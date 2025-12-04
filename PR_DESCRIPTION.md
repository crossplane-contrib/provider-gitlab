# Add `publicJobs` field and deprecate `publicBuilds` for Project resources

## Summary

This PR adds the `publicJobs` field to Project resources and marks `publicBuilds` as deprecated, aligning with GitLab's API naming conventions. The implementation maintains full backward compatibility while providing a clear migration path for users.

## Motivation

GitLab's API uses `public_jobs` to control pipeline visibility, but this provider historically used `publicBuilds`. This inconsistency creates confusion and doesn't align with GitLab's current terminology (GitLab deprecated "builds" in favor of "jobs" terminology).

## Changes

### API Changes

- **Added** `publicJobs` field to `ProjectParameters` in both namespaced and cluster-scoped APIs
- **Marked** `publicBuilds` as `// DEPRECATED` in Go code comments
- Both fields are `*bool` to support tri-state (unset/true/false)

### Core Implementation

#### 1. Precedence Logic (`pkg/common/helper.go`)

```go
func ResolvePublicJobsSetting(publicBuilds, publicJobs *bool) (*bool, bool)
```

- `publicJobs` takes precedence when both fields are set
- Returns the resolved value and whether the deprecated field was used
- Includes comprehensive unit tests (8 test cases)

#### 2. Late Initialization

Projects now synchronize both fields during late initialization:

- If neither field is set: both initialized from GitLab's `PublicJobs` value
- If only `publicBuilds` is set: synced to `publicJobs`
- If only `publicJobs` is set: synced to `publicBuilds` (for backward compat)
- If both are set: no synchronization needed

#### 3. Update Logic

The `isProjectUpToDate` function now uses `ResolvePublicJobsSetting()` to determine the effective value, ensuring precedence rules are respected during updates.

#### 4. Create/Edit Operations

Both cluster-scoped and namespaced controllers properly handle both fields when creating or updating GitLab projects.

### Documentation

- Updated `README.md` with deprecation notice
- Added migration guide with examples
- Documented precedence rules

### Testing

All E2E tests passing:

| Scenario               | User Input                              | GitLab Value         | Status |
|------------------------|-----------------------------------------|----------------------|--------|
| New field only         | `publicJobs: true`                      | `public_jobs: true`  | ✅      |
| New field only         | `publicJobs: false`                     | `public_jobs: false` | ✅      |
| Backward compatibility | `publicBuilds: true`                    | `public_jobs: true`  | ✅      |
| Precedence test        | `publicBuilds: true, publicJobs: false` | `public_jobs: false` | ✅      |
| Migration scenario     | `publicBuilds: true, publicJobs: true`  | `public_jobs: true`  | ✅      |

## Migration Guide

### For Existing Users

If you're currently using `publicBuilds`:

```yaml
# Before
spec:
  forProvider:
    publicBuilds: true

# After - recommended
spec:
  forProvider:
    publicJobs: true

# During migration - both work
spec:
  forProvider:
    publicBuilds: true  # Will be synced to publicJobs
    publicJobs: true    # Takes precedence
```

### Timeline

1. **Now**: Both fields work, `publicBuilds` marked as deprecated
2. **In the future**: Remove `publicBuilds` field entirely

## Backward Compatibility

✅ **Fully backward compatible**

- Existing manifests using `publicBuilds` continue to work
- Late initialization automatically syncs both fields
- No user action required immediately

## Breaking Changes

None. This is a non-breaking change.

## Dependencies

- Updated `gitlab.com/gitlab-org/api/client-go` from v0.137.0 to v0.160.0
  - **Why not v1.x?** Version 1.0.0+ introduced breaking changes, primarily around `ListOptions` changing from type aliasing to composition. This would require updating all list operations across the provider.
  - v0.160.0 is the latest stable v0.x release and provides all features needed for this PR
  
### Note on GitLab Client's `PublicJobs` Field

The GitLab Go client has an inconsistency:
- **`CreateProjectOptions`**: Only has `PublicBuilds` (marked as deprecated with comment "use PublicJobs instead", but `PublicJobs` field doesn't exist in this struct)
- **`EditProjectOptions`**: Has both `PublicJobs` and `PublicBuilds` (both functional)

This means:
- When **creating** projects: We must use `PublicBuilds` field (no alternative available)
- When **updating** projects: We use `PublicJobs` field (the non-deprecated option)

Our implementation uses `resolvePublicJobsValue()` to determine the effective value and maps it to the appropriate field for each operation.

## Checklist

- [x] API types updated (namespaced and cluster-scoped)
- [x] Late initialization logic implemented
- [x] Precedence logic with unit tests
- [x] IsUpToDate comparison updated
- [x] Create/Edit operations handle both fields
- [x] Code generation run for cluster-scoped resources
- [x] Documentation updated (README)
- [x] E2E tests passing (5 scenarios verified)
- [x] Backward compatibility verified

## Related Issues

Fixes #[issue-number] (if applicable)

## Additional Notes

- The deprecation follows Crossplane's best practices for field deprecation
- No warning messages in status to avoid confusion (late init handles migration transparently)
- The implementation pattern can be reused for future API field migrations
