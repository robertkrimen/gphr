# gphr
--
    import "github.com/robertkrimen/gphr"

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

#### type Binary

```go
type Binary struct {
	Filename string
	Name     string
	GOOS     string
	GOARCH   string
	Asset    github.ReleaseAsset
}
```


#### func  NewBinary

```go
func NewBinary(target string) *Binary
```

#### func (*Binary) Dash

```go
func (bn *Binary) Dash() string
```

#### func (*Binary) Extension

```go
func (bn *Binary) Extension() string
```

#### func (*Binary) Identifier

```go
func (bn *Binary) Identifier() string
```

#### func (*Binary) Match

```go
func (bn *Binary) Match(asset string) bool
```

#### func (*Binary) Underscore

```go
func (bn *Binary) Underscore() string
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
