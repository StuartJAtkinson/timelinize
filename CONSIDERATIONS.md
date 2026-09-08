# Considerations

- PR #198 ("Dockerfile: pin libvips to v8.18.4 to fix broken build") is CLEAN/MERGEABLE with all checks passing, but it lives on upstream `timelinize/timelinize` (github.com/timelinize/timelinize/pull/198) and `gh pr merge 198` is rejected: "GraphQL: StuartJAtkinson does not have the correct permissions to execute `MergePullRequest`" — fourth identical rejection (retried again 2026-09-07, head unchanged at `43b9758`, still CLEAN/MERGEABLE). Should the merge phase stop targeting upstream PRs entirely and instead cherry-pick this one-line Dockerfile change onto `StuartJAtkinson/timelinize` main (the repo you do own), or should the project be benched?

- PR #197 ("whatsapp: support 2-digit-year date format and locale-less placeholders", github.com/timelinize/timelinize/pull/197) is MERGEABLE but `mergeStateStatus: UNSTABLE` — its only check, `license/cla`, is permanently pending because the contributor hasn't signed the CLA, and it is an upstream PR this account can't merge anyway. Nothing on this side can move it. Should the merge phase skip PRs whose only blocker is an unsigned CLA, or keep re-checking them each run?

