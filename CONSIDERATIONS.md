# Considerations

- PR #198 ("Dockerfile: pin libvips to v8.18.4 to fix broken build") passed all checks (license/cla, lint, test) against current head 43b9758, but `gh pr merge 198 --merge --delete-branch` was rejected: "StuartJAtkinson does not have the correct permissions to execute `MergePullRequest`". Needs a human with merge rights (or updated repo permissions) to complete the merge.
