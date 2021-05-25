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
)

type ScopeCommit struct {
	firstLine string
	hash      plumbing.Hash
}

func (commit *ScopeCommit) String() string {
	return fmt.Sprintf("%s\t%s", commit.hash, commit.firstLine)
}

type ScopeBranch struct {
	name                string
	firstLinesToCommits map[string]*ScopeCommit
	commits             []*ScopeCommit
}

func newScopeBranch(branchName string, repo *git.Repository) (*ScopeBranch, error) {
//	gitBranch, err := repo.Branch(branchName)
//	if err != nil {
//		return nil, fmt.Errorf("Unable to access branch %s: %s", branchName, err.Error())
//	}
//	startHash, err := repo.ResolveRevision(plumbing.Revision(gitBranch.Merge))
	startHash, err := repo.ResolveRevision(plumbing.Revision(branchName))
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve revision for %s: %s", branchName, err.Error())
	}
	fmt.Printf("WATERMELON1: branchName = %s, resolving revision to %s\n", branchName, startHash)
	commitIter, err := repo.Log(&git.LogOptions{From: *startHash})
	if err != nil {
		return nil, fmt.Errorf("Unable to get git log for git repository in current directory: %s", err.Error())
	}
	defer commitIter.Close()
	scopeBranch := &ScopeBranch{name: branchName,
		firstLinesToCommits: make(map[string]*ScopeCommit),
		commits:             make([]*ScopeCommit, 0)}
	for {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("Error while iterating through git commits: %s", err.Error())
		}
		firstLine := GetFirstLineOfString(commit.Message)
		scopeCommit := &ScopeCommit{firstLine: firstLine, hash: commit.Hash}
		scopeBranch.firstLinesToCommits[firstLine] = scopeCommit
		scopeBranch.commits = append(scopeBranch.commits, scopeCommit)
	}
	return scopeBranch, nil
}

func (scopeBranch *ScopeBranch) Print(out io.Writer) {
	for i := range scopeBranch.commits {
		commit := scopeBranch.commits[i]
		fmt.Fprintf(out, "%s\n", commit.String())
	}
}

type ScopeBranchDiffCommit struct {
	*ScopeCommit
	branches []string
}

func (commit *ScopeBranchDiffCommit) String() string {
	return fmt.Sprintf("%s\t%s\t%s", commit.hash, commit.firstLine, commit.branches)
}

type ScopeBranchDiff struct {
	branches[]string
	commits []*ScopeBranchDiffCommit
}

func createDiff(repo *git.Repository, branchNames []string) (*ScopeBranchDiff, error) {
	var err error
	branches := make([]*ScopeBranch, len(branchNames))
	for i := range branches {
		branches[i], err = newScopeBranch(branchNames[i], repo)
		if err != nil {
			return nil, fmt.Errorf("Unable to create scope branch for source branch %s: %s",
				branchNames[i], err.Error())
		}
	}
	findCommit := func(curBranchIndex int, commit *ScopeCommit) bool {
		for i := 0; i < curBranchIndex; i++ {
			_, ok := branches[i].firstLinesToCommits[commit.firstLine]
			if ok {
				return false
			}
		}
		return true
	}
	diff := &ScopeBranchDiff{branches: make([]string, len(branchNames)), commits: make([]*ScopeBranchDiffCommit, 0)}
	copy(diff.branches, branchNames)
	for i := range branches {
		scopeBranch := branches[i]
		for j := range scopeBranch.commits {
			commit := scopeBranch.commits[j]
			if findCommit(i, commit) {
				fmt.Printf("Found commit %s in a previous scopeBranch\n", commit)
			} else {
				fmt.Printf("Did not find commit %s in a previous scopeBranch\n", commit)
				diffCommit := &ScopeBranchDiffCommit{ScopeCommit: commit, branches: make([]string, 1)}
				diffCommit.branches[0] = scopeBranch.name
				for k := i + 1; k < len(branches); k++ {
					_, ok := branches[k].firstLinesToCommits[commit.firstLine]
					if ok {
						diffCommit.branches = append(diffCommit.branches, branches[k].name)
					}
				}
				diff.commits = append(diff.commits, diffCommit)
			}
		}
	}
	return diff, nil
}

func (diff *ScopeBranchDiff) Print(out io.Writer) {
	for i := range diff.commits {
		commit := diff.commits[i]
		fmt.Fprintf(out, "%s\n", commit.String())
	}
}

func doDiff(out io.Writer, branchNames []string) error {
	if len(branchNames) < 2 {
		return fmt.Errorf("You must specify at least 2 branch names.\n")
	}
	duplicateBranchName := FindDuplicate(branchNames)
	if (duplicateBranchName != nil) {
		return fmt.Errorf("Found a duplicate branch name: %s.\n", *duplicateBranchName)
	}
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("Unable to open the git repository in the current directory: %s",
			err.Error())
	}
	diff, err := createDiff(repo, branchNames)
	if err != nil {
		return fmt.Errorf("Unable to create diff: %s", err.Error())
	}
	diff.Print(out)
	return nil
}

//	func (scopeBranch *ScopeBranch) Diff(otherBranch *ScopeBranch) *ScopeBranchDiff {
//		i := 0
//		scopeBranchCommitsLen := len(scopeBranch.commits)
//		otherBranchCommitsLen := len(otherBranch.commits)
//		diff := &ScopeBranchDiff{commits: make([]*ScopeBranchDiffCommit, 0)}
//		for {
//		numProcessed := 0
//		if i < scopeBranchCommitsLen {
//		numProcessed++
//		commit := scopeBranch.commits[i]
//		otherCommit := otherBranch.descriptionsToCommits[commit.firstLine]
//		if otherCommit == nil {
//		diff.commits = append(diff.commits,
//		&ScopeBranchDiffCommit{commit, []string{otherBranch.name}})
//	}
//	}
//		if i < otherBranchCommitsLen {
//		numProcessed++
//		otherCommit := otherBranch.commits[i]
//		commit := scopeBranch.descriptionsToCommits[otherCommit.firstLine]
//		if commit == nil {
//		diff.commits = append(diff.commits,
//		&ScopeBranchDiffCommit{otherCommit, []string{scopeBranch.name}})
//	} else {
//		fmt.Printf("WATERMELON1: %s matches %s\n", otherCommit.String(), commit)
//	}
//	}
//		if numProcessed == 0 {
//		break
//	}
//		i++
//	}
//		return diff
//	}

//		func newScopeBranch(branchName string, repo *git.Repository) (*ScopeBranch, error) {
//
//	srcBranch, err := newScopeBranch(srcBranchName, repo)
//	if err != nil {
//		return fmt.Errorf("Unable to create scope branch for source branch %s: %s",
//			srcBranchName, err.Error())
//	}
//	fmt.Fprintf(out, "** srcBranch\n")
//	srcBranch.Print(out)
//
//	dstBranch, err := newScopeBranch(dstBranchName, repo)
//	if err != nil {
//		return fmt.Errorf("Unable to create scope branch for destination branch %s: %s",
//			dstBranchName, err.Error())
//	}
//	fmt.Fprintf(out, "** dstBranch\n")
//	dstBranch.Print(out)
//
//	fmt.Fprintf(out, "** diff\n")
//	diff := dstBranch.Diff(srcBranch)
//	diff.Print(out)
//
//	return nil
//}
