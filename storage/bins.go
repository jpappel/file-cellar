package storage

import (
	"file-cellar/shared"
	"fmt"
	"io"
)

// A location for storing for files
type Bin struct {
	Name         string                                  // bin name
	Url          string                                  // the base url for files stored in this bin
	OpenFiles    map[shared.FileIdentifier]io.ReadCloser // files currently opened by this bin
	Driver       Driver
	DriverParams map[string]string // Params to be passed to the storage driver
	stats        Stats
}

func (b *Bin) Get(id shared.FileIdentifier) (io.ReadCloser, error) {
	return b.Driver.Get(id, b.Url)
}

// TODO: implement
func (b Bin) Upload(f *shared.UploadFile) error {
	return nil
}

// TODO: implement
func (b Bin) Delete(id shared.FileIdentifier) error {
	return nil
}

// TODO: implement
func (b Bin) FileStatus(id shared.FileIdentifier) (shared.FileStatus, error) {
	return shared.FileOk, nil
}

func (b Bin) Stats() Stats {
	return b.stats
}

func (b Bin) String() string {
	return fmt.Sprintf("Bin %s [%v]:%s", b.Name, b.Url, b.Driver)
}
