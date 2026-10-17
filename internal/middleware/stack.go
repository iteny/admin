package middleware

import "runtime/debug"

func stack() string {
	return string(debug.Stack())
}
