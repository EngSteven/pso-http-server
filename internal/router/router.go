package router

import "github.com/EngSteven/pso-http-server/internal/types"

type Router struct {
	routes map[string]types.HandlerFunc
}

func NewRouter() *Router {
	return &Router{routes: make(map[string]types.HandlerFunc)}
}

func (r *Router) Handle(path string, handler types.HandlerFunc) {
	r.routes[path] = handler
}

func (r *Router) Match(path string) types.HandlerFunc {
	if h, ok := r.routes[path]; ok {
		return h
	}
	return nil
}
