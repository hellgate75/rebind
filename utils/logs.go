// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"errors"
)

type FileIndex struct {
	Path	string
	Size	int64
	Index	int
}

func FileToIndex(path string, extension string) (FileIndex, error) {
	fi, err := os.Stat(path)
	if err != nil {
		var index int = 0
		txt:=strings.Split(path, ".")
		last := txt[len(txt)-1]
		if strings.ToLower(last) != strings.ToLower(extension) {
			num, err := strconv.Atoi(last)
			if err != nil {
				index = -1
			} else {
				index = num
			}
		}
		if index < 0 {
			return FileIndex{}, errors.New(fmt.Sprintf("Invalid file format: %s", path))
		}
		return FileIndex{
			Path: path,
			Size: fi.Size(),
			Index: index,
		}, nil
	}
	return FileIndex{}, err
}
