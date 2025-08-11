package config

import (
	"fmt"
	"path/filepath"
)

const rootDir = "/media/mahdi/Cloud/Happle"
const applicationDir = "com.helium.settings/payload"
const usersDir = "users"

func GetPath(file string) string {
	return filepath.Join(rootDir, applicationDir, file)
}

func GetUserPath(phone string, file string) string {
	pp := filepath.Join(rootDir, applicationDir, usersDir, phone, file)
	fmt.Println(pp)
	return pp
}
