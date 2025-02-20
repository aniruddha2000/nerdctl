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

package volume

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/nerdctl/pkg/api/types"
)

func Prune(ctx context.Context, client *containerd.Client, options types.VolumePruneOptions) error {
	if !options.Force {
		var confirm string
		msg := "This will remove all local volumes not used by at least one container."
		msg += "\nAre you sure you want to continue? [y/N] "
		fmt.Fprintf(options.Stdout, "WARNING! %s", msg)
		fmt.Fscanf(options.Stdin, "%s", &confirm)

		if strings.ToLower(confirm) != "y" {
			return nil
		}
	}
	volStore, err := Store(options.GOptions.Namespace, options.GOptions.DataRoot, options.GOptions.Address)
	if err != nil {
		return err
	}
	volumes, err := volStore.List(false)
	if err != nil {
		return err
	}

	containers, err := client.Containers(ctx)
	if err != nil {
		return err
	}
	usedVolumes, err := usedVolumes(ctx, containers)
	if err != nil {
		return err
	}
	var removeNames []string // nolint: prealloc
	for _, volume := range volumes {
		if _, ok := usedVolumes[volume.Name]; ok {
			continue
		}
		removeNames = append(removeNames, volume.Name)
	}
	removedNames, err := volStore.Remove(removeNames)
	if err != nil {
		return err
	}
	if len(removedNames) > 0 {
		fmt.Fprintln(options.Stdout, "Deleted Volumes:")
		for _, name := range removedNames {
			fmt.Fprintln(options.Stdout, name)
		}
		fmt.Fprintln(options.Stdout, "")
	}
	return nil
}
