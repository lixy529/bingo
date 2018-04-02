// 路由相关类
//   变更历史
//     2017-02-07  lixiaoya  新建
package bingo

import (
	"github.com/lixy529/bingo/utils"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"runtime"
)

type ShellFunc func()

// RouterInfo 路由信息
type RouterInfo struct {
	controllerType reflect.Type
	method         string
}

// RouterTab 路由表
type RouterTab struct {
	fixedRouters   map[string]RouterInfo // 固定路由列表
	regularRouters map[string]RouterInfo // 正则路由列表
	autoRouters    map[string]RouterInfo // 自动路由列表
	shellRouters   map[string]ShellFunc  // 脚本路由列表

	maxPathCnt int           // 路由最大路径个数，比如/aa/bb/cc，则值为3
	minPathCnt int           // 路由最小路径个数，不能小于2
	reqTimeout time.Duration // 请求超时时间
}

// NewRouterTab 实例化一个路由表
//   参数
//
//   返回
//     路由表对象
func NewRouterTab() *RouterTab {
	rt := &RouterTab{}
	rt.maxPathCnt = 0
	rt.minPathCnt = 99999
	return rt
}

// SetReqTimeout 设置请求超时时间
//   参数
//     reqTimeout: 请求超时时间
//   返回
//     void
func (rt *RouterTab) SetReqTimeout(reqTimeout time.Duration) {
	if reqTimeout <= 0 {
		rt.reqTimeout = 10
	}
	rt.reqTimeout = reqTimeout
}

// AddFixed 添加固定路由，路径不区分大小，同一路径设置多次，后面会覆盖前面
//   参数
//     pattern: 路由请求路径
//     c:       控制器对象地址
//     method:  控制器方法名
//     args:    支持路径前面拼接此信息，比如路径上带上版本v3.0
//   返回
//     void
func (rt *RouterTab) AddFixed(pattern string, c ControllerInterface, method string, args ...string) {
	reflectVal := reflect.ValueOf(c)
	t := reflect.Indirect(reflectVal).Type()
	routeInfo := RouterInfo{}
	routeInfo.controllerType = t
	routeInfo.method = method
	if rt.fixedRouters == nil {
		rt.fixedRouters = make(map[string]RouterInfo)
	}
	pattern = strings.ToLower(strings.TrimRight(pattern, "/"))
	if pattern == "" {
		pattern = "/"
	}
	if len(args) > 0 {
		ext := strings.Trim(args[0], "/")
		if ext != "" {
			pattern = "/" + strings.ToLower(ext) + pattern
		}
	}
	rt.fixedRouters[pattern] = routeInfo
	cnt := rt.getPathCnt(pattern)
	if rt.maxPathCnt < cnt {
		rt.maxPathCnt = cnt
	}
	if cnt > 1 && rt.minPathCnt > cnt {
		rt.minPathCnt = cnt
	}
}

// AddRegular 添加正则路由，同一路径设置多次，后面会覆盖前面
//   参数
//     pattern: 路由请求路径正则表达式
//     c:       控制器对象地址
//     method:  控制器方法名
//   返回
//     void
func (rt *RouterTab) AddRegular(pattern string, c ControllerInterface, method string) {
	reflectVal := reflect.ValueOf(c)
	t := reflect.Indirect(reflectVal).Type()
	routeInfo := RouterInfo{}
	routeInfo.controllerType = t
	routeInfo.method = method
	if rt.regularRouters == nil {
		rt.regularRouters = make(map[string]RouterInfo)
	}
	rt.regularRouters[pattern] = routeInfo
}

