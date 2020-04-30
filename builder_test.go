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

var requester spi.Requester

func init() {
	// see: StudentController.java
	r, err := messaging.Builder().
		DataMimeType("application/json").
		ConnectTCP("127.0.0.1", 7878).
		Build(context.Background())
	if err != nil {
		panic(err)
	}
	requester = r
}

func TestRetrieve(t *testing.T) {
	defer requester.Close()

	err := requester.Route("student.v1.noop.%s", "hello").Retrieve()
	assert.NoError(t, err, "request failed")
}

func TestRetrieveMono(t *testing.T) {
	defer requester.Close()
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
		Metadata([]string{}, "application/json").
		Data(student).
		RetrieveMono().
		BlockTo(context.Background(), &result)
	assert.NoError(t, err, "request failed")
	fmt.Printf("result: %+v\n", result)
}

func TestRetrieveFlux(t *testing.T) {
	defer requester.Close()
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

func TestRetrieveFluxAsSlice(t *testing.T) {
	defer requester.Close()
	students := make([]Student, 0)
	err := requester.Route("students.v1").RetrieveFlux().BlockToSlice(context.Background(), &students)
	assert.NoError(t, err, "retrieve flux failed")
	for _, next := range students {
		fmt.Println(next)
	}
}
