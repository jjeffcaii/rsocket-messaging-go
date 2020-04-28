package messaging_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jjeffcaii/rsocket-messaging-go"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/stretchr/testify/require"
)

type AuthInfo struct {
	App    string `json:"app"`
	Access string `json:"access"`
}

type Student struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Birth string
}

func TestRequestSpec_Data(t *testing.T) {
	a := AuthInfo{
		App:    "test_app",
		Access: "test_access",
	}
	result := Student{}
	sk, err := rsocket.Connect().
		DataMimeType(extension.ApplicationJSON.String()).
		MetadataMimeType(extension.MessageCompositeMetadata.String()).
		Transport("tcp://127.0.0.1:7878").
		Start(context.Background())
	require.NoError(t, err)

	err = messaging.NewRequester(sk).
		Route("student.v1").
		Metadata(a, extension.ApplicationJSON.String()).
		Data(1).
		RetrieveMono().
		BlockTo(context.Background(), &result)
	require.NoError(t, err)
	fmt.Printf("%+v\n", result)
}
