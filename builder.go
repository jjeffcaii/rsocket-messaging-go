package messaging

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/jjeffcaii/rsocket-messaging-go/internal"
	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

type fnRawMake = func(io.Writer) error

type RequestBuilder struct {
	setupMeta    []fnRawMake
	setupData    interface{}
	dataMimeType string
	tpUrl        string
	tpOpts       []rsocket.TransportOpts
}

func (b *RequestBuilder) ConnectTCP(host string, port int, opts ...rsocket.TransportOpts) *RequestBuilder {
	b.tpUrl = fmt.Sprintf("tcp://%s:%d", host, port)
	b.tpOpts = append(b.tpOpts, opts...)
	return b
}

func (b *RequestBuilder) Build(ctx context.Context) (requester spi.Requester, err error) {
	var metadata []byte
	var data []byte
	enc, ok := internal.LoadEncoder(b.dataMimeType)
	if !ok {
		err = errors.Errorf("no encoder for mime type %s", b.dataMimeType)
		return
	}
	data, err = enc(b.setupData)
	if err != nil {
		return
	}
	if len(b.setupMeta) > 0 {
		bf := bytes.Buffer{}
		for _, it := range b.setupMeta {
			err = it(&bf)
			if err != nil {
				return
			}
		}
		metadata = bf.Bytes()
	}

	setup := payload.New(data, metadata)
	rs, err := rsocket.Connect().
		MetadataMimeType(extension.MessageCompositeMetadata.String()).
		DataMimeType(b.dataMimeType).
		SetupPayload(setup).
		Transport(b.tpUrl, b.tpOpts...).
		Start(ctx)
	if err != nil {
		return
	}
	requester = internal.NewRequester(rs, b.dataMimeType)
	return
}

func (b *RequestBuilder) SetupRoute(route string, args ...interface{}) *RequestBuilder {
	b.setupMeta = append(b.setupMeta, func(writer io.Writer) (err error) {
		r, err := internal.MkString(route, args...)
		if err != nil {
			return
		}
		raw, err := extension.EncodeRouting(r)
		if err != nil {
			return
		}
		_, err = extension.NewCompositeMetadata(extension.MessageRouting.String(), raw).WriteTo(writer)
		return
	})
	return b
}

func (b *RequestBuilder) SetupMetadata(metadata interface{}, mimeType string) *RequestBuilder {
	b.setupMeta = append(b.setupMeta, func(writer io.Writer) (err error) {
		enc, ok := internal.LoadEncoder(mimeType)
		if !ok {
			err = errors.Errorf("no encoder for mime type %s", mimeType)
			return
		}
		b, err := enc(metadata)
		if err != nil {
			return
		}
		_, err = extension.NewCompositeMetadata(mimeType, b).WriteTo(writer)
		return
	})
	return b
}

func (b *RequestBuilder) DataMimeType(mimeType string) *RequestBuilder {
	b.dataMimeType = mimeType
	return b
}

func (b *RequestBuilder) SetupData(data interface{}) *RequestBuilder {
	b.setupData = data
	return b
}

func Builder() *RequestBuilder {
	return &RequestBuilder{
		dataMimeType: extension.ApplicationJSON.String(),
	}
}
