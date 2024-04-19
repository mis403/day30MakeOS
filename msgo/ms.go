package msgo

import (
	"fmt"
	"github.com/mis403/msgo/render"
	"html/template"
	"log"
	"net/http"
	"sync"
)

// 定义一个常量表示任意请求方法
const ANY = "ANY"

// HandlerFunc 是一个处理 HTTP 请求的函数类型
type HandlerFunc func(ctx *Context)
type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

// router 结构体表示一个路由器，包含不同组的路由信息
type router struct {
	groups []*routerGroup // 存储不同组的路由信息
}

// handleFuncMap k 为 路由地址，v为嵌套map k为请求方式（any / get） v为对应方法（支持同一个路径下不同的访问/hello get/post访问 ）
type routerGroup struct {
	groupName          string                                 // 组名
	handleFuncMap      map[string]map[string]HandlerFunc      // 请求路径对应的处理函数映射
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc // 请求路径对应的处理的中间件
	handlerMethodMap   map[string][]string                    // 请求路径对应的允许的请求方法映射
	treeNode           *treeNode
	middlewares        []MiddlewareFunc
}

func (r *routerGroup) MiddlewareHandler(middlewareFunc ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewareFunc...)
}

// name: routerName method: requestType
func (r *routerGroup) methodHandler(handlerFunc HandlerFunc, name string, method string, ctx *Context) {
	//前置中间件
	if r.middlewares != nil {
		for _, middleware := range r.middlewares {
			handlerFunc = middleware(handlerFunc)
		}
	}
	//设置路由组级别的中间件
	funcMiddle := r.middlewaresFuncMap[name][method]
	if funcMiddle != nil {
		for _, middlewareFunc := range funcMiddle {
			handlerFunc = middlewareFunc(handlerFunc)
		}
	}
	handlerFunc(ctx)

}

// 初始化一个路由组
func (r *router) Group(name string) *routerGroup {
	// 创建一个新的路由组
	g := &routerGroup{
		groupName:          name,
		handleFuncMap:      make(map[string]map[string]HandlerFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		handlerMethodMap:   make(map[string][]string),
		treeNode: &treeNode{
			name:     "",
			children: make([]*treeNode, 0),
		},
	}
	// 将新的路由组添加到路由器中
	r.groups = append(r.groups, g)
	return g
}

// Engine 结构体表示一个引擎，包含一个路由器
type Engine struct {
	*router
	funcMap    template.FuncMap
	HTMLRender render.HTMLRender
	pool       sync.Pool
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// 加载模板
func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHtmlTemplate(t)

}
func (e *Engine) SetHtmlTemplate(template *template.Template) {
	e.HTMLRender = render.HTMLRender{
		Template: template,
	}
}

// ServeHTTP 方法用于处理 HTTP 请求
func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = writer
	ctx.R = request
	e.httpRequestHandler(ctx, writer, request)
	e.pool.Put(ctx)
}

// handle 方法用于向路由组中添加处理函数，并处理重复添加的情况
func (r *routerGroup) handle(path, method string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	// 检查该路由路径是否已经存在
	_, ok := r.handleFuncMap[path]
	if !ok {
		// 如果不存在，则初始化处理函数映射
		r.handleFuncMap[path] = make(map[string]HandlerFunc)
		r.middlewaresFuncMap[path] = make(map[string][]MiddlewareFunc)
	}
	// 检查该请求方法是否已经存在
	_, ok = r.handleFuncMap[path][method]
	if ok {
		// 如果已经存在，则抛出异常
		panic("有重复的路由")
	}
	// 向路由组中添加处理函数
	r.handleFuncMap[path][method] = handlerFunc

	//向路由组中添加中间件
	r.middlewaresFuncMap[path][method] = append(r.middlewaresFuncMap[path][method], middlewareFunc...)
	//将path放入前缀树
	r.treeNode.Put(path)
}

// Any 方法用于向路由组中添加处理任意请求方法的处理函数
func (r *routerGroup) Any(path string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(path, ANY, handlerFunc, middlewareFunc...)
}

// Get 方法用于向路由组中添加处理 GET 请求方法的处理函数
func (r *routerGroup) Get(path string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(path, http.MethodGet, handlerFunc, middlewareFunc...)
}

// Post 方法用于向路由组中添加处理 POST 请求方法的处理函数
func (r *routerGroup) Post(path string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(path, http.MethodPost, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Delete(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodDelete, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Put(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPut, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Patch(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPatch, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Options(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodOptions, handlerFunc, middlewareFunc...)
}
func (r *routerGroup) Head(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodHead, handlerFunc, middlewareFunc...)
}

// NewEngine 函数用于创建一个新的 Engine 实例
func NewEngine() *Engine {
	engine := &Engine{
		router: &router{},
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

// Run 方法用于运行 HTTP 服务器并监听端口
func (e *Engine) allocateContext() any {

	return &Context{engine: e}
}

// Run 方法用于运行 HTTP 服务器并监听端口
func (e *Engine) Run() {
	// 将 Engine 实例注册为 HTTP 处理器
	http.Handle("/", e)
	// 监听端口并启动 HTTP 服务器
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (e *Engine) httpRequestHandler(ctx *Context, writer http.ResponseWriter, request *http.Request) {
	// 获取请求的方法
	method := request.Method

	// 遍历路由器中的所有路由组  k 为index v为 router
	for _, g := range e.router.groups {
		//request.RequestURI 如果路径上有参数识别不了 换成
		routerName := SubStringLast(request.URL.Path, "/"+g.groupName)
		node := g.treeNode.Get(routerName)
		if node != nil && node.isEnd {

			// 尝试获取对应请求方法的处理函数
			handle, ok := g.handleFuncMap[node.routerName][ANY]
			// 如果找到了 ANY 方法的处理函数，则直接调用
			if ok {
				g.methodHandler(handle, node.routerName, ANY, ctx)
				return
			}
			// 尝试获取对应请求方法的处理函数
			handle, ok = g.handleFuncMap[node.routerName][method]
			// 如果找到了对应请求方法的处理函数，则直接调用
			if ok {
				g.methodHandler(handle, node.routerName, method, ctx)
				return
			}
			// 如果请求方法不匹配，则返回 405 Method Not Allowed 错误
			writer.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(writer, "%s, %s not allowed \n", request.URL, method)
		}

		/*	// 遍历路由组中的每个路由路径 k 为 路由地址，v为嵌套map k为请求方式（any / get） v为对应方法
			for path, methodHandle := range g.handleFuncMap {
				// 构建完整的 URL
				url := "/" + g.groupName + path //组名加上路径 /user/add
				// 如果请求的 URL 与当前路由路径匹配
				if request.RequestURI == url {
					// 创建一个上下文对象

				}
			}*/
	}
	writer.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(writer, "%s  not found \n", request.RequestURI)
}
