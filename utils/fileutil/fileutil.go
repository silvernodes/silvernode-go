package fileutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CurrentDir() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	ret := GetDirByPath(path)
	return ret
}

func ListDir(dir string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	directory, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	suffix = strings.ToUpper(suffix)
	for _, fi := range directory {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

func WalkDir(dir, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix)
	err = filepath.Walk(dir, func(filename string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

func MakeDir(dir string) {
	os.MkdirAll(dir, os.ModePerm)
}

func GetDirByPath(path string) string {
	fix := "/"
	if strings.Contains(path, "\\") && !strings.Contains(path, "/") {
		fix = "\\"
	}
	splitstring := strings.Split(path, fix)
	size := len(splitstring)
	dir := strings.TrimRight(path, splitstring[size-1])
	ret := strings.Replace(dir, "\\", "/", size-1)
	return ret
}

func LoadFile(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func SaveFile(filePath string, content string) error {
	data := []byte(content)
	return ioutil.WriteFile(filePath, data, 0644)
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
