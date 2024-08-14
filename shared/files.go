package shared

import (
	"fmt"
	"io"
	"strings"
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
	RelPath         string    // the path of a file relative to its bin's base url
	UploadTimestamp time.Time // date-time of file upload
	BinId           int
}

type UploadFile struct {
	Resource *io.ReadCloser // an object which allows reading of a file resource
	Size     int64          // size of the file in bytes
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

func (f File) String() string {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("Name: %s\n", f.Name))
	builder.WriteString(fmt.Sprintf("Hash: %s\n", f.Hash))
	builder.WriteString(fmt.Sprintf("RelPath: %s\n", f.RelPath))
	builder.WriteString(fmt.Sprintf("UploadTimestamp: %v\n", f.UploadTimestamp))
	builder.WriteString(fmt.Sprintf("BinId: %d\n", f.BinId))

	return builder.String()
}
