package gphr

import (
	"path/filepath"
	"regexp"

	"github.com/google/go-github/github"
)

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

type Binary struct {
	Path    string              // ../../example/example_linux_386
	Name    string              // example_linux_386
	Program string              // example
	GOOS    string              // linux
	GOARCH  string              // 386
	Asset   github.ReleaseAsset //
}

func NewBinary(path string) *Binary {
	bn := &Binary{}
	name := filepath.Base(path)
	if match := MatchBinary.FindStringSubmatch(name); match != nil {
		return &Binary{
			Path:    path,
			Name:    name,
			Program: match[1],
			GOOS:    match[2],
			GOARCH:  match[3],
		}
	} else {
		return &Binary{
			Path: path,
			Name: name,
		}
	}
	return bn
}

func (bn *Binary) Underscore() string {
	filename := bn.Program + "_" + bn.GOOS + "_" + bn.GOARCH
	if extension := bn.Extension(); extension != "" {
		filename += extension
	}
	return filename
}

func (bn *Binary) Dash() string {
	filename := bn.Identifier()
	if extension := bn.Extension(); extension != "" {
		filename += extension
	}
	return filename
}

func (bn *Binary) Extension() string {
	if bn.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func (bn *Binary) Identifier() string {
	return bn.Program + "-" + bn.GOOS + "-" + bn.GOARCH
}

func (bn *Binary) Match(asset string) bool {
	if match := MatchBinary.FindStringSubmatch(asset); match != nil {
		if bn.Program == match[1] && bn.GOOS == match[2] && bn.GOARCH == match[3] {
			return true
		}
	}
	return false
}
