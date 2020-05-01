package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/extension"
)

var (
	errInvalidUnmarshalTarget = errors.New("invalid unmarshal target")
	errRequireStringPtr       = errors.New("require string ptr")
	errNilInterface           = errors.New("cannot unmarshal to nil target")
	_codecs                   = make(map[string]codec)
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

func LoadCodec(mimeType string) (enc FnMarshal, dec FnUnmarshal, ok bool) {
	found, ok := _codecs[mimeType]
	if !ok {
		return
	}
	enc = found.enc
	dec = found.dec
	return
}

func UnmarshalWithMimeType(raw []byte, v interface{}, mimeType string) error {
	_, dec, ok := LoadCodec(mimeType)
	if ok {
		return dec(raw, v)
	}
	// TODO: default unmarshal
	if v == nil {
		return errNilInterface
	}
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		return errInvalidUnmarshalTarget
	}
	value := reflect.ValueOf(v)
	switch typ.Elem().Kind() {
	case reflect.String:
		value.Elem().SetString(string(raw))
	case reflect.Slice:
		if typ.Elem().Elem().Kind() == reflect.Uint8 {
			return errInvalidUnmarshalTarget
		}
		value.Elem().SetBytes(raw)
	default:
		return errInvalidUnmarshalTarget
	}
	return errors.Errorf("cannot unmarshal for mime type %s", mimeType)
}

func MarshalWithMimeType(v interface{}, mimeType string) (raw []byte, err error) {
	enc, _, ok := LoadCodec(mimeType)
	if ok {
		raw, err = enc(v)
		return
	}
	switch vv := v.(type) {
	case []byte:
		raw = vv
	case string:
		raw = []byte(vv)
	case io.Reader:
		raw, err = ioutil.ReadAll(vv)
	default:
		err = errors.Errorf("cannot marshal for mime type %s", mimeType)
	}
	return
}
