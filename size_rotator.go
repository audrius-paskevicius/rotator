package rotator

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	defaultRotationSize = 1024 * 1024 * 10
	defaultMaxRotation  = 999
)

// SizeRotator is file writer which rotates files by size
type SizeRotator struct {
	path         string     // base file path
	totalSize    int64      // current file size
	file         *os.File   // current file
	mutex        sync.Mutex // lock
	RotationSize int64      // size threshold of the rotation
	MaxRotation  int        // maximum count of the rotation
}

// RightPad2Len https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

// LeftPad2Len https://github.com/DaddyOh/golang-samples/blob/master/pad.go
func LeftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}

// Write bytes to the file. If binaries exceeds rotation threshold,
// it will automatically rotate the file.
func (r *SizeRotator) Write(bytes []byte) (n int, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.file == nil {
		// Check file existence
		stat, _ := os.Lstat(r.path)
		if stat != nil {
			// Update initial size by file size
			r.totalSize = stat.Size()
		}
	}
	// Do rotate when size exceeded
	if r.totalSize+int64(len(bytes)) > r.RotationSize {
		// Get available file name to be rotated
		for i := 1; i <= r.MaxRotation; i++ {
			fext := filepath.Ext(r.path)
			renamedPath := ""
			if len(fext) != 0 {
				renamedPath = strings.TrimRight(r.path, fext) + "_" + LeftPad2Len(strconv.Itoa(i), "0", len(strconv.Itoa(r.MaxRotation))) + fext
			} else {
				renamedPath = r.path + "_" + LeftPad2Len(strconv.Itoa(i), "0", len(strconv.Itoa(r.MaxRotation)))
			}
			stat, _ := os.Lstat(renamedPath)
			if stat == nil {
				if r.file != nil {
					// reset file reference
					r.file.Close()
					r.file = nil
				}
				err := os.Rename(r.path, renamedPath)
				if err != nil {
					return 0, err
				}
				break
			}
			if i == r.MaxRotation {
				return 0, errors.New("Rotation count has been exceeded")
			}
		}
	}
	if r.file == nil {
		r.file, err = os.OpenFile(r.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return 0, err
		}
		// Switch current date
		r.totalSize = 0
	}
	n, err = r.file.Write(bytes)
	r.totalSize += int64(n)
	return n, err
}

// WriteString writes strings to the file. If binaries exceeds rotation threshold,
// it will automatically rotate the file.
func (r *SizeRotator) WriteString(str string) (n int, err error) {
	return r.Write([]byte(str))
}

// Close the file
func (r *SizeRotator) Close() error {
	return r.file.Close()
}

// NewSizeRotator creates new writer of the file
func NewSizeRotator(path string) *SizeRotator {
	return &SizeRotator{
		path:         path,
		RotationSize: defaultRotationSize,
		MaxRotation:  defaultMaxRotation,
	}
}
