package messaging

import (
	"errors"

	"github.com/jjeffcaii/rsocket-messaging-go/internal"
)

// TODO: responder-side router support

type RouteHandler = func(*RouteContext) error

type Router struct {
	routers *internal.PathTrie
}

type RouteContext struct {
	v *internal.PathVariables
}

func (c RouteContext) Variable(name string) (string, bool) {
	return c.v.Get(name)
}

func (c RouteContext) VariableOrDefault(name string, defaultValue string) string {
	return c.v.GetOrDefault(name, defaultValue)
}

func (c RouteContext) VariableOrCompute(name string, compute func() string) string {
	return c.v.GetOrCompute(name, compute)
}

func (r *Router) Route(path string, handler RouteHandler) (err error) {
	return r.routers.AddPath(path, handler)
}

func (r *Router) Fire(path string) error {
	v, h, ok := r.routers.Find(path)
	if !ok {
		return errors.New("no router")
	}
	if h == nil {
		return errors.New("no handler")
	}
	c := &RouteContext{
		v: v,
	}
	return h.(RouteHandler)(c)
}

func NewRouter() *Router {
	return &Router{
		routers: internal.NewPathTrie(),
	}
}
