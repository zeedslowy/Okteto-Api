// Copyright 2022 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package okteto

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/okteto/okteto/pkg/config"
	oktetoContext "github.com/okteto/okteto/pkg/context"
	"github.com/okteto/okteto/pkg/filesystem"
	"github.com/okteto/okteto/pkg/k8s/kubeconfig"
	oktetoLog "github.com/okteto/okteto/pkg/log"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// InitContextWithDeprecatedToken initializes the okteto context if an old fashion exists and it matches the current kubernetes context
// this function is to make "okteto context" transparent to current Okteto Enterprise users, but it can be removed when people upgrade
func InitContextWithDeprecatedToken() {
	if !filesystem.FileExists(config.GetTokenPathDeprecated()) {
		return
	}

	defer os.RemoveAll(config.GetTokenPathDeprecated())
	token, err := getTokenFromOktetoHome()
	if err != nil {
		oktetoLog.Infof("error accessing deprecated okteto token '%s': %v", config.GetTokenPathDeprecated(), err)
		return
	}

	k8sContext := oktetoContext.UrlToKubernetesContext(token.URL)
	if kubeconfig.CurrentContext(config.GetKubeconfigPath()) != k8sContext {
		return
	}

	ctxStore := oktetoContext.ContextStore()
	if _, ok := ctxStore.Contexts[token.URL]; ok {
		return
	}

	certificateBytes, err := os.ReadFile(config.GetCertificatePath())
	if err != nil {
		oktetoLog.Infof("error reading okteto certificate: %v", err)
		return
	}

	ctxStore.Contexts[token.URL] = &oktetoContext.OktetoContext{
		Name:        token.URL,
		Namespace:   kubeconfig.CurrentNamespace(config.GetKubeconfigPath()),
		Token:       token.Token,
		Builder:     token.Buildkit,
		Certificate: base64.StdEncoding.EncodeToString(certificateBytes),
		IsOkteto:    true,
		UserID:      token.ID,
	}
	ctxStore.CurrentContext = token.URL

	if err := oktetoContext.NewContextConfigWriter().Write(); err != nil {
		oktetoLog.Infof("error writing okteto context: %v", err)
	}
}

func GetK8sClient() (*kubernetes.Clientset, *rest.Config, error) {
	if oktetoContext.Context().Cfg == nil {
		return nil, nil, fmt.Errorf("okteto context not initialized")
	}
	c, config, err := getK8sClientWithApiConfig(oktetoContext.Context().Cfg)
	if err == nil {
		oktetoContext.Context().SetClusterType(config.Host)
	}
	return c, config, err
}

// GetDynamicClient returns a kubernetes dynamic client for the current okteto context
func GetDynamicClient() (dynamic.Interface, *rest.Config, error) {
	if oktetoContext.Context().Cfg == nil {
		return nil, nil, fmt.Errorf("okteto context not initialized")
	}
	return getDynamicClient(oktetoContext.Context().Cfg)
}

// GetDiscoveryClient return a kubernetes discovery client for the current okteto context
func GetDiscoveryClient() (discovery.DiscoveryInterface, *rest.Config, error) {
	if oktetoContext.Context().Cfg == nil {
		return nil, nil, fmt.Errorf("okteto context not initialized")
	}
	return getDiscoveryClient(oktetoContext.Context().Cfg)
}

// GetSanitizedUsername returns the username of the authenticated user sanitized to be DNS compatible
func GetSanitizedUsername() string {
	octx := oktetoContext.Context()
	return reg.ReplaceAllString(strings.ToLower(octx.Username), "-")
}

func RemoveSchema(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	return strings.TrimPrefix(u.String(), fmt.Sprintf("%s://", u.Scheme))
}

func AddSchema(oCtx string) string {
	parsedUrl, err := url.Parse(oCtx)
	if err == nil {
		if parsedUrl.Scheme == "" {
			parsedUrl.Scheme = "https"
		}
		return parsedUrl.String()
	}
	return oCtx
}
