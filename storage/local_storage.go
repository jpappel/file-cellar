package storage

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type LocalDriver struct {
	knownRoots map[string]bool
	stats      Stats
	BaseDriver
}

func NewLocalDriver() *LocalDriver {
	d := new(LocalDriver)
	d.name = "LocalDriver"
	d.id = -1
	d.knownRoots = make(map[string]bool)
	return d
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

func (d *LocalDriver) Get(baseUrl string, id FileIdentifier) (io.ReadCloser, error) {
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

func (d *LocalDriver) Upload(baseUrl string, f *UploadFile) error {
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

func (d *LocalDriver) Delete(baseUrl string, id FileIdentifier) error {
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

func (d *LocalDriver) Status(baseUrl string, id FileIdentifier) (FileStatus, error) {
	ok, err := d.rootKnown(baseUrl)
	if !ok {
		return FileUnknownError, err
	}

	info, err := os.Stat(filepath.Join(baseUrl, string(id)))
	if err != nil {
		log.Printf("Driver: %s: failed to get status: %v\n", id, err)
		return FileUnreadable, err
	}

	mode := info.Mode()
	if !mode.IsRegular() {
		return FileUnreadable, errors.New("incorrect file mode, expected regular file")
	}
	if mode&0444 == 0 {
		return FileUnknownError, errors.New("incorrect file permissions, expected read access")
	}

	return FileOk, nil
}

func (d *LocalDriver) Stats() Stats {
	return d.stats
}

func (d *LocalDriver) String() string {
	return fmt.Sprintf("%s:%v", d.name, d.knownRoots)
}

func init() {
	registeredDrivers = append(registeredDrivers, &LocalDriver{})
}
