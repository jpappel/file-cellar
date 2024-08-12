package drivers

import (
	"file-cellar/shared"
	"io"
)

type Driver interface {
	FileGet(shared.FileIdentifier) (io.ReadCloser, error)
	FileUpload(*shared.File) (string, error)
	FileDelete(shared.FileIdentifier) error
	FileStatus(shared.FileIdentifier) (shared.FileStatus, error)
	Status() string
	String() string
}
