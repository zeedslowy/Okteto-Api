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
	Build     map[string]model.BuildInfo `json:"build" jsonschema:"title=build,description=A list of images to build as part of your development environment."`
	Context   string                     `json:"context" jsonschema:"title=context,description=The build context. Relative paths are relative to the location of the Okteto Manifest (default: .),example=api"`
	Namespace string                     `json:"namespace" jsonschema:"title=namespace,description=The namespace where the development environment is deployed. By default, it takes the current okteto context namespace. You can use an environment variable to replace the namespace field, or any part of it: namespace: $DEV_NAMESPACE"`
	Image     string                     `json:"image" jsonschema:"title=image,description=The name of the image to build and push. In clusters that have Okteto installed, this is optional (if not specified, the Okteto Registry is used)."`
	Icon      string                     `json:"icon" jsonschema:"title=icon,description=Sets the icon that will be shown in the Okteto UI. The supported values for icons are listed below.,default=default,enum=default,enum=container,enum=dashboard,enum=database,enum=function,enum=graph,enum=storage,enum=launchdarkly,enum=mongodb,enum=gcp,enum=aws,enum=okteto"`
	// TODO: Dev breaks due to recursion of Dev.Services being an array of []*Dev
	//Dev       map[string]model.Dev       `json:"dev" jsonschema:"title=dev,description=A list of development containers to define the behavior of okteto up and synchronize your code in your development environment."`
	// TODO: deploy
	// TODO: the library doesn't allow oneof_ref and say what type they are! See: https://github.com/invopop/jsonschema/issues/68
	Destroy      interface{}                 `json:"destroy" jsonschema:"title=destroy,oneof_type=object;array,description=Allows destroying resources created by your development environment. Can be either a list of commands or an object (destroy.image, destroy.commands) which in this case will execute remotely."`
	Dependencies map[string]model.Dependency `json:"dependencies" jsonschema:"title=dependencies,description=Repositories you want to deploy as part of your development environment. This feature is only supported in clusters that have Okteto installed."`
	// TODO: make sure all are covered: https://www.okteto.com/docs/reference/manifest/#example
}

func GenerateJsonSchema() *jsonschema.Schema {
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	r.Anonymous = true
	r.AllowAdditionalProperties = false
	r.RequiredFromJSONSchemaTags = false

	schema := r.Reflect(&Manifest{})
	schema.ID = "https://okteto.com/schemas/okteto-manifest.json"
	schema.Title = "Okteto Manifest"
	schema.Required = []string{}

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
