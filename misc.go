package messaging

import "github.com/jjeffcaii/rsocket-messaging-go/internal"

func RegisterCodec(mimeType string, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) error {
	return internal.RegisterCodec(mimeType, marshal, unmarshal)
}
