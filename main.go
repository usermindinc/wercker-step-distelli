package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-errors/errors"
	"gopkg.in/yaml.v2"
	"fmt"
)

var releaseFilename = getenv("WERCKER_DISTELLI_RELEASEFILENAME", "usermind-release.txt")

var distelli = path.Join(os.Getenv("WERCKER_STEP_ROOT"), "DistelliCLI", "bin", "distelli")
var gitBranch = os.Getenv("WERCKER_GIT_BRANCH")
var gitCommit = os.Getenv("WERCKER_GIT_COMMIT")

var DeployConflict = errors.Errorf("Both host and environment are set")
var DeployMissing = errors.Errorf("Host or environment must be set")
var BadReleaseFile = errors.Errorf("Release file is invalid")

func getenv(key, value string) string {
	v := os.Getenv(key)
	if v == "" {
		v = value
	}
	return v
}

func checkBranches() bool {
	branches := os.Getenv("WERCKER_DISTELLI_BRANCHES")

	if branches == "" {
		return true
	}

	for _, branch := range strings.Split(branches, ",") {
		if branch == gitBranch {
			return true
		}
	}

	log.Printf("Current branch %s not in permitted set %s, skipping distelli step.", gitBranch, branches)
	return false
}

func checkManifest() (string, string, error) {
	manifest := os.Getenv("WERCKER_DISTELLI_MANIFEST")
	if manifest == "" {
		return "", "", errors.Errorf("manifest must be set")
	}

	if _, err := os.Stat(manifest); err != nil {
		return "", "", errors.Errorf("manifest file %s not found", manifest)
	}

	dirname, basename := path.Split(manifest)
	return dirname, basename, nil
}

func checkCredentials() error {
	accessKey := os.Getenv("WERCKER_DISTELLI_ACCESSKEY")
	secretKey := os.Getenv("WERCKER_DISTELLI_SECRETKEY")

	if accessKey == "" || secretKey == "" {
		return errors.Errorf("Access key and secret key are required.")
	}

	os.Setenv("DISTELLI_TOKEN", accessKey)
	os.Setenv("DISTELLI_SECRET", secretKey)

	return nil
}

func locateAppName() (string, error) {
	app := os.Getenv("WERCKER_DISTELLI_APPLICATION")

	if app == "" {
		dirname, basename, err := checkManifest()
		if err != nil {
			return "", err
		}

		file, err := os.Open(path.Join(dirname, basename))
		if err != nil {
			return "", err
		}
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return "", err
		}

		var doc map[string]interface{}
		err = yaml.Unmarshal(bytes, &doc)
		if err != nil {
			return "", err
		}

		for key := range doc {
			app = key
			break
		}
	}

	return app, nil
}

func locateReleaseID(buildURL string) (string, error) {
	var releaseID string
	app, err := locateAppName()
	if err != nil {
		return "", err
	}

	output, err := invoke("list", "releases", "-n", app, "-f", "csv")
	if err != nil {
		return "", err
	}

	reader := csv.NewReader(output)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		description := row[3]
		if strings.Contains(description, buildURL) {
			releaseID = row[1]
			break
		}
	}

	if releaseID == "" {
		return "", errors.Errorf("Unable to locate release for build %s in app %s", buildURL, app)
	}

	return releaseID, nil
}

func loadReleaseID() (string, error) {
	releaseID := os.Getenv("WERCKER_DISTELLI_RELEASE")

	if releaseID == "" {
		if stat, err := os.Stat(releaseFilename); err == nil {
			if stat.Size() <= 0 {
				return "", errors.New(BadReleaseFile)
			}

			releaseFile, err := os.Open(releaseFilename)
			if err != nil {
				return "", errors.WrapPrefix(err, fmt.Sprintf("Error opening file %s", releaseFilename), 0)
			}
			defer releaseFile.Close()

			reader := bufio.NewReader(releaseFile)
			releaseID, err = reader.ReadString('\n')
			if err != nil && err != errors.Is(err, io.EOF) {
				return "", errors.WrapPrefix(err, fmt.Sprintf("Error reading file %s", releaseFilename), 0)
			}
		}
	}

	return releaseID, nil
}

