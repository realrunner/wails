package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

// findGoFilePath checks the current working directory and any parents for `filename`
func findGoFilePath(filename string) (string, error) {
	var err error
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	abswd, err := filepath.Abs(wd)
	if err != nil {
		return "", err
	}
	var paths = strings.Split(abswd, string(os.PathSeparator))
	for {
		path, err := filepath.Abs(string(os.PathSeparator) + filepath.Join(append(paths, []string{filename}...)...))
		if err != nil {
			return "", fmt.Errorf("unable to load go.mod at %s. %v", path, err)
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if len(paths) == 0 {
				return "", fmt.Errorf("unable to find go.mod")
			}
			paths = paths[:len(paths)-1]
			continue
		}
		return path, nil
	}
}

func GetWailsVersion() (*semver.Version, error) {
	var FS = NewFSHelper()
	var result *semver.Version

	// Load file
	var err error
	goModFile, err := findGoFilePath("go.mod")
	if err != nil {
		return nil, err
	}
	goMod, err := FS.LoadAsString(goModFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to load go.mod")
	}

	// Find wails version
	versionRegexp := regexp.MustCompile(`.*github.com/wailsapp/wails.*(v\d+.\d+.\d+(?:-pre\d+)?)`)
	versions := versionRegexp.FindStringSubmatch(goMod)

	if len(versions) != 2 {
		return nil, fmt.Errorf("Unable to determine Wails version")
	}

	version := versions[1]
	result, err = semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse Wails version: %s", version)
	}
	return result, nil

}

func GetCurrentVersion() (*semver.Version, error) {
	result, err := semver.NewVersion(Version)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse Wails version: %s", Version)
	}
	return result, nil
}

func GoModOutOfSync() (bool, error) {
	gomodversion, err := GetWailsVersion()
	if err != nil {
		return true, err
	}
	currentVersion, err := GetCurrentVersion()
	if err != nil {
		return true, err
	}
	result := !currentVersion.Equal(gomodversion)
	return result, nil
}

func UpdateGoModVersion() error {
	currentVersion, err := GetCurrentVersion()
	if err != nil {
		return err
	}
	currentVersionString := currentVersion.String()

	requireLine := "-require=github.com/wailsapp/wails@v" + currentVersionString

	// Issue: go mod edit -require=github.com/wailsapp/wails@1.0.2-pre5
	helper := NewProgramHelper()
	command := []string{"go", "mod", "edit", requireLine}
	return helper.RunCommandArray(command)

}
