package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

if os.Getenv("CI"){
	var gitBranch = os.Getenv("WERCKER_GIT_BRANCH")
	var gitCommit = os.Getenv("WERCKER_GIT_COMMIT")
	var distelliApp = os.Getenv("DISTELLI_APP")
	var distelliKey = os.Getenv("DISTELLI_API_KEY")
	var gitLink = fmt.Sprintf("%s/%s/%s/commit", os.Getenv("WERCKER_GIT_DOMAIN"), os.Getenv("WERCKER_GIT_OWNER"), os.Getenv("WERCKER_GIT_REPOSITORY")
} else {
	var gitCommit = "HEAD"
	var distelliApp = "skylab"
	var distelliKey = "jly93qxpswrqzc3t7b5hphy7tfj21761ajd8a"
	var gitLink = fmt.Sprintf("github.com/usermindinc/%s/commit", distelliApp)
}

type Release struct {
	Release_version string
	Commit struct {
		Branch string
		Commit_id string
		Url string
	}
}

type Body struct {
	Releases []Release
}

func getLastCommit() (string, err) {
	var b Body

	resp, err := http.Get(fmt.Sprintf("https://api.distelli.com/umdevs/apps/%s/releases?apiToken=%s&max_results=1&order=desc", distelliApp, apiToken)
	if err != nil {
		return "", err
	}
	
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err := json.Unmarshal(body, &b)
	if err != nil {
		return "", err
	}

	return b.Releases[0].Commit.Commit_id
}

func generateReleaseNotes() (err) {
	format := fmt.Sprintf("%%s%%n    %%<(16,trunc)%%an %s/%%h%%n", gitLink)
	prev_id := getLastCommit()

	commits, err := exec.Command(fmt.Sprintf("git log %s.. --no-merges --format=\"%s\"", prev_id, format)).Output()
	if err != nil {
		return err
	}

	notes, err := os.Create("release_notes")
	if err != nil {
		return err
	}

	defer f.Close()

	f.WriteString("This release contains the following changes:\n\n")
	f.Sync()
	f.WriteString(notes)
	f.Sync()
}


