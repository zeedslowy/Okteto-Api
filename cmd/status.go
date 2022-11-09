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

package cmd

import (
	"context"
	"fmt"

	contextCMD "github.com/okteto/okteto/cmd/context"
	"github.com/okteto/okteto/cmd/utils"
	"github.com/okteto/okteto/pkg/cmd/status"
	"github.com/okteto/okteto/pkg/config"
	oktetoErrors "github.com/okteto/okteto/pkg/errors"
	"github.com/okteto/okteto/pkg/okteto"
	"github.com/okteto/okteto/pkg/syncthing"
	"github.com/spf13/cobra"
)

// Status returns the status of the synchronization process
func Status() *cobra.Command {
	var devPath string
	var namespace string
	var k8sContext string
	var showInfo bool
	var watch bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of the synchronization process",
		Args:  utils.MaximumNArgsAccepted(1, "https://okteto.com/docs/reference/cli/#status"),
		RunE: func(cmd *cobra.Command, args []string) error {

			if okteto.InDevContainer() {
				return oktetoErrors.ErrNotInDevContainer
			}

			ctx := context.Background()

			manifestOpts := contextCMD.ManifestOptions{Filename: devPath, Namespace: namespace, K8sContext: k8sContext}
			manifest, err := contextCMD.LoadManifestWithContext(ctx, manifestOpts)
			if err != nil {
				return err
			}

			dev, err := utils.GetDevDetachMode(manifest, []string{})
			if err != nil {
				return err
			}

			status, err := config.GetState(dev)
			if err != nil {
				return err
			}
			if status == "synchronizing" {
				sy, err := syncthing.Load(dev)
				if err == nil && isSynchronized(ctx, sy) {
					status = "ready"
				}
			}
			fmt.Printf("{\"status\": \"%s\"}\n", status)
			return nil
		},
	}
	cmd.Flags().StringVarP(&devPath, "file", "f", utils.DefaultManifest, "path to the manifest file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace where the up command is executing")
	cmd.Flags().StringVarP(&k8sContext, "context", "c", "", "context where the up command is executing")
	cmd.Flags().BoolVarP(&showInfo, "info", "i", false, "show syncthing links for troubleshooting the synchronization service")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for changes")
	return cmd
}

func isSynchronized(ctx context.Context, sy *syncthing.Syncthing) bool {
	progress, err := status.Run(ctx, sy)
	if err != nil {
		return false
	}
	if progress == 100 {
		return true
	}
	return false
}
