# gphr
--
gphr uploads Go programs (as binaries) to GitHub Releases.

https://github.com/blog/1547-release-your-software

Before you can upload, you'll have to create a GitHub token:
https://github.com/blog/1509-personal-api-tokens

Go binaries are not terribly small, so gphr also does the work of cleaning up
after itself, deleting old binaries when a new one can take its place (you can
override this behavior with --keep). A binary is of the form
`<name>_$GOOS_$GOARCH` (with an optional `.exe` at the end for Windows)

If you need help cross-compiling, try gnat: https://github.com/robertkrimen/gnat

### Install

    go get github.com/robertkrimen/gphr/gphr

### Usage

    gphr <options> ...

          -token=""
                The token to use when authenticating to GitHub:
                https://github.com/blog/1509-personal-api-tokens
                If <token> starts with a "!", then this is a command that outputs the
                token instead. You can also specify a token via the $GPHR_TOKEN
                environment variable.
          -debug=false:     Print out debugging information.
          -dry-run=false:   Do not actually modify the remote repository,
                            just show what would be done instead.

        release <assets>
          -repository="":   The repository (e.g. github.com/alice/example)
          -force=false:     Overwrite assets if they already exist.
          -keep=false:      Do NOT delete assets of same kind in other releases.

          Create a release (if necessary) and upload one or more assets to it.
          If no <repository> is given, default to the current GitHub remote (gphr
          will look at the "origin" remote first, then at the "github" remote, if nothing
          GitHub-ish is found.
          This command should be run from within a git repository.
          gphr will make sure that the tag/commit pair at <repository> matches the local
          tag/commit pair.

            gphr release example_linux_386 example_darwin_386 example_windows_386.exe

            gphr release --force example_linux_amd64

--
**godocdown** http://github.com/robertkrimen/godocdown
