# Check version, download and update

**Goal**: As a user, I can quickly check if there is a later version available
and, if there is, download the new version and update the local `pathman`.

## Part 1: Check version

A new option `--check` is added to the `pathman version` subcommand. This
instructs `pathman` to check for the latest GitHub release and print:

- the currently installed version
- the latest release
- and a message saying that they are on the latest version or a later version
  is available.
- if a later version is available, it suggests using the (new) `update`
  subcommand.

Additionally, an `--all` option is added to the `pathman version` subcommand to
list all available stable versions from GitHub releases, with the current
version clearly marked.

## Part 2: Update subcommand, pathman update

The update command leads the user through an interactive TUI dialog that:

- Verifies if a later version is available, printing both versions
- If one is available - asks if the user wants to upgrade to the new version
- If yes - asks if the user wants to create a backup of the current version
- Downloads the later version from the GitHub release with a progress bar
  showing download progress (percentage and MB transferred)
- Verifies the download integrity by computing and comparing SHA256 checksums
  against the checksums.txt file from the release (with user feedback)
- Extracts the binary from the tar.gz archive to a temporary directory
- Checks that the current executable can be unlinked (validates permissions of
  the folder, the executable, and ownership)
- If it is possible, creates an optional backup (if user chose to), then
  unlinks the old executable (temporarily changing permissions if needed)
- Copies the new release into the same location with appropriate permissions
  (0755)
- If successful, asks if the user wants to delete the downloaded archive
- If not successful, informs the user that the download succeeded but the
  replacement did not, preserving the downloaded files for debugging

**Options**:

- `--yes` fully automates the update (automatic yes to all prompts: update
  confirmation, backup creation, and cleanup)
- `--include-prereleases` allows updating to prerelease versions (not
  recommended for production use; requires explicit typing to prevent
  accidental use)

## Part 3: --yes abbreviated to -y

Programmers are used to `-y` being a shortcode for `--yes`. The `--yes` is
used in a few places in the codebase. Ensure the shortcode works.

## Part 4: --json option for the version subcommand (and --version long-option)

The `version` subcommand should already take the `--check`, `--all`, and
`--include-prereleases` qualifier options. In this part, we further extend the
subcommand to accept a `--json` qualifier option, which asks for the output to
be generated as a simple JSON object. 

Suggested format:
```json
{
    "version": "{VERSION-STRING}",
    "latest": "{VERSION-STRING}",
    "include-prereleases": "{BOOL}",
    "all-versions": [
        "{VERSION-STRING-1}", 
        "{VERSION-STRING-2}", 
        "{VERSION-STRING-3}" 
    ]
}
```

## Part 5: Add field `source` to the JSON output of `pathman version --json`

Add an additional field `source` to the output of the `version` subcommand
when the `--json` option is specified. 

```json
{
    "version": "{VERSION-STRING}",
    "source": "{GIT REPO as HTTPS URL}"
}
```

The `source` should always be included and describes the origin of the 
code i.e. the git-repo that was checked out when the build was made.

In addition, the workflow should be updated to capture the name of the
repo and bake it into the code. If it is a local build it should fall 
back to the default value (https://github.com/sfkleach/nutmeg-run)

Additional note: This is preparatory work for creating a self-updating
convention for any standalone executable.

