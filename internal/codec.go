package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/extension"
)

var (
	errRequireStringPtr = errors.New("require string ptr")
	_codecs             = make(map[string]codec)
)

type (
	writeable   = func(io.Writer) error
	FnMarshal   = func(interface{}) ([]byte, error)
	FnUnmarshal = func([]byte, interface{}) error
)

type codec struct {
	enc FnMarshal
	dec FnUnmarshal
}

func init() {
	_ = RegisterCodec(extension.ApplicationJSON.String(), json.Marshal, json.Unmarshal)
	_ = RegisterCodec(extension.ApplicationXML.String(), xml.Marshal, xml.Unmarshal)
	_ = RegisterCodec(extension.TextPlain.String(), func(v interface{}) ([]byte, error) {
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
	}, func(bytes []byte, v interface{}) (err error) {
		ptr, ok := v.(*string)
		if !ok {
			err = errRequireStringPtr
			return
		}
		*ptr = string(bytes)
		return
	})
}

func RegisterCodec(mimeType string, encoder FnMarshal, decoder FnUnmarshal) error {
	if encoder == nil {
		return errors.New("encoder is nil")
	}
	if decoder == nil {
		return errors.New("decoder is nil")
	}
	_codecs[mimeType] = codec{
		enc: encoder,
		dec: decoder,
	}
	return nil
}

func LoadCodec(mimeType string) (FnMarshal, FnUnmarshal, error) {
	found, ok := _codecs[mimeType]
	if !ok {
		return nil, nil, errors.Errorf("no such codec for mime type %s", mimeType)
	}
	return found.enc, found.dec, nil
}
