package render

import (
	"encoding/xml"
	"net/http"
)

type XML struct {
	Data any
}

func (x *XML) Render(w http.ResponseWriter) error {
	x.WriteContentType(w)
	err := xml.NewEncoder(w).Encode(x.Data)
	return err
}
func (x *XML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/xml; charset=utf-8")
}
