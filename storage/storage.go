package storage

import (
	"fmt"
	"io"
	"net/url"
)

type Backend interface {
	New(name string) (Record, error)
}

type Record interface {
	io.Writer
	io.Closer
}

// Parse the given URI and return an appropriate writer to store the written content.
func Parse(uri string) (Backend, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "file":
		return NewFileBackend(u.Path)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}
