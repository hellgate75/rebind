// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package net

import (
	"errors"
	"fmt"
	"github.com/hellgate75/rebind/log"
	"io/ioutil"
	"net"
	"time"
)

type PipeHandler func(message string)

type PipeType byte

const (
	DEFAULT_LISTEN_ADDRESS string   = "127.0.0.1"
	DEFAULT_ANSWER_ADDRESS string   = "127.0.0.1"
	MIN_PORT_NUMBER        int      = 50
	MAX_PORT_NUMBER        int      = 25000
	PIPE_IN                PipeType = iota + 1
	PIPE_OUT
	PIPE_INOUT
	debugLevel level = iota + 1
	infoLevel
	warnLevel
	errorLevel
	fatalLevel
)

func (t *PipeType) String() string {
	switch *t {
	case PIPE_IN:
		return "Input Net Pipe"
	case PIPE_OUT:
		return "Output Net Pipe"
	case PIPE_INOUT:
		return "Input/Output Net Pipe"
	default:
		return "Unknown Net Pipe"
	}
}

type NetPipe interface {
	Start() error
	Stop()
	IsRunning() bool
	GetInputChannel() (<-chan []byte, error)
	GetOutputChannel() (chan []byte, error)
	Write(d []byte) (int, error)
}

type level byte

func (l *level) String() string {
	switch *l {
	case debugLevel:
		return "DEBUG"
	case infoLevel:
		return "INFO "
	case warnLevel:
		return "WARN "
	case errorLevel:
		return "ERROR"
	case fatalLevel:
		return "FATAL"
	default:
		return "NONE "
	}
}

type pipe struct {
	listenPort    int
	answerPort    int
	listenAddress string
	answerAddress string
	listener      net.Listener
	inChan        chan []byte
	outChan       chan []byte
	_pType        PipeType
	_active       bool
	_handler      PipeHandler
	_log          log.Logger
}

func (p *pipe) log(level level, m string, args ...interface{}) {
	message := m
	if len(args) > 0 {
		message = fmt.Sprintf(m, args...)
	}
	if p._log != nil {
		switch level {
		case debugLevel:
			p._log.Debug(message)
			break
		case warnLevel:
			p._log.Warn(message)
			break
		case errorLevel:
			p._log.Error(message)
			break
		case fatalLevel:
			p._log.Fatal(message)
			break
		default:
			p._log.Info(message)
		}
	} else {
		fmt.Println(fmt.Sprintf("[%s] %s"), level.String(), message)
	}
}

func (p *pipe) Start() error {
	if p._pType == PIPE_OUT {
		p.log(errorLevel, "NetPipe.Start: Unable to start listener for an out pipe")
		return errors.New("NetPipe.Start: Unable to start listener for an out pipe")
	}
	var internalError error
	defer func() {
		if r := recover(); r != nil {
			internalError = errors.New(fmt.Sprintf("NetPipe.Start: Runtime error: %v", r))
		}
	}()
	p.log(infoLevel, "NetPipe.Start -> Start listening on port: %v", p.listenPort)
	p._active = true
	p.inChan = make(chan []byte)
	if p._handler == nil {
		p.outChan = make(chan []byte)
	} else {
		p.outChan = nil
	}
	p.listener, internalError = net.Listen("tcp", fmt.Sprintf("%s:%v", p.listenAddress, p.listenPort))
	if internalError != nil {
		return nil
	}
	go p.writeOnChannelRead()
	go p.readOnOutNP()
	return internalError
}

func (p *pipe) Stop() {
	if p._pType == PIPE_OUT {
		p.log(errorLevel, "NetPipe.Stop: Unable to stop listener for an out pipe")
		return
	}
	p.log(infoLevel, "NetPipe.Stop -> Stop listening on port: %v", p.listenPort)
	p._active = false
	time.Sleep(500 * time.Millisecond)
	defer func() {
		if p.inChan != nil {
			close(p.inChan)
		}
		p.inChan = nil
		if p.outChan != nil {
			close(p.outChan)
		}
		p.outChan = nil
		if p.listener != nil {
			p.listener.Close()
		}
		p.listener = nil
	}()
}

