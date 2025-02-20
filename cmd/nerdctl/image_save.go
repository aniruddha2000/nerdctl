/*
   Copyright The containerd Authors.

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

package main

import (
	"fmt"

	"github.com/containerd/nerdctl/pkg/api/types"
	"github.com/containerd/nerdctl/pkg/clientutil"
	"github.com/containerd/nerdctl/pkg/cmd/image"
	"github.com/spf13/cobra"
)

func newSaveCommand() *cobra.Command {
	var saveCommand = &cobra.Command{
		Use:               "save",
		Args:              cobra.MinimumNArgs(1),
		Short:             "Save one or more images to a tar archive (streamed to STDOUT by default)",
		Long:              "The archive implements both Docker Image Spec v1.2 and OCI Image Spec v1.0.",
		RunE:              saveAction,
		ValidArgsFunction: saveShellComplete,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}
	saveCommand.Flags().StringP("output", "o", "", "Write to a file, instead of STDOUT")

	// #region platform flags
	// platform is defined as StringSlice, not StringArray, to allow specifying "--platform=amd64,arm64"
	saveCommand.Flags().StringSlice("platform", []string{}, "Export content for a specific platform")
	saveCommand.RegisterFlagCompletionFunc("platform", shellCompletePlatforms)
	saveCommand.Flags().Bool("all-platforms", false, "Export content for all platforms")
	// #endregion

	return saveCommand
}

func processImageSaveOptions(cmd *cobra.Command) (types.ImageSaveOptions, error) {
	globalOptions, err := processRootCmdFlags(cmd)
	if err != nil {
		return types.ImageSaveOptions{}, err
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return types.ImageSaveOptions{}, err
	}
	allPlatforms, err := cmd.Flags().GetBool("all-platforms")
	if err != nil {
		return types.ImageSaveOptions{}, err
	}
	platform, err := cmd.Flags().GetStringSlice("platform")
	if err != nil {
		return types.ImageSaveOptions{}, err
	}

	return types.ImageSaveOptions{
		GOptions:     globalOptions,
		Stdout:       cmd.OutOrStdout(),
		AllPlatforms: allPlatforms,
		Output:       output,
		Platform:     platform,
	}, err
}

func saveAction(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires at least 1 argument")
	}
	options, err := processImageSaveOptions(cmd)
	if err != nil {
		return err
	}

	client, ctx, cancel, err := clientutil.NewClient(cmd.Context(), options.GOptions.Namespace, options.GOptions.Address)
	if err != nil {
		return err
	}
	defer cancel()

	return image.Save(ctx, client, args, options)
}

func saveShellComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// show image names
	return shellCompleteImageNames(cmd)
}
