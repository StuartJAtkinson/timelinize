# Considerations

- PR #198 ("Dockerfile: pin libvips to v8.18.4 to fix broken build") is CLEAN/MERGEABLE with all checks passing, but it lives on upstream `timelinize/timelinize` (github.com/timelinize/timelinize/pull/198) and `gh pr merge 198` is rejected: "GraphQL: StuartJAtkinson does not have the correct permissions to execute `MergePullRequest`" — third identical rejection. Should the merge phase stop targeting upstream PRs entirely and instead cherry-pick this one-line Dockerfile change onto `StuartJAtkinson/timelinize` main (the repo you do own), or should the project be benched?

