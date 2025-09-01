package config

import (
	"fmt"
	"path/filepath"
)

const (
	root        = "/app/iris/"
	application = "com.iris.photos"
	users       = "users"
	Metadata    = "metadata"
	Version     = "v2"
)

// GetRootDir returns the base root directory path.
func GetRootDir() string {
	return root
}

// GetPath returns a file path within the application directory.
func GetPath(file string) string {
	return filepath.Join(root, application, file)
}

// GetUserPath returns a file path specific to a user.
func GetUserPath(phone string, file string) string {
	pp := filepath.Join(root, application, users, phone, file)
	fmt.Println(pp)
	return pp
}
