package routers

import (
	"github.com/bingo"
	"github.com/bingo/demo/controllers"
	"github.com/bingo/demo/controllers/api"
)

func init() {
	// 固定路由
	bingo.Router.AddFixed("/demo", &controllers.DemoController{}, "IndexAction")
	bingo.Router.AddFixed("/demo/demo1", &controllers.DemoController{}, "Demo1Action")
	bingo.Router.AddFixed("/v2.1/demo/demo2", &controllers.DemoController{}, "Demo2Action")

	// 正则路由
	bingo.Router.AddRegular("^/$", &controllers.DemoController{}, "IndexAction")
	bingo.Router.AddRegular("^/index$", &controllers.DemoController{}, "DemoAction")

	// 自动路由
	bingo.Router.AddAuto(&controllers.DemoController{})
	bingo.Router.AddAuto(&api.UserController{}, true, "v2.1")

	// 脚本路由
	bingo.Router.AddShell("index", controllers.IndexAction,)
	bingo.Router.AddShell("cache", controllers.CacheAction)
}
