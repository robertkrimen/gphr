/*
gphr uploads Go programs (as binaries) to GitHub Releases.

    go get github.com/robertkrimen/gphr

https://github.com/robertkrimen/gphr

*/
package gphr

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

func (gh *GitHub) UploadReleaseAsset(owner, repository string, release int, name string, file *os.File) (*github.ReleaseAsset, *github.Response, error) {
	url_, err := url.Parse(fmt.Sprintf("repos/%s/%s/releases/%d/assets", owner, repository, release))
	if err != nil {
		return nil, nil, err
	}
	query := url_.Query()
	query.Add("name", name)
	url_.RawQuery = query.Encode()

	stat, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	if stat.IsDir() {
		return nil, nil, errors.New("invalid asset: is a directory")
	}

	rq, err := gh.Client.NewUploadRequest(url_.String(), file, stat.Size(), "application/octet-stream")
	if err != nil {
		return nil, nil, err
	}

	asset := new(github.ReleaseAsset)
	rp, err := gh.Client.Do(rq, asset)
	if err != nil {
		return nil, rp, err
	}
	return asset, rp, err
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

func (gh *GitHub) GetAssetURL(program, platform string) (string, error) {
	releases, err := gh.GetReleases()
	if err != nil {
		return "", err
	}

	binary := NewBinary(program + "_" + platform)

	for _, release := range releases {
		for _, asset := range release.Assets {
			if binary.Match(*asset.Name) {
				return "https://" + gh.Location() + "/releases/download/" + *release.TagName + "/" + *asset.Name, nil
			}
		}
	}

	return "", nil
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
