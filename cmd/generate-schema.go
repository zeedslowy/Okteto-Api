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
			schema := GenerateJsonSchema()
			err := SaveSchema(schema, output)

			return err
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Path to the file where the json schema will be stored")
	return cmd
}

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

func GenerateJsonSchema() *jsonschema.Schema {
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	r.Anonymous = true

	schema := r.Reflect(&Manifest{})
	schema.ID = "https://okteto.com/schemas/okteto-manifest.json"
	schema.Title = "Okteto Manifest"
	schema.Required = []string{""}

	return schema
}

func SaveSchema(schema *jsonschema.Schema, outputFilePath string) error {
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFilePath, schemaBytes, 0644)
	if err != nil {
		return err
	}
	oktetoLog.Success("okteto json schema has been generated and stored in %s", output)

	return nil
}
