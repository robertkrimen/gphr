package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"

	"github.com/robertkrimen/gphr"
)

// /bin/sh -c ...
// cmd /C ...
// git describe --exact-match --tags

func getTarget(target string) (host, owner, repository, program string, err error) {
	if target != "" {
		return gphr.GetTarget(target)
	} else {
		owner, repository, err = gitGetGitHubURL()
		if err != nil {
			return "", "", "", "", err
		}
		if owner == "" {
			return "", "", "", "", fmt.Errorf("getTarget: gitGetGitHuBURL: FIXME")
		}
		host = "github.com"
	}
	return
}

func client(owner, repository string) (*gphr.GitHub, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	cache := httpcache.NewMemoryCache()
	client := httpcache.NewTransport(cache).Client()

	gh := gphr.NewGitHub(owner, repository, client, token)
	return gh, nil
}

var matchBuiltPackage = regexp.MustCompile(`(?m)^#\s*\n^#\s*(.*)\s*\n^#\s*\n`)

func getToken() (string, error) {
	token := *flags.main.token

	if token == "" {
		token = os.ExpandEnv("$GPHR_TOKEN")
	}

	lg.dbg("token = %s", token)

	if token == "" {
		return "", nil
	}

	if token[0] != '!' {
		return token, nil
	}

	token = token[1:]
	cmd := exec.Command("/bin/sh", "-c", token)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

var matchBinary = gphr.MatchBinary

func main() {
	flags.main_.Parse(os.Args[1:])
	if *flags.main.dryRun {
		*flags.main.debug = true
	}

	var cl *github.Client

	err := func() error {
		switch command := flags.main_.Arg(0); command {

		case "release":
			flags.release_.Parse(flags.main_.Args()[1:])

			owner, repository := "", ""

			repository = *flags.release.repository
			if repository != "" {
				match := strings.SplitN(repository, "/", 4)
				switch len(match) {
				case 2:
					owner, repository = match[0], match[1] // alice/example
				default:
					owner, repository = match[1], match[2] // github.com/alice/example
				case 0, 1:
					return lg.error("cannot determine GitHub repository (owner/repository) from: %s", repository)
				}
			} else {
				var err error
				owner, repository, err = gitGetGitHubURL()
				if err != nil {
					return err
				}
				if owner == "" {
					return lg.error("cannot determine GitHub repository from: git config --get remote.origin.url")
				}
			}

			gh, err := client(owner, repository)
			if err != nil {
				return err
			}
			cl = gh.Client

			// Get the tag for HEAD
			tag, err := gitGetTag()
			if err != nil {
				return err
			}
			lg.dbg("tag = %s", tag)

			if tag == "" {
				commit, _ := gitGetTagCommit("HEAD")
				return lg.error("HEAD (%s) is not tagged", commit)
			}

			// Get the commit for the tag
			tagCommit, err := gitGetTagCommit(tag)
			if err != nil {
				return err
			}
			lg.dbg("tagCommit = %s", tagCommit)

			// Make sure arguments look like release assets
			// (Are in the form of *_$GOOOS_$GOARCH, etc.)
			err = nil
			var binaries []*gphr.Binary
			for _, argument := range flags.release_.Args() {
				if match := matchBinary.FindStringSubmatch(argument); match != nil {
					binaries = append(binaries, &gphr.Binary{
						Filename: argument,
						Name:     match[1],
						GOOS:     match[2],
						GOARCH:   match[3],
					})
				} else {
					lg.err("%q: not a binary?\n", argument)
					err = lg.error("trying to upload 1 or more non-binary.Assets")
				}
			}
			if err != nil {
				return err
			}
			if len(binaries) == 0 {
				return lg.error("no binaries to upload")
			}

			// Get a list of releases
			releases, err := gh.GetReleases()
			if err != nil {
				return err
			}

			// Find the release that matches the tag at HEAD
			var release *gphr.Release
			for _, tmp := range releases {
				if *tmp.TagName == tag {
					release = tmp
					break
				}
			}

			checkTag := func(tag string) error {
				commit, _, err := gh.GetCommit(tag)
				if err != nil {
					return err
				}
				if commit == "" {
					return lg.error("tag %q does not exist in the remote repository", tag)
				}
				if tagCommit != commit {
					return lg.error("tag %q (%s) does not match %q in the local repository", commit, tag, tagCommit)
				}
				return nil
			}

			if release == nil {
				// If there is no release that matches
				err := checkTag(tag)
				if err != nil {
					return err
				}

				lg.dbg("create release => %s", tag)

				if *flags.main.dryRun {
					return nil
				}

				// Create a release for tag
				// TODO This will error out if tag does not exist, right?
				release = &gphr.Release{}
				release.TagName = &tag
				release_, _, err := gh.Client.Repositories.CreateRelease(owner, repository, &release.RepositoryRelease)
				if err != nil {
					return err
				}
				release.ID = release_.ID
			} else {
				// Otherwise, we found a release, make sure the commit matches what we have for the tag
				err := checkTag(*release.TagName)
				if err != nil {
					return err
				}
			}

			// Get the assets for the release
			assets, err := gh.GetReleaseAssets(release.RepositoryRelease)
			if err != nil {
				return err
			}

			err = nil
			for _, binary := range binaries {
				for _, asset := range assets {
					if binary.Match(*asset.Name) {
						if *flags.release.force {
							binary.Asset = asset
						} else {
							lg.err("%s: an asset of the same kind already exists (%s)", binary.Filename, *asset.Name)
							err = lg.error("1 or more assets with the same name already exist")
						}
					}
				}
			}
			if err != nil {
				return err
			}

			for _, binary := range binaries {
				file, err := os.Open(binary.Filename)
				if err != nil {
					return err
				}
				defer file.Close()

				if binary.Asset.ID != nil {
					lg.dbg("delete asset => %s (%s)", *binary.Asset.Name, *binary.Asset.URL)
					if !*flags.main.dryRun {
						response, err := gh.Client.Repositories.DeleteReleaseAsset(owner, repository, *binary.Asset.ID)
						if err != nil {
							if response == nil || response.StatusCode != 404 {
								return err
							}
						}
					}
				}

				tmp, _ := file.Stat()
				size := tmp.Size()

				lg.dbg("upload asset => %s (%d)", binary.Filename, size)

				if *flags.main.dryRun {
					continue
				}

				log("Uploading %s (%d)", binary.Filename, size)

				asset, _, err := gh.Client.Repositories.UploadReleaseAsset(owner, repository, *release.ID, &github.UploadOptions{Name: binary.Filename}, file)
				if err != nil {
					return err
				}
				binary.Asset = *asset
			}

			if !*flags.release.keep {
				// Perhaps a small race condition when deleting
				releases, err = gh.GetReleases()
				if err != nil {
					return err
				}

				err = nil
				for _, release := range releases {
					for _, asset := range release.Assets {
						for _, binary := range binaries {
							if *binary.Asset.ID == *asset.ID {
								break
							} else if binary.Match(*asset.Name) {
								lg.dbg("delete asset => %s (%s)", *binary.Asset.Name, *binary.Asset.URL)
								response, err := gh.Client.Repositories.DeleteReleaseAsset(owner, repository, *asset.ID)
								if err != nil {
									if response == nil || response.StatusCode != 404 {
										lg.err("unable to delete (legacy) asset")
										err = lg.error("1 or more (legacy) assets were not deleted")
									}
								}
							}
						}

					}
				}
			}

		case "get":
			flags.get_.Parse(flags.main_.Args()[1:])

			targetOrProgram := flags.get_.Arg(0)
			if targetOrProgram == "" {
				return lg.error("get: FIXME")
			}

			target := targetOrProgram
			if targetOrProgram != "" && strings.Index(targetOrProgram, "/") == -1 {
				target = "" // targetOrProgram is a program
			} else {
				targetOrProgram = "" // targetOrProgram is a target
			}

			_, owner, repository, program, err := getTarget(target)
			if err != nil {
				return err
			}

			if targetOrProgram != "" {
				program = targetOrProgram
			}

			binary := gphr.NewBinary(program)
			if binary.GOOS == "" {
				binary.GOOS = runtime.GOOS
				binary.GOARCH = runtime.GOARCH
			}

			base := "https://github.com/" + owner + "/" + repository

			response, err := http.Get(base + "/releases/latest")
			if err != nil {
				return err
			}
			if response.StatusCode != 200 {
				return lg.error("!200")
			}

			try := func(from, name, to string, asset bool) (bool, error) {
				if name == "" {
					name = to
				}
				request, err := http.NewRequest("GET", from, nil)
				if err != nil {
					return false, err
				}
				if asset {
					request.Header.Add("Accept", "application/octet-stream")
				}
				response, err := new(http.Client).Do(request)
				if response.StatusCode != 200 {
					return false, nil
				}
				if err != nil {
					return false, err
				}

				if *flags.main.dryRun {
					lg.dbg("download asset => %s => %s", name, to)
					return true, nil
				}

				defer response.Body.Close()

				log("Downloading %s => %s (%d)", name, to, response.ContentLength)
				file, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					return false, err
				}
				defer file.Close()

				_, err = io.Copy(file, response.Body)
				if err != nil {
					return false, err
				}

				return true, nil
			}

			if match := regexp.MustCompile(`/[^/]+/[^/]+/releases/[^/]+/([^/]+)$`).FindStringSubmatch(response.Request.URL.Path); match != nil {
				name := match[1]
				base := base + "/releases/download/" + name + "/"

				if binary.Filename != "" {
					done, err := try(base+binary.Filename, "", binary.Filename, false)
					if err != nil {
						return err
					}
					if done {
						return nil
					}
				}

				done, err := try(base+binary.Underscore(), "", binary.Underscore(), false)
				if err != nil {
					return err
				}
				if done {
					return nil
				}

				done, err = try(base+binary.Dash(), "", binary.Dash(), false)
				if err != nil {
					return err
				}
				if done {
					return nil
				}
			}

			gh, err := client(owner, repository)
			if err != nil {
				return err
			}
			cl = gh.Client

			releases, err := gh.GetReleases()
			if err != nil {
				return nil
			}

			for _, release := range releases {
				for _, asset := range release.Assets {
					if binary.Match(*asset.Name) {
						filename := *asset.Name
						if !*flags.get.preserve {
							if binary.GOOS == runtime.GOOS && binary.GOARCH == runtime.GOARCH {
								filename = binary.Name
							}
						}

						_, err := try(*asset.URL, *asset.Name, filename, true)
						if err != nil {
							return err
						}

						return nil
					}
				}
			}

			log("Nothing found for %s in %s", binary.Identifier(), gh.Location())

		case "list":
			_, owner, repository, _, err := getTarget(flags.main_.Arg(1))
			if err != nil {
				return err
			}

			gh, err := client(owner, repository)
			if err != nil {
				return err
			}
			cl = gh.Client

			releases, err := gh.GetReleases()
			if err != nil {
				return err
			}

			if len(releases) == 0 {
				log("There are no releases for %s", gh.Location())
				return nil
			}

			found := false
			for _, release := range releases {
				for _, asset := range release.Assets {
					if matchBinary.MatchString(*asset.Name) {
						found = true
						log("%v %v", *asset.Name, *release.TagName)
					}
				}
			}

			if !found {
				log("There are no assets (or no gphr assets) for %s", gh.Location())
				return nil
			}

		case "test":

			owner, repository, err := gitGetGitHubURL()
			log("owner=%s repository=%s err=%v\n", owner, repository)

			tag, err := gitGetTag()
			log("tag=%s err=%v\n", tag, err)

			commit, err := gitGetTagCommit(tag)
			log("commit=%s err=%v\n", commit, err)

			gh, err := client(owner, repository)
			if err != nil {
				return err
			}
			cl = gh.Client

			if tag != "" {
				commit, _, err = gh.GetCommit(tag)
				log("commit=%s err=%v\n", commit, err)
			}

		case "":
			usage()
			return nil

		default:
			return lg.error("invalid command: %s", command)
		}

		return nil
	}()

	if cl != nil {
		if !cl.Rate.Reset.Time.IsZero() {
			lg.dbg("client.Rate.Remaining = %d (%.1f minutes)", cl.Rate.Remaining, cl.Rate.Reset.Time.Sub(time.Now()).Minutes())
		}
	}

	if err != nil {
		lg.err("%s", err.Error())
		os.Exit(1)
	}
}
