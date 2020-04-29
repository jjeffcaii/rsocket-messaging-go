package messaging

import "github.com/jjeffcaii/rsocket-messaging-go/internal"

func RegisterEncoder(mimeType string, encoder func(interface{}) ([]byte, error)) {
	internal.RegisterEncoder(mimeType, encoder)
}
