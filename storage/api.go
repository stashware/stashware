package storage

import (
	"io/ioutil"
	"path"

	"github.com/ricochet2200/go-disk-usage/du"
)

func DiskFreeBytes(dir string) int64 {
	return int64(du.NewDiskUsage(dir).Free())
}

func WriteFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename, data, 0644)
}

func ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func PathJoin(dir, filename string) string {
	return path.Join(dir, filename)
}
