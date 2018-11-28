package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
)

type FileDriver struct {
	InitialVersion semver.Version

	File       string
	FileBumped string
}

// Bump ...
func (driver *FileDriver) Bump(bump version.Bump) (semver.Version, error) {

	var newVersion semver.Version

	currentVersion, exists, err := driver.readVersion()
	if err != nil {
		return semver.Version{}, err
	}

	if !exists {
		currentVersion = driver.InitialVersion
	}

	newVersion = bump.Apply(currentVersion)

	_, err = driver.writeVersion(newVersion)
	if err != nil {
		return semver.Version{}, err
	}

	return newVersion, nil
}

// Set ...
func (driver *FileDriver) Set(newVersion semver.Version) error {

	_, err := driver.writeVersion(newVersion)
	if err != nil {
		return err
	}

	return nil
}

// Check ...
func (driver *FileDriver) Check(cursor *semver.Version) ([]semver.Version, error) {

	currentVersion, exists, err := driver.readVersion()
	if err != nil {
		return nil, err
	}

	if !exists {
		return []semver.Version{driver.InitialVersion}, nil
	}

	if cursor == nil || currentVersion.GTE(*cursor) {
		return []semver.Version{currentVersion}, nil
	}

	return []semver.Version{}, nil
}

func (driver *FileDriver) readVersion() (semver.Version, bool, error) {
	var currentVersionStr string
	versionFile, err := os.Open(driver.File)
	if err != nil {
		if os.IsNotExist(err) {
			return semver.Version{}, false, nil
		}

		return semver.Version{}, false, err
	}

	defer versionFile.Close()

	_, err = fmt.Fscanf(versionFile, "%s", &currentVersionStr)
	if err != nil {
		return semver.Version{}, false, err
	}

	currentVersion, err := semver.Parse(currentVersionStr)
	if err != nil {
		return semver.Version{}, false, err
	}

	return currentVersion, true, nil
}

func (driver *FileDriver) writeVersion(newVersion semver.Version) (bool, error) {

	err := os.MkdirAll(
		filepath.Dir(driver.FileBumped),
		os.ModePerm,
	)
	if err != nil {
		return false, err
	}

	err = ioutil.WriteFile(driver.FileBumped, []byte(newVersion.String()+"\n"), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}
