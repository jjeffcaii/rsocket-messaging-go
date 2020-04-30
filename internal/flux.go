package internal

import (
	"context"
	"reflect"

	"github.com/jjeffcaii/rsocket-messaging-go/spi"
	"github.com/pkg/errors"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
)

var (
	errRequireChan     = errors.New("require a chan")
	errRequireSlicePtr = errors.New("require a slice ptr")
)

type simpleFlux struct {
	flux.Flux
	dec func([]byte, interface{}) error
}

func (s simpleFlux) BlockToChan(ctx context.Context, to interface{}) (err error) {
	typ := reflect.TypeOf(to)
	if typ.Kind() != reflect.Chan {
		err = errRequireChan
		return
	}
	elem := typ.Elem()
	ch := reflect.ValueOf(to)
	done := make(chan struct{})
	s.
		DoFinally(func(s rx.SignalType) {
			close(done)
		}).
		Subscribe(ctx, rx.OnNext(func(input payload.Payload) {
			newVal := reflect.New(elem)
			if err := s.dec(input.Data(), newVal.Interface()); err != nil {
				panic(err)
			}
			sending := reflect.ValueOf(newVal.Elem().Interface())
			ch.Send(sending)
		}), rx.OnError(func(e error) {
			err = e
		}))
	<-done
	return
}

func (s simpleFlux) BlockToSlice(ctx context.Context, to interface{}) (err error) {
	typ := reflect.TypeOf(to)
	if typ.Kind() != reflect.Ptr {
		return errRequireSlicePtr
	}
	if typ.Elem().Kind() != reflect.Slice {
		return errRequireSlicePtr
	}
	typ = typ.Elem().Elem()

	valuePtr := reflect.ValueOf(to)
	value := valuePtr.Elem()

	done := make(chan struct{})
	s.
		DoFinally(func(s rx.SignalType) {
			close(done)
		}).
		Subscribe(ctx, rx.OnNext(func(input payload.Payload) {
			newVal := reflect.New(typ)
			if err := s.dec(input.Data(), newVal.Interface()); err != nil {
				panic(err)
			}
			sending := reflect.ValueOf(newVal.Elem().Interface())
			value.Set(reflect.Append(value, sending))
		}), rx.OnError(func(e error) {
			err = e
		}))
	<-done
	return
}

type mustErrFlux struct {
	flux.Flux
}

func (m mustErrFlux) BlockToChan(ctx context.Context, to interface{}) (err error) {
	typ := reflect.TypeOf(to)
	if typ.Kind() != reflect.Chan {
		err = errRequireChan
		return
	}
	_, err = m.BlockLast(ctx)
	return
}

func (m mustErrFlux) BlockToSlice(ctx context.Context, to interface{}) error {
	panic("implement me")
}

func NewFluxWithDecoder(origin flux.Flux, dec func([]byte, interface{}) error) spi.Flux {
	return &simpleFlux{
		Flux: origin,
		dec:  dec,
	}
}

func NewFluxWithError(err error) spi.Flux {
	return &mustErrFlux{
		Flux: flux.Error(err),
	}
}
