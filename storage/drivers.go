package storage

import (
	"file-cellar/shared"
	"io"
)

type Driver interface {
	Get(id shared.FileIdentifier, baseUrl string) (io.ReadCloser, error)
	Upload(f *shared.File, baseUrl string) error
	Delete(id shared.FileIdentifier, baseUrl string) error
	Status(id shared.FileIdentifier) (shared.FileStatus, error)
	Stats() Stats
	String() string
}
