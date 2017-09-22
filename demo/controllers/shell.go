// controller shell
//   变更历史
//     2017-03-21  lixiaoya  新建
package controllers

import (
	"github.com/bingo"
	"github.com/bingo/demo/models"
	"log"
	"time"
)

// IndexAction
func IndexAction() {
	log.Println("start index...")
	for {
		log.Println("run once...")
		time.Sleep(1 * time.Second)

		if bingo.ObjApp.StopSrv(5) {
			break
		}
	}
}

// IndexAction
func CacheAction() {
	log.Println("start cache...")
	for {
		log.Println("run once...")

		cacheName := "memcache"
		m := &models.DemoModel{}
		val, err := m.CacheTest(cacheName)
		if err != nil {
			log.Printf("Cache error, %s" + err.Error())
			return
		}

		log.Printf("val = %s", val)
		time.Sleep(1 * time.Second)

		if bingo.ObjApp.StopSrv(5) {
			break
		}
	}
}
