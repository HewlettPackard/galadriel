# Release Procedure

This document outlines our process for releasing and managing versions of Galadriel.

## Active Versions

The Galadriel project actively support and maintain both the current and the previous major versions. 
Development is primarily done on the `main` branch, but we also create branches for versioned releases as needed.

## Release Branches

For each release, we create a dedicated branch named `release/vX.Y.Z`, where `X` is the major version, `Y` is the minor version, and `Z` is the patch version. The origin point for a new release branch depends on the type of release:

* **Patch release for older minor release series**: The new release branch is created from the most recent patch release branch in the same minor release series. For example, if the current release is v1.5.z and we're preparing v1.4.5, we'd base the new branch on `release/v1.4.4`.
* **Security release for current minor release series**: The new release branch is created from the previous release branch in the same minor release series. For example, if the current release is v1.5.0 and we're preparing v1.5.1, we'd base the new branch on `release/v1.5.0`.
* **Scheduled patch release for current minor release series or scheduled minor release**: The new release branch is created from a selected commit on the `main` branch.

If a bug fix needs to be backported, the corresponding patch is cherry-picked or backported to a PR against the version branch, which is maintained similarly to the `main` branch. We ensure that the CHANGELOG is updated in both `main` and the version branch to reflect the new release.

## Release Procedure

Our release process is driven by tags. Once maintainers are ready for a release, they push a tag referencing the release commit. The rest of the process is handled by the CI/CD pipeline, but it's crucial to monitor the pipeline for any errors. If an error occurs, the release is aborted.

Before authorizing a release, maintainers must thoroughly review the proposed release commit to ensure compatibility, thorough testing, and safety/security. If there's any doubt or hesitation, the release should not proceed.

The decision to release at a specific commit hash must be approved by a majority vote of the maintainers. If a release could potentially endanger the project or its users, it should be put on hold until all maintainers have had the chance to review and decide on the matter.

## Release Checklist

Here is a list of steps that must be followed in the specified order when releasing:

1. Ensure all intended changes are fully merged.
2. Designate a specific commit as the release candidate.
3. Open an issue titled "Release X.Y.Z", including the release candidate commit hash.
4. Create the release branch as per the guidelines above.
5. Create a draft pull request against the release branch with updates to the CHANGELOG. 
6. Create an annotated tag named `vX.Y.Z` (where `X.Y.Z` is the software's semantic version number) against the release candidate.
7. Push the annotated tag to the repository, monitoring the build until completion.
8. On the GitHub releases page, copy the release notes, click edit, and paste them back in.
9. Open a PR targeted at the `main` branch with:
   * A cherry-pick of the changelog commit from the latest release, so the `main` branch changelog contains all release notes.
   * An update to the software version to the next projected version.
10. Close the GitHub issue created for the release process.
