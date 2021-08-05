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

package deploykeys

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/crypto/ssh"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
)

const (
	errNotDeployKey = "managed resource is not a Gitlab Deploy Key custom resource"
	errCreateFailed = "cannot create Gitlab Deploy Key"
	errUpdateFailed = "cannot update Gitlab Deploy Key"
	errDeleteFailed = "cannot delete Gitlab Deploy Key"
)

// SetupDeployKey adds a controller that reconciles Project Deploy Key.
func SetupDeployKey(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DeployKeyKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.DeployKey{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DeployKeyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewDeployKeyClient}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.DeployKeyClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)
	if !ok {
		return nil, errors.New(errNotDeployKey)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.DeployKeyClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDeployKey)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	deployKeyID, err := strconv.Atoi(externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.New(errNotDeployKey)
	}

	deployKey, _, _ := e.client.GetDeployKey(*cr.Spec.ForProvider.ProjectID, deployKeyID)
	if deployKey == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	current := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, deployKey)

	cr.Status.AtProvider = projects.GenerateDeployKeyObservation(deployKey)
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isDeployKeyUpToDate(&cr.Spec.ForProvider, deployKey),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDeployKey)
	}

	connectionDetails := managed.ConnectionDetails{}
	if cr.Spec.ForProvider.Key == nil {
		publicKey, privateKey, err := createPrivateKey()
		if err != nil {
			return managed.ExternalCreation{}, errors.New(errCreateFailed)
		}
		cr.Spec.ForProvider.Key = publicKey
		connectionDetails["ssh-privatekey"] = privateKey
	}

	deployKey, _, err := e.client.AddDeployKey(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateAddDeployKeyOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, strconv.Itoa(deployKey.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true, ConnectionDetails: connectionDetails}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DeployKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDeployKey)
	}
	_, _, err := e.client.UpdateDeployKey(
		*cr.Spec.ForProvider.ProjectID,
		cr.Status.AtProvider.ID,
		projects.GenerateUpdateDeployKeyOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DeployKey)
	if !ok {
		return errors.New(errNotDeployKey)
	}

	_, err := e.client.DeleteDeployKey(
		*cr.Spec.ForProvider.ProjectID,
		cr.Status.AtProvider.ID,
		gitlab.WithContext(ctx),
	)
	return errors.Wrap(err, errDeleteFailed)
}

// lateInitialize fills the empty fields in the projecthook spec with the
// values seen in gitlab.DeployKey.
func lateInitialize(in *v1alpha1.DeployKeyParameters, deploykey *gitlab.DeployKey) { // nolint:gocyclo
	if deploykey == nil {
		return
	}

	if in.CanPush == nil {
		in.CanPush = deploykey.CanPush
	}
}

// isDeployKeyUpToDate checks whether there is a change in any of the modifiable fields.
func isDeployKeyUpToDate(p *v1alpha1.DeployKeyParameters, g *gitlab.DeployKey) bool { // nolint:gocyclo
	if !cmp.Equal(p.Title, g.Title) {
		return false
	}

	if !cmp.Equal(p.CanPush, g.CanPush) {
		return false
	}

	return true
}

func createPrivateKey() (public *string, private []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// PublicKey
	publicKey := &privateKey.PublicKey
	publicKeySSH, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyOpenSSH := ssh.MarshalAuthorizedKey(publicKeySSH)
	publicKeyString := strings.TrimSuffix(string(publicKeyOpenSSH), "\n")

	// PrivateKey
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDER,
	}

	privateKeyPEM := pem.EncodeToMemory(&privateKeyBlock)

	return &publicKeyString, privateKeyPEM, nil
}