func saveReleaseID(releaseID string) error {
	releaseFile, err := os.Create(releaseFilename)
	if err != nil {
		return errors.WrapPrefix(err, "Error saving release ID", 0)
	}
	defer releaseFile.Close()

	_, err = releaseFile.WriteString(releaseID + "\n")
	return err
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		err = errors.Errorf("Error running command \"%v %v\": %v", name, strings.Join(args, " "), err)
	}
	return err
}

func invoke(args ...string) (*bytes.Buffer, error) {
	dirname, _, err := checkManifest()
	if err != nil {
		return nil, err
	}

	// Distelli 1.88 assumes manifest is in CWD
	oldCwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// If dirname is blank, don't try to CD
	if dirname != "" {
		if err = os.Chdir(dirname); err != nil {
			return nil, err
		}
	}

	// Wercker checks us out to a commit, not a branch name (sensible, since the
	// branch may have moved on). Distelli doesn't handle this well. We won't have
	// any local branches (except master), so create one with an appropriate name.

	// Checkout the commit to ensure the branch is not current
	if err = run("git", "checkout", "-q", gitCommit); err != nil {
		return nil, err
	}

	// Force update the branch name
	if err = run("git", "branch", "-f", gitBranch, gitCommit); err != nil {
		return nil, err
	}

	// Switch to the branch
	if err = run("git", "checkout", "-q", gitBranch); err != nil {
		return nil, err
	}

	var b bytes.Buffer
	cmd := exec.Command(distelli, args...)
	cmd.Stdout = &b

	if err = cmd.Run(); err != nil {
		return nil, errors.Errorf("Error executing distelli %s\n%\n%s", strings.Join(args, " "), err.Error(), b.String())
	}

	if err = os.Chdir(oldCwd); err != nil {
		return nil, err
	}

	return &b, nil
}

func push(buildURL string) error {
	_, basename, err := checkManifest()
	if err != nil {
		return err
	}

	_, err = invoke("push", "-f", basename, "-m", buildURL)
	if err != nil {
		return err
	}

	releaseID, err := locateReleaseID(buildURL)
	if err != nil {
		return err
	}

	return saveReleaseID(releaseID)
}

func deploy(description string) error {
	args := []string{"deploy"}

	environment := os.Getenv("WERCKER_DISTELLI_ENVIRONMENT")
	host := os.Getenv("WERCKER_DISTELLI_HOST")

	if environment != "" {
		if host != "" {
			return errors.New(DeployConflict)
		}
		args = append(args, "-e", environment)
	} else if host != "" {
		args = append(args, "-h", host)
	} else {
		return errors.New(DeployMissing)
	}

	_, basename, err := checkManifest()
	if err != nil {
		return err
	}

	args = append(args, "-y", "-f", basename, "-m", description)

	releaseID, err := loadReleaseID()
	if err != nil {
		return err
	}
	if releaseID != "" {
		args = append(args, "-r", releaseID)
	}

	wait := strings.ToLower(os.Getenv("WERCKER_DISTELLI_WAIT")) != "false"
	if !wait {
		args = append(args, "-q")
	}

	// A lovely piece of excrement to satisfy the type system.
	stupidity := make([]interface{}, len(args))
	for i, arg := range args {
		stupidity[i] = arg
	}
	log.Println(stupidity...)

	buffer, err := invoke(args...)
	if err != nil {
		return errors.WrapPrefix(err, "Error invoking distelli", 0)
	}
	output := buffer.String()

	if strings.Contains(output, "Deployment Failed") {
		return errors.Errorf(output)
	}

	log.Printf(output)
	return nil
}

func printAndExit(err error) {
	if err != nil {
		if err, ok := err.(*errors.Error); ok {
			log.Fatalln(err.ErrorStack())
		} else {
			panic(err)
		}
	}
}

func main() {
	log.SetFlags(0)

	err := run(distelli, "version")
	if err != nil {
		printAndExit(err)
	}

	if !checkBranches() {
		return
	}
	err = checkCredentials()
	if err != nil {
		printAndExit(err)
	}

	command := os.Getenv("WERCKER_DISTELLI_COMMAND")
	buildURL := os.Getenv("WERCKER_BUILD_URL")
	deployURL := os.Getenv("WERCKER_DEPLOY_URL")

	switch command {
	case "":
		log.Fatalln("command must be set")
	case "push":
		err = push(buildURL)
	case "deploy":
		err = deploy(deployURL)
	default:
		log.Fatalf("unknown command: %s\n", command)
	}

	if err != nil {
		printAndExit(err)
	}
}
