package messaging_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jjeffcaii/rsocket-messaging-go"
	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	// see: StudentController.java
	requester, err := messaging.Builder().
		DataMimeType("application/json").
		ConnectTCP("127.0.0.1", 7878).
		Build(context.Background())
	assert.NoError(t, err, "build requester failed")
	defer requester.Close()

	result := struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{}

	err = requester.
		Route("student.v1.upsert").
		Metadata([]string{}, "application/json").
		Data(struct {
			Name  string `json:"name"`
			Birth string `json:"birth"`
		}{
			Name:  "Foobar",
			Birth: "2020-04-28",
		}).
		RetrieveMono().
		BlockTo(context.Background(), &result)
	assert.NoError(t, err, "request failed")
	fmt.Printf("result: %+v\n", result)

	err = requester.Route("student.v1.noop.%s", "hello").Retrieve()
	assert.NoError(t, err, "request failed")
	// lang:java
}