// AddAuto 添加自动路由，路径不区分大小
//   参数
//     c:    控制器对象地址
//     args: 其它信息，支持以下两个字段
//       1): bool型，路径上是否带上包名
//       2): string型，支持路径前面拼接此信息，比如路径上带上版本v3.0, /v3.0/api/user/index
//   返回
//     void
func (rt *RouterTab) AddAuto(c ControllerInterface, args ...interface{}) {
	// 路径上是否带上包名
	usePackage := false
	if len(args) > 0 {
		usePackage, _ = args[0].(bool)
	}

	// 路径里拼接的信息，比如路径上带上版本信息
	ext := ""
	if len(args) > 1 {
		ext, _ = args[1].(string)
		ext = strings.Trim(ext, "/")
	}

	reflectVal := reflect.ValueOf(c)
	t := reflect.Indirect(reflectVal).Type()
	cList := strings.Split(t.String(), ".")
	pName := cList[0]
	cName := cList[len(cList)-1]
	n := len(cName)
	if n <= 10 {
		return
	}

	var cNamePart string
	if usePackage {
		cNamePart = "/" + pName
	}

	// 类名必须是Controller结尾
	if cName[n-10:] == "Controller" {
		cNamePart += "/" + cName[:n-10]
	} else {
		cNamePart += "/" + cName
	}

	rType := reflectVal.Type()
	for i := 0; i < rType.NumMethod(); i++ {
		method := rType.Method(i).Name
		n := len(method)
		// 类名必须是Action结尾
		if n > 6 && method[n-6:] == "Action" {
			methodPart := method[:n-6]
			pattern := strings.ToLower(cNamePart + "/" + methodPart)

			routeInfo := RouterInfo{}
			routeInfo.controllerType = t
			routeInfo.method = method
			if rt.autoRouters == nil {
				rt.autoRouters = make(map[string]RouterInfo)
			}
			if ext != "" {
				pattern = "/" + strings.ToLower(ext) + pattern
			}

			rt.autoRouters[pattern] = routeInfo
			cnt := rt.getPathCnt(pattern)
			if rt.maxPathCnt < cnt {
				rt.maxPathCnt = cnt
			}
			if cnt > 1 && rt.minPathCnt > cnt {
				rt.minPathCnt = cnt
			}
		}
	}
}

// ServeHTTP 实现http.Handler接口，匹配路由顺序：固定路由 => 自动路由 => 正则路由
//   参数
//     w: ResponseWriter对象
//     r: Request对象
//   返回
//     void
func (rt *RouterTab) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			stack := utils.Stack()
			Flogger.Errorf("path[%s] err[%v] stack[%v]", r.URL.Path, err, stack)
			accessLog(r, http.StatusInternalServerError)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError) // 500
			return
		}
	}()

	if AppCfg.ServerCfg.MaxGoCnt > 0 {
		curGoCnt := runtime.NumGoroutine()
		if curGoCnt > AppCfg.ServerCfg.MaxGoCnt {
			Flogger.Errorf("curGoCnt[%d] maxGoCnt[%d]", curGoCnt, AppCfg.ServerCfg.MaxGoCnt)
			accessLog(r, http.StatusBadGateway)
			http.Error(w, "Internal Server Error", http.StatusBadGateway) // 502
			return
		}
	}

	// 处理url path
	realPath := utils.DelRepeat(r.URL.Path, '/')
	if realPath != "/" {
		realPath = strings.TrimRight(realPath, "/")
	}

	// 静态路由
	if rt.staticRouter(w, r, realPath) {
		return
	}

	var runRouter reflect.Type
	var routeInfo RouterInfo
	var param map[string]string
	ok := false
	var urlPath string
	urlPath = strings.ToLower(realPath)
	curPathCnt := rt.getPathCnt(urlPath)

	// 固定路由
	routeInfo, ok = rt.fixedRouters[urlPath]
	if ok {
		goto RUNNING
	}

	// 自动路由
	routeInfo, ok = rt.autoRouters[urlPath]
	if ok {
		goto RUNNING
	}

	// 正则路由，正则路由是否区分大小要看正则表达如果写
	routeInfo, ok = rt.regularMatch(realPath)
	if ok {
		goto RUNNING
	}

	// 全路径未匹配到
	for i := rt.maxPathCnt - 1; i >= rt.minPathCnt; i-- {
		if i > curPathCnt {
			continue;
		}
		urlPath = rt.getPattern(urlPath, i)
		if urlPath == "" {
			break
		}

		// 固定路由
		routeInfo, ok = rt.fixedRouters[urlPath]
		if ok {
			param = rt.getParam(realPath, i)
			goto RUNNING
		}

		// 自动路由
		routeInfo, ok = rt.autoRouters[urlPath]
		if ok {
			param = rt.getParam(realPath, i)
			goto RUNNING
		}

		// 正则路由，正则路由是否区分大小要看正则表达如果写
		routeInfo, ok = rt.regularMatch(realPath)
		if ok {
			param = rt.getParam(realPath, i)
			goto RUNNING
		}
	}

	accessLog(r, http.StatusNotFound)
	http.NotFound(w, r)
	//httpStatus = http.StatusNotFound
	return

