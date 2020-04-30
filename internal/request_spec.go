package internal

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/jjeffcaii/rsocket-messaging-go/spi"
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
	RegisterEncoder(extension.TextPlain.String(), func(v interface{}) ([]byte, error) {
		switch vv := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return []byte(fmt.Sprintf("%d", vv)), nil
		case float32, float64:
			return []byte(fmt.Sprintf("%f", vv)), nil
		case []byte:
			return vv, nil
		case string:
			return []byte(vv), nil
		case fmt.Stringer:
			return []byte(vv.String()), nil
		default:
			return []byte(fmt.Sprintf("%v", vv)), nil
		}
	})
}

type requestSpec struct {
	parent *requester
	m      []Writeable
	d      func() ([]byte, error)
}

func (p *requestSpec) Metadata(metadata interface{}, mimeType string) spi.RequestSpec {
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

func (p *requestSpec) Data(data interface{}) spi.RequestSpec {
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

func (p *requestSpec) mkRequest() (payload.Payload, error) {
	var (
		data     []byte
		metadata []byte
	)
	if p.m != nil {
		bf := bytes.Buffer{}
		for _, writeable := range p.m {
			if err := writeable(&bf); err != nil {
				return nil, err
			}
		}
		metadata = bf.Bytes()
	}
	if p.d != nil {
		var err error
		d, err := p.d()
		if err != nil {
			return nil, err
		}
		data = d
	}
	return payload.New(data, metadata), nil
}

func (p *requestSpec) Retrieve() error {
	req, err := p.mkRequest()
	if err != nil {
		return err
	}
	p.parent.socket.FireAndForget(req)
	return nil
}

func (p *requestSpec) RetrieveMono() spi.Mono {
	req, err := p.mkRequest()
	if err != nil {
		return NewMonoWithError(err)
	}
	res := p.parent.socket.RequestResponse(req)
	return NewMonoWithDecoder(res, p.decode)
}

func (p *requestSpec) decode(raw []byte, v interface{}) error {
	// TODO: support decoders
	return json.Unmarshal(raw, v)
}

func (p *requestSpec) RetrieveFlux() spi.Flux {
	req, err := p.mkRequest()
	if err != nil {
		return NewFluxWithError(err)
	}
	origin := p.parent.socket.RequestStream(req)
	return NewFluxWithDecoder(origin, p.decode)
}

func RegisterEncoder(mimeType string, encoder FnEncode) {
	_defaultEncodes[mimeType] = encoder
}

func LoadEncoder(mimeType string) (FnEncode, bool) {
	found, ok := _defaultEncodes[mimeType]
	return found, ok
}
