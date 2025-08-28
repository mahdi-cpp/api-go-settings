package config

import (
	"fmt"
	"path/filepath"
)

const rootDir = "/app/iris/"
const applicationDir = "com.iris.settings"
const usersDir = "users"

const versionCode = 2

// GetRootDir returns the base root directory path.
func GetRootDir() string {
	return rootDir
}

// GetPath returns a file path within the application directory.
func GetPath(file string) string {
	return filepath.Join(rootDir, applicationDir, file)
}

// GetUserPath returns a file path specific to a user.
func GetUserPath(phone string, file string) string {
	pp := filepath.Join(rootDir, applicationDir, usersDir, phone, file)
	fmt.Println(pp)
	return pp
}
