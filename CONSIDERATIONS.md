# Considerations

- The Go module path is still `github.com/timelinize/timelinize` and every import in the repo uses it. "Never reference the original again" is now true of remotes, `gh` and the merge phase, but renaming the module would rewrite every import in the codebase. Rename it to `github.com/StuartJAtkinson/timelinize`, or leave the module path as the historical name?
