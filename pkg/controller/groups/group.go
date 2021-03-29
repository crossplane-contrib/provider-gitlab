/*
Copyright 2020 The Crossplane Authors.

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
	"context"
	"reflect"
	"strconv"

	"github.com/xanzy/go-gitlab"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/groups/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/groups"
)

const (
	errNotGroup         = "managed resource is not a Gitlab Group custom resource"
	errGetFailed        = "cannot get Gitlab Group"
	errKubeUpdateFailed = "cannot update Gitlab Group custom resource"
	errCreateFailed     = "cannot create Gitlab Group"
	errUpdateFailed     = "cannot update Gitlab Group"
	errDeleteFailed     = "cannot delete Gitlab Group"
)

// SetupGroup adds a controller that reconciles Groups.
func SetupGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.GroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Group{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.GroupKubernetesGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: groups.NewGroupClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) groups.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return nil, errors.New(errNotGroup)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client groups.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroup)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	grp, _, err := e.client.GetGroup(meta.GetExternalName(cr), nil)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(groups.IsErrorGroupNotFound, err), errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	groups.LateInitialize(&cr.Spec.ForProvider, grp)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(context.Background(), cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = groups.GenerateObservation(grp)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        groups.IsGroupUpToDate(&cr.Spec.ForProvider, grp),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
		ConnectionDetails:       managed.ConnectionDetails{"runnersToken": []byte(grp.RunnersToken)},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroup)
	}

	cr.Status.SetConditions(xpv1.Creating())
	grp, _, err := e.client.CreateGroup(groups.GenerateCreateGroupOptions(cr.Name, &cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	err = e.updateExternalName(cr, grp)
	return managed.ExternalCreation{}, errors.Wrap(err, errKubeUpdateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroup)
	}

	_, _, err := e.client.UpdateGroup(meta.GetExternalName(cr), groups.GenerateEditGroupOptions(cr.Name, &cr.Spec.ForProvider), gitlab.WithContext(ctx))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, errors.Wrap(err, errKubeUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return errors.New(errNotGroup)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteGroup(meta.GetExternalName(cr), gitlab.WithContext(ctx))
	return errors.Wrap(err, errDeleteFailed)
}

func (e *external) updateExternalName(cr *v1alpha1.Group, grp *gitlab.Group) error {
	meta.SetExternalName(cr, strconv.Itoa(grp.ID))
	return e.kube.Update(context.Background(), cr)
}
