package services

import (
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/rest/services/v1"
	"net/http"
)

type RestService interface {
	Create(w http.ResponseWriter, r *http.Request)
	Read(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func NewV1DnsRootRestService(pipe net.NetPipe, store registry.Store, logger log.Logger) RestService {
	return &v1.DnsRootService{
		Pipe:  pipe,
		Store: store,
		Log:   logger,
	}
}

func NewV1DnsGroupsRestService(pipe net.NetPipe, store registry.Store, logger log.Logger) RestService {
	return &v1.DnsGroupsService{
		Pipe:  pipe,
		Store: store,
		Log:   logger,
	}
}

func NewV1DnsGroupRestService(pipe net.NetPipe, store registry.Store, logger log.Logger) RestService {
	return &v1.DnsGroupService{
		Pipe:  pipe,
		Store: store,
		Log:   logger,
	}
}

func NewV1DnsGroupResourcesRestService(pipe net.NetPipe, store registry.Store, logger log.Logger) RestService {
	return &v1.DnsGroupResourcesService{
		Pipe:  pipe,
		Store: store,
		Log:   logger,
	}
}
