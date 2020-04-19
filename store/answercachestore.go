// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"
	"fmt"
	"github.com/hellgate75/rebind/model"
	rErrrors "github.com/hellgate75/rebind/rerrors"
	"github.com/hellgate75/rebind/utils"
	"golang.org/x/net/dns/dnsmessage"
	"sync"
	"time"
)

var DEFAULT_ANSWER_TIME_TO_LIVE time.Duration = 5 * time.Minute

type AnswersCacheStore interface {
	Get(key string) ([]dnsmessage.Resource, rErrrors.Error)
	Set(key string, log ...dnsmessage.Resource) rErrrors.Error
	Remove(key string) rErrrors.Error
	Trim() rErrrors.Error
}

type AnswersCacheStoreData struct {
	sync.RWMutex
	store map[string][]model.AnswerBlock
	path  string
}

func (b *AnswersCacheStoreData) Get(key string) ([]dnsmessage.Resource, rErrrors.Error) {
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
	var out = make([]dnsmessage.Resource, 0)
	for _, v := range val {
		if v.IsValid() {
			out = append(out, v.Answer...)
		}
	}
	out = utils.RemoveDuplicatesInResourceList(out)
	return out, internalErr
}

func (b *AnswersCacheStoreData) Set(key string, log ...dnsmessage.Resource) rErrrors.Error {
	var internalErr rErrrors.Error
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if _, ok := b.store[key]; ok {
		b.store[key] = append(b.store[key], model.AnswerBlock{
			Created: time.Now(),
			TTL:     DEFAULT_ANSWER_TIME_TO_LIVE,
			Answer:  log,
		})
	} else {
		b.store[key] = []model.AnswerBlock{model.AnswerBlock{
			Created: time.Now(),
			TTL:     DEFAULT_ANSWER_TIME_TO_LIVE,
			Answer:  log,
		}}
	}
	return internalErr
}

func (b *AnswersCacheStoreData) Remove(key string) rErrrors.Error {
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

func (b *AnswersCacheStoreData) Trim() rErrrors.Error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	for key, value := range b.store {
		newVal := make([]model.AnswerBlock, 0)
		for _, t := range value {
			if t.IsValid() {
				newVal = append(newVal, t)
			}
		}
		value = newVal

		if len(value) == 0 {
			delete(b.store, key)
		} else {
			b.store[key] = value
		}
	}
	if err != nil {
		return rErrrors.New(err, 55,
			rErrrors.StoreProcessErrorType)
	}
	return nil
}

func NewAnswersCacheStore() AnswersCacheStore {
	return &AnswersCacheStoreData{
		store: make(map[string][]model.AnswerBlock),
	}
}
