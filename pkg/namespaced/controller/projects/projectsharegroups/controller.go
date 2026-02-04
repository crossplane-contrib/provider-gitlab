package projectsharegroups

import (
	"context"
	"strconv"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
	"github.com/crossplane-contrib/provider-gitlab/pkg/namespaced/clients/projects"
)

const (
	errNotProjectShareGroup = "managed resource is not a ProjectShareGroup"
	errGetProject           = "cannot get project from GitLab"
	errShareProject         = "cannot share project with group"
	errUnshareProject       = "cannot unshare project from group"
)

// SetupProjectShareGroup adds a controller that reconciles ProjectShareGroups.
func SetupProjectShareGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ProjectShareGroupGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProjectClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ProjectShareGroupGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.ProjectShareGroupList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ProjectShareGroup{}).
		Complete(r)
}

// SetupProjectShareGroupGated adds a controller with CRD gate support.
func SetupProjectShareGroupGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupProjectShareGroup(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.ProjectShareGroupGroupVersionKind.String())
		}
	}, v1alpha1.ProjectShareGroupGroupVersionKind)
	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ProjectShareGroup)
	if !ok {
		return nil, errors.New(errNotProjectShareGroup)
	}

	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &external{client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	client projects.Client
}

// Observe gets the Project, then looks for the Group in the 'shared_with_groups' list.
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ProjectShareGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProjectShareGroup)
	}

	if cr.Spec.ForProvider.ProjectID == nil || cr.Spec.ForProvider.GroupID == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	p, _, err := e.client.GetProject(*cr.Spec.ForProvider.ProjectID, nil, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errGetProject)
	}

	targetGroupID, err := strconv.ParseInt(*cr.Spec.ForProvider.GroupID, 10, 64)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "GroupID must be an integer")
	}

	var foundShare *gitlab.ProjectSharedWithGroup
	for i, share := range p.SharedWithGroups {
		if share.GroupID == targetGroupID {
			foundShare = &p.SharedWithGroups[i]
			break
		}
	}

	if foundShare == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider.ID = *cr.Spec.ForProvider.ProjectID + "-" + *cr.Spec.ForProvider.GroupID
	cr.SetConditions(xpv1.Available())

	isUpToDate := int64(cr.Spec.ForProvider.AccessLevel) == foundShare.GroupAccessLevel

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: isUpToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ProjectShareGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProjectShareGroup)
	}

	groupID, err := strconv.ParseInt(*cr.Spec.ForProvider.GroupID, 10, 64)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "GroupID must be an integer")
	}

	opt := &gitlab.ShareWithGroupOptions{
		GroupID:     gitlab.Ptr(groupID),
		GroupAccess: gitlab.Ptr(gitlab.AccessLevelValue(cr.Spec.ForProvider.AccessLevel)),
	}
	if cr.Spec.ForProvider.ExpiresAt != nil {
		opt.ExpiresAt = cr.Spec.ForProvider.ExpiresAt
	}

	_, err = e.client.ShareProjectWithGroup(*cr.Spec.ForProvider.ProjectID, opt, gitlab.WithContext(ctx))

	return managed.ExternalCreation{}, errors.Wrap(err, errShareProject)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, err := e.Delete(ctx, mg)
	return managed.ExternalUpdate{}, err
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.ProjectShareGroup)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotProjectShareGroup)
	}

	groupID, err := strconv.ParseInt(*cr.Spec.ForProvider.GroupID, 10, 64)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "GroupID must be an integer")
	}

	_, err = e.client.DeleteSharedProjectFromGroup(*cr.Spec.ForProvider.ProjectID, groupID, gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errUnshareProject)
	}
	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
