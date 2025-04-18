package hsp

import (
	"fmt"
	"strings"
)

type Adddress struct {
	Host string
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

	return &Adddress{
		Host: addr,
		Route: route,
	}, nil
}

func (a *Adddress) String() string {
	return fmt.Sprintf("%s:%s", a.Host, HSP_PORT)
}

