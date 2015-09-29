package main

import (
	"fmt"
	"log"
	"testing"
)

func TestNotes(t *testing.T) {
	gitCommit = "HEAD"
	distelliApp = "skylab"
	distelliKey = "jly93qxpswrqzc3t7b5hphy7tfj21761ajd8a"
	gitLink = fmt.Sprintf("github.com/usermindinc/%s/commit", distelliApp)

	err := generateReleaseNotes()
	if err != nil{
		log.Print(err)
		t.Fail()
	}
}