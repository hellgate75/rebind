// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package log

import (
	errs "errors"
	"fmt"
	"github.com/hellgate75/rebind/rerrors"
	"github.com/hellgate75/rebind/utils"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	updateInterval time.Duration = 2 * time.Second
	checkInterval  time.Duration = 2 * time.Second
)

var (
	sepString = fmt.Sprintf("%c", os.PathSeparator)
)

type RotationCallBack func()

type LogRotator interface {
	IsEnabled() bool
	Hook(msgLen int64) rerrors.Error
	GetDefaultWriter() (io.Writer, bool)
	UpdateCallBack(callback RotationCallBack)
}

type _rotator struct {
	sync.Mutex
	folder      *os.File
	fileName    string
	rotateLen   int
	maxSize     int64
	enabled     bool
	callback    RotationCallBack
	files       []*utils.FileIndex
	currentFile *utils.FileIndex
	lastUpdate  time.Time
	lastCheck   time.Time
	writer      io.Writer
}

func (r *_rotator) IsEnabled() bool {
	return r.enabled
}

func (r *_rotator) UpdateCallBack(callback RotationCallBack) {
	r.callback = callback
}

func (r *_rotator) GetDefaultWriter() (io.Writer, bool) {
	if r.writer != nil {
		return r.writer, true
	}
	return nil, false
}

func (r *_rotator) Hook(msgLen int64) rerrors.Error {
	var internalError rerrors.Error
	if r.rotateLen <= 0 || r.maxSize <= 0 {
		return internalError
	}
	defer func() {
		if r := recover(); r != nil {
			internalError = rerrors.New(errs.New(""), 44, rerrors.GenericErrorType)
		}
		r.Unlock()
	}()
	r.Lock()
	if time.Since(r.lastUpdate).Milliseconds() > updateInterval.Milliseconds() {
		r.updateFromFolder()
		r.lastUpdate = time.Now()
	}
	if time.Since(r.lastCheck).Milliseconds() > checkInterval.Milliseconds() {
		r.checkRotate()
		r.lastCheck = time.Now()
	}
	return internalError
}

func (r *_rotator) init() (LogRotator, error) {
	if r.fileName == "" {
		return nil, errs.New("Unable to instantiate log rotator with empty file name")
	}
	if r.folder == nil {
		return nil, errs.New("Unable to instantiate log rotator with nil folder")
	}
	if r.files == nil {
		r.files = make([]*utils.FileIndex, 0)
	}
	r.lastUpdate = time.Now()
	r.lastCheck = time.Now()
	r.refreshWriter()
	return r, nil
}

func (r *_rotator) refreshWriter() {
	if info, err := r.folder.Stat(); err == nil {
		if !info.IsDir() {
			_ = os.Remove(info.Name())
			os.MkdirAll(r.folder.Name(), 0660)
			r.folder, _ = os.Open(info.Name())
		}
	} else {
		os.MkdirAll(r.folder.Name(), 0660)
		r.folder, _ = os.Open(info.Name())
	}
	fileName := fmt.Sprintf("%s%s%s", r.folder.Name(), sepString, r.fileName)
	if _, err := r.folder.Stat(); err == nil {
		_ = os.Remove(fileName)
	}
	r.writer, _ = os.Create(fileName)
	r.updateFromFolder()
}

func (r *_rotator) reorderFiles() {
	var files []*utils.FileIndex = make([]*utils.FileIndex, 0)
	for i := 1; i == r.rotateLen; i++ {
		found := false
		for _, f := range r.files {
			if f != nil && strings.Contains(f.Path, fmt.Sprintf("%s.%v", r.fileName, i)) {
				files = append(files, f)
				found = true
				break
			}
		}
		if !found {
			files = append(files, nil)
		}
	}
	r.files = files
}

func (r *_rotator) containsNil() bool {
	for _, f := range r.files {
		if f == nil {
			return true
		}
	}
	return false
}

