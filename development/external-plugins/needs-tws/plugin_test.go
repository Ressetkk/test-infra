package main

import (
	"testing"

	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"go.uber.org/zap"
	"k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
	"k8s.io/test-infra/prow/repoowners"
)

type fakeAliases struct {
	ownersAliases
	Aliases repoowners.RepoAliases
}

type fakeGitClientFactory struct {
	git.ClientFactory
}

type fakeRepoClient struct {
	git.RepoClient
}

func (f fakeAliases) LoadOwnersAliases(l *zap.SugaredLogger, basedir, filename string) (repoowners.RepoAliases, error) {
	return f.Aliases, nil
}

func (f fakeRepoClient) Directory() string {
	return ""
}

func (f fakeGitClientFactory) ClientFor(org, repo string) (git.RepoClient, error) {
	return fakeRepoClient{}, nil
}

func Test_HandlePullRequest(t *testing.T) {
	SHA := "9448a2cb0a3915ac956685de8ffb3d4ef55fbc05"
	twsLabel := "org/repo#101:do-not-merge/missing-docs-review"
	testcases := []struct {
		name                string
		event               github.PullRequestEvent
		commit              github.RepositoryCommit
		IssueLabelsAdded    []string
		IssueLabelsExisting []string
		IssueLabelsRemoved  []string
		Reviews             []github.Review
	}{
		{
			name: "pr_opened, files changed, add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
			IssueLabelsAdded: []string{twsLabel},
		},
		{
			name: "pr_opened, files not, changed, do not add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
		},
		{
			name: "pr_synchronize, files changed, add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
				SHA: SHA,
			},
			IssueLabelsAdded: []string{twsLabel},
		},
		{
			name: "pr_synchronize, files changed, already has a label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
				SHA: SHA,
			},
			IssueLabelsExisting: []string{twsLabel},
		},
		{
			name: "pr_synchronize, files not changed, do not add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionSynchronize,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
		},
		{
			name: "pr_opened is a draft",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionOpened,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
					Draft: true,
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
		},
		{
			name: "pr_closed",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionClosed,
			},
		},
		{
			name: "pr_labeled, review approved, label already removed",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionLabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
			Reviews: []github.Review{
				{
					User:  github.User{Login: "reviewer"},
					State: github.ReviewStateApproved,
				},
			},
		},
		{
			name: "pr_unlabeled, review unapproved, label already present",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionUnlabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
			Reviews: []github.Review{
				{
					User:  github.User{Login: "reviewer"},
					State: github.ReviewStateChangesRequested,
				},
			},
			IssueLabelsExisting: []string{twsLabel},
		},
		{
			name: "pr_unlabeled, review unapproved, add removed label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionUnlabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
			Reviews: []github.Review{
				{
					User:  github.User{Login: "reviewer"},
					State: github.ReviewStateChangesRequested,
				},
			},
			IssueLabelsAdded: []string{twsLabel},
		},
		{
			name: "pr_labeled, review approved, remove added label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionLabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
			Reviews: []github.Review{
				{
					User:  github.User{Login: "reviewer"},
					State: github.ReviewStateApproved,
				},
			},
			IssueLabelsExisting: []string{twsLabel},
			IssueLabelsRemoved:  []string{twsLabel},
		},
		{
			name: "pr_unlabeled, 2 reviews, one approved one changes requested, files changed, add label",
			event: github.PullRequestEvent{
				Action: github.PullRequestActionLabeled,
				PullRequest: github.PullRequest{
					Number: 101,
					Head: github.PullRequestBranch{
						SHA: SHA,
					},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"},
				},
			},
			commit: github.RepositoryCommit{
				Files: []github.CommitFile{
					{
						Filename: "README.md",
					},
				},
			},
			Reviews: []github.Review{
				{
					User:  github.User{Login: "reviewer"},
					State: github.ReviewStateApproved,
				},
				{
					User:  github.User{Login: "reviewer2"},
					State: github.ReviewStateChangesRequested,
				},
			},
			IssueLabelsAdded: []string{twsLabel},
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			l := externalplugin.NewLogger().With("test", c.name)
			defer l.Sync()
			fc := fakegithub.NewFakeClient()
			a := fakeAliases{
				Aliases: repoowners.RepoAliases{
					"technical-writers": {
						"reviewer":  {},
						"reviewer2": {},
					}},
			}
			p := PluginBackend{
				ghc: fc,
				oac: a,
				gcf: fakeGitClientFactory{},
			}
			fc.Commits[SHA] = c.commit
			fc.IssueLabelsExisting = c.IssueLabelsExisting
			fc.Reviews[c.event.PullRequest.Number] = c.Reviews
			err := p.handlePullRequest(l, c.event)
			if err != nil {
				t.Errorf("handlePullRequest() returned error: %v", err)
			}
			if got, want := len(fc.IssueLabelsAdded), len(c.IssueLabelsAdded); got != want {
				t.Errorf("case %s, IssueLabelsAdded mismatch - got %d, want %d.", c.name, got, want)
			}
			if got, want := len(fc.IssueLabelsRemoved), len(c.IssueLabelsRemoved); got != want {
				t.Errorf("case %s, IssueLabelsRemoved mismatch - got %d, want %d.", c.name, got, want)
			}
		})
	}
}

func Test_HandlePullRequestReview(t *testing.T) {
	testcases := []struct {
		name      string
		event     github.ReviewEvent
		assignees []string
		labels    []string
	}{
		{
			name: "not a submitted review",
			event: github.ReviewEvent{
				Action: github.ReviewActionDismissed,
			},
		},
		{
			name:      "pr review approved and assigned, remove label",
			assignees: []string{"org/repo#101:reviewer"},
			labels:    []string{"org/repo#101:do-not-merge/missing-docs-review"},
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name: "pr review made by author",
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "pr-author"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name: "pr approve not made by required reviewer",
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateApproved,
					User:  github.User{Login: "bad-reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
		{
			name:      "pr changes requested by a reviewer, assign a reviewer",
			assignees: []string{"org/repo#101:reviewer"},
			event: github.ReviewEvent{
				Action: github.ReviewActionSubmitted,
				Review: github.Review{
					State: github.ReviewStateChangesRequested,
					User:  github.User{Login: "reviewer"},
				},
				Repo: github.Repo{
					Name:  "repo",
					Owner: github.User{Login: "org"}},
				PullRequest: github.PullRequest{
					Number:    101,
					User:      github.User{Login: "pr-author"},
					Assignees: []github.User{},
				},
			},
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			fc := fakegithub.NewFakeClient()
			l := externalplugin.NewLogger().With("test", c.name)
			defer l.Sync()
			a := fakeAliases{
				Aliases: repoowners.RepoAliases{
					"technical-writers": {
						"reviewer": {},
					},
				}}
			fc.Collaborators = []string{"reviewer"}
			fc.IssueLabelsExisting = c.labels
			p := PluginBackend{
				ghc: fc,
				oac: a,
				gcf: fakeGitClientFactory{},
			}
			err := p.handlePullRequestReview(l, c.event)
			if err != nil {
				t.Errorf("handlePullRequestReview() returned an error where it shouldn't: %v", err)
			}
			if got, want := len(fc.AssigneesAdded), len(c.assignees); got != want {
				t.Errorf("case %s, number of assignees is wrong. got %d, want %d", c.name, got, want)
			}
			if got, want := len(fc.IssueLabelsRemoved), len(c.labels); got != want {
				t.Errorf("case %s, didn't remove a label where it should have been removed. got %d want %d", c.name, got, want)
			}
		})
	}
}
