// demo的shell入口
//   变更历史
//     2017-03-31  lixiaoya  新建
package main

import (
	"flag"
	"github.com/bingo"
	_ "github.com/bingo/cache/memcache"
	_ "github.com/bingo/cache/redis"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/bingo/demo/routers"
)

func main() {
	pattern := flag.String("p", "", "please input pattern")
	flag.Parse()

	bingo.ObjApp.RunShell(*pattern)

	return
}
