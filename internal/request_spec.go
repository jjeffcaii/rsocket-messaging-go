package internal

import (
	"bytes"
	"io"

	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

type requestSpec struct {
	parent *requester
	m      []writeable
	d      func() ([]byte, error)
}

func (p *requestSpec) Metadata(metadata interface{}, mimeType string) spi.RequestSpec {
	p.m = append(p.m, func(writer io.Writer) (err error) {
		enc, _, err := p.parent.getCodec()
		if err != nil {
			err = errors.Wrap(err, "encode metadata failed")
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
		enc, _, err := p.parent.getCodec()
		if err != nil {
			err = errors.Wrap(err, "encode data failed")
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
	_, dec, err := p.parent.getCodec()
	if err != nil {
		return err
	}
	return dec(raw, v)
}

func (p *requestSpec) RetrieveFlux() spi.Flux {
	req, err := p.mkRequest()
	if err != nil {
		return NewFluxWithError(err)
	}
	origin := p.parent.socket.RequestStream(req)
	return NewFluxWithDecoder(origin, p.decode)
}