RUNNING:
	runRouter = routeInfo.controllerType
	vc := reflect.New(runRouter)
	objController, ok := vc.Interface().(ControllerInterface)
	if !ok {
		Flogger.Errorf("path[%s] err[controller is not ControllerInterface]", r.URL.Path)
		accessLog(r, http.StatusInternalServerError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError) // 500
	}

	runMethod := routeInfo.method
	objController.Init(w, r, runRouter.Name(), runMethod, param)
	chanRes := make(chan bool)
	httpStatus := 0
	var mu sync.Mutex

	// 超时设置
	f := func() {
		mu.Lock()
		defer mu.Unlock()
		if httpStatus == 0 {
			httpStatus = http.StatusBadGateway
			close(chanRes)
		}
	}
	t := time.AfterFunc(rt.reqTimeout*time.Second, f)

	// 处理Action
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := utils.Stack()
				Flogger.Errorf("path[%s] err[%v] stack[%v]", r.URL.Path, err, stack)

				mu.Lock()
				defer mu.Unlock()
				if httpStatus == 0 {
					t.Stop()
					httpStatus = http.StatusInternalServerError
					close(chanRes)
				}
			}
		}()

		objController.Prepare()
		res := objController.Filter()
		if res {
			var in []reflect.Value
			method := vc.MethodByName(runMethod)
			method.Call(in)
		}
		objController.Finish()

		mu.Lock()
		defer mu.Unlock()
		if httpStatus == 0 {
			httpStatus = http.StatusOK
			close(chanRes)
		}
	}()

	select {
	case <-chanRes:
		if httpStatus == http.StatusBadGateway {
			accessLog(r, httpStatus)
			http.Error(w, "Bad Gateway", httpStatus) // 502
		} else if httpStatus == http.StatusInternalServerError {
			accessLog(r, httpStatus)
			http.Error(w, "Internal Server Error", httpStatus) // 500
		}
	}
	if httpStatus == http.StatusOK {
		objController.Show()
	}

	objController.UnInit()
	return
}

// regularMatch 正则路由匹配
//   参数
//     urlPath: 访问路径
//   返回
//     匹配成功返回路由信息，否则返回匹配失败
func (rt *RouterTab) regularMatch(urlPath string) (RouterInfo, bool) {
	var (
		pattern    string
		routerInfo RouterInfo
	)
	for pattern, routerInfo = range rt.regularRouters {
		m, _ := regexp.MatchString(pattern, urlPath)
		if m {
			return routerInfo, true
		}
	}
	return routerInfo, false
}

// staticRouter 静态路由
//   参数
//     w:       ResponseWriter对象
//     w:       Request对象
//     urlPath: 请求路径
//   返回
//     匹配成功返回true，否则返回false
func (rt *RouterTab) staticRouter(w http.ResponseWriter, r *http.Request, urlPath string) bool {
	if r.Method != "GET" && r.Method != "HEAD" {
		return false
	}
	filePath, isMatch := rt.checkStaticFile(urlPath)
	if !isMatch {
		return false
	}

	if filePath == "" {
		return false
	}

	http.ServeFile(w, r, filePath)

	return true
}

// checkStaticFile 验证静态文件
//   参数
//     urlPath: 请求路径
//   返回
//     文件路径、是否匹配路由
func (rt *RouterTab) checkStaticFile(urlPath string) (string, bool) {
	var file string

	requestPath := filepath.ToSlash(filepath.Clean(urlPath))

	// favicon.ico、robots.txt文件单独处理
	if requestPath == "/favicon.ico" || requestPath == "/robots.txt" {
		file = path.Join(AppRoot, requestPath)
		r, err := utils.IsFile(file)
		if !r || err != nil {
			return file, false
		}
		return file, true
	}

	// 静态文件
	for _, staticDir := range AppCfg.WebCfg.StaticDir {
		staticDir = strings.TrimRight(staticDir, "/")
		if requestPath != staticDir && !strings.HasPrefix(requestPath, staticDir+"/") {
			continue
		}

		file := path.Join(AppRoot, requestPath)
		// 如果是文件夹，拼上index.html
		isDir, err := utils.IsDir(file)
		if err != nil {
			return "", false
		} else if isDir {
			file = filepath.Join(file, "index.html")
		}
		return file, true
	}

	return "", false
}

