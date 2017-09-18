package routers

import (
	"legitlab.letv.cn/uc_tp/goweb"
	"legitlab.letv.cn/uc_tp/goweb/demo/controllers"
	"legitlab.letv.cn/uc_tp/goweb/demo/controllers/api"
)

func init() {
	// 固定路由
	goweb.Router.AddFixed("/demo", &controllers.DemoController{}, "IndexAction")
	goweb.Router.AddFixed("/demo/demo1", &controllers.DemoController{}, "Demo1Action")
	goweb.Router.AddFixed("/v2.1/demo/demo2", &controllers.DemoController{}, "Demo2Action")

	// 正则路由
	goweb.Router.AddRegular("^/$", &controllers.DemoController{}, "IndexAction")
	goweb.Router.AddRegular("^/index$", &controllers.DemoController{}, "DemoAction")

	// 自动路由
	goweb.Router.AddAuto(&controllers.DemoController{})
	goweb.Router.AddAuto(&api.UserController{}, true, "v2.1")

	// 脚本路由
	goweb.Router.AddShell("index", controllers.IndexAction,)
	goweb.Router.AddShell("cache", controllers.CacheAction)
}
