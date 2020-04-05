// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"fmt"
	"net"
	rErrrors "github.com/hellgate75/rebind/errors"
	"sync"
	"errors"
	"time"
)

type LogRecord struct {
	Time	time.Time
	Caller	net.UDPAddr
}

func NewLogRecord(caller net.UDPAddr) LogRecord {
	return LogRecord{
		Time: time.Now(),
		Caller: caller,
	}
}

type LogStore interface {
	Get(key string) ([]LogRecord, rErrrors.Error)
	Set(key string, log LogRecord) rErrrors.Error
	Remove(key string) rErrrors.Error
}

type _callCacheStore struct {
	sync.RWMutex
	store map[string][]LogRecord
}

func (b *_callCacheStore) Get(key string) ([]LogRecord, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("Key " + key + " doesn't exist"), int64(20), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(21), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	val, ok := b.store[key]
	if ok {
		internalErr = nil
	}
	return val, internalErr
}

func (b *_callCacheStore) Set(key string, log LogRecord) rErrrors.Error {
	var internalErr rErrrors.Error
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if _, ok := b.store[key]; ok {
		b.store[key] = append(b.store[key], log)
	} else {
		b.store[key] = []LogRecord{log}
	}
	return internalErr
}

func (b *_callCacheStore) Remove(key string) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key " + key + " doesn't exist"), int64(24), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(25), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	_, ok := b.store[key]
	if ok {
		delete(b.store, key)
		internalErr = nil
	}
	return internalErr
}

func NewLogStore() LogStore{
	return &_callCacheStore{
		store: make(map[string][]LogRecord),
	}
}