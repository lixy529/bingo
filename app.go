// 项目入口类
//   变更历史
//     2017-02-06  lixiaoya  新建
package bingo

import (
	"fmt"
	"github.com/lixy529/bingo/gracefcgi"
	"github.com/lixy529/bingo/gracehttp"
	"github.com/lixy529/bingo/utils"
	"log"
	"os"
	"time"
)

const (
	VERSION = "1.0.0"
	PROD = "prod"
	DEV = "dev"
	MaxMem = 64 << 20 // 64M
)

var (
	ObjApp  *App       // 应用全局对象
	Router  *RouterTab // 路由表
	isShell bool       // 是否以shell形式启动
)

func init() {
	// create application
	ObjApp = NewApp()
	Router = NewRouterTab()
	isShell = false
}

type App struct {
	chanStop chan bool // 关闭服务，shell形式启动使用
}

func NewApp() *App {
	app := &App{
		chanStop: make(chan bool)}
	return app
}

// Run 应用的入口函数
func (app *App) Run() {
	app.beforeRun()

	addr := fmt.Sprintf("%s:%d", AppCfg.ServerCfg.Addr, AppCfg.ServerCfg.Port)
	log.Printf("Start server, addr[%s] pid[%d]>>>", addr, os.Getpid())
	Flogger.Infof("Start server, addr[%s] pid[%d]>>>", addr, os.Getpid())

	if AppCfg.ServerCfg.IsFcgi {
		// fastcgi启动
		Flogger.Info("Server start use fcgi.")
		err := gracefcgi.NewServer(AppCfg.ServerCfg.Addr, AppCfg.ServerCfg.Port, Router, AppCfg.ServerCfg.ShutTimeout).ListenAndServe()
		if err != nil {
			Flogger.Errorf("Start server by fcgi failed. err: %s", err.Error())
			log.Printf("Start server by fcgi failed. err: %s", err.Error())
		}
	} else if AppCfg.ServerCfg.UseGrace {
		// grace启动
		Flogger.Info("Server start use grace.")
		srv := gracehttp.NewServer(addr, Router, AppCfg.ServerCfg.ReqTimeout, AppCfg.ServerCfg.WriteTimeout, AppCfg.ServerCfg.ShutTimeout)
		if AppCfg.ServerCfg.Secure {
			err := srv.ListenAndServeTLS(AppCfg.ServerCfg.CertFile, AppCfg.ServerCfg.KeyFile)
			if err != nil {
				Flogger.Errorf("Start server by https failed. err: %s", err.Error())
				log.Printf("Start server by https failed. err: %s", err.Error())
			}
		} else {
			err := srv.ListenAndServe()
			if err != nil {
				Flogger.Errorf("Start server by http failed. err: %s", err.Error())
				log.Printf("Start server by http failed. err: %s", err.Error())
			}
		}
	} else {
		// 原生模式启动
		Flogger.Info("Server start use native.")
		srv := NewWebServer(addr, Router, AppCfg.ServerCfg.ReqTimeout, AppCfg.ServerCfg.WriteTimeout, AppCfg.ServerCfg.ShutTimeout)
		if AppCfg.ServerCfg.Secure {
			err := srv.ListenAndServeTLS(AppCfg.ServerCfg.CertFile, AppCfg.ServerCfg.KeyFile)
			if err != nil {
				Flogger.Errorf("Start server by https failed. err: %s", err.Error())
				log.Printf("Start server by https failed. err: %s", err.Error())
			}
		} else {
			err := srv.ListenAndServe()
			if err != nil {
				Flogger.Errorf("Start server by http failed. err: %s", err.Error())
				log.Printf("Start server by http failed. err: %s", err.Error())
			}
		}
	}

	Flogger.Infof("Server stop, pid[%d] >>>", os.Getpid())
	app.afterRun()
	log.Printf("Server stop, pid[%d] >>>\n", os.Getpid())
}

// RunShell 以shell模式运行
//   参数
//     pattern: 路由请求路径
//   返回
//
func (app *App) RunShell(pattern string) {
	pid := os.Getpid()
	log.Printf("Start shell server, pid[%d]", pid)
	isShell = true
	app.beforeRun()

	// 捕获信号
	go func() {
		_, sigName := utils.HandleSignals()
		log.Printf("App: Pid %d received %s.\n", pid, sigName)
		app.chanStop <- true
	}()

	f := Router.matchShell(pattern)
	if f == nil {
		log.Println("App: Match shell router failed.")
	} else {
		f()
	}

	app.afterRun()
	log.Printf("App: Stop shell server, pid[%d]", pid)
}

// BeforeRun 运行run前初始函数
func (app *App) beforeRun() {
	// 设置请求的超时时间
	Router.SetReqTimeout(AppCfg.ServerCfg.ReqTimeout)

	for _, f := range inits {
		if err := f(); err != nil {
			Flogger.Errorf("beforeRun: %s", err.Error())
			panic(err)
		}
	}

	return
}

// AfterRun  运行run后销毁函数
func (app *App) afterRun() {
	// 资源释放
	for _, f := range unInits {
		f()
	}

	return
}

// StopSrv 判断是否停止服务
//   参数
//
//   返回
//     false-不停止 true-停止
func (app *App) StopSrv(timeOut int) bool {
	select {
	case <- app.chanStop:
		return true
	case <-time.After(time.Duration(timeOut) * time.Second):
		return false
	}

	return false
}
