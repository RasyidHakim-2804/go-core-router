package core

type MiddlewareAndStatus struct {
	Middleware MiddlewareInterface
	status     bool
}

func generateSliceMiddlewareAndStatus(slice []MiddlewareAndStatus, middleware MiddlewareInterface, status bool) []MiddlewareAndStatus {
	notExists := true

	for i := range slice {
		if slice[i].Middleware == middleware {
			slice[i].status = status
			notExists = false
			break
		}
	}

	if notExists {
		slice = append(slice, MiddlewareAndStatus{middleware, status})
	}

	return slice
}
