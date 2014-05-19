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

    go get github.com/robertkrimen/gphr


### Download

If you're on system without a Go environment, you can download a gphr executable
(Mac OS, Linux, Windows) with curl:

    curl -OJL https://gphr-io.appspot.com/

### Usage

    gphr [-token=""] [-debug=false] [-dry-run=false] <command> ...

        -token=""
            The token to use when accessing GitHub:
            https://github.com/blog/1509-personal-api-tokens

            If <token> starts with a "!", then this is a command that outputs the
            token instead. You can also specify a token via the GPHR_TOKEN
            environment variable.

         -debug=false
            Print out debugging information.

         -dry-run=false
            Do not actually modify the remote repository, just show what would be done instead.

    gphr release [-repository=""] [-force=false] [-keep=false] <assets>

        -repository=""
            The repository (e.g. github.com/alice/example).

        -force=false
            Overwrite assets if they already exist.

        -keep=false
            Do NOT delete assets of same kind in other releases.

        Create a release (if none already exists) and upload one or more assets to it.
        If no <repository> is given, default to the current GitHub remote (gphr
        will look at the "origin" remote first, then at the "github" remote, if nothing
        GitHub-ish is found).

        This command should be run from within a git repository.

        gphr will make sure that the tag/commit pair at <repository> matches the local
        tag/commit pair.

            gphr release example_linux_386 example_darwin_386 example_windows_386.exe

            gphr release --force example_linux_amd64


### Workflow

The workflow for a release:

    1. Determine the GitHub owner/repository from the local repository (if not explicity given).

        git config --get remote.origin.url
        git config --get remote.github.url # If "origin" is unsuccessful

    2. Determine the tag for HEAD in the local repository. This is the target tag.

        git describe --tags --exact-match

    3. Determine the commit for the target tag.

        git rev-list <tag>

    4. Find the release that matches the target tag. If found, then make sure the tag commit
    is the same in both the local and remote repositories. This is the target release.

    5. If no release was found, then create a release for the target tag. Again, make sure the
    tag commit is the same in both the local and remote repositories. This is the target release.

    6. Upload assets to the target release.

    7. Delete matching assets from other releases.

--
**godocdown** http://github.com/robertkrimen/godocdown
