package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// https://github.com/example/example.git
// git@github.com:example/example.git

var matchGitHubURL = regexp.MustCompile(`^(?:git@github.com(?:-[^:]+)?:|https?://github.com/)([^/]+)/(.*)(\.git)$`)

// git config --get remote.origin.url

func gitGetGitHubURL() (string, string, error) {

	try := func(remote string) (string, error) {
		cmd := exec.Command("git", "config", "--get", fmt.Sprintf("remote.%s.url", remote))
		output, err := cmd.CombinedOutput()
		lg.dbg("git config --get remote.%s.url:\n%s---", remote, string(output))
		if err != nil {
			if len(output) == 0 {
				// Probably no remote
				// FIXME Add check to see if we're in a git repository
				return "", nil
			}
			return "", fmt.Errorf("git: %v: %s", err, firstLine(output))
		}
		return string(bytes.TrimSpace(output)), nil
	}

	for _, remote := range []string{"origin", "github"} {
		remote, err := try(remote)
		if err != nil {
			return "", "", err
		}
		if match := matchGitHubURL.FindStringSubmatch(remote); match != nil {
			return match[1], match[2], nil
		}
	}
	return "", "", nil
}

// git describe --tags --exact-match

func gitGetTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	output, err := cmd.CombinedOutput()
	lg.dbg("git describe --tags --exact-match:\n%s---", string(output))
	if err != nil {
		if bytes.HasPrefix(output, []byte("fatal: no tag exactly matches")) {
			output = nil
		} else if bytes.HasPrefix(output, []byte("fatal: No names found, cannot describe anything")) {
			output = nil
		} else {
			return "", fmt.Errorf("git: %v: %s", err, firstLine(output))
		}
	}
	tags := strings.Fields(string(output))
	if len(tags) == 0 {
		return "", nil
	}
	return tags[0], nil
}

func firstLine(lines []byte) string {
	if len(lines) == 0 {
		return ""
	}
	scanner := bufio.NewScanner(bytes.NewReader(lines))
	scanner.Scan()
	return scanner.Text()
}

// git rev-list <tag>

func gitGetTagCommit(tag string) (string, error) {
	cmd := exec.Command("git", "rev-list", tag)
	output, err := cmd.CombinedOutput()
	lg.dbg("git rev-list:\n%s---", string(output))
	if err != nil {
		return "", fmt.Errorf("git: %v: %s", err, firstLine(output))
	}
	return firstLine(output), nil
}

//func gitGetProgramName() (string, error) {
//    return "gphr", nil
//    cmd := exec.Command("go", "build", "-n")
//    output, err := cmd.Output()
//    if err != nil {
//        return "", err
//    }
//    if match := matchBuiltPackage.FindAllSubmatch(output, -1); match != nil {
//        pkg := string(match[len(match)-1][1])
//        return filepath.Base(pkg), nil
//    }
//    return "", nil // TODO err?
//}
