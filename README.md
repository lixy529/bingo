bingo
======

一个golang的web开发框架，实际项目中已使用，供大家一起学习优化。

框架的使用可以参考demo代码

main函数调用bingo.ObjApp.Run()启动服务

如果未传入配置文件，默认为$APPROOT/config/app.conf


install
-------

go get https://github.com/lixy529/bingo

demo
------

demo代码

环境变量里设置APPROOT为demo的目录，就可运行demo

### 环境变量配置demo实例：

export PRJROOT=$HOME/goyard

export GOPATH=$PRJROOT

export GOBIN=$PRJROOT/bin

export PATH=$PATH:$GOBIN

export APPROOT=$HOME/goyard/src/demo

export APPCONFIG=app.conf # 如果配置从环境变量读，未配置默认为app.conf

### 编译

go install demo

会在$PRJROOT/bin下生成demo文件

### 运行

$PRJROOT/bin/demo
