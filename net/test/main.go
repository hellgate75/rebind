// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/net"
	"sync"
	"time"
)

func main() {
	logger := log.NewLogger("test netpipe", log.DEBUG)
	pipe1, err3 := net.NewOutputPipe(395, nil, logger)
	if err3 != nil {
		fmt.Printf("Error creating pipe stream 1: %v", err3)
		return
	}
	pipe2, err4 := net.NewInputOutputPipe(395, 395, nil, logger)
	if err4 != nil {
		fmt.Printf("Error creating pipe stream 2: %v", err4)
		return
	}
	defer pipe2.Stop()

	go pipe2.Start()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		time.Sleep(4 * time.Second)
		fmt.Println("Pipe1: Sending data ...")
		n, err := pipe1.Write([]byte("Good afternoon!!"))
		fmt.Printf("Pipe1: Sent %v byte(s)\n", n)
		if err != nil {
			fmt.Printf("Pipe1: Write Error: %v\n", err)
		}
		fmt.Println("Pipe1: Complete - Exit!!")
		wg.Done()
	}()

	go func() {
		fmt.Println("Pipe2: Listening for clients ...")
		for {
			outChan, _ := pipe2.GetOutputChannel()
			select {
			case msg, ok := <-outChan:
				if ok {
					fmt.Printf("Pipe2: Message: %s\n", string(msg))
					fmt.Println("Pipe2: Complete - Exit!!")
					wg.Done()
					pipe2.Stop()
					return
				} else {
					fmt.Println("Pipe2: Message: Error")
				}
			case <-time.After(10 * time.Second):
				break
			default:
			}
		}
	}()

	//time.Sleep(5 * time.Second)
	wg.Wait()
}
