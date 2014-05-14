/*
gphr uploads Go programs (as binaries) to GitHub Releases.

https://github.com/blog/1547-release-your-software

Install

    go get github.com/robertkrimen/gphr/gphr

Usage

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

*/
package gphr

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

func MatchTarget(target string) (host, owner, repository, program string) {
	match := strings.SplitN(target, "/", 4)
	switch len(match) {
	case 1:
		host = match[0]
	case 2:
		host, owner = match[0], match[1]
	case 3:
		host, owner, repository = match[0], match[1], match[2]
	default:
		return match[0], match[1], match[2], match[3]
	}
	return
}

func GetTarget(target string) (host, owner, repository, program string, err error) {
	if target != "" {
		host, owner, repository, program = MatchTarget(target)
		if host != "github.com" {
			return "", "", "", "", fmt.Errorf("invalid target: %s: not a github.com URL", target)
		}
		if repository == "" {
			return "", "", "", "", fmt.Errorf("invalid target: %s: missing repository", target)
		}
	}
	return
}

var MatchBinary = regexp.MustCompile(`^(.*)[_-](darwin|dragonfly|freebsd|linux|netbsd|openbsd|plan9|windows)[_-](386|amd64|arm)(?:\.exe)?$`)

// darwin/386
// dragonfly/386
// dragonfly/amd64
// freebsd/386
// freebsd/amd64
// freebsd/arm
// linux/386
// linux/amd64
// linux/arm
// netbsd/386
// netbsd/amd64
// netbsd/arm
// openbsd/386
// openbsd/amd64
// plan9/386
// plan9/amd64
// windows/386

type GitHub struct {
	Owner      string
	Repository string
	Client     *github.Client
}

func NewGitHub(owner, repository string, client *http.Client, token string) *GitHub {

	if client == nil {
		client = http.DefaultClient
	}

	if token != "" {
		client.Transport = &oauth.Transport{
			Token:     &oauth.Token{AccessToken: token},
			Transport: client.Transport,
		}
	}

	gh := &GitHub{
		Owner:      owner,
		Repository: repository,
		Client:     github.NewClient(client),
	}

	return gh
}

func (gh *GitHub) Location() string {
	return fmt.Sprintf("github.com/%s/%s", gh.Owner, gh.Repository)
}

func (gh *GitHub) TagExists(tag string) (bool, error) {
	_, response, err := gh.Client.Git.GetRef(gh.Owner, gh.Repository, "tags/"+tag)
	if response != nil {
		if response.StatusCode == 404 {
			return false, nil
		}
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (gh *GitHub) GetCommit(digest string) (string, *github.RepositoryCommit, error) {
	commit, response, err := gh.Client.Repositories.GetCommit(gh.Owner, gh.Repository, digest)
	if response != nil {
		if response.StatusCode == 404 {
			return "", nil, nil
		}
	}
	if err != nil {
		return "", nil, err
	}
	return *commit.SHA, commit, nil
}

func (gh *GitHub) GetReleases() ([]*Release, error) {
	client := gh.Client
	owner := gh.Owner
	repository := gh.Repository

	var releases []*Release
	_, err := pages(func(options *github.ListOptions) (*github.Response, error) {
		start := len(releases)
		tmp, response, err := client.Repositories.ListReleases(owner, repository, options)
		if err != nil {
			return nil, err
		}
		for _, item := range tmp {
			releases = append(releases, &Release{item, nil})
		}

		for _, release := range releases[start:] {
			_, err = pages(func(options *github.ListOptions) (*github.Response, error) {
				tmp, response, err := client.Repositories.ListReleaseAssets(owner, repository, *release.ID, options)
				if err != nil {
					return nil, err
				}
				release.Assets = append(release.Assets, tmp...)

				return response, nil
			})
			if err != nil {
				return nil, err
			}
		}

		return response, nil
	})
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(_sortReleaseByTime(releases)))

	return releases, nil
}

func (gh *GitHub) GetReleaseAssets(release github.RepositoryRelease) ([]github.ReleaseAsset, error) {
	client := gh.Client
	owner := gh.Owner
	repository := gh.Repository

	var assets []github.ReleaseAsset

	_, err := pages(func(options *github.ListOptions) (*github.Response, error) {
		tmp, response, err := client.Repositories.ListReleaseAssets(owner, repository, *release.ID, options)
		if err != nil {
			return nil, err
		}
		assets = append(assets, tmp...)

		return response, nil
	})
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func pages(inner func(*github.ListOptions) (*github.Response, error)) (*github.Response, error) {

	options := github.ListOptions{}

	for {

		response, err := inner(&options)
		if err != nil {
			return response, err
		}

		if response.NextPage == 0 {
			return response, nil
		}

		options.Page = response.NextPage
	}

	return nil, nil
}

type _sortReleaseByTime []*Release

func (a _sortReleaseByTime) Len() int      { return len(a) }
func (a _sortReleaseByTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a _sortReleaseByTime) Less(i, j int) bool {
	return a[i].CreatedAt.UnixNano() < a[j].CreatedAt.UnixNano()
}

type Release struct {
	github.RepositoryRelease
	Assets []github.ReleaseAsset
}
