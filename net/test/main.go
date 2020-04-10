// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/hellgate75/rebind/net"
	"sync"
	"time"
)

func main() {
	pipe1, err3 := net.New(394, 395)
	if err3 != nil {
		fmt.Printf("Error creating pipe stream 1: %v", err3)
		return
	}
	defer pipe1.Stop()
	pipe2, err4 := net.New(395, 394)
	if err4 != nil {
		fmt.Printf("Error creating pipe stream 2: %v", err4)
		return
	}
	defer pipe2.Stop()

	go pipe1.Start()
	go pipe2.Start()
	//if err3 != nil {
	//	fmt.Printf("Error starting pipe stream 1: %v\n", err3)
	//	return
	//}
	//
	//if err4 != nil {
	//	fmt.Printf("Error starting pipe stream 2: %v\n", err4)
	//	return
	//}

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
		pipe1.Stop()
	}()

	go func() {
		fmt.Println("Pipe2: Listening for clients ...")
		for {
			select {
			case msg, ok := <-pipe2.GetOutputChannel():
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
