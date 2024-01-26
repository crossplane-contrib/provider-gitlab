package deploykeys

import (
	"context"
	"strconv"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	crpc "github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	controller "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	errNotDeployKey     = "managed resource is not a Gitlab deploy key custom resource"
	errNotFound         = "404 project deploy key not found"
	errGetFail          = "cannot get Gitlab deploy key"
	errCreateFail       = "cannot create Gitlab deploy key"
	errUpdateFail       = "cannot update Gitlab deploy key"
	errDeleteFail       = "cannot delete Gitlab deploy key"
	errKeyMissing       = "missing key ref value"
	errIDNotAnInt       = "external-name is not an int"
	errProjectIDMissing = "missing project ID"
)

type external struct {
	kube   client.Client
	client projects.DeployKeyClient
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(clientConfig clients.Config) projects.DeployKeyClient
}

// SetupDeployKey adds a controller that reconciles ProjectDeployKey.
func SetupDeployKey(manager controller.Manager, o crpc.Options) error {
	name := managed.ControllerName(v1alpha1.DeployKeyKind)

	connector := &connector{kube: manager.GetClient(), newGitlabClientFn: newDeployKeyClient}

	reconciler := managed.NewReconciler(manager,
		resource.ManagedKind(v1alpha1.DeployKeyGroupVersionKind),
		managed.WithExternalConnecter(connector),
		managed.WithInitializers(managed.NewDefaultProviderConfig(manager.GetClient())),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(manager.GetEventRecorderFor(name))))

	return controller.NewControllerManagedBy(manager).
		Named(name).
		For(&v1alpha1.DeployKey{}).
		Complete(reconciler)
}

func (c *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha1.DeployKey)

	if !ok {
		return nil, errors.New(errNotDeployKey)
	}

	config, err := clients.GetConfig(ctx, c.kube, cr)

	if err != nil {
		return nil, err
	}

	return &external{kube: c.kube, client: c.newGitlabClientFn(*config)}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)

	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDeployKey)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	id, err := strconv.Atoi(meta.GetExternalName(cr))

	if err != nil {
		return managed.ExternalObservation{}, errors.New(errIDNotAnInt)
	}

	dk, res, err := e.client.GetDeployKey(
		*cr.Spec.ForProvider.ProjectID,
		id,
	)

	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFail)
	}

	currentState := cr.Spec.ForProvider.DeepCopy()
	lateInitializeProjectDeployKey(&cr.Spec.ForProvider, dk)
	isLateInitialized := !cmp.Equal(currentState, &cr.Spec.ForProvider)

	cr.Status.AtProvider = v1alpha1.DeployKeyObservation{
		ID:        &dk.ID,
		CreatedAt: clients.TimeToMetaTime(dk.CreatedAt),
	}

	cr.Status.SetConditions(xpv1.Available())
	isUpToDate := isUpToDate(cr, dk)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate,
		ResourceLateInitialized: isLateInitialized,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)

	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDeployKey)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	keySecretRef := cr.Spec.ForProvider.KeySecretRef

	namespacedName := types.NamespacedName{
		Namespace: keySecretRef.Namespace,
		Name:      keySecretRef.Name,
	}

	secret := &corev1.Secret{}
	err := e.kube.Get(ctx, namespacedName, secret)

	if err != nil {
		return managed.ExternalCreation{},
			errors.Wrap(err, errKeyMissing)
	}

	keyResponse, _, err := e.client.AddDeployKey(
		*cr.Spec.ForProvider.ProjectID,
		generateCreateOptions(string(secret.Data[keySecretRef.Key]), &cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFail)
	}

	id := strconv.Itoa(keyResponse.ID)
	meta.SetExternalName(cr, id)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)

	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDeployKey)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	idString := meta.GetExternalName(cr)
	id, err := strconv.Atoi(idString)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errIDNotAnInt)
	}

	_, _, er := e.client.UpdateDeployKey(
		cr.Spec.ForProvider.ProjectID,
		id,
		generateUpdateOptions(cr),
	)

	return managed.ExternalUpdate{}, errors.Wrap(er, errUpdateFail)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DeployKey)

	if !ok {
		return errors.New(errDeleteFail)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return errors.New(errProjectIDMissing)
	}

	keyIDString := meta.GetExternalName(cr)
	keyID, err := strconv.Atoi(keyIDString)

	if err != nil {
		return errors.Wrap(err, errIDNotAnInt)
	}

	_, err = e.client.DeleteDeployKey(
		*cr.Spec.ForProvider.ProjectID,
		keyID,
	)

	return errors.Wrap(err, errDeleteFail)
}

func lateInitializeProjectDeployKey(local *v1alpha1.DeployKeyParameters, external *gitlab.ProjectDeployKey) {
	if external == nil {
		return
	}

	if local.CanPush == nil {
		local.CanPush = &external.CanPush
	}
}

func generateCreateOptions(externalName string, params *v1alpha1.DeployKeyParameters) *gitlab.AddDeployKeyOptions {
	return &gitlab.AddDeployKeyOptions{
		Key:     &externalName,
		Title:   &params.Title,
		CanPush: params.CanPush,
	}
}

func newDeployKeyClient(clientConfig clients.Config) projects.DeployKeyClient {
	return clients.NewClient(clientConfig).DeployKeys
}

func generateUpdateOptions(customResourse *v1alpha1.DeployKey) *gitlab.UpdateDeployKeyOptions {
	return &gitlab.UpdateDeployKeyOptions{
		Title:   &customResourse.Spec.ForProvider.Title,
		CanPush: customResourse.Spec.ForProvider.CanPush,
	}
}

func isUpToDate(cr *v1alpha1.DeployKey, dk *gitlab.ProjectDeployKey) bool {
	isCanPushUpToDate := pointer.BoolEqual(cr.Spec.ForProvider.CanPush, &dk.CanPush)
	isTitleUpToDate := cr.Spec.ForProvider.Title == dk.Title

	return isCanPushUpToDate && isTitleUpToDate
}
