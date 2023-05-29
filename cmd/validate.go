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
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

// Validate validates a Okteto Manifest file
func Validate() *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.MaximumNArgs(1),
		Use:   "validate [manifest]",
		Short: "Validate a Okteto Manifest file",
		RunE: func(cmd *cobra.Command, args []string) error {
			var manifestFile string
			if len(args) > 0 {
				manifestFile = args[0]
			} else {
				// look for okteto.yml or okteto.yaml
				if _, err := os.Stat("okteto.yml"); err == nil {
					manifestFile = "okteto.yml"
				} else if _, err := os.Stat("okteto.yaml"); err == nil {
					manifestFile = "okteto.yaml"
				} else {
					return errors.New("unable to locate manifest file: okteto.yml or okteto.yaml")
				}
			}

			manifest, err := ioutil.ReadFile(manifestFile)
			if err != nil {
				return err
			}

			var obj interface{}
			err = yaml.Unmarshal(manifest, &obj)
			if err != nil {
				return err
			}

			schema := GenerateJsonSchema()

			// Load JSON schema
			jsonLoader := gojsonschema.NewGoLoader(schema)

			// Load JSON document
			documentLoader := gojsonschema.NewGoLoader(obj)

			// Validate JSON document
			result, err := gojsonschema.Validate(jsonLoader, documentLoader)
			if err != nil {
				return err
			}

			if !result.Valid() {
				fmt.Printf("The document is not valid. See errors :\n")
				for _, desc := range result.Errors() {
					fmt.Printf("- %s\n", desc)
				}
			} else {
				fmt.Printf("The document is valid.\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Path to the file where the json schema will be stored")
	return cmd
}
