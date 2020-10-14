/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"log"

	"github.com/BetssonGroup/dev-tool/pkg/git"
	"github.com/BetssonGroup/dev-tool/pkg/image"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a docker image",
	Long:  `Build  docker image.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current commit
		rev, err := git.GetCommit(".")
		if err != nil {
			log.Fatalf("Error getting commit: %s", err)
		}
		err = image.BuildImage("Dockerfile", ".", fmt.Sprintf("imagename:%s", rev))
		if err != nil {
			log.Fatalf("build error - %s", err)
		}
	},
}

func init() {
	imageCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
