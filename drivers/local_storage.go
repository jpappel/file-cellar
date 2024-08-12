package drivers

import (
	"file-cellar/shared"
	"io"
	"log"
	"os"
	"path/filepath"
)

type LocalDriver struct {
	root string
}

// Create a driver for storage and retrieval of files at a root directory
//
// if the directory does not exist, it is created
func CreateLocal(root string) *LocalDriver {
	driverRoot, err := filepath.Abs(root)
	if err != nil {
		log.Fatal("Failed to get root directory\n", err)
	}

	driver := LocalDriver{driverRoot}
	err = os.MkdirAll(driver.root, 0755)
	if err != nil {
		log.Fatal(err)
	}

	return &driver
}

func (d *LocalDriver) FileGet(id shared.FileIdentifier) (io.ReadCloser, error) {
	path := filepath.Join(d.root, string(id))
	return os.Open(path)
}

func (d *LocalDriver) FileUpload(f *shared.File) (string, error) {
	os.Create(d.root)
	return "", nil
}

func (d *LocalDriver) FileDelete(id shared.FileIdentifier) error {
	path := filepath.Join(d.root, string(id))
	return os.Remove(path)
}

func (d *LocalDriver) FileStatus(id shared.FileIdentifier) (shared.FileStatus, error) {
	return shared.FileOk, nil
}

func (d *LocalDriver) String() string {
	return d.root
}
