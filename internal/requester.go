package internal

import (
	"fmt"
	"io"

	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
)

type requester struct {
	encoders     map[string]FnEncode
	dataMimeType string
	socket       rsocket.RSocket
}

func (p *requester) Route(route string, args ...interface{}) spi.RequestSpec {
	return &requestSpec{
		parent: p,
		m: []Writeable{
			func(writer io.Writer) (err error) {
				b, err := extension.EncodeRouting(fmt.Sprintf(route, args...))
				if err != nil {
					return
				}
				_, err = extension.NewCompositeMetadata(extension.MessageRouting.String(), b).WriteTo(writer)
				if err != nil {
					return
				}
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

func (p requester) getDataEncoder() (FnEncode, bool) {
	return p.getEncoder(p.dataMimeType)
}

func (p requester) getEncoder(mimeType string) (enc FnEncode, ok bool) {
	enc, ok = p.encoders[mimeType]
	return
}

func NewRequester(socket rsocket.RSocket, dataMimeType string) *requester {
	encoders := make(map[string]FnEncode)
	for k, v := range _defaultEncodes {
		encoders[k] = v
	}
	return &requester{
		encoders:     encoders,
		dataMimeType: dataMimeType,
		socket:       socket,
	}
}
