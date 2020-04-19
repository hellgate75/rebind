// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package rerrors

type ErrorType int

const (
	UndefinedType    ErrorType = 0
	GenericErrorType ErrorType = iota + 1
	DnsMessageErrorType
	DnsProcessingErrorType
	RestMessageErrorType
	RestProcessingErrorType
	StoreLoadErrorType
	StoreSaveErrorType
	StoreCreateErrorType
	StoreProcessErrorType
	ConfigLoadErrorType
)

// Interface that describe a cross application error
type Error interface {
	// Returns Error Code Value
	Code() int64
	// Returns Error Category Type
	Type() ErrorType
	// Returns thrown error interface
	Error() error
}

type _error struct {
	err     error
	code    int64
	errType ErrorType
}

func (err *_error) Code() int64 {
	return err.code
}

func (err *_error) Type() ErrorType {
	return err.errType
}

func (err *_error) Error() error {
	return err.err
}

func New(err error, code int64, errType ErrorType) Error {
	return &_error{
		err:     err,
		errType: errType,
		code:    code,
	}
}
