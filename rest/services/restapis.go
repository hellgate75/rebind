package services

import (
	"github.com/gorilla/mux"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"net/http"
)

func CreateApiEndpoints(router *mux.Router,
	authFunc func(h http.HandlerFunc) http.HandlerFunc,
	dnsHandler func(serv RestService) http.HandlerFunc,
	pipe net.NetPipe,
	store registry.Store,
	logger log.Logger,
	hostBaseUrl string) {
	//Create v1 APIs
	createV1ApiEndpoints(router, authFunc, dnsHandler, pipe, store, logger, hostBaseUrl)
}

func createV1ApiEndpoints(router *mux.Router,
	authFunc func(h http.HandlerFunc) http.HandlerFunc,
	dnsHandler func(serv RestService) http.HandlerFunc,
	pipe net.NetPipe,
	store registry.Store,
	logger log.Logger,
	hostBaseUrl string) {
	v1DnsRootRest := NewV1DnsRootRestService(pipe, store, logger, hostBaseUrl)
	v1GroupsRest := NewV1DnsGroupsRestService(pipe, store, logger, hostBaseUrl)
	v1GroupRest := NewV1DnsGroupRestService(pipe, store, logger, hostBaseUrl)
	v1DnsGroupResourcesRest := NewV1DnsGroupResourcesRestService(pipe, store, logger, hostBaseUrl)
	v1DnsGroupResourceDetailsRest := NewV1DnsGroupResourceDetailsRestService(pipe, store, logger, hostBaseUrl)
	//Adding entry point for zones queries (PUT, POST, DEL, GET)
	router.HandleFunc("/v1/dns", authFunc(dnsHandler(v1DnsRootRest))).Methods("GET", "POST", "PUT", "DELETE")
	//Adding entry point for groups queries (PUT, POST, DEL, GET)
	router.HandleFunc("/v1/dns/groups", authFunc(dnsHandler(v1GroupsRest))).Methods("GET", "POST", "PUT", "DELETE")
	//Adding entry point for spcific group queries (PUT, POST, DEL, GET)
	router.HandleFunc("/v1/dns/group/{group:[a-zA-Z0-9]+}", authFunc(dnsHandler(v1GroupRest))).Methods("GET", "POST", "PUT", "DELETE")
	//Adding entry point for specific group queries (PUT, POST, DEL, GET)
	router.HandleFunc("/v1/dns/group/{group:[a-zA-Z0-9]+}/resources", authFunc(dnsHandler(v1DnsGroupResourcesRest))).Methods("GET", "POST", "PUT", "DELETE")
	//Adding entry point for specific group queries (PUT, POST, DEL, GET)
	router.HandleFunc("/v1/dns/group/{group:[a-zA-Z0-9]+}/resources/{resource:[a-zA-Z0-9]+}", authFunc(dnsHandler(v1DnsGroupResourceDetailsRest))).Methods("GET", "POST", "PUT", "DELETE")
}
