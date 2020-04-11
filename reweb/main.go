// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/reweb/rest"
	"net/http"
	"os"
)

var rwDirPath = flag.String("rwdir", "/var/dns", "dns storage dir")
var listenIP = flag.String("listen-ip", "8.8.8.8", "http server ip")
var listenPort = flag.Int("listen-port", 9000, "http server port")
var dnsPipeIP = flag.String("dns-pipe-ip", "127.0.0.1", "tcp dns pipe ip")
var dnsPipePort = flag.Int("dns-pipe-port", 953, "tcp dns pipe port")
var tlsCert = flag.String("tsl-cert", "", "tls certificate file path")
var tlsKey = flag.String("tsl-key", "", "tls certificate key file path")

const (
	internalListenPort int = 954
	internalDialPort   int = 953
)

//TODO: Give Life to Logger
var logger log.Logger = log.NewLogger("re-web", log.INFO)

func main() {
	flag.Parse()
	if err := os.MkdirAll(*rwDirPath, 0666); err != nil {
		logger.Errorf("create rwdirpath: %v error: %v", *rwDirPath, err)
		return
	}
	logger.Info("starting re-web server ...")
	pipe, err := net.NewInputOutputPipe(internalListenPort, internalDialPort, nil, logger)
	if err != nil {
		logger.Fatalf("Unable to create NetPipe in listen: %v and bind: %v\n", internalListenPort, internalDialPort)
	}
	rest := rest.NewRestService(pipe)

	dnsHandler := func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			//TODO: Filter Method by request path: r.RequestURI
			switch r.Method {
			case http.MethodPost:
				rest.Create(w, r)
			case http.MethodGet:
				rest.Read(w, r)
			case http.MethodPut:
				rest.Update(w, r)
			case http.MethodDelete:
				rest.Delete(w, r)
			}
		}
	}

	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		// authentication intercepting
		var _ = "intercept"
		return func(w http.ResponseWriter, r *http.Request) {
			h(w, r)
		}
	}

	rtr := mux.NewRouter()
	//Adding entry point for generic queries (GET)
	rtr.HandleFunc("/dns", withAuth(dnsHandler())).Methods("GET", "POST")
	//Adding entry point for zones queries (PUT, POST, DEL, GET)
	rtr.HandleFunc("/dns/zones", withAuth(dnsHandler())).Methods("GET", "PUT")
	//Adding entry point for spcific zone queries (PUT, POST, DEL, GET)
	rtr.HandleFunc("/dns/zone/{name:[a-zA-Z0-9]+}/profile", withAuth(dnsHandler())).Methods("GET", "POST", "PUT", "DELETE")

	http.Handle("/", rtr)
	if *tlsCert == "" || *tlsKey == "" {
		logger.Infof("RestService start-up:: Starting server in simple mode on ip: %s and port: %v\n", *listenIP, *listenPort)
		err = http.ListenAndServe(fmt.Sprintf("%s:%v", *listenIP, *listenPort), nil)
	} else {
		logger.Infof("RestService start-up:: Starting server in simple mode on ip: %s and port: %v\n", *listenIP, *listenPort)
		logger.Infof("RestService start-up:: Using certificate file: %s and certticate key file: %v..\n", *tlsCert, *tlsKey)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%v", *listenIP, *listenPort), *tlsCert, *tlsKey, nil)
	}
	if err != nil {
		logger.Fatalf("RestService start-up:: Error listening on s:%v - Error: %v\n", *listenIP, *listenPort, err)
	}
}
