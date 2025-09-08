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

package runners

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/cluster/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	runners "github.com/crossplane-contrib/provider-gitlab/pkg/clients/runners"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/users"
	shared "github.com/crossplane-contrib/provider-gitlab/pkg/controller/shared/groups/runners"
)

// SetupRunner adds a controller that reconciles runners.
func SetupRunner(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RunnerGroupKind)

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:              mgr.GetClient(),
			newGitlabClientFn: runners.NewRunnerClient,
			newRunnerClientFn: users.NewRunnerClient,
		}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}
	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RunnerGroupVersionKind),
		reconcilerOpts...)

	if err := mgr.Add(statemetrics.NewMRStateRecorder(
		mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &v1alpha1.RunnerList{}, o.MetricOptions.PollStateMetricInterval)); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Runner{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) runners.RunnerClient
	newRunnerClientFn func(cfg clients.Config) users.RunnerClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Runner)
	if !ok {
		return nil, errors.New(shared.ErrNotRunner)
	}

	cfg, err := clients.ResolveProviderConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &shared.External{Client: c.newGitlabClientFn(*cfg), UserRunnerClient: c.newRunnerClientFn(*cfg)}, nil
}

// SetupRunnerGated adds a controller with CRD gate support.
func SetupRunnerGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := SetupRunner(mgr, o); err != nil {
			mgr.GetLogger().Error(err, "unable to setup reconciler", "gvk", v1alpha1.RunnerGroupVersionKind.String())
		}
	}, v1alpha1.RunnerGroupVersionKind)
	return nil
}
