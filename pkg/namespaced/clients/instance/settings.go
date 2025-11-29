package instance

import (
	"strings"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// ApplicationSettingsClient defines Gitlab Application Settings service operations
type ApplicationSettingsClient interface {
	GetSettings(options ...gitlab.RequestOptionFunc) (*gitlab.Settings, *gitlab.Response, error)
	UpdateSettings(opt *gitlab.UpdateSettingsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Settings, *gitlab.Response, error)
}

// NewApplicationSettingsClient returns a new Gitlab Application Settings service
func NewApplicationSettingsClient(cfg common.Config) gitlab.SettingsServiceInterface {
	git := common.NewClient(cfg)
	return git.Settings
}

// GenerateUpdateApplicationSettingsOptions generates GitLab Settings update options from the desired state
func GenerateUpdateApplicationSettingsOptions(p *v1alpha1.ApplicationSettingsParameters) *gitlab.UpdateSettingsOptions {
	return &gitlab.UpdateSettingsOptions{
		AbuseNotificationEmail:                                p.AbuseNotificationEmail,
		AdminMode:                                             p.AdminMode,
		AfterSignOutPath:                                      p.AfterSignOutPath,
		AfterSignUpText:                                       p.AfterSignUpText,
		AkismetAPIKey:                                         p.AkismetAPIKey,
		AkismetEnabled:                                        p.AkismetEnabled,
		AllowAccountDeletion:                                  p.AllowAccountDeletion,
		AllowAllIntegrations:                                  p.AllowAllIntegrations,
		AllowedIntegrations:                                   p.AllowedIntegrations,
		AllowGroupOwnersToManageLDAP:                          p.AllowGroupOwnersToManageLDAP,
		AllowLocalRequestsFromSystemHooks:                     p.AllowLocalRequestsFromSystemHooks,
		AllowLocalRequestsFromWebHooksAndServices:             p.AllowLocalRequestsFromWebHooksAndServices,
		AllowProjectCreationForGuestAndBelow:                  p.AllowProjectCreationForGuestAndBelow,
		AllowRunnerRegistrationToken:                          p.AllowRunnerRegistrationToken,
		ArchiveBuildsInHumanReadable:                          p.ArchiveBuildsInHumanReadable,
		ASCIIDocMaxIncludes:                                   p.ASCIIDocMaxIncludes,
		AssetProxyAllowlist:                                   p.AssetProxyAllowlist,
		AssetProxyEnabled:                                     p.AssetProxyEnabled,
		AssetProxySecretKey:                                   p.AssetProxySecretKey,
		AssetProxyURL:                                         p.AssetProxyURL,
		AuthorizedKeysEnabled:                                 p.AuthorizedKeysEnabled,
		AutoBanUserOnExcessiveProjectsDownload:                p.AutoBanUserOnExcessiveProjectsDownload,
		AutocompleteUsers:                                     p.AutocompleteUsers,
		AutocompleteUsersUnauthenticated:                      p.AutocompleteUsersUnauthenticated,
		AutoDevOpsDomain:                                      p.AutoDevOpsDomain,
		AutoDevOpsEnabled:                                     p.AutoDevOpsEnabled,
		AutomaticPurchasedStorageAllocation:                   p.AutomaticPurchasedStorageAllocation,
		BulkImportConcurrentPipelineBatchLimit:                p.BulkImportConcurrentPipelineBatchLimit,
		BulkImportEnabled:                                     p.BulkImportEnabled,
		BulkImportMaxDownloadFileSize:                         p.BulkImportMaxDownloadFileSize,
		CanCreateGroup:                                        p.CanCreateGroup,
		CheckNamespacePlan:                                    p.CheckNamespacePlan,
		CIJobLiveTraceEnabled:                                 p.CIJobLiveTraceEnabled,
		CIMaxIncludes:                                         p.CIMaxIncludes,
		CIMaxTotalYAMLSizeBytes:                               p.CIMaxTotalYAMLSizeBytes,
		CIPartitionsSizeLimit:                                 p.CIPartitionsSizeLimit,
		CommitEmailHostname:                                   p.CommitEmailHostname,
		ConcurrentBitbucketImportJobsLimit:                    p.ConcurrentBitbucketImportJobsLimit,
		ConcurrentBitbucketServerImportJobsLimit:              p.ConcurrentBitbucketServerImportJobsLimit,
		ConcurrentGitHubImportJobsLimit:                       p.ConcurrentGitHubImportJobsLimit,
		ContainerExpirationPoliciesEnableHistoricEntries:      p.ContainerExpirationPoliciesEnableHistoricEntries,
		ContainerRegistryCleanupTagsServiceMaxListSize:        p.ContainerRegistryCleanupTagsServiceMaxListSize,
		ContainerRegistryDeleteTagsServiceTimeout:             p.ContainerRegistryDeleteTagsServiceTimeout,
		ContainerRegistryExpirationPoliciesCaching:            p.ContainerRegistryExpirationPoliciesCaching,
		ContainerRegistryExpirationPoliciesWorkerCapacity:     p.ContainerRegistryExpirationPoliciesWorkerCapacity,
		ContainerRegistryImportCreatedBefore:                  p.ContainerRegistryImportCreatedBefore,
		ContainerRegistryImportMaxRetries:                     p.ContainerRegistryImportMaxRetries,
		ContainerRegistryImportMaxStepDuration:                p.ContainerRegistryImportMaxStepDuration,
		ContainerRegistryImportMaxTagsCount:                   p.ContainerRegistryImportMaxTagsCount,
		ContainerRegistryImportStartMaxRetries:                p.ContainerRegistryImportStartMaxRetries,
		ContainerRegistryImportTargetPlan:                     p.ContainerRegistryImportTargetPlan,
		ContainerRegistryTokenExpireDelay:                     p.ContainerRegistryTokenExpireDelay,
		CustomHTTPCloneURLRoot:                                p.CustomHTTPCloneURLRoot,
		DNSRebindingProtectionEnabled:                         p.DNSRebindingProtectionEnabled,
		DSAKeyRestriction:                                     p.DSAKeyRestriction,
		DeactivateDormantUsers:                                p.DeactivateDormantUsers,
		DeactivateDormantUsersPeriod:                          p.DeactivateDormantUsersPeriod,
		DecompressArchiveFileTimeout:                          p.DecompressArchiveFileTimeout,
		DefaultArtifactsExpireIn:                              p.DefaultArtifactsExpireIn,
		DefaultBranchName:                                     p.DefaultBranchName,
		DefaultBranchProtectionDefaults:                       p.DefaultBranchProtectionDefaults,
		DefaultCiConfigPath:                                   p.DefaultCiConfigPath,
		DefaultGroupVisibility:                                (*gitlab.VisibilityValue)(p.DefaultGroupVisibility),
		DefaultPreferredLanguage:                              p.DefaultPreferredLanguage,
		DefaultProjectCreation:                                p.DefaultProjectCreation,
		DefaultProjectDeletionProtection:                      p.DefaultProjectDeletionProtection,
		DefaultProjectVisibility:                              (*gitlab.VisibilityValue)(p.DefaultProjectVisibility),
		DefaultProjectsLimit:                                  p.DefaultProjectsLimit,
		DefaultSnippetVisibility:                              (*gitlab.VisibilityValue)(p.DefaultSnippetVisibility),
		DefaultSyntaxHighlightingTheme:                        p.DefaultSyntaxHighlightingTheme,
		DelayedGroupDeletion:                                  p.DelayedGroupDeletion,
		DelayedProjectDeletion:                                p.DelayedProjectDeletion,
		DeleteInactiveProjects:                                p.DeleteInactiveProjects,
		DeleteUnconfirmedUsers:                                p.DeleteUnconfirmedUsers,
		DeletionAdjournedPeriod:                               p.DeletionAdjournedPeriod,
		DiagramsnetEnabled:                                    p.DiagramsnetEnabled,
		DiagramsnetURL:                                        p.DiagramsnetURL,
		DiffMaxFiles:                                          p.DiffMaxFiles,
		DiffMaxLines:                                          p.DiffMaxLines,
		DiffMaxPatchBytes:                                     p.DiffMaxPatchBytes,
		DisableFeedToken:                                      p.DisableFeedToken,
		DisableAdminOAuthScopes:                               p.DisableAdminOAuthScopes,
		DisableOverridingApproversPerMergeRequest:             p.DisableOverridingApproversPerMergeRequest,
		DisablePersonalAccessTokens:                           p.DisablePersonalAccessTokens,
		DisabledOauthSignInSources:                            p.DisabledOauthSignInSources,
		DomainAllowlist:                                       p.DomainAllowlist,
		DomainDenylist:                                        p.DomainDenylist,
		DomainDenylistEnabled:                                 p.DomainDenylistEnabled,
		DownstreamPipelineTriggerLimitPerProjectUserSHA:       p.DownstreamPipelineTriggerLimitPerProjectUserSHA,
		DuoFeaturesEnabled:                                    p.DuoFeaturesEnabled,
		ECDSAKeyRestriction:                                   p.ECDSAKeyRestriction,
		ECDSASKKeyRestriction:                                 p.ECDSASKKeyRestriction,
		EKSAccessKeyID:                                        p.EKSAccessKeyID,
		EKSAccountID:                                          p.EKSAccountID,
		EKSIntegrationEnabled:                                 p.EKSIntegrationEnabled,
		EKSSecretAccessKey:                                    p.EKSSecretAccessKey,
		Ed25519KeyRestriction:                                 p.Ed25519KeyRestriction,
		Ed25519SKKeyRestriction:                               p.Ed25519SKKeyRestriction,
		ElasticsearchAWS:                                      p.ElasticsearchAWS,
		ElasticsearchAWSAccessKey:                             p.ElasticsearchAWSAccessKey,
		ElasticsearchAWSRegion:                                p.ElasticsearchAWSRegion,
		ElasticsearchAWSSecretAccessKey:                       p.ElasticsearchAWSSecretAccessKey,
		ElasticsearchAnalyzersKuromojiEnabled:                 p.ElasticsearchAnalyzersKuromojiEnabled,
		ElasticsearchAnalyzersSmartCNEnabled:                  p.ElasticsearchAnalyzersSmartCNEnabled,
		ElasticsearchClientRequestTimeout:                     p.ElasticsearchClientRequestTimeout,
		ElasticsearchIndexedFieldLengthLimit:                  p.ElasticsearchIndexedFieldLengthLimit,
		ElasticsearchIndexedFileSizeLimitKB:                   p.ElasticsearchIndexedFileSizeLimitKB,
		ElasticsearchIndexing:                                 p.ElasticsearchIndexing,
		ElasticsearchLimitIndexing:                            p.ElasticsearchLimitIndexing,
		ElasticsearchMaxBulkConcurrency:                       p.ElasticsearchMaxBulkConcurrency,
		ElasticsearchMaxBulkSizeMB:                            p.ElasticsearchMaxBulkSizeMB,
		ElasticsearchMaxCodeIndexingConcurrency:               p.ElasticsearchMaxCodeIndexingConcurrency,
		ElasticsearchNamespaceIDs:                             p.ElasticsearchNamespaceIDs,
		ElasticsearchPassword:                                 p.ElasticsearchPassword,
		ElasticsearchPauseIndexing:                            p.ElasticsearchPauseIndexing,
		ElasticsearchProjectIDs:                               p.ElasticsearchProjectIDs,
		ElasticsearchReplicas:                                 p.ElasticsearchReplicas,
		ElasticsearchRequeueWorkers:                           p.ElasticsearchRequeueWorkers,
		ElasticsearchRetryOnFailure:                           p.ElasticsearchRetryOnFailure,
		ElasticsearchSearch:                                   p.ElasticsearchSearch,
		ElasticsearchShards:                                   p.ElasticsearchShards,
		ElasticsearchURL:                                      p.ElasticsearchURL,
		ElasticsearchUsername:                                 p.ElasticsearchUsername,
		ElasticsearchWorkerNumberOfShards:                     p.ElasticsearchWorkerNumberOfShards,
		EmailAdditionalText:                                   p.EmailAdditionalText,
		EmailAuthorInBody:                                     p.EmailAuthorInBody,
		EmailConfirmationSetting:                              p.EmailConfirmationSetting,
		EmailRestrictions:                                     p.EmailRestrictions,
		EmailRestrictionsEnabled:                              p.EmailRestrictionsEnabled,
		EnableArtifactExternalRedirectWarningPage:             p.EnableArtifactExternalRedirectWarningPage,
		EnabledGitAccessProtocol:                              p.EnabledGitAccessProtocol,
		EnforceCIInboundJobTokenScopeEnabled:                  p.EnforceCIInboundJobTokenScopeEnabled,
		EnforceNamespaceStorageLimit:                          p.EnforceNamespaceStorageLimit,
		EnforcePATExpiration:                                  p.EnforcePATExpiration,
		EnforceSSHKeyExpiration:                               p.EnforceSSHKeyExpiration,
		EnforceTerms:                                          p.EnforceTerms,
		ExternalAuthClientCert:                                p.ExternalAuthClientCert,
		ExternalAuthClientKey:                                 p.ExternalAuthClientKey,
		ExternalAuthClientKeyPass:                             p.ExternalAuthClientKeyPass,
		ExternalAuthorizationServiceDefaultLabel:              p.ExternalAuthorizationServiceDefaultLabel,
		ExternalAuthorizationServiceEnabled:                   p.ExternalAuthorizationServiceEnabled,
		ExternalAuthorizationServiceTimeout:                   p.ExternalAuthorizationServiceTimeout,
		ExternalAuthorizationServiceURL:                       p.ExternalAuthorizationServiceURL,
		ExternalPipelineValidationServiceTimeout:              p.ExternalPipelineValidationServiceTimeout,
		ExternalPipelineValidationServiceToken:                p.ExternalPipelineValidationServiceToken,
		ExternalPipelineValidationServiceURL:                  p.ExternalPipelineValidationServiceURL,
		FailedLoginAttemptsUnlockPeriodInMinutes:              p.FailedLoginAttemptsUnlockPeriodInMinutes,
		FileTemplateProjectID:                                 p.FileTemplateProjectID,
		FirstDayOfWeek:                                        p.FirstDayOfWeek,
		FlocEnabled:                                           p.FlocEnabled,
		GeoNodeAllowedIPs:                                     p.GeoNodeAllowedIPs,
		GeoStatusTimeout:                                      p.GeoStatusTimeout,
		GitRateLimitUsersAlertlist:                            p.GitRateLimitUsersAlertlist,
		GitTwoFactorSessionExpiry:                             p.GitTwoFactorSessionExpiry,
		GitalyTimeoutDefault:                                  p.GitalyTimeoutDefault,
		GitalyTimeoutFast:                                     p.GitalyTimeoutFast,
		GitalyTimeoutMedium:                                   p.GitalyTimeoutMedium,
		GitlabDedicatedInstance:                               p.GitlabDedicatedInstance,
		GitlabEnvironmentToolkitInstance:                      p.GitlabEnvironmentToolkitInstance,
		GitlabShellOperationLimit:                             p.GitlabShellOperationLimit,
		GitpodEnabled:                                         p.GitpodEnabled,
		GitpodURL:                                             p.GitpodURL,
		GitRateLimitUsersAllowlist:                            p.GitRateLimitUsersAllowlist,
		GloballyAllowedIPs:                                    p.GloballyAllowedIPs,
		GrafanaEnabled:                                        p.GrafanaEnabled,
		GrafanaURL:                                            p.GrafanaURL,
		GravatarEnabled:                                       p.GravatarEnabled,
		GroupDownloadExportLimit:                              p.GroupDownloadExportLimit,
		GroupExportLimit:                                      p.GroupExportLimit,
		GroupImportLimit:                                      p.GroupImportLimit,
		GroupOwnersCanManageDefaultBranchProtection:           p.GroupOwnersCanManageDefaultBranchProtection,
		GroupRunnerTokenExpirationInterval:                    p.GroupRunnerTokenExpirationInterval,
		HTMLEmailsEnabled:                                     p.HTMLEmailsEnabled,
		HashedStorageEnabled:                                  p.HashedStorageEnabled,
		HelpPageDocumentationBaseURL:                          p.HelpPageDocumentationBaseURL,
		HelpPageHideCommercialContent:                         p.HelpPageHideCommercialContent,
		HelpPageSupportURL:                                    p.HelpPageSupportURL,
		HelpPageText:                                          p.HelpPageText,
		HelpText:                                              p.HelpText,
		HideThirdPartyOffers:                                  p.HideThirdPartyOffers,
		HomePageURL:                                           p.HomePageURL,
		HousekeepingEnabled:                                   p.HousekeepingEnabled,
		HousekeepingOptimizeRepositoryPeriod:                  p.HousekeepingOptimizeRepositoryPeriod,
		ImportSources:                                         p.ImportSources,
		InactiveProjectsDeleteAfterMonths:                     p.InactiveProjectsDeleteAfterMonths,
		InactiveProjectsMinSizeMB:                             p.InactiveProjectsMinSizeMB,
		InactiveProjectsSendWarningEmailAfterMonths:           p.InactiveProjectsSendWarningEmailAfterMonths,
		IncludeOptionalMetricsInServicePing:                   p.IncludeOptionalMetricsInServicePing,
		InProductMarketingEmailsEnabled:                       p.InProductMarketingEmailsEnabled,
		InvisibleCaptchaEnabled:                               p.InvisibleCaptchaEnabled,
		IssuesCreateLimit:                                     p.IssuesCreateLimit,
		JiraConnectApplicationKey:                             p.JiraConnectApplicationKey,
		JiraConnectPublicKeyStorageEnabled:                    p.JiraConnectPublicKeyStorageEnabled,
		JiraConnectProxyURL:                                   p.JiraConnectProxyURL,
		KeepLatestArtifact:                                    p.KeepLatestArtifact,
		KrokiEnabled:                                          p.KrokiEnabled,
		KrokiFormats:                                          p.KrokiFormats,
		KrokiURL:                                              p.KrokiURL,
		LocalMarkdownVersion:                                  p.LocalMarkdownVersion,
		LockDuoFeaturesEnabled:                                p.LockDuoFeaturesEnabled,
		LockMembershipsToLDAP:                                 p.LockMembershipsToLDAP,
		LoginRecaptchaProtectionEnabled:                       p.LoginRecaptchaProtectionEnabled,
		MailgunEventsEnabled:                                  p.MailgunEventsEnabled,
		MailgunSigningKey:                                     p.MailgunSigningKey,
		MaintenanceMode:                                       p.MaintenanceMode,
		MaintenanceModeMessage:                                p.MaintenanceModeMessage,
		MavenPackageRequestsForwarding:                        p.MavenPackageRequestsForwarding,
		MaxArtifactsSize:                                      p.MaxArtifactsSize,
		MaxAttachmentSize:                                     p.MaxAttachmentSize,
		MaxDecompressedArchiveSize:                            p.MaxDecompressedArchiveSize,
		MaxExportSize:                                         p.MaxExportSize,
		MaxImportRemoteFileSize:                               p.MaxImportRemoteFileSize,
		MaxImportSize:                                         p.MaxImportSize,
		MaxLoginAttempts:                                      p.MaxLoginAttempts,
		MaxNumberOfRepositoryDownloads:                        p.MaxNumberOfRepositoryDownloads,
		MaxNumberOfRepositoryDownloadsWithinTimePeriod:        p.MaxNumberOfRepositoryDownloadsWithinTimePeriod,
		MaxPagesSize:                                          p.MaxPagesSize,
		MaxPersonalAccessTokenLifetime:                        p.MaxPersonalAccessTokenLifetime,
		MaxSSHKeyLifetime:                                     p.MaxSSHKeyLifetime,
		MaxTerraformStateSizeBytes:                            p.MaxTerraformStateSizeBytes,
		MaxYAMLDepth:                                          p.MaxYAMLDepth,
		MaxYAMLSizeBytes:                                      p.MaxYAMLSizeBytes,
		MetricsMethodCallThreshold:                            p.MetricsMethodCallThreshold,
		MinimumPasswordLength:                                 p.MinimumPasswordLength,
		MirrorAvailable:                                       p.MirrorAvailable,
		MirrorCapacityThreshold:                               p.MirrorCapacityThreshold,
		MirrorMaxCapacity:                                     p.MirrorMaxCapacity,
		MirrorMaxDelay:                                        p.MirrorMaxDelay,
		NPMPackageRequestsForwarding:                          p.NPMPackageRequestsForwarding,
		NotesCreateLimit:                                      p.NotesCreateLimit,
		NotifyOnUnknownSignIn:                                 p.NotifyOnUnknownSignIn,
		NugetSkipMetadataURLValidation:                        p.NugetSkipMetadataURLValidation,
		OutboundLocalRequestsAllowlistRaw:                     p.OutboundLocalRequestsAllowlistRaw,
		OutboundLocalRequestsWhitelist:                        p.OutboundLocalRequestsWhitelist,
		PackageMetadataPURLTypes:                              p.PackageMetadataPURLTypes,
		PackageRegistryAllowAnyoneToPullOption:                p.PackageRegistryAllowAnyoneToPullOption,
		PackageRegistryCleanupPoliciesWorkerCapacity:          p.PackageRegistryCleanupPoliciesWorkerCapacity,
		PagesDomainVerificationEnabled:                        p.PagesDomainVerificationEnabled,
		PasswordAuthenticationEnabledForGit:                   p.PasswordAuthenticationEnabledForGit,
		PasswordAuthenticationEnabledForWeb:                   p.PasswordAuthenticationEnabledForWeb,
		PasswordNumberRequired:                                p.PasswordNumberRequired,
		PasswordSymbolRequired:                                p.PasswordSymbolRequired,
		PasswordUppercaseRequired:                             p.PasswordUppercaseRequired,
		PasswordLowercaseRequired:                             p.PasswordLowercaseRequired,
		PerformanceBarAllowedGroupPath:                        p.PerformanceBarAllowedGroupPath,
		PersonalAccessTokenPrefix:                             p.PersonalAccessTokenPrefix,
		PlantumlEnabled:                                       p.PlantumlEnabled,
		PlantumlURL:                                           p.PlantumlURL,
		PipelineLimitPerProjectUserSha:                        p.PipelineLimitPerProjectUserSha,
		PollingIntervalMultiplier:                             p.PollingIntervalMultiplier,
		PreventMergeRequestsAuthorApproval:                    p.PreventMergeRequestsAuthorApproval,
		PreventMergeRequestsCommittersApproval:                p.PreventMergeRequestsCommittersApproval,
		ProjectDownloadExportLimit:                            p.ProjectDownloadExportLimit,
		ProjectExportEnabled:                                  p.ProjectExportEnabled,
		ProjectExportLimit:                                    p.ProjectExportLimit,
		ProjectImportLimit:                                    p.ProjectImportLimit,
		ProjectJobsAPIRateLimit:                               p.ProjectJobsAPIRateLimit,
		ProjectRunnerTokenExpirationInterval:                  p.ProjectRunnerTokenExpirationInterval,
		ProjectsAPIRateLimitUnauthenticated:                   p.ProjectsAPIRateLimitUnauthenticated,
		PrometheusMetricsEnabled:                              p.PrometheusMetricsEnabled,
		ProtectedCIVariables:                                  p.ProtectedCIVariables,
		PseudonymizerEnabled:                                  p.PseudonymizerEnabled,
		PushEventActivitiesLimit:                              p.PushEventActivitiesLimit,
		PushEventHooksLimit:                                   p.PushEventHooksLimit,
		PyPIPackageRequestsForwarding:                         p.PyPIPackageRequestsForwarding,
		RSAKeyRestriction:                                     p.RSAKeyRestriction,
		RateLimitingResponseText:                              p.RateLimitingResponseText,
		RawBlobRequestLimit:                                   p.RawBlobRequestLimit,
		RecaptchaEnabled:                                      p.RecaptchaEnabled,
		RecaptchaPrivateKey:                                   p.RecaptchaPrivateKey,
		RecaptchaSiteKey:                                      p.RecaptchaSiteKey,
		ReceiveMaxInputSize:                                   p.ReceiveMaxInputSize,
		ReceptiveClusterAgentsEnabled:                         p.ReceptiveClusterAgentsEnabled,
		RememberMeEnabled:                                     p.RememberMeEnabled,
		RepositoryChecksEnabled:                               p.RepositoryChecksEnabled,
		RepositorySizeLimit:                                   p.RepositorySizeLimit,
		RepositoryStorages:                                    p.RepositoryStorages,
		RepositoryStoragesWeighted:                            p.RepositoryStoragesWeighted,
		RequireAdminApprovalAfterUserSignup:                   p.RequireAdminApprovalAfterUserSignup,
		RequireAdminTwoFactorAuthentication:                   p.RequireAdminTwoFactorAuthentication,
		RequirePersonalAccessTokenExpiry:                      p.RequirePersonalAccessTokenExpiry,
		RequireTwoFactorAuthentication:                        p.RequireTwoFactorAuthentication,
		RestrictedVisibilityLevels:                            clients.StringPtrSliceToVisibilityValuePtrSlice(p.RestrictedVisibilityLevels),
		RunnerTokenExpirationInterval:                         p.RunnerTokenExpirationInterval,
		SearchRateLimit:                                       p.SearchRateLimit,
		SearchRateLimitUnauthenticated:                        p.SearchRateLimitUnauthenticated,
		SecretDetectionRevocationTokenTypesURL:                p.SecretDetectionRevocationTokenTypesURL,
		SecretDetectionTokenRevocationEnabled:                 p.SecretDetectionTokenRevocationEnabled,
		SecretDetectionTokenRevocationToken:                   p.SecretDetectionTokenRevocationToken,
		SecretDetectionTokenRevocationURL:                     p.SecretDetectionTokenRevocationURL,
		SecurityApprovalPoliciesLimit:                         p.SecurityApprovalPoliciesLimit,
		SecurityPolicyGlobalGroupApproversEnabled:             p.SecurityPolicyGlobalGroupApproversEnabled,
		SecurityTXTContent:                                    p.SecurityTXTContent,
		SendUserConfirmationEmail:                             p.SendUserConfirmationEmail,
		SentryClientsideDSN:                                   p.SentryClientsideDSN,
		SentryDSN:                                             p.SentryDSN,
		SentryEnvironment:                                     p.SentryEnvironment,
		ServiceAccessTokensExpirationEnforced:                 p.ServiceAccessTokensExpirationEnforced,
		SessionExpireDelay:                                    p.SessionExpireDelay,
		SharedRunnersEnabled:                                  p.SharedRunnersEnabled,
		SharedRunnersMinutes:                                  p.SharedRunnersMinutes,
		SharedRunnersText:                                     p.SharedRunnersText,
		SidekiqJobLimiterCompressionThresholdBytes:            p.SidekiqJobLimiterCompressionThresholdBytes,
		SidekiqJobLimiterLimitBytes:                           p.SidekiqJobLimiterLimitBytes,
		SidekiqJobLimiterMode:                                 p.SidekiqJobLimiterMode,
		SignInText:                                            p.SignInText,
		SignupEnabled:                                         p.SignupEnabled,
		SilentAdminExportsEnabled:                             p.SilentAdminExportsEnabled,
		SilentModeEnabled:                                     p.SilentModeEnabled,
		SlackAppEnabled:                                       p.SlackAppEnabled,
		SlackAppID:                                            p.SlackAppID,
		SlackAppSecret:                                        p.SlackAppSecret,
		SlackAppSigningSecret:                                 p.SlackAppSigningSecret,
		SlackAppVerificationToken:                             p.SlackAppVerificationToken,
		SnippetSizeLimit:                                      p.SnippetSizeLimit,
		SnowplowAppID:                                         p.SnowplowAppID,
		SnowplowCollectorHostname:                             p.SnowplowCollectorHostname,
		SnowplowCookieDomain:                                  p.SnowplowCookieDomain,
		SnowplowDatabaseCollectorHostname:                     p.SnowplowDatabaseCollectorHostname,
		SnowplowEnabled:                                       p.SnowplowEnabled,
		SourcegraphEnabled:                                    p.SourcegraphEnabled,
		SourcegraphPublicOnly:                                 p.SourcegraphPublicOnly,
		SourcegraphURL:                                        p.SourcegraphURL,
		SpamCheckAPIKey:                                       p.SpamCheckAPIKey,
		SpamCheckEndpointEnabled:                              p.SpamCheckEndpointEnabled,
		SpamCheckEndpointURL:                                  p.SpamCheckEndpointURL,
		StaticObjectsExternalStorageAuthToken:                 p.StaticObjectsExternalStorageAuthToken,
		StaticObjectsExternalStorageURL:                       p.StaticObjectsExternalStorageURL,
		SuggestPipelineEnabled:                                p.SuggestPipelineEnabled,
		TerminalMaxSessionTime:                                p.TerminalMaxSessionTime,
		Terms:                                                 p.Terms,
		ThrottleAuthenticatedAPIEnabled:                       p.ThrottleAuthenticatedAPIEnabled,
		ThrottleAuthenticatedAPIPeriodInSeconds:               p.ThrottleAuthenticatedAPIPeriodInSeconds,
		ThrottleAuthenticatedAPIRequestsPerPeriod:             p.ThrottleAuthenticatedAPIRequestsPerPeriod,
		ThrottleAuthenticatedDeprecatedAPIEnabled:             p.ThrottleAuthenticatedDeprecatedAPIEnabled,
		ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds:     p.ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds,
		ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod:   p.ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod,
		ThrottleAuthenticatedFilesAPIEnabled:                  p.ThrottleAuthenticatedFilesAPIEnabled,
		ThrottleAuthenticatedFilesAPIPeriodInSeconds:          p.ThrottleAuthenticatedFilesAPIPeriodInSeconds,
		ThrottleAuthenticatedFilesAPIRequestsPerPeriod:        p.ThrottleAuthenticatedFilesAPIRequestsPerPeriod,
		ThrottleAuthenticatedGitLFSEnabled:                    p.ThrottleAuthenticatedGitLFSEnabled,
		ThrottleAuthenticatedGitLFSPeriodInSeconds:            p.ThrottleAuthenticatedGitLFSPeriodInSeconds,
		ThrottleAuthenticatedGitLFSRequestsPerPeriod:          p.ThrottleAuthenticatedGitLFSRequestsPerPeriod,
		ThrottleAuthenticatedPackagesAPIEnabled:               p.ThrottleAuthenticatedPackagesAPIEnabled,
		ThrottleAuthenticatedPackagesAPIPeriodInSeconds:       p.ThrottleAuthenticatedPackagesAPIPeriodInSeconds,
		ThrottleAuthenticatedPackagesAPIRequestsPerPeriod:     p.ThrottleAuthenticatedPackagesAPIRequestsPerPeriod,
		ThrottleAuthenticatedWebEnabled:                       p.ThrottleAuthenticatedWebEnabled,
		ThrottleAuthenticatedWebPeriodInSeconds:               p.ThrottleAuthenticatedWebPeriodInSeconds,
		ThrottleAuthenticatedWebRequestsPerPeriod:             p.ThrottleAuthenticatedWebRequestsPerPeriod,
		ThrottleIncidentManagementNotificationEnabled:         p.ThrottleIncidentManagementNotificationEnabled,
		ThrottleIncidentManagementNotificationPerPeriod:       p.ThrottleIncidentManagementNotificationPerPeriod,
		ThrottleIncidentManagementNotificationPeriodInSeconds: p.ThrottleIncidentManagementNotificationPeriodInSeconds,
		ThrottleProtectedPathsEnabled:                         p.ThrottleProtectedPathsEnabled,
		ThrottleProtectedPathsPeriodInSeconds:                 p.ThrottleProtectedPathsPeriodInSeconds,
		ThrottleProtectedPathsRequestsPerPeriod:               p.ThrottleProtectedPathsRequestsPerPeriod,
		ThrottleUnauthenticatedAPIEnabled:                     p.ThrottleUnauthenticatedAPIEnabled,
		ThrottleUnauthenticatedAPIPeriodInSeconds:             p.ThrottleUnauthenticatedAPIPeriodInSeconds,
		ThrottleUnauthenticatedAPIRequestsPerPeriod:           p.ThrottleUnauthenticatedAPIRequestsPerPeriod,
		ThrottleUnauthenticatedDeprecatedAPIEnabled:           p.ThrottleUnauthenticatedDeprecatedAPIEnabled,
		ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds:   p.ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds,
		ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod: p.ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod,
		ThrottleUnauthenticatedFilesAPIEnabled:                p.ThrottleUnauthenticatedFilesAPIEnabled,
		ThrottleUnauthenticatedFilesAPIPeriodInSeconds:        p.ThrottleUnauthenticatedFilesAPIPeriodInSeconds,
		ThrottleUnauthenticatedFilesAPIRequestsPerPeriod:      p.ThrottleUnauthenticatedFilesAPIRequestsPerPeriod,
		ThrottleUnauthenticatedGitLFSEnabled:                  p.ThrottleUnauthenticatedGitLFSEnabled,
		ThrottleUnauthenticatedGitLFSPeriodInSeconds:          p.ThrottleUnauthenticatedGitLFSPeriodInSeconds,
		ThrottleUnauthenticatedGitLFSRequestsPerPeriod:        p.ThrottleUnauthenticatedGitLFSRequestsPerPeriod,
		ThrottleUnauthenticatedPackagesAPIEnabled:             p.ThrottleUnauthenticatedPackagesAPIEnabled,
		ThrottleUnauthenticatedPackagesAPIPeriodInSeconds:     p.ThrottleUnauthenticatedPackagesAPIPeriodInSeconds,
		ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod:   p.ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod,
		ThrottleUnauthenticatedWebEnabled:                     p.ThrottleUnauthenticatedWebEnabled,
		ThrottleUnauthenticatedWebPeriodInSeconds:             p.ThrottleUnauthenticatedWebPeriodInSeconds,
		ThrottleUnauthenticatedWebRequestsPerPeriod:           p.ThrottleUnauthenticatedWebRequestsPerPeriod,
		TimeTrackingLimitToHours:                              p.TimeTrackingLimitToHours,
		TwoFactorGracePeriod:                                  p.TwoFactorGracePeriod,
		UnconfirmedUsersDeleteAfterDays:                       p.UnconfirmedUsersDeleteAfterDays,
		UniqueIPsLimitEnabled:                                 p.UniqueIPsLimitEnabled,
		UniqueIPsLimitPerUser:                                 p.UniqueIPsLimitPerUser,
		UniqueIPsLimitTimeWindow:                              p.UniqueIPsLimitTimeWindow,
		UpdateRunnerVersionsEnabled:                           p.UpdateRunnerVersionsEnabled,
		UpdatingNameDisabledForUsers:                          p.UpdatingNameDisabledForUsers,
		UsagePingEnabled:                                      p.UsagePingEnabled,
		UsagePingFeaturesEnabled:                              p.UsagePingFeaturesEnabled,
		UseClickhouseForAnalytics:                             p.UseClickhouseForAnalytics,
		UserDeactivationEmailsEnabled:                         p.UserDeactivationEmailsEnabled,
		UserDefaultExternal:                                   p.UserDefaultExternal,
		UserDefaultInternalRegex:                              p.UserDefaultInternalRegex,
		UserDefaultsToPrivateProfile:                          p.UserDefaultsToPrivateProfile,
		UserEmailLookupLimit:                                  p.UserEmailLookupLimit,
		UserOauthApplications:                                 p.UserOauthApplications,
		UserShowAddSSHKeyMessage:                              p.UserShowAddSSHKeyMessage,
		UsersGetByIDLimit:                                     p.UsersGetByIDLimit,
		UsersGetByIDLimitAllowlistRaw:                         p.UsersGetByIDLimitAllowlistRaw,
		ValidRunnerRegistrars:                                 p.ValidRunnerRegistrars,
		VersionCheckEnabled:                                   p.VersionCheckEnabled,
		WebIDEClientsidePreviewEnabled:                        p.WebIDEClientsidePreviewEnabled,
		WhatsNewVariant:                                       p.WhatsNewVariant,
		WikiPageMaxContentBytes:                               p.WikiPageMaxContentBytes,
		AdminNotificationEmail:                                p.AdminNotificationEmail,
		AllowLocalRequestsFromHooksAndServices:                p.AllowLocalRequestsFromHooksAndServices,
		AssetProxyWhitelist:                                   p.AssetProxyWhitelist,
		DefaultBranchProtection:                               p.DefaultBranchProtection,
		HousekeepingBitmapsEnabled:                            p.HousekeepingBitmapsEnabled,
		HousekeepingFullRepackPeriod:                          p.HousekeepingFullRepackPeriod,
		HousekeepingGcPeriod:                                  p.HousekeepingGcPeriod,
		HousekeepingIncrementalRepackPeriod:                   p.HousekeepingIncrementalRepackPeriod,
		PerformanceBarAllowedGroupID:                          p.PerformanceBarAllowedGroupID,
		PerformanceBarEnabled:                                 p.PerformanceBarEnabled,
		ThrottleUnauthenticatedEnabled:                        p.ThrottleUnauthenticatedEnabled,
		ThrottleUnauthenticatedPeriodInSeconds:                p.ThrottleUnauthenticatedPeriodInSeconds,
		ThrottleUnauthenticatedRequestsPerPeriod:              p.ThrottleUnauthenticatedRequestsPerPeriod,
	}
}

// GenerateApplicationSettingsObservation generates Gitlab Application Settings observation from gitlab.Settings
func GenerateApplicationSettingsObservation(g *gitlab.Settings) v1alpha1.ApplicationSettingsObservation {
	return v1alpha1.ApplicationSettingsObservation{
		AbuseNotificationEmail:                                g.AbuseNotificationEmail,
		AdminMode:                                             g.AdminMode,
		AfterSignOutPath:                                      g.AfterSignOutPath,
		AfterSignUpText:                                       g.AfterSignUpText,
		AkismetAPIKey:                                         g.AkismetAPIKey,
		AkismetEnabled:                                        g.AkismetEnabled,
		AllowAccountDeletion:                                  g.AllowAccountDeletion,
		AllowAllIntegrations:                                  g.AllowAllIntegrations,
		AllowedIntegrations:                                   g.AllowedIntegrations,
		AllowGroupOwnersToManageLDAP:                          g.AllowGroupOwnersToManageLDAP,
		AllowLocalRequestsFromSystemHooks:                     g.AllowLocalRequestsFromSystemHooks,
		AllowLocalRequestsFromWebHooksAndServices:             g.AllowLocalRequestsFromWebHooksAndServices,
		AllowProjectCreationForGuestAndBelow:                  g.AllowProjectCreationForGuestAndBelow,
		AllowRunnerRegistrationToken:                          g.AllowRunnerRegistrationToken,
		ArchiveBuildsInHumanReadable:                          g.ArchiveBuildsInHumanReadable,
		ASCIIDocMaxIncludes:                                   g.ASCIIDocMaxIncludes,
		AssetProxyAllowlist:                                   g.AssetProxyAllowlist,
		AssetProxyEnabled:                                     g.AssetProxyEnabled,
		AssetProxySecretKey:                                   g.AssetProxySecretKey,
		AssetProxyURL:                                         g.AssetProxyURL,
		AuthorizedKeysEnabled:                                 g.AuthorizedKeysEnabled,
		AutoBanUserOnExcessiveProjectsDownload:                g.AutoBanUserOnExcessiveProjectsDownload,
		AutocompleteUsers:                                     g.AutocompleteUsers,
		AutocompleteUsersUnauthenticated:                      g.AutocompleteUsersUnauthenticated,
		AutoDevOpsDomain:                                      g.AutoDevOpsDomain,
		AutoDevOpsEnabled:                                     g.AutoDevOpsEnabled,
		AutomaticPurchasedStorageAllocation:                   g.AutomaticPurchasedStorageAllocation,
		BulkImportConcurrentPipelineBatchLimit:                g.BulkImportConcurrentPipelineBatchLimit,
		BulkImportEnabled:                                     g.BulkImportEnabled,
		BulkImportMaxDownloadFileSize:                         g.BulkImportMaxDownloadFileSize,
		CanCreateGroup:                                        g.CanCreateGroup,
		CheckNamespacePlan:                                    g.CheckNamespacePlan,
		CIJobLiveTraceEnabled:                                 g.CIJobLiveTraceEnabled,
		CIMaxIncludes:                                         g.CIMaxIncludes,
		CIMaxTotalYAMLSizeBytes:                               g.CIMaxTotalYAMLSizeBytes,
		CIPartitionsSizeLimit:                                 g.CIPartitionsSizeLimit,
		CommitEmailHostname:                                   g.CommitEmailHostname,
		ConcurrentBitbucketImportJobsLimit:                    g.ConcurrentBitbucketImportJobsLimit,
		ConcurrentBitbucketServerImportJobsLimit:              g.ConcurrentBitbucketServerImportJobsLimit,
		ConcurrentGitHubImportJobsLimit:                       g.ConcurrentGitHubImportJobsLimit,
		ContainerExpirationPoliciesEnableHistoricEntries:      g.ContainerExpirationPoliciesEnableHistoricEntries,
		ContainerRegistryCleanupTagsServiceMaxListSize:        g.ContainerRegistryCleanupTagsServiceMaxListSize,
		ContainerRegistryDeleteTagsServiceTimeout:             g.ContainerRegistryDeleteTagsServiceTimeout,
		ContainerRegistryExpirationPoliciesCaching:            g.ContainerRegistryExpirationPoliciesCaching,
		ContainerRegistryExpirationPoliciesWorkerCapacity:     g.ContainerRegistryExpirationPoliciesWorkerCapacity,
		ContainerRegistryImportCreatedBefore:                  g.ContainerRegistryImportCreatedBefore,
		ContainerRegistryImportMaxRetries:                     g.ContainerRegistryImportMaxRetries,
		ContainerRegistryImportMaxStepDuration:                g.ContainerRegistryImportMaxStepDuration,
		ContainerRegistryImportMaxTagsCount:                   g.ContainerRegistryImportMaxTagsCount,
		ContainerRegistryImportStartMaxRetries:                g.ContainerRegistryImportStartMaxRetries,
		ContainerRegistryImportTargetPlan:                     g.ContainerRegistryImportTargetPlan,
		ContainerRegistryTokenExpireDelay:                     g.ContainerRegistryTokenExpireDelay,
		CustomHTTPCloneURLRoot:                                g.CustomHTTPCloneURLRoot,
		DNSRebindingProtectionEnabled:                         g.DNSRebindingProtectionEnabled,
		DSAKeyRestriction:                                     g.DSAKeyRestriction,
		DeactivateDormantUsers:                                g.DeactivateDormantUsers,
		DeactivateDormantUsersPeriod:                          g.DeactivateDormantUsersPeriod,
		DecompressArchiveFileTimeout:                          g.DecompressArchiveFileTimeout,
		DefaultArtifactsExpireIn:                              g.DefaultArtifactsExpireIn,
		DefaultBranchName:                                     g.DefaultBranchName,
		DefaultBranchProtectionDefaults:                       g.DefaultBranchProtectionDefaults,
		DefaultCiConfigPath:                                   g.DefaultCiConfigPath,
		DefaultGroupVisibility:                                (string)(g.DefaultGroupVisibility),
		DefaultPreferredLanguage:                              g.DefaultPreferredLanguage,
		DefaultProjectCreation:                                g.DefaultProjectCreation,
		DefaultProjectDeletionProtection:                      g.DefaultProjectDeletionProtection,
		DefaultProjectVisibility:                              (string)(g.DefaultProjectVisibility),
		DefaultProjectsLimit:                                  g.DefaultProjectsLimit,
		DefaultSnippetVisibility:                              (string)(g.DefaultSnippetVisibility),
		DefaultSyntaxHighlightingTheme:                        g.DefaultSyntaxHighlightingTheme,
		DelayedGroupDeletion:                                  g.DelayedGroupDeletion,
		DelayedProjectDeletion:                                g.DelayedProjectDeletion,
		DeleteInactiveProjects:                                g.DeleteInactiveProjects,
		DeleteUnconfirmedUsers:                                g.DeleteUnconfirmedUsers,
		DeletionAdjournedPeriod:                               g.DeletionAdjournedPeriod,
		DiagramsnetEnabled:                                    g.DiagramsnetEnabled,
		DiagramsnetURL:                                        g.DiagramsnetURL,
		DiffMaxFiles:                                          g.DiffMaxFiles,
		DiffMaxLines:                                          g.DiffMaxLines,
		DiffMaxPatchBytes:                                     g.DiffMaxPatchBytes,
		DisableFeedToken:                                      g.DisableFeedToken,
		DisableAdminOAuthScopes:                               g.DisableAdminOAuthScopes,
		DisableOverridingApproversPerMergeRequest:             g.DisableOverridingApproversPerMergeRequest,
		DisablePersonalAccessTokens:                           g.DisablePersonalAccessTokens,
		DisabledOauthSignInSources:                            g.DisabledOauthSignInSources,
		DomainAllowlist:                                       g.DomainAllowlist,
		DomainDenylist:                                        g.DomainDenylist,
		DomainDenylistEnabled:                                 g.DomainDenylistEnabled,
		DownstreamPipelineTriggerLimitPerProjectUserSHA:       g.DownstreamPipelineTriggerLimitPerProjectUserSHA,
		DuoFeaturesEnabled:                                    g.DuoFeaturesEnabled,
		ECDSAKeyRestriction:                                   g.ECDSAKeyRestriction,
		ECDSASKKeyRestriction:                                 g.ECDSASKKeyRestriction,
		EKSAccessKeyID:                                        g.EKSAccessKeyID,
		EKSAccountID:                                          g.EKSAccountID,
		EKSIntegrationEnabled:                                 g.EKSIntegrationEnabled,
		EKSSecretAccessKey:                                    g.EKSSecretAccessKey,
		Ed25519KeyRestriction:                                 g.Ed25519KeyRestriction,
		Ed25519SKKeyRestriction:                               g.Ed25519SKKeyRestriction,
		ElasticsearchAWS:                                      g.ElasticsearchAWS,
		ElasticsearchAWSAccessKey:                             g.ElasticsearchAWSAccessKey,
		ElasticsearchAWSRegion:                                g.ElasticsearchAWSRegion,
		ElasticsearchAWSSecretAccessKey:                       g.ElasticsearchAWSSecretAccessKey,
		ElasticsearchAnalyzersKuromojiEnabled:                 g.ElasticsearchAnalyzersKuromojiEnabled,
		ElasticsearchAnalyzersSmartCNEnabled:                  g.ElasticsearchAnalyzersSmartCNEnabled,
		ElasticsearchClientRequestTimeout:                     g.ElasticsearchClientRequestTimeout,
		ElasticsearchIndexedFieldLengthLimit:                  g.ElasticsearchIndexedFieldLengthLimit,
		ElasticsearchIndexedFileSizeLimitKB:                   g.ElasticsearchIndexedFileSizeLimitKB,
		ElasticsearchIndexing:                                 g.ElasticsearchIndexing,
		ElasticsearchLimitIndexing:                            g.ElasticsearchLimitIndexing,
		ElasticsearchMaxBulkConcurrency:                       g.ElasticsearchMaxBulkConcurrency,
		ElasticsearchMaxBulkSizeMB:                            g.ElasticsearchMaxBulkSizeMB,
		ElasticsearchMaxCodeIndexingConcurrency:               g.ElasticsearchMaxCodeIndexingConcurrency,
		ElasticsearchNamespaceIDs:                             g.ElasticsearchNamespaceIDs,
		ElasticsearchPassword:                                 g.ElasticsearchPassword,
		ElasticsearchPauseIndexing:                            g.ElasticsearchPauseIndexing,
		ElasticsearchProjectIDs:                               g.ElasticsearchProjectIDs,
		ElasticsearchReplicas:                                 g.ElasticsearchReplicas,
		ElasticsearchRequeueWorkers:                           g.ElasticsearchRequeueWorkers,
		ElasticsearchRetryOnFailure:                           g.ElasticsearchRetryOnFailure,
		ElasticsearchSearch:                                   g.ElasticsearchSearch,
		ElasticsearchShards:                                   g.ElasticsearchShards,
		ElasticsearchURL:                                      g.ElasticsearchURL,
		ElasticsearchUsername:                                 g.ElasticsearchUsername,
		ElasticsearchWorkerNumberOfShards:                     g.ElasticsearchWorkerNumberOfShards,
		EmailAdditionalText:                                   g.EmailAdditionalText,
		EmailAuthorInBody:                                     g.EmailAuthorInBody,
		EmailConfirmationSetting:                              g.EmailConfirmationSetting,
		EmailRestrictions:                                     g.EmailRestrictions,
		EmailRestrictionsEnabled:                              g.EmailRestrictionsEnabled,
		EnableArtifactExternalRedirectWarningPage:             g.EnableArtifactExternalRedirectWarningPage,
		EnabledGitAccessProtocol:                              g.EnabledGitAccessProtocol,
		EnforceCIInboundJobTokenScopeEnabled:                  g.EnforceCIInboundJobTokenScopeEnabled,
		EnforceNamespaceStorageLimit:                          g.EnforceNamespaceStorageLimit,
		EnforcePATExpiration:                                  g.EnforcePATExpiration,
		EnforceSSHKeyExpiration:                               g.EnforceSSHKeyExpiration,
		EnforceTerms:                                          g.EnforceTerms,
		ExternalAuthClientCert:                                g.ExternalAuthClientCert,
		ExternalAuthClientKey:                                 g.ExternalAuthClientKey,
		ExternalAuthClientKeyPass:                             g.ExternalAuthClientKeyPass,
		ExternalAuthorizationServiceDefaultLabel:              g.ExternalAuthorizationServiceDefaultLabel,
		ExternalAuthorizationServiceEnabled:                   g.ExternalAuthorizationServiceEnabled,
		ExternalAuthorizationServiceTimeout:                   g.ExternalAuthorizationServiceTimeout,
		ExternalAuthorizationServiceURL:                       g.ExternalAuthorizationServiceURL,
		ExternalPipelineValidationServiceTimeout:              g.ExternalPipelineValidationServiceTimeout,
		ExternalPipelineValidationServiceToken:                g.ExternalPipelineValidationServiceToken,
		ExternalPipelineValidationServiceURL:                  g.ExternalPipelineValidationServiceURL,
		FailedLoginAttemptsUnlockPeriodInMinutes:              g.FailedLoginAttemptsUnlockPeriodInMinutes,
		FileTemplateProjectID:                                 g.FileTemplateProjectID,
		FirstDayOfWeek:                                        g.FirstDayOfWeek,
		FlocEnabled:                                           g.FlocEnabled,
		GeoNodeAllowedIPs:                                     g.GeoNodeAllowedIPs,
		GeoStatusTimeout:                                      g.GeoStatusTimeout,
		GitRateLimitUsersAlertlist:                            g.GitRateLimitUsersAlertlist,
		GitTwoFactorSessionExpiry:                             g.GitTwoFactorSessionExpiry,
		GitalyTimeoutDefault:                                  g.GitalyTimeoutDefault,
		GitalyTimeoutFast:                                     g.GitalyTimeoutFast,
		GitalyTimeoutMedium:                                   g.GitalyTimeoutMedium,
		GitlabDedicatedInstance:                               g.GitlabDedicatedInstance,
		GitlabEnvironmentToolkitInstance:                      g.GitlabEnvironmentToolkitInstance,
		GitlabShellOperationLimit:                             g.GitlabShellOperationLimit,
		GitpodEnabled:                                         g.GitpodEnabled,
		GitpodURL:                                             g.GitpodURL,
		GitRateLimitUsersAllowlist:                            g.GitRateLimitUsersAllowlist,
		GloballyAllowedIPs:                                    g.GloballyAllowedIPs,
		GrafanaEnabled:                                        g.GrafanaEnabled,
		GrafanaURL:                                            g.GrafanaURL,
		GravatarEnabled:                                       g.GravatarEnabled,
		GroupDownloadExportLimit:                              g.GroupDownloadExportLimit,
		GroupExportLimit:                                      g.GroupExportLimit,
		GroupImportLimit:                                      g.GroupImportLimit,
		GroupOwnersCanManageDefaultBranchProtection:           g.GroupOwnersCanManageDefaultBranchProtection,
		GroupRunnerTokenExpirationInterval:                    g.GroupRunnerTokenExpirationInterval,
		HTMLEmailsEnabled:                                     g.HTMLEmailsEnabled,
		HashedStorageEnabled:                                  g.HashedStorageEnabled,
		HelpPageDocumentationBaseURL:                          g.HelpPageDocumentationBaseURL,
		HelpPageHideCommercialContent:                         g.HelpPageHideCommercialContent,
		HelpPageSupportURL:                                    g.HelpPageSupportURL,
		HelpPageText:                                          g.HelpPageText,
		HelpText:                                              g.HelpText,
		HideThirdPartyOffers:                                  g.HideThirdPartyOffers,
		HomePageURL:                                           g.HomePageURL,
		HousekeepingEnabled:                                   g.HousekeepingEnabled,
		HousekeepingOptimizeRepositoryPeriod:                  g.HousekeepingOptimizeRepositoryPeriod,
		ImportSources:                                         g.ImportSources,
		InactiveProjectsDeleteAfterMonths:                     g.InactiveProjectsDeleteAfterMonths,
		InactiveProjectsMinSizeMB:                             g.InactiveProjectsMinSizeMB,
		InactiveProjectsSendWarningEmailAfterMonths:           g.InactiveProjectsSendWarningEmailAfterMonths,
		IncludeOptionalMetricsInServicePing:                   g.IncludeOptionalMetricsInServicePing,
		InProductMarketingEmailsEnabled:                       g.InProductMarketingEmailsEnabled,
		InvisibleCaptchaEnabled:                               g.InvisibleCaptchaEnabled,
		IssuesCreateLimit:                                     g.IssuesCreateLimit,
		JiraConnectApplicationKey:                             g.JiraConnectApplicationKey,
		JiraConnectPublicKeyStorageEnabled:                    g.JiraConnectPublicKeyStorageEnabled,
		JiraConnectProxyURL:                                   g.JiraConnectProxyURL,
		KeepLatestArtifact:                                    g.KeepLatestArtifact,
		KrokiEnabled:                                          g.KrokiEnabled,
		KrokiFormats:                                          g.KrokiFormats,
		KrokiURL:                                              g.KrokiURL,
		LocalMarkdownVersion:                                  g.LocalMarkdownVersion,
		LockDuoFeaturesEnabled:                                g.LockDuoFeaturesEnabled,
		LockMembershipsToLDAP:                                 g.LockMembershipsToLDAP,
		LoginRecaptchaProtectionEnabled:                       g.LoginRecaptchaProtectionEnabled,
		MailgunEventsEnabled:                                  g.MailgunEventsEnabled,
		MailgunSigningKey:                                     g.MailgunSigningKey,
		MaintenanceMode:                                       g.MaintenanceMode,
		MaintenanceModeMessage:                                g.MaintenanceModeMessage,
		MavenPackageRequestsForwarding:                        g.MavenPackageRequestsForwarding,
		MaxArtifactsSize:                                      g.MaxArtifactsSize,
		MaxAttachmentSize:                                     g.MaxAttachmentSize,
		MaxDecompressedArchiveSize:                            g.MaxDecompressedArchiveSize,
		MaxExportSize:                                         g.MaxExportSize,
		MaxImportRemoteFileSize:                               g.MaxImportRemoteFileSize,
		MaxImportSize:                                         g.MaxImportSize,
		MaxLoginAttempts:                                      g.MaxLoginAttempts,
		MaxNumberOfRepositoryDownloads:                        g.MaxNumberOfRepositoryDownloads,
		MaxNumberOfRepositoryDownloadsWithinTimePeriod:        g.MaxNumberOfRepositoryDownloadsWithinTimePeriod,
		MaxPagesSize:                                          g.MaxPagesSize,
		MaxPersonalAccessTokenLifetime:                        g.MaxPersonalAccessTokenLifetime,
		MaxSSHKeyLifetime:                                     g.MaxSSHKeyLifetime,
		MaxTerraformStateSizeBytes:                            g.MaxTerraformStateSizeBytes,
		MaxYAMLDepth:                                          g.MaxYAMLDepth,
		MaxYAMLSizeBytes:                                      g.MaxYAMLSizeBytes,
		MetricsMethodCallThreshold:                            g.MetricsMethodCallThreshold,
		MinimumPasswordLength:                                 g.MinimumPasswordLength,
		MirrorAvailable:                                       g.MirrorAvailable,
		MirrorCapacityThreshold:                               g.MirrorCapacityThreshold,
		MirrorMaxCapacity:                                     g.MirrorMaxCapacity,
		MirrorMaxDelay:                                        g.MirrorMaxDelay,
		NPMPackageRequestsForwarding:                          g.NPMPackageRequestsForwarding,
		NotesCreateLimit:                                      g.NotesCreateLimit,
		NotifyOnUnknownSignIn:                                 g.NotifyOnUnknownSignIn,
		NugetSkipMetadataURLValidation:                        g.NugetSkipMetadataURLValidation,
		OutboundLocalRequestsAllowlistRaw:                     g.OutboundLocalRequestsAllowlistRaw,
		OutboundLocalRequestsWhitelist:                        g.OutboundLocalRequestsWhitelist,
		PackageMetadataPURLTypes:                              g.PackageMetadataPURLTypes,
		PackageRegistryAllowAnyoneToPullOption:                g.PackageRegistryAllowAnyoneToPullOption,
		PackageRegistryCleanupPoliciesWorkerCapacity:          g.PackageRegistryCleanupPoliciesWorkerCapacity,
		PagesDomainVerificationEnabled:                        g.PagesDomainVerificationEnabled,
		PasswordAuthenticationEnabledForGit:                   g.PasswordAuthenticationEnabledForGit,
		PasswordAuthenticationEnabledForWeb:                   g.PasswordAuthenticationEnabledForWeb,
		PasswordNumberRequired:                                g.PasswordNumberRequired,
		PasswordSymbolRequired:                                g.PasswordSymbolRequired,
		PasswordUppercaseRequired:                             g.PasswordUppercaseRequired,
		PasswordLowercaseRequired:                             g.PasswordLowercaseRequired,
		PerformanceBarAllowedGroupPath:                        g.PerformanceBarAllowedGroupPath,
		PersonalAccessTokenPrefix:                             g.PersonalAccessTokenPrefix,
		PlantumlEnabled:                                       g.PlantumlEnabled,
		PlantumlURL:                                           g.PlantumlURL,
		PipelineLimitPerProjectUserSha:                        g.PipelineLimitPerProjectUserSha,
		PollingIntervalMultiplier:                             g.PollingIntervalMultiplier,
		PreventMergeRequestsAuthorApproval:                    g.PreventMergeRequestsAuthorApproval,
		PreventMergeRequestsCommittersApproval:                g.PreventMergeRequestsCommittersApproval,
		ProjectDownloadExportLimit:                            g.ProjectDownloadExportLimit,
		ProjectExportEnabled:                                  g.ProjectExportEnabled,
		ProjectExportLimit:                                    g.ProjectExportLimit,
		ProjectImportLimit:                                    g.ProjectImportLimit,
		ProjectJobsAPIRateLimit:                               g.ProjectJobsAPIRateLimit,
		ProjectRunnerTokenExpirationInterval:                  g.ProjectRunnerTokenExpirationInterval,
		ProjectsAPIRateLimitUnauthenticated:                   g.ProjectsAPIRateLimitUnauthenticated,
		PrometheusMetricsEnabled:                              g.PrometheusMetricsEnabled,
		ProtectedCIVariables:                                  g.ProtectedCIVariables,
		PseudonymizerEnabled:                                  g.PseudonymizerEnabled,
		PushEventActivitiesLimit:                              g.PushEventActivitiesLimit,
		PushEventHooksLimit:                                   g.PushEventHooksLimit,
		PyPIPackageRequestsForwarding:                         g.PyPIPackageRequestsForwarding,
		RSAKeyRestriction:                                     g.RSAKeyRestriction,
		RateLimitingResponseText:                              g.RateLimitingResponseText,
		RawBlobRequestLimit:                                   g.RawBlobRequestLimit,
		RecaptchaEnabled:                                      g.RecaptchaEnabled,
		RecaptchaPrivateKey:                                   g.RecaptchaPrivateKey,
		RecaptchaSiteKey:                                      g.RecaptchaSiteKey,
		ReceiveMaxInputSize:                                   g.ReceiveMaxInputSize,
		ReceptiveClusterAgentsEnabled:                         g.ReceptiveClusterAgentsEnabled,
		RememberMeEnabled:                                     g.RememberMeEnabled,
		RepositoryChecksEnabled:                               g.RepositoryChecksEnabled,
		RepositorySizeLimit:                                   g.RepositorySizeLimit,
		RepositoryStorages:                                    g.RepositoryStorages,
		RepositoryStoragesWeighted:                            g.RepositoryStoragesWeighted,
		RequireAdminApprovalAfterUserSignup:                   g.RequireAdminApprovalAfterUserSignup,
		RequireAdminTwoFactorAuthentication:                   g.RequireAdminTwoFactorAuthentication,
		RequirePersonalAccessTokenExpiry:                      g.RequirePersonalAccessTokenExpiry,
		RequireTwoFactorAuthentication:                        g.RequireTwoFactorAuthentication,
		RestrictedVisibilityLevels:                            clients.VisibilityValueSliceToStringSlice(g.RestrictedVisibilityLevels),
		RunnerTokenExpirationInterval:                         g.RunnerTokenExpirationInterval,
		SearchRateLimit:                                       g.SearchRateLimit,
		SearchRateLimitUnauthenticated:                        g.SearchRateLimitUnauthenticated,
		SecretDetectionRevocationTokenTypesURL:                g.SecretDetectionRevocationTokenTypesURL,
		SecretDetectionTokenRevocationEnabled:                 g.SecretDetectionTokenRevocationEnabled,
		SecretDetectionTokenRevocationToken:                   g.SecretDetectionTokenRevocationToken,
		SecretDetectionTokenRevocationURL:                     g.SecretDetectionTokenRevocationURL,
		SecurityApprovalPoliciesLimit:                         g.SecurityApprovalPoliciesLimit,
		SecurityPolicyGlobalGroupApproversEnabled:             g.SecurityPolicyGlobalGroupApproversEnabled,
		SecurityTXTContent:                                    g.SecurityTXTContent,
		SendUserConfirmationEmail:                             g.SendUserConfirmationEmail,
		SentryClientsideDSN:                                   g.SentryClientsideDSN,
		SentryDSN:                                             g.SentryDSN,
		SentryEnvironment:                                     g.SentryEnvironment,
		ServiceAccessTokensExpirationEnforced:                 g.ServiceAccessTokensExpirationEnforced,
		SessionExpireDelay:                                    g.SessionExpireDelay,
		SharedRunnersEnabled:                                  g.SharedRunnersEnabled,
		SharedRunnersMinutes:                                  g.SharedRunnersMinutes,
		SharedRunnersText:                                     g.SharedRunnersText,
		SidekiqJobLimiterCompressionThresholdBytes:            g.SidekiqJobLimiterCompressionThresholdBytes,
		SidekiqJobLimiterLimitBytes:                           g.SidekiqJobLimiterLimitBytes,
		SidekiqJobLimiterMode:                                 g.SidekiqJobLimiterMode,
		SignInText:                                            g.SignInText,
		SignupEnabled:                                         g.SignupEnabled,
		SilentAdminExportsEnabled:                             g.SilentAdminExportsEnabled,
		SilentModeEnabled:                                     g.SilentModeEnabled,
		SlackAppEnabled:                                       g.SlackAppEnabled,
		SlackAppID:                                            g.SlackAppID,
		SlackAppSecret:                                        g.SlackAppSecret,
		SlackAppSigningSecret:                                 g.SlackAppSigningSecret,
		SlackAppVerificationToken:                             g.SlackAppVerificationToken,
		SnippetSizeLimit:                                      g.SnippetSizeLimit,
		SnowplowAppID:                                         g.SnowplowAppID,
		SnowplowCollectorHostname:                             g.SnowplowCollectorHostname,
		SnowplowCookieDomain:                                  g.SnowplowCookieDomain,
		SnowplowDatabaseCollectorHostname:                     g.SnowplowDatabaseCollectorHostname,
		SnowplowEnabled:                                       g.SnowplowEnabled,
		SourcegraphEnabled:                                    g.SourcegraphEnabled,
		SourcegraphPublicOnly:                                 g.SourcegraphPublicOnly,
		SourcegraphURL:                                        g.SourcegraphURL,
		SpamCheckAPIKey:                                       g.SpamCheckAPIKey,
		SpamCheckEndpointEnabled:                              g.SpamCheckEndpointEnabled,
		SpamCheckEndpointURL:                                  g.SpamCheckEndpointURL,
		StaticObjectsExternalStorageAuthToken:                 g.StaticObjectsExternalStorageAuthToken,
		StaticObjectsExternalStorageURL:                       g.StaticObjectsExternalStorageURL,
		SuggestPipelineEnabled:                                g.SuggestPipelineEnabled,
		TerminalMaxSessionTime:                                g.TerminalMaxSessionTime,
		Terms:                                                 g.Terms,
		ThrottleAuthenticatedAPIEnabled:                       g.ThrottleAuthenticatedAPIEnabled,
		ThrottleAuthenticatedAPIPeriodInSeconds:               g.ThrottleAuthenticatedAPIPeriodInSeconds,
		ThrottleAuthenticatedAPIRequestsPerPeriod:             g.ThrottleAuthenticatedAPIRequestsPerPeriod,
		ThrottleAuthenticatedDeprecatedAPIEnabled:             g.ThrottleAuthenticatedDeprecatedAPIEnabled,
		ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds:     g.ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds,
		ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod:   g.ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod,
		ThrottleAuthenticatedFilesAPIEnabled:                  g.ThrottleAuthenticatedFilesAPIEnabled,
		ThrottleAuthenticatedFilesAPIPeriodInSeconds:          g.ThrottleAuthenticatedFilesAPIPeriodInSeconds,
		ThrottleAuthenticatedFilesAPIRequestsPerPeriod:        g.ThrottleAuthenticatedFilesAPIRequestsPerPeriod,
		ThrottleAuthenticatedGitLFSEnabled:                    g.ThrottleAuthenticatedGitLFSEnabled,
		ThrottleAuthenticatedGitLFSPeriodInSeconds:            g.ThrottleAuthenticatedGitLFSPeriodInSeconds,
		ThrottleAuthenticatedGitLFSRequestsPerPeriod:          g.ThrottleAuthenticatedGitLFSRequestsPerPeriod,
		ThrottleAuthenticatedPackagesAPIEnabled:               g.ThrottleAuthenticatedPackagesAPIEnabled,
		ThrottleAuthenticatedPackagesAPIPeriodInSeconds:       g.ThrottleAuthenticatedPackagesAPIPeriodInSeconds,
		ThrottleAuthenticatedPackagesAPIRequestsPerPeriod:     g.ThrottleAuthenticatedPackagesAPIRequestsPerPeriod,
		ThrottleAuthenticatedWebEnabled:                       g.ThrottleAuthenticatedWebEnabled,
		ThrottleAuthenticatedWebPeriodInSeconds:               g.ThrottleAuthenticatedWebPeriodInSeconds,
		ThrottleAuthenticatedWebRequestsPerPeriod:             g.ThrottleAuthenticatedWebRequestsPerPeriod,
		ThrottleIncidentManagementNotificationEnabled:         g.ThrottleIncidentManagementNotificationEnabled,
		ThrottleIncidentManagementNotificationPerPeriod:       g.ThrottleIncidentManagementNotificationPerPeriod,
		ThrottleIncidentManagementNotificationPeriodInSeconds: g.ThrottleIncidentManagementNotificationPeriodInSeconds,
		ThrottleProtectedPathsEnabled:                         g.ThrottleProtectedPathsEnabled,
		ThrottleProtectedPathsPeriodInSeconds:                 g.ThrottleProtectedPathsPeriodInSeconds,
		ThrottleProtectedPathsRequestsPerPeriod:               g.ThrottleProtectedPathsRequestsPerPeriod,
		ThrottleUnauthenticatedAPIEnabled:                     g.ThrottleUnauthenticatedAPIEnabled,
		ThrottleUnauthenticatedAPIPeriodInSeconds:             g.ThrottleUnauthenticatedAPIPeriodInSeconds,
		ThrottleUnauthenticatedAPIRequestsPerPeriod:           g.ThrottleUnauthenticatedAPIRequestsPerPeriod,
		ThrottleUnauthenticatedDeprecatedAPIEnabled:           g.ThrottleUnauthenticatedDeprecatedAPIEnabled,
		ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds:   g.ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds,
		ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod: g.ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod,
		ThrottleUnauthenticatedFilesAPIEnabled:                g.ThrottleUnauthenticatedFilesAPIEnabled,
		ThrottleUnauthenticatedFilesAPIPeriodInSeconds:        g.ThrottleUnauthenticatedFilesAPIPeriodInSeconds,
		ThrottleUnauthenticatedFilesAPIRequestsPerPeriod:      g.ThrottleUnauthenticatedFilesAPIRequestsPerPeriod,
		ThrottleUnauthenticatedGitLFSEnabled:                  g.ThrottleUnauthenticatedGitLFSEnabled,
		ThrottleUnauthenticatedGitLFSPeriodInSeconds:          g.ThrottleUnauthenticatedGitLFSPeriodInSeconds,
		ThrottleUnauthenticatedGitLFSRequestsPerPeriod:        g.ThrottleUnauthenticatedGitLFSRequestsPerPeriod,
		ThrottleUnauthenticatedPackagesAPIEnabled:             g.ThrottleUnauthenticatedPackagesAPIEnabled,
		ThrottleUnauthenticatedPackagesAPIPeriodInSeconds:     g.ThrottleUnauthenticatedPackagesAPIPeriodInSeconds,
		ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod:   g.ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod,
		ThrottleUnauthenticatedWebEnabled:                     g.ThrottleUnauthenticatedWebEnabled,
		ThrottleUnauthenticatedWebPeriodInSeconds:             g.ThrottleUnauthenticatedWebPeriodInSeconds,
		ThrottleUnauthenticatedWebRequestsPerPeriod:           g.ThrottleUnauthenticatedWebRequestsPerPeriod,
		TimeTrackingLimitToHours:                              g.TimeTrackingLimitToHours,
		TwoFactorGracePeriod:                                  g.TwoFactorGracePeriod,
		UnconfirmedUsersDeleteAfterDays:                       g.UnconfirmedUsersDeleteAfterDays,
		UniqueIPsLimitEnabled:                                 g.UniqueIPsLimitEnabled,
		UniqueIPsLimitPerUser:                                 g.UniqueIPsLimitPerUser,
		UniqueIPsLimitTimeWindow:                              g.UniqueIPsLimitTimeWindow,
		UpdateRunnerVersionsEnabled:                           g.UpdateRunnerVersionsEnabled,
		UpdatingNameDisabledForUsers:                          g.UpdatingNameDisabledForUsers,
		UsagePingEnabled:                                      g.UsagePingEnabled,
		UsagePingFeaturesEnabled:                              g.UsagePingFeaturesEnabled,
		UseClickhouseForAnalytics:                             g.UseClickhouseForAnalytics,
		UserDeactivationEmailsEnabled:                         g.UserDeactivationEmailsEnabled,
		UserDefaultExternal:                                   g.UserDefaultExternal,
		UserDefaultInternalRegex:                              g.UserDefaultInternalRegex,
		UserDefaultsToPrivateProfile:                          g.UserDefaultsToPrivateProfile,
		UserEmailLookupLimit:                                  g.UserEmailLookupLimit,
		UserOauthApplications:                                 g.UserOauthApplications,
		UserShowAddSSHKeyMessage:                              g.UserShowAddSSHKeyMessage,
		UsersGetByIDLimit:                                     g.UsersGetByIDLimit,
		UsersGetByIDLimitAllowlistRaw:                         g.UsersGetByIDLimitAllowlistRaw,
		ValidRunnerRegistrars:                                 g.ValidRunnerRegistrars,
		VersionCheckEnabled:                                   g.VersionCheckEnabled,
		WebIDEClientsidePreviewEnabled:                        g.WebIDEClientsidePreviewEnabled,
		WhatsNewVariant:                                       g.WhatsNewVariant,
		WikiPageMaxContentBytes:                               g.WikiPageMaxContentBytes,
		AdminNotificationEmail:                                g.AdminNotificationEmail,
		AllowLocalRequestsFromHooksAndServices:                g.AllowLocalRequestsFromHooksAndServices,
		AssetProxyWhitelist:                                   g.AssetProxyWhitelist,
		DefaultBranchProtection:                               g.DefaultBranchProtection,
		HousekeepingBitmapsEnabled:                            g.HousekeepingBitmapsEnabled,
		HousekeepingFullRepackPeriod:                          g.HousekeepingFullRepackPeriod,
		HousekeepingGcPeriod:                                  g.HousekeepingGcPeriod,
		HousekeepingIncrementalRepackPeriod:                   g.HousekeepingIncrementalRepackPeriod,
		PerformanceBarAllowedGroupID:                          g.PerformanceBarAllowedGroupID,
		PerformanceBarEnabled:                                 g.PerformanceBarEnabled,
		ThrottleUnauthenticatedEnabled:                        g.ThrottleUnauthenticatedEnabled,
		ThrottleUnauthenticatedPeriodInSeconds:                g.ThrottleUnauthenticatedPeriodInSeconds,
		ThrottleUnauthenticatedRequestsPerPeriod:              g.ThrottleUnauthenticatedRequestsPerPeriod,
	}
}

// IsApplicationSettingsUpToDate checks whether the observed state matches the desired state
func IsApplicationSettingsUpToDate(p *v1alpha1.ApplicationSettingsParameters, g *gitlab.Settings) bool {
	if !clients.IsComparableEqualToComparablePtr(p.AbuseNotificationEmail, g.AbuseNotificationEmail) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AdminMode, g.AdminMode) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AfterSignOutPath, g.AfterSignOutPath) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AfterSignUpText, g.AfterSignUpText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AkismetAPIKey, g.AkismetAPIKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AkismetEnabled, g.AkismetEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowAccountDeletion, g.AllowAccountDeletion) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowAllIntegrations, g.AllowAllIntegrations) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.AllowedIntegrations, g.AllowedIntegrations) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowGroupOwnersToManageLDAP, g.AllowGroupOwnersToManageLDAP) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowLocalRequestsFromSystemHooks, g.AllowLocalRequestsFromSystemHooks) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowLocalRequestsFromWebHooksAndServices, g.AllowLocalRequestsFromWebHooksAndServices) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowProjectCreationForGuestAndBelow, g.AllowProjectCreationForGuestAndBelow) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowRunnerRegistrationToken, g.AllowRunnerRegistrationToken) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ArchiveBuildsInHumanReadable, g.ArchiveBuildsInHumanReadable) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ASCIIDocMaxIncludes, g.ASCIIDocMaxIncludes) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.AssetProxyAllowlist, g.AssetProxyAllowlist) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AssetProxyEnabled, g.AssetProxyEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AssetProxySecretKey, g.AssetProxySecretKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AssetProxyURL, g.AssetProxyURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AuthorizedKeysEnabled, g.AuthorizedKeysEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AutoBanUserOnExcessiveProjectsDownload, g.AutoBanUserOnExcessiveProjectsDownload) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AutocompleteUsers, g.AutocompleteUsers) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AutocompleteUsersUnauthenticated, g.AutocompleteUsersUnauthenticated) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AutoDevOpsDomain, g.AutoDevOpsDomain) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AutoDevOpsEnabled, g.AutoDevOpsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AutomaticPurchasedStorageAllocation, g.AutomaticPurchasedStorageAllocation) {
		return false
	}
	if !clients.IsDefaultBranchProtectionDefaultsPtrEqualToDefaultsPtr(p.DefaultBranchProtectionDefaults, g.DefaultBranchProtectionDefaults) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.BulkImportConcurrentPipelineBatchLimit, g.BulkImportConcurrentPipelineBatchLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.BulkImportEnabled, g.BulkImportEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.BulkImportMaxDownloadFileSize, g.BulkImportMaxDownloadFileSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CanCreateGroup, g.CanCreateGroup) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CheckNamespacePlan, g.CheckNamespacePlan) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CIJobLiveTraceEnabled, g.CIJobLiveTraceEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CIMaxIncludes, g.CIMaxIncludes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CIMaxTotalYAMLSizeBytes, g.CIMaxTotalYAMLSizeBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CIPartitionsSizeLimit, g.CIPartitionsSizeLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CommitEmailHostname, g.CommitEmailHostname) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ConcurrentBitbucketImportJobsLimit, g.ConcurrentBitbucketImportJobsLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ConcurrentBitbucketServerImportJobsLimit, g.ConcurrentBitbucketServerImportJobsLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ConcurrentGitHubImportJobsLimit, g.ConcurrentGitHubImportJobsLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerExpirationPoliciesEnableHistoricEntries, g.ContainerExpirationPoliciesEnableHistoricEntries) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryCleanupTagsServiceMaxListSize, g.ContainerRegistryCleanupTagsServiceMaxListSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryDeleteTagsServiceTimeout, g.ContainerRegistryDeleteTagsServiceTimeout) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryExpirationPoliciesCaching, g.ContainerRegistryExpirationPoliciesCaching) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryExpirationPoliciesWorkerCapacity, g.ContainerRegistryExpirationPoliciesWorkerCapacity) {
		return false
	}
	if !clients.IsTimePtrEqualToTimePtr(p.ContainerRegistryImportCreatedBefore, g.ContainerRegistryImportCreatedBefore) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryImportMaxRetries, g.ContainerRegistryImportMaxRetries) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryImportMaxStepDuration, g.ContainerRegistryImportMaxStepDuration) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryImportMaxTagsCount, g.ContainerRegistryImportMaxTagsCount) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryImportStartMaxRetries, g.ContainerRegistryImportStartMaxRetries) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryImportTargetPlan, g.ContainerRegistryImportTargetPlan) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ContainerRegistryTokenExpireDelay, g.ContainerRegistryTokenExpireDelay) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.CustomHTTPCloneURLRoot, g.CustomHTTPCloneURLRoot) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DNSRebindingProtectionEnabled, g.DNSRebindingProtectionEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DSAKeyRestriction, g.DSAKeyRestriction) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DeactivateDormantUsers, g.DeactivateDormantUsers) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DeactivateDormantUsersPeriod, g.DeactivateDormantUsersPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DecompressArchiveFileTimeout, g.DecompressArchiveFileTimeout) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultArtifactsExpireIn, g.DefaultArtifactsExpireIn) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultBranchName, g.DefaultBranchName) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultCiConfigPath, g.DefaultCiConfigPath) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultGroupVisibility, string(g.DefaultGroupVisibility)) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultPreferredLanguage, g.DefaultPreferredLanguage) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultProjectCreation, g.DefaultProjectCreation) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultProjectDeletionProtection, g.DefaultProjectDeletionProtection) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultProjectVisibility, string(g.DefaultProjectVisibility)) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultProjectsLimit, g.DefaultProjectsLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultSnippetVisibility, string(g.DefaultSnippetVisibility)) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultSyntaxHighlightingTheme, g.DefaultSyntaxHighlightingTheme) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DelayedGroupDeletion, g.DelayedGroupDeletion) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DelayedProjectDeletion, g.DelayedProjectDeletion) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DeleteInactiveProjects, g.DeleteInactiveProjects) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DeleteUnconfirmedUsers, g.DeleteUnconfirmedUsers) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DeletionAdjournedPeriod, g.DeletionAdjournedPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DiagramsnetEnabled, g.DiagramsnetEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DiagramsnetURL, g.DiagramsnetURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DiffMaxFiles, g.DiffMaxFiles) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DiffMaxLines, g.DiffMaxLines) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DiffMaxPatchBytes, g.DiffMaxPatchBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DisableFeedToken, g.DisableFeedToken) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DisableAdminOAuthScopes, g.DisableAdminOAuthScopes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DisableOverridingApproversPerMergeRequest, g.DisableOverridingApproversPerMergeRequest) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DisablePersonalAccessTokens, g.DisablePersonalAccessTokens) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.DisabledOauthSignInSources, g.DisabledOauthSignInSources) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.DomainAllowlist, g.DomainAllowlist) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.DomainDenylist, g.DomainDenylist) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DomainDenylistEnabled, g.DomainDenylistEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DownstreamPipelineTriggerLimitPerProjectUserSHA, g.DownstreamPipelineTriggerLimitPerProjectUserSHA) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DuoFeaturesEnabled, g.DuoFeaturesEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ECDSAKeyRestriction, g.ECDSAKeyRestriction) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ECDSASKKeyRestriction, g.ECDSASKKeyRestriction) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EKSAccessKeyID, g.EKSAccessKeyID) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EKSAccountID, g.EKSAccountID) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EKSIntegrationEnabled, g.EKSIntegrationEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EKSSecretAccessKey, g.EKSSecretAccessKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.Ed25519KeyRestriction, g.Ed25519KeyRestriction) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.Ed25519SKKeyRestriction, g.Ed25519SKKeyRestriction) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchAWS, g.ElasticsearchAWS) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchAWSAccessKey, g.ElasticsearchAWSAccessKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchAWSRegion, g.ElasticsearchAWSRegion) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchAWSSecretAccessKey, g.ElasticsearchAWSSecretAccessKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchAnalyzersKuromojiEnabled, g.ElasticsearchAnalyzersKuromojiEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchAnalyzersSmartCNEnabled, g.ElasticsearchAnalyzersSmartCNEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchClientRequestTimeout, g.ElasticsearchClientRequestTimeout) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchIndexedFieldLengthLimit, g.ElasticsearchIndexedFieldLengthLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchIndexedFileSizeLimitKB, g.ElasticsearchIndexedFileSizeLimitKB) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchIndexing, g.ElasticsearchIndexing) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchLimitIndexing, g.ElasticsearchLimitIndexing) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchMaxBulkConcurrency, g.ElasticsearchMaxBulkConcurrency) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchMaxBulkSizeMB, g.ElasticsearchMaxBulkSizeMB) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchMaxCodeIndexingConcurrency, g.ElasticsearchMaxCodeIndexingConcurrency) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.ElasticsearchNamespaceIDs, g.ElasticsearchNamespaceIDs) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchPassword, g.ElasticsearchPassword) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchPauseIndexing, g.ElasticsearchPauseIndexing) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.ElasticsearchProjectIDs, g.ElasticsearchProjectIDs) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchReplicas, g.ElasticsearchReplicas) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchRequeueWorkers, g.ElasticsearchRequeueWorkers) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchRetryOnFailure, g.ElasticsearchRetryOnFailure) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchSearch, g.ElasticsearchSearch) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchShards, g.ElasticsearchShards) {
		return false
	}
	if p.ElasticsearchURL != nil {
		// ElasticsearchURL is a comma separated string in gitlab
		splitUrls := strings.Split(*p.ElasticsearchURL, ",")
		if !clients.IsComparableSliceEqualToComparableSlicePtr(&splitUrls, g.ElasticsearchURL) {
			return false
		}
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchUsername, g.ElasticsearchUsername) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ElasticsearchWorkerNumberOfShards, g.ElasticsearchWorkerNumberOfShards) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EmailAdditionalText, g.EmailAdditionalText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EmailAuthorInBody, g.EmailAuthorInBody) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EmailConfirmationSetting, g.EmailConfirmationSetting) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EmailRestrictions, g.EmailRestrictions) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EmailRestrictionsEnabled, g.EmailRestrictionsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnableArtifactExternalRedirectWarningPage, g.EnableArtifactExternalRedirectWarningPage) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnabledGitAccessProtocol, g.EnabledGitAccessProtocol) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnforceCIInboundJobTokenScopeEnabled, g.EnforceCIInboundJobTokenScopeEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnforceNamespaceStorageLimit, g.EnforceNamespaceStorageLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnforcePATExpiration, g.EnforcePATExpiration) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnforceSSHKeyExpiration, g.EnforceSSHKeyExpiration) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.EnforceTerms, g.EnforceTerms) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthClientCert, g.ExternalAuthClientCert) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthClientKey, g.ExternalAuthClientKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthClientKeyPass, g.ExternalAuthClientKeyPass) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthorizationServiceDefaultLabel, g.ExternalAuthorizationServiceDefaultLabel) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthorizationServiceEnabled, g.ExternalAuthorizationServiceEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthorizationServiceTimeout, g.ExternalAuthorizationServiceTimeout) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalAuthorizationServiceURL, g.ExternalAuthorizationServiceURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalPipelineValidationServiceTimeout, g.ExternalPipelineValidationServiceTimeout) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalPipelineValidationServiceToken, g.ExternalPipelineValidationServiceToken) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ExternalPipelineValidationServiceURL, g.ExternalPipelineValidationServiceURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.FailedLoginAttemptsUnlockPeriodInMinutes, g.FailedLoginAttemptsUnlockPeriodInMinutes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.FileTemplateProjectID, g.FileTemplateProjectID) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.FirstDayOfWeek, g.FirstDayOfWeek) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.FlocEnabled, g.FlocEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GeoNodeAllowedIPs, g.GeoNodeAllowedIPs) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GeoStatusTimeout, g.GeoStatusTimeout) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.GitRateLimitUsersAlertlist, g.GitRateLimitUsersAlertlist) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitTwoFactorSessionExpiry, g.GitTwoFactorSessionExpiry) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitalyTimeoutDefault, g.GitalyTimeoutDefault) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitalyTimeoutFast, g.GitalyTimeoutFast) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitalyTimeoutMedium, g.GitalyTimeoutMedium) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitlabDedicatedInstance, g.GitlabDedicatedInstance) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitlabEnvironmentToolkitInstance, g.GitlabEnvironmentToolkitInstance) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitlabShellOperationLimit, g.GitlabShellOperationLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitpodEnabled, g.GitpodEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GitpodURL, g.GitpodURL) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.GitRateLimitUsersAllowlist, g.GitRateLimitUsersAllowlist) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GloballyAllowedIPs, g.GloballyAllowedIPs) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GrafanaEnabled, g.GrafanaEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GrafanaURL, g.GrafanaURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GravatarEnabled, g.GravatarEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GroupDownloadExportLimit, g.GroupDownloadExportLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GroupExportLimit, g.GroupExportLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GroupImportLimit, g.GroupImportLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GroupOwnersCanManageDefaultBranchProtection, g.GroupOwnersCanManageDefaultBranchProtection) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.GroupRunnerTokenExpirationInterval, g.GroupRunnerTokenExpirationInterval) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HTMLEmailsEnabled, g.HTMLEmailsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HashedStorageEnabled, g.HashedStorageEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HelpPageDocumentationBaseURL, g.HelpPageDocumentationBaseURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HelpPageHideCommercialContent, g.HelpPageHideCommercialContent) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HelpPageSupportURL, g.HelpPageSupportURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HelpPageText, g.HelpPageText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HelpText, g.HelpText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HideThirdPartyOffers, g.HideThirdPartyOffers) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HomePageURL, g.HomePageURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HousekeepingEnabled, g.HousekeepingEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HousekeepingOptimizeRepositoryPeriod, g.HousekeepingOptimizeRepositoryPeriod) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.ImportSources, g.ImportSources) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.InactiveProjectsDeleteAfterMonths, g.InactiveProjectsDeleteAfterMonths) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.InactiveProjectsMinSizeMB, g.InactiveProjectsMinSizeMB) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.InactiveProjectsSendWarningEmailAfterMonths, g.InactiveProjectsSendWarningEmailAfterMonths) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.IncludeOptionalMetricsInServicePing, g.IncludeOptionalMetricsInServicePing) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.InProductMarketingEmailsEnabled, g.InProductMarketingEmailsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.InvisibleCaptchaEnabled, g.InvisibleCaptchaEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.IssuesCreateLimit, g.IssuesCreateLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.JiraConnectApplicationKey, g.JiraConnectApplicationKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.JiraConnectPublicKeyStorageEnabled, g.JiraConnectPublicKeyStorageEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.JiraConnectProxyURL, g.JiraConnectProxyURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.KeepLatestArtifact, g.KeepLatestArtifact) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.KrokiEnabled, g.KrokiEnabled) {
		return false
	}
	if !clients.IsMapStringToComparableEqualToMapStringToComparablePtr(p.KrokiFormats, g.KrokiFormats) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.KrokiURL, g.KrokiURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.LocalMarkdownVersion, g.LocalMarkdownVersion) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.LockDuoFeaturesEnabled, g.LockDuoFeaturesEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.LockMembershipsToLDAP, g.LockMembershipsToLDAP) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.LoginRecaptchaProtectionEnabled, g.LoginRecaptchaProtectionEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MailgunEventsEnabled, g.MailgunEventsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MailgunSigningKey, g.MailgunSigningKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaintenanceMode, g.MaintenanceMode) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaintenanceModeMessage, g.MaintenanceModeMessage) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MavenPackageRequestsForwarding, g.MavenPackageRequestsForwarding) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxArtifactsSize, g.MaxArtifactsSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxAttachmentSize, g.MaxAttachmentSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxDecompressedArchiveSize, g.MaxDecompressedArchiveSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxExportSize, g.MaxExportSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxImportRemoteFileSize, g.MaxImportRemoteFileSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxImportSize, g.MaxImportSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxLoginAttempts, g.MaxLoginAttempts) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxNumberOfRepositoryDownloads, g.MaxNumberOfRepositoryDownloads) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxNumberOfRepositoryDownloadsWithinTimePeriod, g.MaxNumberOfRepositoryDownloadsWithinTimePeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxPagesSize, g.MaxPagesSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxPersonalAccessTokenLifetime, g.MaxPersonalAccessTokenLifetime) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxSSHKeyLifetime, g.MaxSSHKeyLifetime) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxTerraformStateSizeBytes, g.MaxTerraformStateSizeBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxYAMLDepth, g.MaxYAMLDepth) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MaxYAMLSizeBytes, g.MaxYAMLSizeBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MetricsMethodCallThreshold, g.MetricsMethodCallThreshold) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MinimumPasswordLength, g.MinimumPasswordLength) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MirrorAvailable, g.MirrorAvailable) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MirrorCapacityThreshold, g.MirrorCapacityThreshold) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MirrorMaxCapacity, g.MirrorMaxCapacity) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.MirrorMaxDelay, g.MirrorMaxDelay) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.NPMPackageRequestsForwarding, g.NPMPackageRequestsForwarding) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.NotesCreateLimit, g.NotesCreateLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.NotifyOnUnknownSignIn, g.NotifyOnUnknownSignIn) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.NugetSkipMetadataURLValidation, g.NugetSkipMetadataURLValidation) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.OutboundLocalRequestsAllowlistRaw, g.OutboundLocalRequestsAllowlistRaw) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.OutboundLocalRequestsWhitelist, g.OutboundLocalRequestsWhitelist) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.PackageMetadataPURLTypes, g.PackageMetadataPURLTypes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PackageRegistryAllowAnyoneToPullOption, g.PackageRegistryAllowAnyoneToPullOption) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PackageRegistryCleanupPoliciesWorkerCapacity, g.PackageRegistryCleanupPoliciesWorkerCapacity) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PagesDomainVerificationEnabled, g.PagesDomainVerificationEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PasswordAuthenticationEnabledForGit, g.PasswordAuthenticationEnabledForGit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PasswordAuthenticationEnabledForWeb, g.PasswordAuthenticationEnabledForWeb) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PasswordNumberRequired, g.PasswordNumberRequired) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PasswordSymbolRequired, g.PasswordSymbolRequired) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PasswordUppercaseRequired, g.PasswordUppercaseRequired) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PasswordLowercaseRequired, g.PasswordLowercaseRequired) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PerformanceBarAllowedGroupPath, g.PerformanceBarAllowedGroupPath) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PersonalAccessTokenPrefix, g.PersonalAccessTokenPrefix) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PlantumlEnabled, g.PlantumlEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PlantumlURL, g.PlantumlURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PipelineLimitPerProjectUserSha, g.PipelineLimitPerProjectUserSha) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PollingIntervalMultiplier, g.PollingIntervalMultiplier) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PreventMergeRequestsAuthorApproval, g.PreventMergeRequestsAuthorApproval) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PreventMergeRequestsCommittersApproval, g.PreventMergeRequestsCommittersApproval) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectDownloadExportLimit, g.ProjectDownloadExportLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectExportEnabled, g.ProjectExportEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectExportLimit, g.ProjectExportLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectImportLimit, g.ProjectImportLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectJobsAPIRateLimit, g.ProjectJobsAPIRateLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectRunnerTokenExpirationInterval, g.ProjectRunnerTokenExpirationInterval) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProjectsAPIRateLimitUnauthenticated, g.ProjectsAPIRateLimitUnauthenticated) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PrometheusMetricsEnabled, g.PrometheusMetricsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ProtectedCIVariables, g.ProtectedCIVariables) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PseudonymizerEnabled, g.PseudonymizerEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PushEventActivitiesLimit, g.PushEventActivitiesLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PushEventHooksLimit, g.PushEventHooksLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PyPIPackageRequestsForwarding, g.PyPIPackageRequestsForwarding) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RSAKeyRestriction, g.RSAKeyRestriction) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RateLimitingResponseText, g.RateLimitingResponseText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RawBlobRequestLimit, g.RawBlobRequestLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RecaptchaEnabled, g.RecaptchaEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RecaptchaPrivateKey, g.RecaptchaPrivateKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RecaptchaSiteKey, g.RecaptchaSiteKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ReceiveMaxInputSize, g.ReceiveMaxInputSize) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ReceptiveClusterAgentsEnabled, g.ReceptiveClusterAgentsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RememberMeEnabled, g.RememberMeEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RepositoryChecksEnabled, g.RepositoryChecksEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RepositorySizeLimit, g.RepositorySizeLimit) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.RepositoryStorages, g.RepositoryStorages) {
		return false
	}
	if !clients.IsMapStringToComparableEqualToMapStringToComparablePtr(p.RepositoryStoragesWeighted, g.RepositoryStoragesWeighted) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RequireAdminApprovalAfterUserSignup, g.RequireAdminApprovalAfterUserSignup) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RequireAdminTwoFactorAuthentication, g.RequireAdminTwoFactorAuthentication) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RequirePersonalAccessTokenExpiry, g.RequirePersonalAccessTokenExpiry) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RequireTwoFactorAuthentication, g.RequireTwoFactorAuthentication) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.RestrictedVisibilityLevels, clients.VisibilityValueSliceToStringSlice(g.RestrictedVisibilityLevels)) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.RunnerTokenExpirationInterval, g.RunnerTokenExpirationInterval) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SearchRateLimit, g.SearchRateLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SearchRateLimitUnauthenticated, g.SearchRateLimitUnauthenticated) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecretDetectionRevocationTokenTypesURL, g.SecretDetectionRevocationTokenTypesURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecretDetectionTokenRevocationEnabled, g.SecretDetectionTokenRevocationEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecretDetectionTokenRevocationToken, g.SecretDetectionTokenRevocationToken) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecretDetectionTokenRevocationURL, g.SecretDetectionTokenRevocationURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecurityApprovalPoliciesLimit, g.SecurityApprovalPoliciesLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecurityPolicyGlobalGroupApproversEnabled, g.SecurityPolicyGlobalGroupApproversEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SecurityTXTContent, g.SecurityTXTContent) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SendUserConfirmationEmail, g.SendUserConfirmationEmail) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SentryClientsideDSN, g.SentryClientsideDSN) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SentryDSN, g.SentryDSN) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SentryEnvironment, g.SentryEnvironment) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ServiceAccessTokensExpirationEnforced, g.ServiceAccessTokensExpirationEnforced) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SessionExpireDelay, g.SessionExpireDelay) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SharedRunnersEnabled, g.SharedRunnersEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SharedRunnersMinutes, g.SharedRunnersMinutes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SharedRunnersText, g.SharedRunnersText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SidekiqJobLimiterCompressionThresholdBytes, g.SidekiqJobLimiterCompressionThresholdBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SidekiqJobLimiterLimitBytes, g.SidekiqJobLimiterLimitBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SidekiqJobLimiterMode, g.SidekiqJobLimiterMode) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SignInText, g.SignInText) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SignupEnabled, g.SignupEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SilentAdminExportsEnabled, g.SilentAdminExportsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SilentModeEnabled, g.SilentModeEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SlackAppEnabled, g.SlackAppEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SlackAppID, g.SlackAppID) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SlackAppSecret, g.SlackAppSecret) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SlackAppSigningSecret, g.SlackAppSigningSecret) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SlackAppVerificationToken, g.SlackAppVerificationToken) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SnippetSizeLimit, g.SnippetSizeLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SnowplowAppID, g.SnowplowAppID) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SnowplowCollectorHostname, g.SnowplowCollectorHostname) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SnowplowCookieDomain, g.SnowplowCookieDomain) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SnowplowDatabaseCollectorHostname, g.SnowplowDatabaseCollectorHostname) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SnowplowEnabled, g.SnowplowEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SourcegraphEnabled, g.SourcegraphEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SourcegraphPublicOnly, g.SourcegraphPublicOnly) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SourcegraphURL, g.SourcegraphURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SpamCheckAPIKey, g.SpamCheckAPIKey) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SpamCheckEndpointEnabled, g.SpamCheckEndpointEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SpamCheckEndpointURL, g.SpamCheckEndpointURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.StaticObjectsExternalStorageAuthToken, g.StaticObjectsExternalStorageAuthToken) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.StaticObjectsExternalStorageURL, g.StaticObjectsExternalStorageURL) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.SuggestPipelineEnabled, g.SuggestPipelineEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.TerminalMaxSessionTime, g.TerminalMaxSessionTime) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.Terms, g.Terms) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedAPIEnabled, g.ThrottleAuthenticatedAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedAPIPeriodInSeconds, g.ThrottleAuthenticatedAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedAPIRequestsPerPeriod, g.ThrottleAuthenticatedAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedDeprecatedAPIEnabled, g.ThrottleAuthenticatedDeprecatedAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds, g.ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod, g.ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedFilesAPIEnabled, g.ThrottleAuthenticatedFilesAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedFilesAPIPeriodInSeconds, g.ThrottleAuthenticatedFilesAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedFilesAPIRequestsPerPeriod, g.ThrottleAuthenticatedFilesAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedGitLFSEnabled, g.ThrottleAuthenticatedGitLFSEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedGitLFSPeriodInSeconds, g.ThrottleAuthenticatedGitLFSPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedGitLFSRequestsPerPeriod, g.ThrottleAuthenticatedGitLFSRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedPackagesAPIEnabled, g.ThrottleAuthenticatedPackagesAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedPackagesAPIPeriodInSeconds, g.ThrottleAuthenticatedPackagesAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedPackagesAPIRequestsPerPeriod, g.ThrottleAuthenticatedPackagesAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedWebEnabled, g.ThrottleAuthenticatedWebEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedWebPeriodInSeconds, g.ThrottleAuthenticatedWebPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleAuthenticatedWebRequestsPerPeriod, g.ThrottleAuthenticatedWebRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleIncidentManagementNotificationEnabled, g.ThrottleIncidentManagementNotificationEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleIncidentManagementNotificationPerPeriod, g.ThrottleIncidentManagementNotificationPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleIncidentManagementNotificationPeriodInSeconds, g.ThrottleIncidentManagementNotificationPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleProtectedPathsEnabled, g.ThrottleProtectedPathsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleProtectedPathsPeriodInSeconds, g.ThrottleProtectedPathsPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleProtectedPathsRequestsPerPeriod, g.ThrottleProtectedPathsRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedAPIEnabled, g.ThrottleUnauthenticatedAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedAPIPeriodInSeconds, g.ThrottleUnauthenticatedAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedAPIRequestsPerPeriod, g.ThrottleUnauthenticatedAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedDeprecatedAPIEnabled, g.ThrottleUnauthenticatedDeprecatedAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds, g.ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod, g.ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedFilesAPIEnabled, g.ThrottleUnauthenticatedFilesAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedFilesAPIPeriodInSeconds, g.ThrottleUnauthenticatedFilesAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedFilesAPIRequestsPerPeriod, g.ThrottleUnauthenticatedFilesAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedGitLFSEnabled, g.ThrottleUnauthenticatedGitLFSEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedGitLFSPeriodInSeconds, g.ThrottleUnauthenticatedGitLFSPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedGitLFSRequestsPerPeriod, g.ThrottleUnauthenticatedGitLFSRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedPackagesAPIEnabled, g.ThrottleUnauthenticatedPackagesAPIEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedPackagesAPIPeriodInSeconds, g.ThrottleUnauthenticatedPackagesAPIPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod, g.ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedWebEnabled, g.ThrottleUnauthenticatedWebEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedWebPeriodInSeconds, g.ThrottleUnauthenticatedWebPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedWebRequestsPerPeriod, g.ThrottleUnauthenticatedWebRequestsPerPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.TimeTrackingLimitToHours, g.TimeTrackingLimitToHours) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.TwoFactorGracePeriod, g.TwoFactorGracePeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UnconfirmedUsersDeleteAfterDays, g.UnconfirmedUsersDeleteAfterDays) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UniqueIPsLimitEnabled, g.UniqueIPsLimitEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UniqueIPsLimitPerUser, g.UniqueIPsLimitPerUser) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UniqueIPsLimitTimeWindow, g.UniqueIPsLimitTimeWindow) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UpdateRunnerVersionsEnabled, g.UpdateRunnerVersionsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UpdatingNameDisabledForUsers, g.UpdatingNameDisabledForUsers) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UsagePingEnabled, g.UsagePingEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UsagePingFeaturesEnabled, g.UsagePingFeaturesEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UseClickhouseForAnalytics, g.UseClickhouseForAnalytics) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserDeactivationEmailsEnabled, g.UserDeactivationEmailsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserDefaultExternal, g.UserDefaultExternal) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserDefaultInternalRegex, g.UserDefaultInternalRegex) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserDefaultsToPrivateProfile, g.UserDefaultsToPrivateProfile) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserEmailLookupLimit, g.UserEmailLookupLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserOauthApplications, g.UserOauthApplications) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UserShowAddSSHKeyMessage, g.UserShowAddSSHKeyMessage) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UsersGetByIDLimit, g.UsersGetByIDLimit) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.UsersGetByIDLimitAllowlistRaw, g.UsersGetByIDLimitAllowlistRaw) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.ValidRunnerRegistrars, g.ValidRunnerRegistrars) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.VersionCheckEnabled, g.VersionCheckEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.WebIDEClientsidePreviewEnabled, g.WebIDEClientsidePreviewEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.WhatsNewVariant, g.WhatsNewVariant) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.WikiPageMaxContentBytes, g.WikiPageMaxContentBytes) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AdminNotificationEmail, g.AdminNotificationEmail) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.AllowLocalRequestsFromHooksAndServices, g.AllowLocalRequestsFromHooksAndServices) {
		return false
	}
	if !clients.IsComparableSliceEqualToComparableSlicePtr(p.AssetProxyWhitelist, g.AssetProxyWhitelist) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.DefaultBranchProtection, g.DefaultBranchProtection) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HousekeepingBitmapsEnabled, g.HousekeepingBitmapsEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HousekeepingFullRepackPeriod, g.HousekeepingFullRepackPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HousekeepingGcPeriod, g.HousekeepingGcPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.HousekeepingIncrementalRepackPeriod, g.HousekeepingIncrementalRepackPeriod) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PerformanceBarAllowedGroupID, g.PerformanceBarAllowedGroupID) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.PerformanceBarEnabled, g.PerformanceBarEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedEnabled, g.ThrottleUnauthenticatedEnabled) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedPeriodInSeconds, g.ThrottleUnauthenticatedPeriodInSeconds) {
		return false
	}
	if !clients.IsComparableEqualToComparablePtr(p.ThrottleUnauthenticatedRequestsPerPeriod, g.ThrottleUnauthenticatedRequestsPerPeriod) {
		return false
	}
	return true
}
