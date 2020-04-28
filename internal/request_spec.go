package internal

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/pkg/errors"
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

type requestSpec struct {
	parent *requester
	m      []Writeable
	d      func() ([]byte, error)
}

func (p *requester) Route(route string, args ...interface{}) *requestSpec {
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

func (p *requestSpec) Metadata(metadata interface{}, mimeType string) *requestSpec {
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

func (p *requestSpec) Data(data interface{}) *requestSpec {
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

func (p *requestSpec) RetrieveMono() Mono {
	var (
		data     []byte
		metadata []byte
	)
	if p.m != nil {
		bf := bytes.Buffer{}
		for _, writeable := range p.m {
			if err := writeable(&bf); err != nil {
				return NewMonoWithError(err)
			}
		}
		metadata = bf.Bytes()
	}
	if p.d != nil {
		var err error
		d, err := p.d()
		if err != nil {
			return NewMonoWithError(err)
		}
		data = d
	}

	req := payload.New(data, metadata)
	res := p.parent.socket.RequestResponse(req)
	return NewMonoWithDecoder(res, func(raw []byte, v interface{}) error {
		return json.Unmarshal(raw, v)
	})
}

func RegisterEncoder(mimeType string, encoder FnEncode) {
	_defaultEncodes[mimeType] = encoder
}
