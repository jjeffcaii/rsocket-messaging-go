package internal

import (
	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

type requestSpec struct {
	parent *requester
	m      []func(*extension.CompositeMetadataBuilder) error
	d      func() ([]byte, error)
}

func (p *requestSpec) Metadata(metadata interface{}, mimeType string) spi.RequestSpec {
	p.m = append(p.m, func(builder *extension.CompositeMetadataBuilder) (err error) {
		b, err := MarshalWithMimeType(metadata, mimeType)
		if err != nil {
			err = errors.Wrap(err, "encode metadata failed")
			return
		}
		builder.Push(mimeType, b)
		return
	})
	return p
}

func (p *requestSpec) Data(data interface{}) spi.RequestSpec {
	p.d = func() (raw []byte, err error) {
		return p.parent.Marshal(data)
	}
	return p
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
	return NewMonoWithDecoder(res, p.parent.Unmarshal)
}

func (p *requestSpec) mkRequest() (payload.Payload, error) {
	var (
		data     []byte
		metadata []byte
	)
	if p.m != nil {
		bu := extension.NewCompositeMetadataBuilder()
		for _, it := range p.m {
			if err := it(bu); err != nil {
				return nil, err
			}
		}
		m, err := bu.Build()
		if err != nil {
			return nil, err

		}
		metadata = m
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

func (p *requestSpec) RetrieveFlux() spi.Flux {
	req, err := p.mkRequest()
	if err != nil {
		return NewFluxWithError(err)
	}
	origin := p.parent.socket.RequestStream(req)
	return NewFluxWithDecoder(origin, p.parent.Unmarshal)
}
