package internal_test

import (
	"testing"
	"time"

	. "github.com/jjeffcaii/rsocket-messaging-go/internal"
	"github.com/stretchr/testify/assert"
)

type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

func TestMarshalWithMimeType(t *testing.T) {
	msg := Message{
		ID:        1234,
		Content:   "foobar",
		CreatedAt: time.Now(),
	}
	b, err := MarshalWithMimeType("foobar", "application/json")
	assert.NoError(t, err, "marshal failed")
	assert.Equal(t, `"foobar"`, string(b), "bad result")
	b, err = MarshalWithMimeType("foobar", "unknown_mime_type")
	assert.NoError(t, err, "marshal failed")
	assert.Equal(t, "foobar", string(b), "bad result")

	b, err = MarshalWithMimeType(msg, "unknown_mime_type")
	assert.Error(t, err, "should marshal fail")
}
