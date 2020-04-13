// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	pnet "github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/rest/services"
	"github.com/hellgate75/rebind/utils"
	"net"
	"net/http"
	"os"
)

var rwDirPath string
var listenIP string
var listenPort int
var dnsPipeIP string
var dnsPipePort int
var dnsPipeResponsePort int
var tlsCert string
var tlsKey string

//TODO: Give Life to Logger
var logger log.Logger = log.NewLogger("re-web", log.INFO)

var defaultForwarders = make([]net.UDPAddr, 0)

func init() {
	logger.Info("Initializing Re-Web Rest Server ....")
	flag.StringVar(&rwDirPath, "rwdir", model.DefaultStorageFolder, "dns storage dir")
	flag.StringVar(&listenIP, "listen-ip", model.DefaultIpAddress, "http server ip")
	flag.IntVar(&listenPort, "listen-port", model.DefaultRestServerPort, "http server port")
	flag.StringVar(&dnsPipeIP, "dns-pipe-ip", model.DefaultDnsPipeAddress, "tcp dns pipe ip")
	flag.IntVar(&dnsPipePort, "dns-pipe-port", model.DefaultDnsPipePort, "tcp dns pipe port")
	flag.IntVar(&dnsPipeResponsePort, "dns-pipe-response-port", model.DefaultDnsPipePort, "tcp dns pipe responses port")
	flag.StringVar(&tlsCert, "tsl-cert", "", "tls certificate file path")
	flag.StringVar(&tlsKey, "tsl-key", "", "tls certificate key file path")
}

func main() {
	logger.Info("Starting Re-Web Rest Server ...")
	flag.Parse()
	if utils.StringsListContainItem("-h", flag.Args(), true) ||
		utils.StringsListContainItem("--help", flag.Args(), true) {
		flag.Usage()
		os.Exit(0)
	}
	if err := os.MkdirAll(rwDirPath, 0666); err != nil {
		logger.Errorf("Create rwdirpath: %v error: %v", rwDirPath, err)
		return
	}

	// Improve list of default group forwarders if provided
	defaultForwarders = append(defaultForwarders, model.DefaultGroupForwarders...)

	// Create network Pipe Stream with the dns server
	pipe, err := pnet.NewInputOutputPipeWith(dnsPipeIP, dnsPipePort, dnsPipeIP, dnsPipeResponsePort, nil, logger)
	if err != nil {
		logger.Fatalf("Unable to create NetPipe in listen: %v and bind: %v/%v\n", dnsPipePort, dnsPipeResponsePort)
	}
	// Create Data Store
	store := registry.NewStore(logger, rwDirPath, defaultForwarders)
	store.Load()

	// Handler stuf for the API service groups
	dnsHandler := func(serv services.RestService) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				serv.Create(w, r)
			case http.MethodGet:
				serv.Read(w, r)
			case http.MethodPut:
				serv.Update(w, r)
			case http.MethodDelete:
				serv.Delete(w, r)
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
	var proto string = "http"
	if tlsCert != "" && tlsKey != "" {
		proto = "https"
	}
	// Creates/Sets API endpoints handlers
	services.CreateApiEndpoints(rtr, withAuth, dnsHandler,
		pipe, store, logger, fmt.Sprintf("%s://%s:%v", proto, listenIP, listenPort))

	//Adding entry point for generic queries (GET)
	http.Handle("/", rtr)

	// Adding TLS certificates if required
	if tlsCert == "" || tlsKey == "" {
		logger.Infof("RestService start-up:: Starting server in simple mode on ip: %s and port: %v\n", listenIP, listenPort)
		err = http.ListenAndServe(fmt.Sprintf("%s:%v", listenIP, listenPort), nil)
	} else {
		logger.Infof("RestService start-up:: Starting server in simple mode on ip: %s and port: %v\n", listenIP, listenPort)
		logger.Infof("RestService start-up:: Using certificate file: %s and certticate key file: %v..\n", tlsCert, tlsKey)
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%v", listenIP, listenPort), tlsCert, tlsKey, nil)
	}
	if err != nil {
		logger.Fatalf("RestService start-up:: Error listening on s:%v - Error: %v\n", listenIP, listenPort, err)
	}
}
