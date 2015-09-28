package main

import (
	"fmt"
	"testing"
)

func TestNotes(*testing.T) {
	gitCommit = "HEAD"
	distelliApp = "skylab"
	distelliKey = "jly93qxpswrqzc3t7b5hphy7tfj21761ajd8a"
	gitLink = fmt.Sprintf("github.com/usermindinc/%s/commit", distelliApp)

	generateReleaseNotes()
}