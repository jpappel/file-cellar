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
	BaseDriverI
}

type BaseDriverI interface {
	SetName(string)
	SetId(int64)
	Name() string
	Id() int64
}

type BaseDriver struct {
	name string
	id   int64
}

func (d *BaseDriver) Name() string {
	return d.name
}

func (d *BaseDriver) SetName(name string) {
	d.name = name
}

func (d *BaseDriver) Id() int64 {
	return d.id
}

func (d *BaseDriver) SetId(id int64) {
	d.id = id
}

func Drivers() []Driver {
	return registeredDrivers
}
