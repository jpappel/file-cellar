package storage

import (
	"context"
	"io"
)

var registeredDrivers []Driver

type Driver interface {
	Get(ctx context.Context, baseUrl string, id FileIdentifier) (io.ReadSeekCloser, error)
	Upload(ctx context.Context, baseUrl string, f *File) error
	Delete(ctx context.Context, baseUrl string, id FileIdentifier) error
	Status(ctx context.Context, baseUrl string, id FileIdentifier) (FileStatus, error)
	Stats() Stats
	SetName(string)
	SetId(int64)
	Name() string
	Id() int64
	String() string
}

func ListDrivers() []Driver {
	return registeredDrivers
}
