package msgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mis403/msgo/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

const defaultMaxMemory = 32 << 20

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	IsValidate            bool
}

func (c *Context) QueryMap(key string) (dict map[string]string) {
	dict, _ = c.GetQueryMap(key)
	return
}
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}
func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	//user[id]=1&user[name]=张三
	dicts := make(map[string]string)
	exist := false
	//k =  user[id]  v = 1
	for k, value := range m {
		//判断是否为map参数    i >= 1 说明`[`不在第一个索引 i的值为[的第一个下标，user[id] i的下标为4 k[0:i] 为user
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			//如果[]中有
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = value[0]
			}
		}
	}
	return dicts, exist
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	values, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return values[0]
}
func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}
func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	strings, ok := c.queryCache[key]
	return strings, ok
}
func (c *Context) QueryArray() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

// 从请求中获取参数信息
func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}
func (c *Context) initPostFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.R
		if err := req.ParseMultipartForm(defaultMaxMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	}
}

// 解析文件
func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

// 下载功能提取
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// 下载
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMaxMemory)
	return c.R.MultipartForm, err
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initPostFormCache()
	values, ok := c.formCache[key]
	return values, ok
}
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initPostFormCache()
	return c.get(c.formCache, key)
}
func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetPostFormMap(key)
	return
}
func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
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
	if http.StatusOK != status {
		c.W.WriteHeader(status)
	}

	return err
}

func (c *Context) DealJson(data any) error {
	body := c.R.Body
	if c.R == nil || body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	if c.DisallowUnknownFields {
		//有未知的字段报错
		decoder.DisallowUnknownFields()
	}

	if c.IsValidate {
		err := validateRequireParam(data, decoder)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(data)
		if err != nil {
			return err
		}
	}

	return validate(data)

}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

func validateRequireParam(data any, decoder *json.Decoder) error {
	if data == nil {
		return nil
	}
	valueOf := reflect.ValueOf(data)
	//判断是否为指针类型
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("no ptr type")
	}
	t := valueOf.Elem().Interface()
	of := reflect.ValueOf(t)
	switch of.Kind() {
	case reflect.Struct: //如果是结构体类型，
		return checkParam(of, data, decoder)
	case reflect.Slice, reflect.Array:
		elem := of.Type().Elem()
		if elem.Kind() == reflect.Struct {
			return checkParamSlice(elem, data, decoder)
		}
	default:
		err := decoder.Decode(data)
		if err != nil {
			return err
		}
	}
	return validate(data)
}
func validate(data any) error {
	return Validator.ValidateStruct(data)
}

func checkParamSlice(of reflect.Type, data any, decoder *json.Decoder) error {
	mapData := make([]map[string]interface{}, 0)
	_ = decoder.Decode(&mapData)         //解析为map 根据key 进行比对
	for i := 0; i < of.NumField(); i++ { //判断map的值有没有对应的key
		field := of.Field(i)
		jsomTag := field.Tag.Get("json")
		tag := field.Tag.Get("must")
		for _, v := range mapData {
			value := v[jsomTag]
			if value == nil && tag == "required" {
				return errors.New(fmt.Sprintf("filed [%s] is not exist", jsomTag))
			}
		}

	}
	marshal, _ := json.Marshal(mapData)
	_ = json.Unmarshal(marshal, data)
	return nil
}
func checkParam(of reflect.Value, data any, decoder *json.Decoder) error {
	mapData := make(map[string]interface{})
	_ = decoder.Decode(&mapData)         //解析为map 根据key 进行比对
	for i := 0; i < of.NumField(); i++ { //判断map的值有没有对应的key
		field := of.Type().Field(i)
		jsomTag := field.Tag.Get("json")
		tag := field.Tag.Get("must")

		value := mapData[jsomTag]
		if value == nil && tag == "required" {
			return errors.New(fmt.Sprintf("filed [%s] is not exist", jsomTag))
		}
	}
	marshal, _ := json.Marshal(mapData)
	_ = json.Unmarshal(marshal, data)
	return nil
}
