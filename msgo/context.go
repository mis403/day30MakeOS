package msgo

import (
	"fmt"
	"github.com/mis403/msgo/render"
	"html/template"
	"net/http"
	"net/url"
)

type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	engine *Engine
}

func (c *Context) HTML(status int, html string) error {

	return c.Render(status, &render.HTML{
		Data:       html,
		IsTemplate: false,
	})
}

func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	// 设置响应头中的 Content-Type 字段为 text/html; charset=utf-8，表示响应的内容类型为 HTML，并且使用 UTF-8 字符编码。
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 创建一个新的模板，并将其赋值给变量 t。模板的名字由传入的 name 参数指定。
	t := template.New(name)

	// 解析指定模式（pattern 参数）的所有文件，并将解析的模板存储在 t 中。
	// 解析的模板可以用来执行后续的渲染操作。同时，它还返回一个 error 类型的值 err，表示解析过程中是否出现了错误。
	t, err := t.ParseGlob(pattern)

	// 如果在解析模板的过程中出现了错误，则直接返回该错误。
	if err != nil {
		return err
	}

	// 执行模板 t，并将结果写入到 c.W（c 的 W 字段）中。这里使用了模板中的 data 数据进行渲染。
	t.Execute(c.W, data)

	// 最后，返回可能在模板执行过程中出现的错误，如果执行成功，则返回的错误值为 nil。
	return err
}

func (c *Context) Template(name string, data any) error {
	return c.Render(http.StatusOK, &render.HTML{
		Data:       data,
		Name:       name,
		Template:   c.engine.HTMLRender.Template,
		IsTemplate: true,
	})
}
func (c *Context) JSON(status int, data any) error {
	// 设置响应头中的 Content-Type 字段为 text/html; charset=utf-8，表示响应的内容类型为 HTML，并且使用 UTF-8 字符编码。
	c.W.Header().Set("Content-Type", "text/json; charset=utf-8")
	c.W.WriteHeader(status)

	return c.Render(status, &render.JSON{Data: data})
}
func (c *Context) XML(status int, data any) error {

	return c.Render(status, &render.XML{
		Data: data,
	})
}
func (c *Context) FileAttachment(filepath, filename string) {

	//判断文件格式
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// 从文件系统下载  filepath是相对于文件系统路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		fmt.Println(old) //old 为url
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

// 重定向
func (c *Context) Redirect(status int, location string) {
	c.Render(status, &render.Redirect{
		Code:     status,
		Request:  c.R,
		Location: location,
	})
}

func (c *Context) String(status int, format string, values ...any) error {

	return c.Render(status, &render.String{
		Format: format,
		Data:   values,
	})
}

func (c *Context) Render(status int, r render.Render) error {
	err := r.Render(c.W)
	c.W.WriteHeader(status)
	return err
}
