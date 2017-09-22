// demo的web入口
//   变更历史
//     2017-02-06  lixiaoya  新建
package main

import (
	"github.com/bingo"
	_ "github.com/bingo/demo/routers"
	_ "github.com/bingo/session/memcache"
)

func main() {
	bingo.ObjApp.Run()
}
