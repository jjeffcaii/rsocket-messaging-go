package messaging_test

import (
	"context"
	"encoding/json"
	"testing"

	messaging "github.com/jjeffcaii/rsocket-messaging-go"
	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {
	requester, err := messaging.Builder().Build()
	assert.NoError(t, err, "build requester failed")

	result := struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}{}

	err = requester.
		Route("create.resource.v1.%d", 1234).
		Metadata("some text", "text/plain").
		Data(struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Birth string `json:"birth"`
		}{
			ID:    77778888,
			Name:  "Foobar",
			Birth: "2020-04-28",
		}).
		RetrieveMono().
		BlockTo(context.Background(), &result)
	assert.NoError(t, err, "request failed")
}
