package internal

import (
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
)

type requester struct {
	encoders     map[string]FnEncode
	dataMimeType string
	socket       rsocket.RSocket
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

func NewRequester(socket rsocket.RSocket) *requester {
	encoders := make(map[string]FnEncode)
	for k, v := range _defaultEncodes {
		encoders[k] = v
	}
	return &requester{
		encoders:     encoders,
		dataMimeType: extension.ApplicationJSON.String(),
		socket:       socket,
	}
}
