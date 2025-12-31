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

package license

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/crossplane-contrib/provider-gitlab/apis/namespaced/instance/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/common"
)

const (
	errFetchFromEndpoint = "cannot fetch license from endpoint"
)

// getLicenseFromSecrets updates the License depending on the provided secrets references
// If there is a LicenseEndpointURL / LicenseEndpointURLSecretRef, it will be used to fetch the license key
// Else if there is a LicenseSecretRef, it will be used to get the license key
// Else the License in the spec will be used
func (e *external) getLicenseFromSecrets(mg resource.Managed, ctx context.Context, params *v1alpha1.LicenseParameters) (managed.ConnectionDetails, error) {
	connectionDetails := managed.ConnectionDetails{}
	hasEndpoint := params.LicenseEndpointURL != nil || params.LicenseEndpointURLSecretRef != nil
	switch {
	case hasEndpoint:
		// Retrieve endpoint url
		url, err := e.getEndpointURL(mg, ctx, params, &connectionDetails)
		if err != nil {
			return nil, err
		}

		// Retrieve optional auth
		auth, err := e.getEndpointAuth(mg, ctx, params, &connectionDetails)
		if err != nil {
			return nil, err
		}

		// Fetch license from endpoint
		license, err := common.FetchFromEndpoint(ctx, common.RequestParameters{
			EndpointURL: url,
			Auth:        auth,
		})
		if err != nil {
			return nil, errors.New(errFetchFromEndpoint)
		}
		connectionDetails[keyLicense] = []byte(license)

	case params.LicenseSecretRef != nil:
		// Retrieve license from secret reference
		license, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, params.LicenseSecretRef)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get license from secret reference")
		}
		connectionDetails[keyLicense] = []byte(*license)
	case params.License != nil:
		// Use license from spec
		connectionDetails[keyLicense] = []byte(*params.License)
	default:
		return nil, errors.New("no license source provided; please specify either LicenseEndpointURL, LicenseEndpointURLSecretRef, LicenseSecretRef, or License in the spec")
	}

	return connectionDetails, nil
}

// getEndpointURL retrieves the LicenseEndpointURL from the secret reference if provided.
// Else it returns the LicenseEndpointURL from the spec.
// It also updates the connectionDetails with the retrieved value.
func (e *external) getEndpointURL(mg resource.Managed, ctx context.Context, params *v1alpha1.LicenseParameters, connectionDetails *managed.ConnectionDetails) (string, error) {
	if connectionDetails == nil {
		return "", errors.New("connectionDetails cannot be nil")
	}

	if params.LicenseEndpointURLSecretRef != nil {
		url, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, params.LicenseEndpointURLSecretRef)
		if err != nil {
			return "", errors.Wrap(err, "cannot get license endpoint URL from secret reference")
		}
		// Do not persist endpoint metadata in connection details; it is only used to fetch the license.
		return *url, nil
	} else if params.LicenseEndpointURL != nil {
		// Do not persist endpoint metadata in connection details; it is only used to fetch the license.
		return *params.LicenseEndpointURL, nil
	}

	return "", errors.New("license endpoint URL must be provided via LicenseEndpointURL or LicenseEndpointURLSecretRef")
}

// getEndpointAuth retrieves the LicenseEndpointUsername, LicenseEndpointPassword and LicenseEndpointToken
// from the secret references if provided.
// Else it returns the values from the spec.
// It also updates the connectionDetails with the retrieved values.
func (e *external) getEndpointAuth(mg resource.Managed, ctx context.Context, params *v1alpha1.LicenseParameters, _ *managed.ConnectionDetails) (*common.AuthParameters, error) {

	auth := &common.AuthParameters{}

	// Get username
	if params.LicenseEndpointUsernameSecretRef != nil {
		username, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, params.LicenseEndpointUsernameSecretRef)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get license endpoint username from secret reference")
		}
		auth.Username = username
	} else if params.LicenseEndpointUsername != nil {
		auth.Username = params.LicenseEndpointUsername
	}

	// Get password
	if params.LicenseEndpointPasswordSecretRef != nil {
		password, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, params.LicenseEndpointPasswordSecretRef)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get license endpoint password from secret reference")
		}
		auth.Password = password
	} else if params.LicenseEndpointPassword != nil {
		auth.Password = params.LicenseEndpointPassword
	}

	// Get token
	if params.LicenseEndpointTokenSecretRef != nil {
		token, err := common.GetTokenValueFromLocalSecret(ctx, e.kube, mg, params.LicenseEndpointTokenSecretRef)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get license endpoint token from secret reference")
		}
		auth.Token = token
	} else if params.LicenseEndpointToken != nil {
		auth.Token = params.LicenseEndpointToken
	}

	return auth, nil
}

// isErrorFetchingLicenseFromEndpoint checks whether the error is due to failure in fetching license from endpoint
func isErrorFetchingLicenseFromEndpoint(err error) bool {
	return err != nil && err.Error() == errFetchFromEndpoint
}
