# gphr
--
    import "github.com/robertkrimen/gphr/gphr"

gphr uploads Go programs (as binaries) to GitHub Releases.

    go get github.com/robertkrimen/gphr

https://github.com/robertkrimen/gphr

## Usage

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
	Path    string              // ../../example/example_linux_386
	Name    string              // example_linux_386
	Program string              // example
	GOOS    string              // linux
	GOARCH  string              // 386
	Asset   github.ReleaseAsset //
}
```


#### func  NewBinary

```go
func NewBinary(path string) *Binary
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

#### func (*GitHub) GetAssetURL

```go
func (gh *GitHub) GetAssetURL(program, platform string) (string, error)
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

#### func (*GitHub) UploadReleaseAsset

```go
func (gh *GitHub) UploadReleaseAsset(owner, repository string, release int, name string, file *os.File) (*github.ReleaseAsset, *github.Response, error)
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
