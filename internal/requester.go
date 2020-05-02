package internal

import (
	"fmt"

	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
)

type requester struct {
	dataMimeType string
	socket       rsocket.RSocket
}

func (p *requester) Route(route string, args ...interface{}) spi.RequestSpec {
	return &requestSpec{
		parent: p,
		m: []func(*extension.CompositeMetadataBuilder) error{
			func(builder *extension.CompositeMetadataBuilder) (err error) {
				b, err := extension.EncodeRouting(fmt.Sprintf(route, args...))
				if err != nil {
					return
				}
				builder.PushWellKnown(extension.MessageRouting, b)
				return
			},
		},
	}
}

func (p *requester) Close() (err error) {
	if c, ok := p.socket.(rsocket.CloseableRSocket); ok {
		err = c.Close()
	}
	return
}

func (p *requester) Unmarshal(raw []byte, v interface{}) error {
	return UnmarshalWithMimeType(raw, v, p.dataMimeType)
}

func (p *requester) Marshal(v interface{}) ([]byte, error) {
	return MarshalWithMimeType(v, p.dataMimeType)
}

func NewRequester(socket rsocket.RSocket, dataMimeType string) *requester {
	return &requester{
		dataMimeType: dataMimeType,
		socket:       socket,
	}
}
