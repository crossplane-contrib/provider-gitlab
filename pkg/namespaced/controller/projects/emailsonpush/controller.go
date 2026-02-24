package emailsonpush

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
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
	errNotEmailsOnPush  = "managed resource is not a EmailsOnPush custom resource"
	errProjectIDMissing = "ProjectID is missing"
	errGetFailed        = "cannot get EmailsOnPush service"
	errSetFailed        = "cannot set EmailsOnPush service"
)

// SetupEmailsOnPush adds a controller that reconciles EmailsOnPush.
func SetupEmailsOnPush(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.EmailsOnPushGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: projects.NewEmailsOnPushClient,
		}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(
		mgr,
		resource.ManagedKind(v1alpha1.EmailsOnPushGroupVersionKind),
		reconcilerOpts...,
	)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(),
		o.Logger,
		o.MetricOptions.MRStateMetrics,
		&v1alpha1.EmailsOnPushList{},
		o.MetricOptions.PollStateMetricInterval,
	)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.EmailsOnPush{}).
		Complete(r)
}

// SetupEmailsOnPushGated adds controller with CRD gate support.
func SetupEmailsOnPushGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupEmailsOnPush(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler",
				"gvk", v1alpha1.EmailsOnPushGroupVersionKind.String())
		}
	}, v1alpha1.EmailsOnPushGroupVersionKind)

	return nil
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg common.Config) projects.EmailsOnPushClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.EmailsOnPush)
	if !ok {
		return nil, errors.New(errNotEmailsOnPush)
	}

	cfg, err := common.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &external{
		kube:   c.kube,
		client: c.newGitlabClientFn(*cfg),
	}, nil
}

type external struct {
	kube   client.Client
	client projects.EmailsOnPushClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.EmailsOnPush)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEmailsOnPush)
	}

	if meta.WasDeleted(cr) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	svc, resp, err := e.client.GetEmailsOnPushService(
		*cr.Spec.ForProvider.ProjectID,
		gitlab.WithContext(ctx),
	)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: false,
			}, nil
		}

		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	before := cr.Spec.ForProvider
	projects.LateInitializeEmailsOnPush(&cr.Spec.ForProvider, svc)

	if before != cr.Spec.ForProvider && e.kube != nil {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "cannot update EmailsOnPush spec after late init")
		}
	}

	cr.Status.AtProvider = projects.GenerateEmailsOnPushObservation(svc)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	// GitLab integration endpoint uses PUT (idempotent),
	// so Create does nothing — Update handles configuration.
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.EmailsOnPush)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEmailsOnPush)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	options := projects.GenerateSetEmailsOnPushOptions(&cr.Spec.ForProvider)

	_, _, err := e.client.SetEmailsOnPushService(
		*cr.Spec.ForProvider.ProjectID,
		options,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errSetFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}
