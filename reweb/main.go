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
	"github.com/hellgate75/rebind/reweb/rest"
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
var tlsCert string
var tlsKey string

const (
	internalListenPort int = 954
	internalDialPort   int = 953
)

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
	defaultForwarders = append(defaultForwarders, model.DefaultGroupForwarders...)
	pipe, err := pnet.NewInputOutputPipe(internalListenPort, internalDialPort, nil, logger)
	if err != nil {
		logger.Fatalf("Unable to create NetPipe in listen: %v and bind: %v\n", internalListenPort, internalDialPort)
	}
	store := registry.NewStore(logger, rwDirPath, defaultForwarders)
	store.Load()
	v1GroupsRest := rest.NewV1DnsGroupsRestService(pipe, store, logger)
	v1GroupRest := rest.NewV1DnsGroupRestService(pipe, store, logger)
	v1DnsRootRest := rest.NewV1DnsRootRestService(pipe, store, logger)

	dnsHandler := func(serv rest.RestService) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			//TODO: Filter Method by request path: r.RequestURI
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
	//Adding entry point for generic queries (GET)
	rtr.HandleFunc("/v1/dns", withAuth(dnsHandler(v1DnsRootRest))).Methods("GET", "POST")
	//Adding entry point for zones queries (PUT, POST, DEL, GET)
	rtr.HandleFunc("/v1/dns/groups", withAuth(dnsHandler(v1GroupsRest))).Methods("GET", "POST")
	//Adding entry point for spcific zone queries (PUT, POST, DEL, GET)
	rtr.HandleFunc("/v1/dns/group/{name:[a-zA-Z0-9]+}/profile", withAuth(dnsHandler(v1GroupRest))).Methods("GET", "POST", "PUT", "DELETE")

	http.Handle("/", rtr)
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
