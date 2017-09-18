bingo
======

一个golang的web开发框架，以学习为目的，大家可以一起学习。
框架的使用可以参考demo代码

main函数调用goweb.ObjApp.Run(*appCfg)启动服务

如果未传入配置文件，默认为$APPROOT/config/app.conf


install
-------

go get https://github.com/lixy529/bingo

demo
------

demo代码

环境变量里设置APPROOT为demo的目录，就可运行demo

### 环境变量配置demo实例：

export GOROOT=$HOME/go

export GOARCH=amd64

export GOOS=linux

export PATH=$PATH:$GOROOT/bin



export PRJROOT=$HOME/sso_go

export APPROOT=$HOME/sso_go/src/legitlab.letv.cn/uc_tp/goweb/demo

export APPCONFIG=app.conf # 如果配置从环境变量读，未配置默认为app.conf

export GOPATH=$PRJROOT

export GOBIN=$PRJROOT/bin

export PATH=$PATH:$GOBIN

### 编译

go install legitlab.letv.cn/uc_tp/goweb/demo

会在$PRJROOT/bin下生成demo文件

### 运行

$PRJROOT/bin/demo

或者

$PRJROOT/bin/demo -c /letv/config/app.conf （手动传配置文件）