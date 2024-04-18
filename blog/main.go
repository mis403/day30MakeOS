package main

import (
	"fmt"
	"github.com/mis403/msgo"
)

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
	g.Get("/hello", func(ctx *msgo.Context) {
		fmt.Fprintf(ctx.W, "%s GET这是一个测试", "www.baidu.com")
	})
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

	engine.Run()
}