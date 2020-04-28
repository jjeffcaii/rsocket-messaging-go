package messaging

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

var errNoDataEncoder = errors.New("cannot find data encoder")
var _defaultEncodes = make(map[string]FnEncode)

type Writeable = func(io.Writer) error
type FnEncode = func(interface{}) ([]byte, error)
type FnDecode = func([]byte, interface{}) error

func init() {
	RegisterEncoder(extension.ApplicationJSON.String(), json.Marshal)
	RegisterEncoder(extension.ApplicationXML.String(), xml.Marshal)
}

type Requester struct {
	encoders     map[string]FnEncode
	dataMimeType string
	socket       rsocket.RSocket
}

func (p *Requester) Close() (err error) {
	if c, ok := p.socket.(rsocket.CloseableRSocket); ok {
		err = c.Close()
	}
	return
}

type RequestSpec struct {
	parent *Requester
	m      []Writeable
	d      func() ([]byte, error)
}

func NewRequester(socket rsocket.RSocket) *Requester {
	encoders := make(map[string]FnEncode)
	for k, v := range _defaultEncodes {
		encoders[k] = v
	}
	return &Requester{
		encoders:     encoders,
		dataMimeType: extension.ApplicationJSON.String(),
		socket:       socket,
	}
}

func (p Requester) getDataEncoder() (FnEncode, bool) {
	return p.getEncoder(p.dataMimeType)
}

func (p Requester) getEncoder(mimeType string) (enc FnEncode, ok bool) {
	enc, ok = p.encoders[mimeType]
	return
}

func (p *Requester) Route(route string, args ...interface{}) *RequestSpec {
	return &RequestSpec{
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

func (p *RequestSpec) Metadata(metadata interface{}, mimeType string) *RequestSpec {
	p.m = append(p.m, func(writer io.Writer) (err error) {
		enc, ok := p.parent.getEncoder(mimeType)
		if !ok {
			err = errors.Errorf("cannot find encoder for mime type %s", mimeType)
			return
		}
		b, err := enc(metadata)
		if err != nil {
			err = errors.Wrap(err, "encode metadata failed")
			return
		}
		_, err = extension.NewCompositeMetadata(mimeType, b).WriteTo(writer)
		return
	})
	return p
}

func (p *RequestSpec) Data(data interface{}) *RequestSpec {
	p.d = func() (raw []byte, err error) {
		enc, ok := p.parent.getDataEncoder()
		if !ok {
			err = errNoDataEncoder
			return
		}
		return enc(data)
	}
	return p
}

func (p *RequestSpec) RetrieveMono() Mono {
	var (
		data     []byte
		metadata []byte
	)
	if p.m != nil {
		bf := bytes.Buffer{}
		for _, writeable := range p.m {
			if err := writeable(&bf); err != nil {
				return withError(err)
			}
		}
		metadata = bf.Bytes()
	}
	if p.d != nil {
		var err error
		d, err := p.d()
		if err != nil {
			return withError(err)
		}
		data = d
	}

	req := payload.New(data, metadata)
	res := p.parent.socket.RequestResponse(req)
	return newExtraMono(res, func(raw []byte, v interface{}) error {
		return json.Unmarshal(raw, v)
	})
}

func RegisterEncoder(mimeType string, encoder FnEncode) {
	_defaultEncodes[mimeType] = encoder
}