func (p *pipe) IsRunning() bool {
	return p._active
}

func (p *pipe) readOnOutNP() {
	defer func() {
		if r := recover(); r != nil {
			p.log(errorLevel, "NetPipe.ReadThread: Runtime error: %v", r)
			p.Stop()
		}
	}()
	for p._active {
		p.log(debugLevel, "NetPipe.ReadThread -> Waiting for connection on port: %v", p.listenPort)
		var conn net.Conn
		var err error
		if conn, err = p.listener.Accept(); err == nil {
			go p.handleRequest(conn)
		} else if p._active {
			p.log(errorLevel, "NetPipe.ReadThread -> Error accepting client on port: %v -> Error: %v", p.listenPort, err)
		}

	}
}
func (p *pipe) handleRequest(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			p.log(errorLevel, "NetPipe.HandleRequest: Runtime error: %v", r)
			p.Stop()
		}
		conn.Close()
	}()
	p.log(debugLevel, "NetPipe.HandleRequest: Handling Conn with client...")
	byteArr, err := ioutil.ReadAll(conn)
	if err == nil {
		p.log(debugLevel, "NetPipe.HandleRequest: Writing Client Request On NetPipe...")
		if p._handler != nil {
			p._handler(string(byteArr))
		} else {
			p.outChan <- byteArr
		}
	} else {
		p.log(errorLevel, "NetPipe.HandleRequest: Reading error: %v", err)
	}
}
func (p *pipe) writeOnChannelRead() {
	defer func() {
		if r := recover(); r != nil {
			p.log(errorLevel, "NetPipe.WriteThread: Runtime error: %v", r)
			p.Stop()
		}
	}()
	for p._active {
		//fmt.Printf("NetPipe.WriteThread: Waiting for message to send at port: %v", p.answerPort)
		select {
		case msg, ok := <-p.inChan:
			if !ok {
				continue
			}
			outConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", p.answerPort))
			if err != nil {
				p.log(errorLevel, "NetPipe.WriteThread: error dialing on port: %v", err)
				panic(err)
			}
			p.log(debugLevel, "NetPipe.WriteThread: connected on port: %v", p.answerPort)
			go p.handleAnswer(outConn, msg)
		case <-time.After(10 * time.Second):
		default:

		}
	}
}

func (p *pipe) handleAnswer(outConn net.Conn, msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			p.log(errorLevel, "NetPipe.HandleAnswer: Runtime error: %v", r)
			p.Stop()
		}
		outConn.Close()
	}()
	if outConn != nil {
		p.log(debugLevel, "NetPipe.HandleAnswer: Sending message at port: %v", p.answerPort)
		_, _ = outConn.Write(msg)
	} else {
		p.log(errorLevel, "NetPipe.HandleAnswer: Failed to acquire writer")
	}
}

func (p *pipe) GetInputChannel() (<-chan []byte, error) {
	if p._pType == PIPE_OUT {
		p.log(errorLevel, "NetPipe.GetInputChannel: Unable to get input pipe for an out pipe")
		return nil, errors.New("NetPipe.GetInputChannel: Unable to get input pipe for an out pipe")
	}
	return p.inChan, nil
}

func (p *pipe) GetOutputChannel() (chan []byte, error) {
	if p._pType != PIPE_INOUT {
		p.log(errorLevel, "NetPipe.GetInputChannel: Unable to get output pipe for a not in-out pipe")
		return nil, errors.New("NetPipe.GetInputChannel: Unable to get output pipe for not in-out pipe")
	}
	if p._handler == nil {
		return p.outChan, nil
	}
	p.log(warnLevel, "NetPipe.GetOutputChannel Answer handled with PipeHandler")
	return nil, errors.New("NetPipe.GetOutputChannel Answer handled with PipeHandler")
}