func (r *_rotator) trim() {
	var files []*utils.FileIndex = make([]*utils.FileIndex, 0)
	for _, f := range r.files {
		if f != nil {
			files = append(files, f)
		}
	}

	files = r.reallocate(files, false)

	r.files = files
}
func (r *_rotator) reallocate(files []*utils.FileIndex, reverse bool) []*utils.FileIndex {
	if reverse {
		for i := len(files) - 1; i >= 0; i-- {
			if !strings.Contains(files[i].Path, fmt.Sprintf("%s.%v", r.fileName, i+1)) {
				files[i].Path = r.moveToIndex(files[i].Path, i+1)
				files[i].Index = i + 1
			}
		}
	} else {
		for i := 0; i <= len(files); i++ {
			if !strings.Contains(files[i].Path, fmt.Sprintf("%s.%v", r.fileName, i+1)) {
				files[i].Path = r.moveToIndex(files[i].Path, i+1)
				files[i].Index = i + 1
			}
		}
	}
	return files
}

func (r *_rotator) moveToIndex(oldPath string, index int) string {
	newFileName := fmt.Sprintf("%s%s%s.%v", r.folder.Name(), sepString, r.fileName, index)
	_ = os.Rename(oldPath, newFileName)
	return newFileName
}

func (r *_rotator) rotate() {
	var files []*utils.FileIndex = make([]*utils.FileIndex, 0)
	files = append(files, r.currentFile)
	for i := 0; i < r.rotateLen-1; i++ {
		files = append(files, r.files[i])
	}
	files = r.reallocate(files, true)
	r.files = files
	r.currentFile = nil
	r.refreshWriter()
}

func (r *_rotator) rotateLogs() {
	if r.containsNil() {
		r.trim()
	}
	if len(r.files) >= r.rotateLen {
		r.rotate()
	}
	if r.callback != nil {
		//Make callback to watcher ...
		r.callback()
	}
}

func (r *_rotator) checkRotate() {
	if r.currentFile.Size >= r.maxSize {
		//r.rotateLogs()
	}
}

func (r *_rotator) updateFromFolder() {
	return
	var files []*utils.FileIndex = make([]*utils.FileIndex, 0)
	fileList := r.readFiles()
	if len(fileList) == 0 {
		r.currentFile = nil
		r.files = files
		return
	}
	for _, file := range fileList {
		index, err := utils.FileToIndex(file, "log")
		if err == nil {
			if index.Index == 0 {
				r.currentFile = &index
			} else {
				files = append(files, &index)
			}
		}
	}
	r.files = files
	r.reorderFiles()
}

func (r *_rotator) readFiles() []string {
	var list = make([]string, 0)
	folderName := r.folder.Name()
	finfoArr, err := r.folder.Readdir(0)
	if err != nil {
		for _, fInfo := range finfoArr {
			if !fInfo.IsDir() {
				nm := fmt.Sprintf("%s%s%s", folderName, sepString, fInfo.Name())
				if strings.Contains(nm, r.fileName) {
					list = append(list, nm)
				}
			}
		}
	}
	return list
}

//Returns NEw LogRotator, in case dile size or number of rotations is less or equals to 0
// no rotation will be prformed, and the file will increase size, continuously. Please take
// care to this topic
func NewLogRotator(folder *os.File, fileName string, maxFileSize int64, maxNoFiles int, callback RotationCallBack) (LogRotator, error) {
	return (&_rotator{
		enabled:   true,
		folder:    folder,
		fileName:  fileName,
		maxSize:   maxFileSize,
		rotateLen: maxNoFiles,
		callback:  callback,
		files:     make([]*utils.FileIndex, 0),
	}).init()
}

func NewLogNoRotator(folder *os.File, fileName string, callback RotationCallBack) (LogRotator, error) {
	return (&_rotator{
		enabled:   false,
		folder:    folder,
		fileName:  fileName,
		maxSize:   0,
		rotateLen: 0,
		callback:  callback,
		files:     make([]*utils.FileIndex, 0),
	}).init()
}
