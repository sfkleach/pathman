# Fix generated .profile snippet

**Goal**: Ensure that the $PATH is safely updated, protecting against errors.

## Background

The current snippet that is inserted into .profile is:

```sh
export PATH=$(pathman path)
```

However, this is incautious. If `pathman` is not on the path then this will
delete $PATH (and lock me out of my account, as I discovered!) It should also
verify that `pathman` completes successfully before modifying the $PATH.

## The fix

- Check the pathman command exists
- Verify `pathman path` returns successfully before updating $PATH
- Ensure the snippet's comment includes 
  - the time and date (in international format, year first) we edited .profile
  - the edit is attributed to `pathman init`.
- Errors should only be printed if we are in an interactive shell.