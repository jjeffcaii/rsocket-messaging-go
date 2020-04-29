package internal

import (
	"fmt"

	"github.com/pkg/errors"
)

var errMakeString = errors.New("make string failed")

func MkString(format string, args ...interface{}) (str string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errMakeString
		}
	}()
	str = fmt.Sprintf(format, args...)
	return
}
