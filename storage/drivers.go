package storage

import (
	"io"
)

var registeredDrivers []Driver

type Driver interface {
	Get(baseUrl string, id FileIdentifier) (io.ReadCloser, error)
	Upload(baseUrl string, f *UploadFile) error
	Delete(baseUrl string, id FileIdentifier) error
	Status(baseUrl string, id FileIdentifier) (FileStatus, error)
	Stats() Stats
	String() string
}

func Drivers() []Driver {
	return registeredDrivers
}
