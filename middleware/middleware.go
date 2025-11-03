package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func Stack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			next = xs[i](next)
		}

		return next
	}
}

func With(mw Middleware, h http.HandlerFunc) http.Handler {
	return mw(h)
}
