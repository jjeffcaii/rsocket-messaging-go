package messaging_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jjeffcaii/rsocket-messaging-go"
	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/stretchr/testify/assert"
)

type Student struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Birth string `json:"birth"`
}

func (s Student) String() string {
	return fmt.Sprintf("Student{ID=%d,Name=%s,Birth=%s}", s.ID, s.Name, s.Birth)
}

func TestSuite(t *testing.T) {
	// see: StudentController.java
	requester, err := messaging.Builder().
		DataMimeType("application/json").
		ConnectTCP("127.0.0.1", 7878).
		Build(context.Background())
	assert.NoError(t, err, "connect failed")
	defer requester.Close()

	testRetrieve(t, requester)
	testRetrieveMono(t, requester)
	testRetrieveFluxChan(t, requester)
	testRetrieveFluxAsSlice(t, requester)
}

func testRetrieve(t *testing.T, requester spi.Requester) {
	err := requester.Route("student.v1.noop.%s", "hello").
		Metadata("test unknown mime type", "message/x.rsocket.authentication.bearer.v0").
		Retrieve()
	assert.NoError(t, err, "request failed")
}

func testRetrieveMono(t *testing.T, requester spi.Requester) {
	result := struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{}

	student := Student{
		Name:  "Foobar",
		Birth: "2020-04-28",
	}
	err := requester.
		Route("student.v1.upsert").
		Data(student).
		RetrieveMono().
		BlockTo(context.Background(), &result)
	assert.NoError(t, err, "request failed")
	fmt.Printf("result: %+v\n", result)
}

func testRetrieveFluxChan(t *testing.T, requester spi.Requester) {
	students := make(chan Student)
	go func() {
		for next := range students {
			fmt.Println("next:", next)
		}
	}()
	err := requester.Route("students.v1").RetrieveFlux().BlockToChan(context.Background(), students)
	close(students)
	assert.NoError(t, err, "retrieve flux failed")
}

func testRetrieveFluxAsSlice(t *testing.T, requester spi.Requester) {
	students := make([]Student, 0)
	err := requester.Route("students.v1").RetrieveFlux().BlockToSlice(context.Background(), &students)
	assert.NoError(t, err, "retrieve flux failed")
	for _, next := range students {
		fmt.Println(next)
	}
}
