package storage

import (
	"errors"
	"file-cellar/shared"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type LocalDriver struct {
	knownRoots map[string]bool
	stats      Stats
}

func NewLocalDriver() *LocalDriver {
	return new(LocalDriver)
}

// Create a local directory and return its absolute path
func createLocal(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal("Failed to get root directory\n", err)
	}

	err = os.MkdirAll(absPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	return absPath
}

func (d *LocalDriver) rootKnown(baseUrl string) (bool, error) {
	_, ok := d.knownRoots[baseUrl]
	var err error = nil
	if !ok {
		err = errors.New("unknown base directory")
	}

	return ok, err
}

func (d *LocalDriver) addRoot(root string) {
	d.knownRoots[root] = true
}

func (d *LocalDriver) Get(baseUrl string, id shared.FileIdentifier) (io.ReadCloser, error) {
	path := filepath.Join(baseUrl, string(id))
	f, err := os.Open(path)
	if err != nil {
		d.stats.Failed++
		log.Printf("Driver: failed to open %s: %v\n", id, err)
	} else {
		d.stats.Downloaded++
	}

	return f, err
}

func (d *LocalDriver) Upload(baseUrl string, f *shared.UploadFile) error {
	ok, err := d.rootKnown(baseUrl)
	if !ok {
		d.stats.Failed++
		return err
	}

	path := filepath.Join(baseUrl, f.RelPath)

	w, err := os.Create(path)
	if err != nil {
		d.stats.Failed++
		log.Printf("Driver: Failed to create %s: %v\n", f.RelPath, err)
		return err
	}

	n, err := io.Copy(w, *f.Resource)
	if err != nil {
		d.stats.Failed++
		log.Printf("Driver: Failed to write file %s: %v\n", f.RelPath, err)
		return err
	}
	if n != f.Size {
		d.stats.Failed++
		log.Printf("Driver: %s: Incorrect number of bytes written: %d != %d\n", f.RelPath, n, f.Size)

		err := os.Remove(path)
		if err != nil {
			log.Printf("Driver: failed cleanup after incorrect write: %s", f.RelPath)
			return err
		}

		return errors.New("incorrect number of bytes written")
	}

	d.stats.Uploaded++

	return nil
}

func (d *LocalDriver) Delete(baseUrl string, id shared.FileIdentifier) error {
	ok, err := d.rootKnown(baseUrl)
	if !ok {
		d.stats.Failed++
		return err
	}

	path := filepath.Join(baseUrl, string(id))
	err = os.Remove(path)
	if err != nil {
		d.stats.Failed++
	} else {
		d.stats.Deleted++
	}
	return err
}

func (d *LocalDriver) Status(baseUrl string, id shared.FileIdentifier) (shared.FileStatus, error) {
	ok, err := d.rootKnown(baseUrl)
	if !ok {
		return shared.FileUnknownError, err
	}

	info, err := os.Stat(filepath.Join(baseUrl, string(id)))
	if err != nil {
		log.Printf("Driver: %s: failed to get status: %v\n", id, err)
		return shared.FileUnreadable, err
	}

	mode := info.Mode()
	if !mode.IsRegular() {
		return shared.FileUnreadable, errors.New("incorrect file mode, expected regular file")
	}
	if mode&0444 == 0 {
		return shared.FileUnknownError, errors.New("incorrect file permissions, expected read access")
	}

	return shared.FileOk, nil
}

func (d *LocalDriver) Stats() Stats {
	return d.stats
}

func (d *LocalDriver) String() string {
	return fmt.Sprintf("LocalDriver:%v", d.knownRoots)
}
