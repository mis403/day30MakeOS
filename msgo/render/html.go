package render

import (
	"github.com/mis403/msgo/internal/bytesconv"
	"html/template"
	"net/http"
)

type HTMLRender struct {
	Template *template.Template
}
type HTML struct {
	Data       any
	Name       string
	Template   *template.Template
	IsTemplate bool
}

func (h *HTML) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	if h.IsTemplate {
		err := h.Template.ExecuteTemplate(w, h.Name, h.Data)
		return err
	}
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}
func (h *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
