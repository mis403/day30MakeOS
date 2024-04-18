package main

import (
	"fmt"
	"github.com/mis403/msgo"
	"log"
	"net/http"
)

func Log(next msgo.HandlerFunc) msgo.HandlerFunc {
	return func(ctx *msgo.Context) {
		fmt.Println("打印XXX")
		next(ctx)
		fmt.Println("返回XXX")
	}
}

type User struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

func main() {

	/*handlerFunc := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "%s 这tm版本不对啊", "www.baidu.com")
	})

	http.Handle("/hello", handlerFunc)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalln(err)
	}*/
	/*	func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello mszlu.com")
	}*/
	engine := msgo.NewEngine()
	g := engine.Group("user")
	g.MiddlewareHandler(func(next msgo.HandlerFunc) msgo.HandlerFunc {
		return func(ctx *msgo.Context) {
			fmt.Println("pre handler")
			next(ctx)
		}
	})
	g.Get("/hello", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s GET这是一个测试", "www.baidu.com")
	})
	g.Get("/hello/get", func(ctx *msgo.Context) {
		fmt.Println("FuncHandler test")
		fmt.Fprintf(ctx.W, "%s GET这是一个测试", "www.baidu.com")
	}, Log)

	g.Post("/hello/post", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s p这是一个测试", "www.baidu.com")
	})

	g.Post("/pp", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s POST这是一个测试", "www.baidu.com")
	})
	g.Post("/info", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s info", "mszlu.com")
	})
	g.Any("/any", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s any", "mszlu.com")
	})
	g.Get("/get/:id", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s get user info path variable", "mszlu.com")
	})
	g.Get("/htmlTemplateGlob", func(ctx *msgo.Context) {
		err := ctx.HTMLTemplateGlob("index.html", "", "tpl/*.html")
		if err != nil {
			log.Println(err)
		}
	})
	//提前将模板加载到内存
	engine.LoadTemplate("tpl/*.html")
	g.Get("/template", func(ctx *msgo.Context) {
		err := ctx.Template("login.html", "")
		if err != nil {
			log.Println(err)
		}
	})

	g.Get("/json", func(ctx *msgo.Context) {
		user := &User{
			Name: "码神",
		}
		err := ctx.JSON(http.StatusOK, user)
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/xml", func(ctx *msgo.Context) {
		user := &User{
			Name: "码神",
			Age:  100,
		}
		err := ctx.XML(http.StatusOK, user)
		if err != nil {
			log.Println(err)
		}
	})
	g.Get("/download", func(ctx *msgo.Context) {
		ctx.FileAttachment("tpl/text.docx", "aaa.docx")
	})
	g.Get("/fs", func(ctx *msgo.Context) {
		ctx.FileFromFS("text.docx", http.Dir("tpl"))
	})
	g.Get("/toRedirect", func(ctx *msgo.Context) {
		ctx.Redirect(http.StatusFound, "/user/hello")
	})
	g.Get("/string", func(ctx *msgo.Context) {
		ctx.String(http.StatusOK, "%s 是由 %s 制作 \n", "goweb框架", "码神之路")

	})

	engine.Run()
}
