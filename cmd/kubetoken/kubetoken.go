// Copyright 2023 The Okteto Authors
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

package kubetoken

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/okteto/okteto/pkg/types"
	"os"

	contextCMD "github.com/okteto/okteto/cmd/context"

	"github.com/okteto/okteto/pkg/errors"
	"github.com/okteto/okteto/pkg/okteto"
	"github.com/spf13/cobra"
)

type KubeTokenSerializer struct {
	KubeToken types.KubeTokenResponse
}

func (k *KubeTokenSerializer) ToJson() (string, error) {
	bytes, err := json.MarshalIndent(k.KubeToken, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func KubeToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubetoken",
		Short: "Print Kubernetes cluster credentials in ExecCredential format.",
		Long: `Print Kubernetes cluster credentials in ExecCredential format.
You can find more information on 'ExecCredential' and 'client side authentication' at (https://kubernetes.io/docs/reference/config-api/client-authentication.v1/) and  https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins`,
		Hidden: true,
		Args:   cobra.NoArgs,
	}

	var namespace string
	var contextName string
	cmd.RunE = func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		err := contextCMD.NewContextCommand().Run(ctx, &contextCMD.ContextOptions{
			Context:   contextName,
			Namespace: namespace,
		})
		if err != nil {
			return err
		}

		octx := okteto.Context()
		if !octx.IsOkteto {
			return errors.ErrContextIsNotOktetoCluster
		}

		c, err := okteto.NewOktetoClient()
		if err != nil {
			return fmt.Errorf("failed to initialize the kubetoken client: %w", err)
		}

		out, err := c.Kubetoken().GetKubeToken(octx.Name, octx.Namespace)
		if err != nil {
			return fmt.Errorf("failed to get the kubetoken: %w", err)
		}

		serializer := &KubeTokenSerializer{
			KubeToken: *out,
		}

		jsonStr, err := serializer.ToJson()
		if err != nil {
			return fmt.Errorf("failed to marshal KubeTokenResponse: %w", err)
		}

		cmd.Print(jsonStr)
		return nil
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "okteto context's namespace")
	cmd.Flags().StringVarP(&contextName, "context", "c", "", "okteto context's name")

	cmd.SetOut(os.Stdout)

	return cmd
}
