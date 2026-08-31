# Considerations

- PR #198 ("Dockerfile: pin libvips to v8.18.4 to fix broken build") passed all checks (license/cla, lint, test) against current head 43b9758, but `gh pr merge 198 --merge --delete-branch` was rejected: "StuartJAtkinson does not have the correct permissions to execute `MergePullRequest`". Needs a human with merge rights (or updated repo permissions) to complete the merge.
- PR #197 ("whatsapp: support 2-digit-year date format and locale-less placeholders") cannot be merged safely: its only check, `license/cla`, has been stuck in PENDING state since 2026-07-14 (contributor hasn't signed the CLA). Needs a human to either prompt the contributor to sign, or otherwise resolve the CLA gate.
