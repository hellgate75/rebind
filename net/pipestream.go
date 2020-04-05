// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package net

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

type NetPipe interface{
	Start() error
	Stop()
	IsRunning() bool
	GetInputChannel() <- chan []byte
	GetOutputChannel() chan []byte
	Write(d []byte) (int, error)
}

type pipe struct {
	listenPort	int
	answerPort	int
	listener	net.Listener
	inChan   	chan []byte
	outChan  	chan []byte
	_active  	bool
}

func (p *pipe) Start() error {
	var internalError error
	defer func(){
		if r := recover(); r != nil {
			internalError = errors.New(fmt.Sprintf("NetPipe.Start: Runtime error: %v", r))
		}
	}()
	fmt.Printf("NetPipe.Start -> Start listening on port: %v\n", p.listenPort)
	p._active = true
	p.inChan = make(chan []byte)
	p.outChan = make(chan []byte)
	p.listener, internalError = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", p.listenPort))
	if internalError != nil {
		return nil
	}
	go p.writeOnChannelRead()
	go p.readOnOutNP()
	return internalError
}

func (p *pipe) Stop() {
	fmt.Printf("NetPipe.Stop -> Stop listening on port: %v\n", p.listenPort)
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
	defer func(){
		if r := recover(); r != nil {
			fmt.Printf("NetPipe.ReadThread: Runtime error: %v\n", r)
			p.Stop()
		}
	}()
	for p._active {
		fmt.Printf("NetPipe.ReadThread -> Waiting for connection on port: %v\n", p.listenPort)
		var conn net.Conn
		var err error
		if conn, err = p.listener.Accept(); err == nil{
			go p.handleRequest(conn)
		} else if p._active {
			fmt.Printf("NetPipe.ReadThread -> Error accepting client on port: %v -> Error: %v\n", p.listenPort, err)
		}

	}
}
func (p *pipe) handleRequest(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("NetPipe.HandleRequest: Runtime error: %v\n", r)
			p.Stop()
		}
		conn.Close()
	}()
	fmt.Println("NetPipe.HandleRequest: Handling Conn with client...")
	byteArr, err := ioutil.ReadAll(conn)
	if err == nil {
		fmt.Println("NetPipe.HandleRequest: Writing Client Request On NetPipe...")
		p.outChan <- byteArr
	} else {
		fmt.Printf("NetPipe.HandleRequest: Reading error: %v\n", err)
		//TODO: Print errors to Logger
	}
}
func (p *pipe) writeOnChannelRead() {
	defer func(){
		if r := recover(); r != nil {
			fmt.Printf("NetPipe.WriteThread: Runtime error: %v\n", r)
			p.Stop()
		}
	}()
	for p._active {
		//fmt.Printf("NetPipe.WriteThread: Waiting for message to send at port: %v\n", p.answerPort)
		select{
		case msg, ok := <- p.inChan:
			if ! ok {
				continue
			}
			outConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", p.answerPort))
			if err != nil {
				panic(err)
			}
			fmt.Printf("NetPipe.WriteThread: connected on port: %v\n", p.answerPort)
			go p.handleAnswer(outConn, msg)
		case <- time.After(10*time.Second):
		default:

		}
	}
}

func (p *pipe) handleAnswer(outConn net.Conn, msg []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("NetPipe.HandleAnswer: Runtime error: %v\n", r)
			p.Stop()
		}
		outConn.Close()
	}()
	if outConn != nil {
		fmt.Printf("NetPipe.HandleAnswer: Sending message at port: %v\n", p.answerPort)
		_, _ = outConn.Write(msg)
	} else {
		fmt.Println("NetPipe.HandleAnswer: Failed to acquire writer")
	}
}


func (p *pipe) GetInputChannel() <- chan []byte {
	return p.inChan
}

func (p *pipe) GetOutputChannel() chan []byte {
	return p.outChan
}

func (p *pipe) Write(d []byte) (int, error) {
	defer func(){
		if r := recover(); r != nil {
			fmt.Printf("NetPipe.Write: Runtime error: %v", r)
			p.Stop()
		}
	}()
	fmt.Printf("NetPipe.Write: Connecting on port: %v\n", p.answerPort)
	outConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", p.answerPort))
	if err != nil {
		panic(err)
	}
	defer outConn.Close()
	var n int
	defer func(){
		if r := recover(); r != nil {
			n = 0
			err = errors.New(fmt.Sprintf("NetPipe.Write: Runtime error: %v", r))
		}
	}()
	if outConn != nil {
		n, err = outConn.Write(d)
	} else {
		err = errors.New("NetPipe.Write: Failed to acquire writer")
	}
	return n, err
}

func New(internalReadPort int,internalWritePort int) (NetPipe, error) {
	if internalReadPort < 50 ||  internalReadPort > 25000 {
		return nil, errors.New(fmt.Sprintf("NetPipe.New: invalid input port: %v", internalReadPort))
	}
	if internalWritePort < 50 ||  internalWritePort > 25000 {
		return nil, errors.New(fmt.Sprintf("NetPipe.New: invalid output port: %v", internalWritePort))
	}
	if internalWritePort == internalReadPort {
		return nil, errors.New("NetPipe.New: input and output port must be different")

	}
	//pr, pw := net.NetPipe()
	//_ = net.TeeReader(pr, fileOut)

	return &pipe{
		listenPort: internalReadPort,
		answerPort: internalWritePort,
	}, nil

}