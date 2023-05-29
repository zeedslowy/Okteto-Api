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

package cmd

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"github.com/okteto/okteto/cmd/utils"
	oktetoLog "github.com/okteto/okteto/pkg/log"
	"github.com/okteto/okteto/pkg/model"
	"github.com/spf13/cobra"
	"os"
)

var output string

// GenerateSchema create the json schema for the okteto manifest
func GenerateSchema() *cobra.Command {
	cmd := &cobra.Command{
		Args:   utils.NoArgsAccepted("https://okteto.com/docs/reference/cli/#generate-schema"),
		Hidden: true,
		Use:    "generate-schema",
		Short:  "Generates the json schema for the okteto manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateJsonSchema(output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Path to the file where the json schema will be stored")
	return cmd
}

//
//type Dev struct {
//	Dev map[string]model.Dev `json:"dev"`
//}

type Manifest struct {
	Build     map[string]model.BuildInfo `json:"build" jsonschema:""`
	Context   string                     `json:"context" jsonschema:""`
	Namespace string                     `json:"namespace" jsonschema:""`
	Image     string                     `json:"image" jsonschema:""`
	Icon      string                     `json:"icon" jsonschema:""`
	//Dev       Dev    `json:"dev" jsonschema:""`
	// TODO: deploy
	// TODO: destroy
	// TODO: dependencies
	// TODO: make sure all are covered: https://www.okteto.com/docs/reference/manifest/#example
}

func generateJsonSchema(output string) error {
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	r.Anonymous = true

	schema := r.Reflect(&Manifest{})
	schema.ID = "https://okteto.com/schemas/okteto-manifest.json"
	schema.Title = "Okteto Manifest"
	schema.Required = []string{""}
	//schema.Properties

	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	if output != "" {
		err := os.WriteFile(output, schemaBytes, 0644)
		if err != nil {
			return err
		}
		oktetoLog.Success("okteto json schema has been generated and stored in %s", output)
	} else {
		oktetoLog.Success("okteto json schema has been generated")
	}

	return nil
}
