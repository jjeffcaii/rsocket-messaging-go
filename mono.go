package messaging

import (
	"context"

	"github.com/rsocket/rsocket-go/rx/mono"
)

type Mono interface {
	mono.Mono
	BlockTo(ctx context.Context, to interface{}) error
}

type mustErrorMono struct {
	mono.Mono
}

func (m mustErrorMono) BlockTo(ctx context.Context, _ interface{}) error {
	_, err := m.Block(ctx)
	return err
}

type extraMono struct {
	mono.Mono
	dec func([]byte, interface{}) error
}

func (e *extraMono) BlockTo(ctx context.Context, to interface{}) (err error) {
	pa, err := e.Block(ctx)
	if err != nil {
		return
	}
	err = e.dec(pa.Data(), to)
	return
}

func withError(err error) Mono {
	return mustErrorMono{
		Mono: mono.Error(err),
	}
}

func newExtraMono(origin mono.Mono, decode FnDecode) Mono {
	return &extraMono{
		Mono: origin,
		dec:  decode,
	}
}
