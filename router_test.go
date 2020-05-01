package messaging_test

import (
	"fmt"
	"testing"

	. "github.com/jjeffcaii/rsocket-messaging-go"
	"github.com/stretchr/testify/assert"
)

func TestRouter_Route(t *testing.T) {
	router := NewRouter()
	err := router.Route("students.{id}", func(c *RouteContext) error {
		id, _ := c.Variable("id")
		fmt.Println("got id:", id)
		assert.Equal(t, "2020", id, "bad result")
		return nil
	})
	assert.NoError(t, err, "bind route failed")
	err = router.Fire("students.2020")
	assert.NoError(t, err, "fire failed")
}
