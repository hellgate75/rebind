// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"
	"fmt"
	rErrrors "github.com/hellgate75/rebind/rerrors"
	"net"
	"sync"
)

type RequestsCacheStore interface {
	Get(key string) ([]net.UDPAddr, rErrrors.Error)
	Set(key string, log net.UDPAddr) rErrrors.Error
	Remove(key string) rErrrors.Error
}

type RequestsCacheStoreData struct {
	sync.RWMutex
	store map[string][]net.UDPAddr
}

func (b *RequestsCacheStoreData) Get(key string) ([]net.UDPAddr, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("Key "+key+" doesn't exist"), int64(20), rErrrors.StoreProcessErrorType)
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

func (b *RequestsCacheStoreData) Set(key string, log net.UDPAddr) rErrrors.Error {
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
		b.store[key] = []net.UDPAddr{log}
	}
	return internalErr
}

func (b *RequestsCacheStoreData) Remove(key string) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key "+key+" doesn't exist"), int64(24), rErrrors.StoreProcessErrorType)
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

func NewRequestsCacheStore() RequestsCacheStore {
	return &RequestsCacheStoreData{
		store: make(map[string][]net.UDPAddr),
	}
}
