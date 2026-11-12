package core

import "reflect"

type MiddlewareAndStatus struct {
	Middleware MiddlewareInterface
	status     bool
}

// NamedMiddleware provides stable identity, including for non-comparable values.
// Use distinct names to register multiple instances of the same middleware type.
type NamedMiddleware struct {
	MiddlewareInterface
	Name string
}

func middlewareKey(m MiddlewareInterface) any {
	if m == nil {
		panic("core: nil middleware")
	}
	v := reflect.ValueOf(m)
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Interface, reflect.Chan:
		if v.IsNil() {
			panic("core: nil middleware")
		}
	}
	if named, ok := m.(NamedMiddleware); ok {
		if named.Name == "" {
			panic("core: empty middleware name")
		}
		middlewareKey(named.MiddlewareInterface)
		return named.Name
	}
	if v.Comparable() {
		return m
	}
	// Non-comparable values use their type as identity; names distinguish instances.
	return v.Type()
}

func generateSliceMiddlewareAndStatus(entries []MiddlewareAndStatus, middleware MiddlewareInterface, status bool) []MiddlewareAndStatus {
	key := middlewareKey(middleware)
	for i := range entries {
		if middlewareKey(entries[i].Middleware) == key {
			entries[i].status = status
			if status {
				entries[i].Middleware = middleware
			}
			return entries
		}
	}
	return append(entries, MiddlewareAndStatus{middleware, status})
}
