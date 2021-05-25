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
	"os"
	"strings"
)

type ScopeBranch struct {
	name    string
	commits map[string]ScopeCommit
}

func newScopeBranch(branchName string, repo *git.Repository) (*ScopeBranch, error) {
	gitBranch, err := repo.Branch(branchName)
	if err != nil {
		return nil, fmt.Errorf("Unable to access branch %s: %s", branchName, err.Error())
	}
	startHash, err := repo.ResolveRevision(plumbing.Revision(gitBranch.Merge)) //gitBranch.Name
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve revision for %s: %s", gitBranch, err.Error())
	}
	commitIter, err := repo.Log(&git.LogOptions{From: *startHash})
	if err != nil {
		return nil, fmt.Errorf("Unable to get git log for git repository in current directory: %s", err.Error())
	}
	defer commitIter.Close()
	scopeBranch := &ScopeBranch{name: branchName,
		commits: make(map[string]ScopeCommit)}
	for {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("Error while iterating through git commits: %s", err.Error())
		}
		firstLine := getFirstLine(commit.Message)
		scopeBranch.commits[firstLine] = ScopeCommit{firstLine: firstLine, hash: commit.Hash}
	}
	return scopeBranch, nil
}

func getFirstLine(input string) string {
	i := strings.Index(input, "\n")
	if i < 0 {
		return input
	}
	return input[:i]
}

type ScopeCommit struct {
	firstLine string
	hash      plumbing.Hash
}

func doDiff(out *os.File, srcBranchName string, dstBranchName string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("Unable to open git repository in current directory: %s",
			err.Error())
	}
	_, err = newScopeBranch(srcBranchName, repo)
	if err != nil {
		return fmt.Errorf("Unable to create scope branch for source branch %s: %s",
			srcBranchName, err.Error())
	}
	_, err = newScopeBranch(dstBranchName, repo)
	if err != nil {
		return fmt.Errorf("Unable to create scope branch for destination branch %s: %s",
			srcBranchName, err.Error())
	}
	return nil
}
