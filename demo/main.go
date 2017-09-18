// demo的web入口
//   变更历史
//     2017-02-06  lixiaoya  新建
package main

import (
	"legitlab.letv.cn/uc_tp/goweb"
	_ "legitlab.letv.cn/uc_tp/goweb/demo/routers"
	_ "legitlab.letv.cn/uc_tp/goweb/session/memcache"
)

func main() {
	goweb.ObjApp.Run()
}
