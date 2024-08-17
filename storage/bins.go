package storage

import (
	"fmt"
	"io"
	"net/url"
)

type pathPair struct {
	Internal string
	External string
}

// A location for storing for files
type Bin struct {
	Id           int64
	Name         string // bin name
	Path         pathPair
	OpenFiles    map[FileIdentifier]io.ReadCloser // files currently opened by this bin
	Driver       Driver
	Redirect     bool              // if bin should Redirect or Download when getting a file
	DriverParams map[string]string // Params to be passed to the storage driver
	stats        Stats
}

// Get a file from a bin
//
// If Bin.Redirect is false returns an io.ReaderCloser, else returns a url for redirection
func (b *Bin) Get(id FileIdentifier) (io.ReadCloser, string, error) {
	if b.Redirect {
		redirectURL, err := url.JoinPath(b.Path.Internal, string(id))
		if err != nil {
			b.stats.Failed++
			return nil, "", nil
		} else {
			b.stats.Redirected++
		}
		return nil, redirectURL, nil
	}

	rc, err := b.Driver.Get(b.Path.Internal, id)
	if err != nil {
		b.stats.Failed++
	} else {
		b.stats.Downloaded++
	}
	return rc, "", err
}

// TODO: implement
func (b Bin) Upload(f *UploadFile) error {
	err := b.Driver.Upload(b.Path.Internal, f)
	if err != nil {
		b.stats.Failed++
	} else {
		b.stats.Uploaded++
	}
	return err
}

// TODO: implement
func (b Bin) Delete(id FileIdentifier) error {
	err := b.Driver.Delete(b.Path.Internal, id)
	if err != nil {
		b.stats.Failed++
	} else {
		b.stats.Deleted++
	}
	return err
}

func (b Bin) FileStatus(id FileIdentifier) (FileStatus, error) {
	return b.Driver.Status(b.Path.Internal, id)
}

func (b Bin) Stats() Stats {
	return b.stats
}

func (b Bin) String() string {
	// TODO: update with more relavent fields
	return fmt.Sprintf("Bin %s [%v]:%s", b.Name, b.Driver, b.Path.Internal)
}
