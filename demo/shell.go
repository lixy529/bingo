// demo的shell入口
//   变更历史
//     2017-03-31  lixiaoya  新建
package main

import (
	"flag"
	"legitlab.letv.cn/uc_tp/goweb"
	_ "legitlab.letv.cn/uc_tp/goweb/cache/memcache"
	_ "legitlab.letv.cn/uc_tp/goweb/cache/redis"
	_ "legitlab.letv.cn/uc_tp/goweb/db/mysql"
	_ "legitlab.letv.cn/uc_tp/goweb/demo/routers"
)

func main() {
	pattern := flag.String("p", "", "please input pattern")
	flag.Parse()

	goweb.ObjApp.RunShell(*pattern)

	return
}
