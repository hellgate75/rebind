// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package utils

import (
	"fmt"
	"github.com/hellgate75/rebind/model"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

var DEFAULT_DOMAIN_NAMES []string = []string{"", "home", "local"}

const DEFAULT_GROUP_NAME = "default"

func SplitDomainsFromHostname(hostName string) []string {
	var out []string = make([]string, 0)
	var domain string
	if strings.Index(hostName, ".") > 0 {
		tokens := strings.Split(hostName, ".")
		if len(tokens) > 1 {
			length := len(tokens) - 2
			domain = tokens[len(tokens)-1]
			if len(tokens) == 2 {
				length += 1
				domain = ""
			}
			for i := length; i >= 1; i-- {
				if len(domain) > 0 {
					domain = tokens[i] + "." + domain
				} else {
					domain = tokens[i]
				}
				out = append(out, domain)
			}
		}
	} else {
		out = append(out, "")
	}
	return out
}

func IsDefaultGroupDomain(domain string) bool {
	return StringsListContainItem(domain, DEFAULT_DOMAIN_NAMES, true)
}

// question to string
func QToString(q dnsmessage.Question) string {
	b := make([]byte, q.Name.Length+2)
	for i := 0; i < int(q.Name.Length); i++ {
		b[i] = q.Name.Data[i]
	}
	b[q.Name.Length] = uint8(q.Type >> 8)
	b[q.Name.Length+1] = uint8(q.Type)

	return string(b)
}

// resource name and type to string
func NtToString(rName dnsmessage.Name, rType dnsmessage.Type) string {
	b := make([]byte, rName.Length+2)
	for i := 0; i < int(rName.Length); i++ {
		b[i] = rName.Data[i]
	}
	b[rName.Length] = uint8(rType >> 8)
	b[rName.Length+1] = uint8(rType)

	return string(b)
}

// resource to string
func RToString(r dnsmessage.Resource) string {
	var sb strings.Builder
	sb.Write(r.Header.Name.Data[:])
	sb.WriteString(r.Header.Type.String())
	sb.WriteString(r.Body.GoString())

	return sb.String()
}

// packet to string
func PToString(p model.Packet) string {
	return fmt.Sprint(p.Message.ID)
}

func UpdAddrToString(a1 net.UDPAddr) string {
	return fmt.Sprintf("%s-%v-%s", a1.IP.String(), a1.Port, a1.Zone)
}

func ResourceToString(a1 dnsmessage.Resource) string {
	return fmt.Sprintf("%s-%s", a1.Header.GoString(), a1.Body.GoString())
}

func RemoveDuplicatesInResourceList(l []dnsmessage.Resource) []dnsmessage.Resource {
	keys := make(map[string]bool)
	var out = make([]dnsmessage.Resource, 0)
	for _, resource := range l {
		addrStr := ResourceToString(resource)
		if _, ok := keys[addrStr]; !ok {
			keys[addrStr] = true
			out = append(out, resource)
		}
	}
	return out
}

func RemoveDuplicatesInUpdAddrList(l []net.UDPAddr) []net.UDPAddr {
	keys := make(map[string]bool)
	var out = make([]net.UDPAddr, 0)
	for _, addr := range l {
		addrStr := UpdAddrToString(addr)
		if _, ok := keys[addrStr]; !ok {
			keys[addrStr] = true
			out = append(out, addr)
		}
	}
	return out
}
