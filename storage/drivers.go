package storage

import (
	"file-cellar/shared"
	"io"
)

type Driver interface {
	Get(baseUrl string, id shared.FileIdentifier) (io.ReadCloser, error)
	Upload(baseUrl string, f *shared.UploadFile) error
	Delete(baseUrl string, id shared.FileIdentifier) error
	Status(baseUrl string, id shared.FileIdentifier) (shared.FileStatus, error)
	Stats() Stats
	String() string
}
