package hsp

import (
	"fmt"
	"strings"
)

type Adddress struct {
	Host  string
	Port string
	Route string
}

func ParseAddress(address string) (*Adddress, error) {
	parts := strings.SplitN(address, "/", 2)

	var route string
	if len(parts) == 1 {
		route = "/"
	} else if len(parts) > 1 {
		route = "/" + strings.Join(parts[1:], "")
	} else {
		return nil, fmt.Errorf("Failed to parse address: %s", address)
	}

	addr := parts[0]

	port := HSP_PORT

	if strings.Contains(addr, ":") {
		p := strings.Split(addr, ":")
		if len(p) >= 2 {
			port = p[len(p) - 1]
			addr = p[0]
		}
	}

	return &Adddress{
		Host:  addr,
		Port: port,
		Route: route,
	}, nil
}

func (a *Adddress) String() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}
