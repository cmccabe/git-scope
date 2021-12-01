/**
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"io"
	"sort"
	"strings"
)

type BranchMap map[string][]string

func (bmap BranchMap) Add(branch string, remote string) {
	remotes := bmap[branch]
	if remotes == nil {
		remotes = make([]string, 0)
	}
	for i := range remotes {
		if remotes[i] == remote {
			return
		}
	}
	bmap[branch] = append(remotes, remote)
}

func (bmap BranchMap) Extras() []string {
	extras := make([]string, 0)
	for branch, remotes := range bmap {
		if len(remotes) == 1 && remotes[0] == "" {
			extras = append(extras, branch)
		}
	}
	sort.Strings(extras)
	return extras
}

func loadBranchMap(repo *git.Repository) (BranchMap, error) {
	bmap := make(BranchMap)
	refs, err := repo.References()
	if err != nil {
		return bmap, fmt.Errorf("Unable to list the references in this repo: %s",
			err.Error())
	}
	err = refs.ForEach(func(ref *plumbing.Reference) error {
//		if ref.Type() == plumbing.SymbolicReference {
//			return nil
//		}
		name := ref.Name().String()
		branch := strings.TrimPrefix(name, "refs/heads/")
		if (branch != name) {
			bmap.Add(branch, "")
		} else {
			remoteBranchName := strings.TrimPrefix(name, "refs/remotes/")
			if (remoteBranchName != name) {
				firstSlash := strings.Index(remoteBranchName, "/")
				if (firstSlash > 0) {
					remote := remoteBranchName[:firstSlash]
					branch := remoteBranchName[firstSlash+1:]
					bmap.Add(branch, remote)
				}
			}
		}
		return nil
	})
	return bmap, nil
}

func doExtras(out io.Writer) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("Unable to open the git repository in the current directory: %s",
			err.Error())
	}
	bmap, err := loadBranchMap(repo)
	if err != nil {
		return fmt.Errorf("Unable to load the branch map: %s", err.Error())
	}
	extras := bmap.Extras()
	for i := range(extras) {
		fmt.Fprintf(out, "%s\n", extras[i])
	}
	return nil
}
