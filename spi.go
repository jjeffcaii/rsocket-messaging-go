package messaging

import (
	"io"

	"github.com/jjeffcaii/rsocket-messaging-go/internal"
)

type Mono = internal.Mono

type Requester interface {
	io.Closer
	Route(route string, args ...interface{}) RequestSpec
}

type RequestSpec interface {
	Metadata(metadata interface{}, mimeType string) RequestSpec
	Data(data interface{}) RequestSpec
	RetrieveMono() Mono
}

func RegisterEncoder(mimeType string, encoder func(interface{}) ([]byte, error)) {
	internal.RegisterEncoder(mimeType, encoder)
}