func (p *pipe) Write(d []byte) (int, error) {
	if p._pType == PIPE_IN {
		p.log(errorLevel, "NetPipe.Start: Unable to write data for an in pipe")
		return 0, errors.New("NetPipe.Start: Unable to write data for an in pipe")
	}
	defer func() {
		if r := recover(); r != nil {
			p.log(errorLevel, "NetPipe.Write: Runtime error: %v", r)
			p.Stop()
		}
	}()
	p.log(debugLevel, "NetPipe.Write: Connecting on port: %v", p.answerPort)
	outConn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", p.listenAddress, p.answerPort))
	if err != nil {
		panic(err)
	}
	defer outConn.Close()
	var n int
	defer func() {
		if r := recover(); r != nil {
			n = 0
			err = errors.New(fmt.Sprintf("NetPipe.Write: Runtime error: %v", r))
		}
	}()
	if outConn != nil {
		n, err = outConn.Write(d)
	} else {
		p.log(errorLevel, "NetPipe.Write: Failed to acquire writer")
		err = errors.New("NetPipe.Write: Failed to acquire writer")
	}
	return n, err
}
func New(pType PipeType, inputPort int, outputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return NewNetPipe(pType, DEFAULT_LISTEN_ADDRESS, inputPort, DEFAULT_ANSWER_ADDRESS, outputPort, handler, logger)
}

func NewNetPipe(pType PipeType, listenAddress string, inputPort int, answerAddress string, outputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	if pType != PIPE_OUT && (inputPort < MIN_PORT_NUMBER || inputPort > MAX_PORT_NUMBER) {
		return nil, errors.New(fmt.Sprintf("NetPipe.New: invalid input port: %v", inputPort))
	}
	if pType != PIPE_IN && (outputPort < MIN_PORT_NUMBER || outputPort > MAX_PORT_NUMBER) {
		return nil, errors.New(fmt.Sprintf("NetPipe.New: invalid output port: %v", outputPort))
	}

	//if outputPort == inputPort {
	//	return nil, errors.New("NetPipe.New: input and output port must be different")
	//}

	if pType != PIPE_IN && pType != PIPE_OUT && pType != PIPE_INOUT {
		return nil, errors.New(fmt.Sprintf("NetPipe.New: Unknown pipe type: %v", pType))
	} else {
		if logger != nil {
			logger.Infof("NetPipe.New: Creating Pipe Type: %s", pType.String())
		} else {
			fmt.Printf("[INFO ] NetPipe.New: Creating Pipe Type: %s\n", pType.String())
		}
	}
	if handler == nil {
		if logger != nil {
			logger.Warn("NetPipe.New: handler is provided, proceeding with input handler ...")
		} else {
			fmt.Println("[WARN ] NetPipe.New: handler is provided, proceeding with input handler ...")
		}
	} else {
		if logger != nil {
			logger.Warn("NetPipe.New: handler is nil, proceeding with input channel ...")
		} else {
			fmt.Println("[WARN ] NetPipe.New: handler is nil, proceeding with input channel ...")
		}
	}

	return &pipe{
		_pType:     pType,
		listenPort: inputPort,
		answerPort: outputPort,
		_handler:   handler,
		_log:       logger,
	}, nil

}
func NewInputPipe(inputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return New(PIPE_IN, inputPort, 0, handler, logger)
}
func NewInputPipeWith(listenIp string, inputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return NewNetPipe(PIPE_IN, listenIp, inputPort, DEFAULT_ANSWER_ADDRESS, 0, handler, logger)
}
func NewOutputPipe(outputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return New(PIPE_OUT, 0, outputPort, handler, logger)
}
func NewOutputPipeWith(answerIp string, outputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return NewNetPipe(PIPE_OUT, DEFAULT_LISTEN_ADDRESS, 0, answerIp, outputPort, handler, logger)
}
func NewInputOutputPipe(inputPort int, outputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return New(PIPE_INOUT, inputPort, outputPort, handler, logger)
}
func NewInputOutputPipeWith(listenIp string, inputPort int, answerIp string, outputPort int, handler PipeHandler, logger log.Logger) (NetPipe, error) {
	return NewNetPipe(PIPE_INOUT, listenIp, inputPort, answerIp, outputPort, handler, logger)
}
