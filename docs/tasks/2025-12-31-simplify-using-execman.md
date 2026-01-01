# Simplification, thanks to execman

## Goal

Remove the features which are superceded by `execman` and revise the documentation to show workflows using execman to manage the update part of the lifecycle. See https://github.com/sfkleach/execman

## Background

We have implemented a sister project `execman` that manages standalone executables, including `pathman`. It allows the download, installation and upgrade of the executables by keeping track of the origin repo and examining releases. A full list of current features are:

- Install executables directly from GitHub releases
- Track installed executables with version and origin information
- List all managed executables with details
- Check for available updates across all executables
- Update executables individually or all at once
- Remove executables and delete files
- Forget executables while keeping files on disk
- Registry maintains metadata for secure updates
- Cross-platform support for Linux and macOS

Because of this, there isn't really any need for:

- `pathman init` to install the binary, that can be left to `execman`.
- `pathman check` as `execman check` does the same.
- `pathman update` as `execman update` does the same.

And the `pathman version` subcommand can be simplified.
