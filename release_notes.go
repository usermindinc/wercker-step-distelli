package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"log"
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
	if distelliKey == "" {
		return "", fmt.Errorf("Distelli API Key not present")
	}

	var b Body

	resp, err := http.Get(fmt.Sprintf("https://api.distelli.com/umdevs/apps/%s/releases?apiToken=%s&max_results=1&order=desc", distelliApp, distelliKey))
	if err != nil {
		return "", err
	}
	
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s error requesting release information for distelli app %s", resp.Status, distelliApp)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &b)
	if err != nil {
		return "", err
	}

	return b.Releases[0].Commit.Commit_id, nil
}

func generateReleaseNotes() (string, error){
	format := fmt.Sprintf("%%s%%n    %%<(16,trunc)%%an %s/%%h%%n", gitLink)
	release_notes := "release_notes"
	
	prev_id, err := getLastCommit()
	if err != nil {
		return err
	}

	log.Printf("Generating release notes for %s since %s", distelliApp, prev_id)

	cmd := exec.Command("git", "log", prev_id + ".." , "--no-merges", "--format=" + format)
	commits, err := cmd.Output()
	if err != nil {
		log.Printf("Git commit %s doesn't exist in this branch", prev_id)
		return err
	}

	notes, err := os.Create(release_notes)
	if err != nil {
		return err
	}

	defer notes.Close()

	notes.WriteString("This release contains the following changes:\n\n")
	notes.Sync()
	notes.Write(commits)
	notes.Sync()

	return release_notes
}