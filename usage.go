/*
gphr uploads Go programs (as binaries) to GitHub Releases.

https://github.com/blog/1547-release-your-software

Before you can upload, you'll have to create a GitHub token: https://github.com/blog/1509-personal-api-tokens

A binary is of the form `<name>_$GOOS_$GOARCH` (with an optional `.exe` at the end for Windows)

If you need help cross-compiling, try gnat: https://github.com/robertkrimen/gnat

Install

    go get github.com/robertkrimen/gphr

Download

If you're on system without a Go environment, you can download a gphr executable (Linux, Mac OS, Windows) with cURL:

    curl -JLO https://gphr-io.appspot.com/

Usage

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

    gphr release [-repository=""] [-force=false] [-keep=true] <assets>

        -repository=""
            The repository (e.g. github.com/alice/example).

        -force=false
            Overwrite assets if they already exist.

        -keep=true
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

Workflow

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

    7. Delete matching assets from other, older releases (if any).

*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type _flags struct {
	main_ *flag.FlagSet
	main  _mainFlags

	release_ *flag.FlagSet
	release  _releaseFlags

	get_ *flag.FlagSet
	get  _getFlags
}

type _mainFlags struct {
	debug  *bool
	dryRun *bool
	token  *string
}

type _releaseFlags struct {
	repository *string
	force      *bool
	keep       *bool
}

type _getFlags struct {
	preserve *bool
}

var flags = func() (flags *_flags) {
	flags = &_flags{
		main_:    flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		release_: flag.NewFlagSet(os.Args[0]+" release", flag.ExitOnError),
		get_:     flag.NewFlagSet(os.Args[0]+" get", flag.ExitOnError),
	}

	var flag *flag.FlagSet

	flag = flags.main_
	flag.Usage = usage
	flags.main.debug = flag.Bool("debug", false, "")
	flags.main.dryRun = flag.Bool("dry-run", false, "")
	flags.main.token = flag.String("token", "", "")

	flag = flags.release_
	flag.Usage = usage
	flags.release.repository = flag.String("repository", "", "")
	flags.release.force = flag.Bool("force", false, "")
	flags.release.keep = flag.Bool("keep", true, "")

	flag = flags.get_
	flag.Usage = usage
	flags.get.preserve = flag.Bool("preserve", false, "")

	return
}()

func usage() {
	fmt.Fprintf(os.Stderr, strings.TrimSpace(`
Usage of %s:

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

    gphr release [-repository=""] [-force=false] [-keep=true] <assets>

        -repository=""
            The repository (e.g. github.com/alice/example).

        -force=false
            Overwrite assets if they already exist.

        -keep=true
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

    `), os.Args[0])
	fmt.Fprintln(os.Stderr, "\n")
}

/*

      -dry-run=false:   Do not actually modify the remote repository or
                        download anything. Show what would be done instead.

   get <repository>
   get <repository>/<target>
   get <repository> <target>
     -preserve=false:  Always preserve the filename of the asset instead of stripping
                       the $GOOS/$GOARCH suffix.

     Download the binary/asset from <repository>.
     If no <target> is given, then default to the same name as the repository.
     By default, get will look for the binary corresponding to the current $GOOS & $GOARCH,
     but you can change this by specifying it via <target>:

       gphr get github.com/alice/example/example_linux_386

   list <repository>

     List all gphr-like assets for <repository>.

//

    1. Make sure the asset arguments actually look like assets, e.g.:
         example_darwin_386
         example_windows_386.exe
         example_linux_amd64

*/
