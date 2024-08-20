package storage

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
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

// TODO: use Type field
type FileInfo struct {
	Name            string    // name of the source
	Hash            string    // hash of the file content
	Type            string    // mimetype of the file content
	Size            int64     // size of the file in bytes
	RelPath         string    // the path of a file relative to its bin's base url
	UploadTimestamp time.Time // date-time of file upload
	Bin             *Bin      // bin storing this file
}

type File struct {
	Data io.ReadSeekCloser // an object which allows reading of a file resource
	FileInfo
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

// Get a relPath that should be associated with a file.
//
// Returns an error if name or fileHash are empty strings.
func GetRelPath(fileName string, fileHash string, uploadTime time.Time) (string, error) {

	if fileName == "" || fileHash == "" {
		return "", errors.New("missing file name or file hash")
	}

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%d", fileName, fileHash, uploadTime.Unix())))
	encoding := base64.URLEncoding.EncodeToString(hash[:])

	return strings.Trim(encoding, "="), nil
}

// Returns if two files are the identical
func (this FileInfo) Equal(other FileInfo) bool {
	return this.Name == other.Name &&
		this.Hash == other.Hash &&
		this.Size == other.Size &&
		this.RelPath == other.RelPath &&
		this.UploadTimestamp.Equal(other.UploadTimestamp) &&
		this.Bin == other.Bin
}

// Returns if two Files could be backed by the same data
func (this FileInfo) Equivalent(other FileInfo) bool {
	return this.Hash == other.Hash &&
		this.Size == other.Size &&
		this.Bin.Id == other.Bin.Id
}

func (f FileInfo) String() string {
	return fmt.Sprintf("%s:%s uploaded at %v size of %d in bin%d with hash of %s", f.Name, f.RelPath, f.UploadTimestamp, f.Size, f.Bin.Id, f.Hash)
}
