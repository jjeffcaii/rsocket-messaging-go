package spi

import (
	"context"
	"io"

	"github.com/rsocket/rsocket-go/rx/mono"
)

type Mono interface {
	mono.Mono
	BlockTo(ctx context.Context, to interface{}) error
}

type Requester interface {
	io.Closer
	Route(route string, args ...interface{}) RequestSpec
}

type RequestSpec interface {
	Metadata(metadata interface{}, mimeType string) RequestSpec
	Data(data interface{}) RequestSpec
	RetrieveMono() Mono
	Retrieve() error
}
