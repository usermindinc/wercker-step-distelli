package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var gitLink = fmt.Sprintf("%s/%s/%s/commit", os.Getenv("WERCKER_GIT_DOMAIN"), os.Getenv("WERCKER_GIT_OWNER"), os.Getenv("WERCKER_GIT_REPOSITORY"))

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

func getLastCommit() (string, error) {
	var b Body

	resp, err := http.Get(fmt.Sprintf("https://api.distelli.com/umdevs/apps/%s/releases?apiToken=%s&max_results=1&order=desc", distelliApp, distelliKey))
	if err != nil {
		return "", err
	}
	
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &b)
	if err != nil {
		return "", err
	}

	return b.Releases[0].Commit.Commit_id, err
}

func generateReleaseNotes() (error){
	format := fmt.Sprintf("%%s%%n    %%<(16,trunc)%%an %s/%%h%%n", gitLink)
	prev_id, err := getLastCommit()
	if err != nil {
		return err
	}

	commits, err := exec.Command(fmt.Sprintf("git log %s.. --no-merges --format=\"%s\"", prev_id, format)).Output()
	if err != nil {
		return err
	}

	notes, err := os.Create("release_notes")
	if err != nil {
		return err
	}

	defer notes.Close()

	notes.WriteString("This release contains the following changes:\n\n")
	notes.Sync()
	notes.Write(commits)
	notes.Sync()

	return nil
}

func test() {
if os.Getenv("CI") == ""{
		gitCommit = "HEAD"
		distelliApp = "skylab"
		distelliKey = "jly93qxpswrqzc3t7b5hphy7tfj21761ajd8a"
		gitLink = fmt.Sprintf("github.com/usermindinc/%s/commit", distelliApp)
	}
}