// getUrlAndParam 根据path获取用于匹配的pattern和参数
// 前面两段做为pattern，后面的做为参数
// 比如url=/user/index/ver/3.0/id/10，则pattern为/user/index，参数为ver=3.0 id=10
//   参数：
//     path: 原始的url
//   返回：
//     pattern: 匹配串
//     param:   参数串
func (rt *RouterTab) getPatternAndParam(path string) (string, map[string]string) {
	if path == "" {
		return "", nil
	}

	urlList := strings.Split(strings.TrimLeft(path, "/"), "/")
	n := len(urlList)
	if n < 3 {
		// 没有参数
		return path, nil
	}

	pattern := "/" + urlList[0] + "/" + urlList[1]
	param := make(map[string]string)
	isEven := n%2 == 0
	for i := 2; i < n; i += 2 {
		if isEven || i < n-1 {
			param[urlList[i]] = urlList[i+1]
		} else {
			param[urlList[i]] = ""
		}
	}

	return pattern, param
}

// getPattern 根据path获取用于匹配的pattern
// 比如url=/user/index/ver/3.0/id/10, cnt=2，则参数为ver=3.0 id=10
//   参数：
//     path: 原始的url
//     cnt:  前面cnt段为路径
//   返回：
//     参数信息
func (rt *RouterTab) getPattern(path string, cnt int) string {
	if path == "" {
		return ""
	}

	urlList := strings.Split(strings.TrimLeft(path, "/"), "/")
	n := len(urlList)
	if n < cnt {
		return ""
	}

	return "/" + strings.Join(urlList[:cnt], "/")
}

// getParam 根据path获取参数
// 比如url=/user/index/ver/3.0/id/10, cnt=2，则pattern为/user/index
//   参数：
//     path: 原始的url
//     cnt:  前面cnt段为路径
//   返回：
//     pattern: 匹配串
func (rt *RouterTab) getParam(path string, cnt int) map[string]string {
	param := make(map[string]string)
	urlList := strings.Split(strings.TrimLeft(path, "/"), "/")
	paramList := urlList[cnt:]
	n := len(paramList)
	isEven := n%2 == 0
	for i := 0; i < n; i += 2 {
		if isEven || i < n-1 {
			param[paramList[i]] = paramList[i+1]
		} else {
			param[paramList[i]] = ""
		}
	}

	return param
}

// accessLog 写访问日志
//   参数
//     r:          http.Request
//     httpStatus: http状态码
//   返回
//     void
func accessLog(r *http.Request, httpStatus int) {
	userAgent := r.Header.Get("User-Agent")
	proxy1 := r.Header.Get("X-Forwarded-For")
	proxy2 := ""
	if AppCfg.ServerCfg.ForwardName != "" {
		proxy2 = r.Header.Get(AppCfg.ServerCfg.ForwardName)
	}

	if httpStatus >= 400 {
		Flogger.Errorf("%s|%s|%s|%d|%s|%s|%s|%s", r.RemoteAddr, r.Method, r.RequestURI, httpStatus, userAgent, r.Host, proxy1, proxy2)
	} else {
		Flogger.Infof("%s|%s|%s|%d|%s|%s|%s|%s", r.RemoteAddr, r.Method, r.RequestURI, httpStatus, userAgent, r.Host, proxy1, proxy2)
	}
}

// AddShell 添加Shell脚本路由
//   参数
//     pattern: 路由请求路径
//     handler: Shell要执行的函数
//   返回
//     void
func (rt *RouterTab) AddShell(pattern string, handler ShellFunc) {
	if rt.shellRouters == nil {
		rt.shellRouters = make(map[string]ShellFunc)
	}

	pattern = strings.ToLower(pattern)
	rt.shellRouters[pattern] = handler
}

// matchShell 匹配一个Shell脚本的路由
//   参数
//     pattern: 路由请求路径
//   返回
//     Shell要执行的函数
func (rt *RouterTab) matchShell(pattern string) ShellFunc {
	if rt.shellRouters == nil {
		return nil
	}

	pattern = strings.ToLower(pattern)
	if f, ok := rt.shellRouters[pattern]; ok {
		return f
	}

	return nil
}

// getPathCnt 获取路由段数
// 比如 /aa/bb/cc，则返回 3
//   参数
//     pattern: 路由请求路径
//   返回
//     获取路由段数
func (rt *RouterTab) getPathCnt(pattern string) int {
	pList := strings.Split(pattern, "/")

	return len(pList) - 1
}
