# gphr
--
    import "github.com/robertkrimen/gphr"

gphr uploads Go programs (as binaries) to GitHub Releases.

https://github.com/blog/1547-release-your-software

### Install

    go get github.com/robertkrimen/gphr/gphr

### Usage

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

## Usage

### Index

[func GetTarget](#)

[func MatchTarget](#)

```go
var MatchBinary = regexp.MustCompile(`^(.*)[_-](darwin|dragonfly|freebsd|linux|netbsd|openbsd|plan9|windows)[_-](386|amd64|arm)(?:\.exe)?$`)
```

#### func  GetTarget

```go
func GetTarget(target string) (host, owner, repository, program string, err error)
```

#### func  MatchTarget

```go
func MatchTarget(target string) (host, owner, repository, program string)
```

#### type GitHub

```go
type GitHub struct {
	Owner      string
	Repository string
	Client     *github.Client
}
```


#### func  NewGitHub

```go
func NewGitHub(owner, repository string, client *http.Client, token string) *GitHub
```

#### func (*GitHub) GetCommit

```go
func (gh *GitHub) GetCommit(digest string) (string, *github.RepositoryCommit, error)
```

#### func (*GitHub) GetReleaseAssets

```go
func (gh *GitHub) GetReleaseAssets(release github.RepositoryRelease) ([]github.ReleaseAsset, error)
```

#### func (*GitHub) GetReleases

```go
func (gh *GitHub) GetReleases() ([]*Release, error)
```

#### func (*GitHub) Location

```go
func (gh *GitHub) Location() string
```

#### func (*GitHub) TagExists

```go
func (gh *GitHub) TagExists(tag string) (bool, error)
```

#### type Release

```go
type Release struct {
	github.RepositoryRelease
	Assets []github.ReleaseAsset
}
```

--
**godocdown** http://github.com/robertkrimen/godocdown
