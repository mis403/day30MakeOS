package msgo

import "net/http"

type Context struct {
	W http.ResponseWriter
	R *http.Request
}
