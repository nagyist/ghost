# Deployment

Releases are automatic on tag push. When you create a new GitHub Release with a semver tag (e.g. `v0.1.0`), the `Release` workflow runs [GoReleaser](https://goreleaser.com/) to build and publish binaries, Docker images, and Linux packages to GitHub Releases, S3, GHCR, and PackageCloud.

## Creating a release

To create a release, go to the repo's [Releases](../../releases) page and click **Draft a new release**, or use the GitHub CLI:

```bash
gh release create v<VERSION> --generate-notes --latest
```

The tag push triggers the `Release` workflow. No manual deploy step is needed.
