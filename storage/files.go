package storage

import (
	"fmt"
	"io"
	"time"
)

type User string
type FileIdentifier string

type FileRequest struct {
	requester string
	id        FileIdentifier
}

type File struct {
	Name            string    // name of the source
	Hash            string    // hash of the file content
	Size            int64     // size of the file in bytes
	RelPath         string    // the path of a file relative to its bin's base url
	UploadTimestamp time.Time // date-time of file upload
	Bin             *Bin      // bin storing this file
}

type UploadFile struct {
	Resource *io.ReadCloser // an object which allows reading of a file resource
	File
}

type FileStatus uint8

const (
	FileOk FileStatus = iota
	FileMissing
	FileUnreadable
	FileUnknownError
)

func (fs FileStatus) String() string {
	return [...]string{"File Ok", "File Missing", "File Unreadable", "File Unknown Error"}[fs]
}

// Returns if two files are the identical
func (this File) Equal(other File) bool {
	return this.Name == other.Name &&
		this.Hash == other.Hash &&
		this.Size == other.Size &&
		this.RelPath == other.RelPath &&
		this.UploadTimestamp.Equal(other.UploadTimestamp) &&
		this.Bin == other.Bin
}

// Returns if two Files could be backed by the same data
func (this File) Equivalent(other File) bool {
	return this.Hash == other.Hash &&
		this.Size == other.Size &&
		this.Bin.Id == other.Bin.Id
}

func (f File) String() string {
	return fmt.Sprintf("%s:%s uploaded at %v size of %d in bin%d with hash of %s", f.Name, f.RelPath, f.UploadTimestamp, f.Size, f.Bin.Id, f.Hash)
}
