// 控制器类的接口及基类
//   变更历史
//     2017-02-06  lixiaoya  新建
package bingo

import (
	"errors"
	"io"
	"io/ioutil"
	"github.com/bingo/session"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"bytes"
	"encoding/json"
	"github.com/bingo/utils"
	"html/template"
	"encoding/xml"
)

// ControllerInterface 所有controller接口
type ControllerInterface interface {
	Init(w http.ResponseWriter, r *http.Request, controllereName, actionName string, param map[string]string)
	UnInit()

	Prepare()     // 在Action函数之前执行，子类可以重写
	Finish()      // 在Action函数之后执行，子类可以重写
	Filter() bool // 在Action函数之前做一次过滤处理，返回true则继续执行Action
	Show()        // 输出数据到页面
}

// Controller 所有Controller的基类
type Controller struct {
	// 请求与响应信息
	Req     Request
	Rsp     Response
	outBuf  bytes.Buffer // 要输出到页面的数据

	// 模板信息
	tplData  map[string]interface{}
	TplFiles []string

	// 路由信息
	ControllerName string
	ActionName     string

	// Session
	CurSession session.SessData
}

// Init 初始化
//   参数
//     w:              ResponseWriter
//     r:              Request
//     controllerName: controller类名
//     actionName:     action函数名
//     param:          path参数
//   返回
//     void
func (c *Controller) Init(w http.ResponseWriter, r *http.Request, controllereName, actionName string, param map[string]string) {
	// 初始化ResponseWriter 和 Response
	c.Req.reSet(r)
	encoding := ""
	if AppCfg.ServerCfg.GzipStatus {
		encoding = RspEncoding(r.Header.Get("Accept-Encoding"))
	}
	c.Rsp.reSet(w, encoding)

	// 设置参数
	c.Req.parseGetParam()
	c.Req.parsePostParam()
	c.Req.parsePathParam(param)

	c.ControllerName = controllereName
	c.ActionName = actionName

	c.tplData = make(map[string]interface{})
}

// UnInit 反初始化
//   参数
//     void
//   返回
//     void
func (c *Controller) UnInit() {
	if c.CurSession != nil {
		c.CurSession.Write()
	}
}

// Prapare 在执行action函数之前执行
//   参数
//     void
//   返回
//     void
func (c *Controller) Prepare() {
}

// Finish 在执行action函数之后执行
//   参数
//     void
//   返回
//     void
func (c *Controller) Finish() {
}

// Filter 在Action函数之前做一次过滤处理
//   参数
//     void
//   返回
//     true-继续执行Action false-不执行Action
func (c *Controller) Filter() bool {
	return true
}

// Show 输出数据到页面
//   参数
//     void
//   返回
//     void
func (c *Controller) Show() {
	out := c.outBuf.Bytes()
	if len(out) > 0 {
		c.Rsp.OutPut(c.outBuf.Bytes())
	}
}

// GetString 根据key获取对应的GET参数
// 如果key对应多个值，则返回第一个值
// c.GetString("bb", "44")
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) GetString(key string, def ...string) string {
	if v, ok := c.Req.getParam[key]; ok {
		return strings.TrimSpace(v[0])
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetStrings 根据key获取对应的GET参数
// 如果key对应多个值，则都返回
// c.GetStrings("bb", []string{"33", "44"})
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) GetStrings(key string, def ...[]string) []string {
	if v, ok := c.Req.getParam[key]; ok {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return []string{}
}

// GetInt 返回int型
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) GetInt(key string, def ...int) (int, error) {
	v, _ := c.Req.getParam[key]
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}

	return strconv.Atoi(strings.TrimSpace(v[0]))
}

// GetBool 返回bool型
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) GetBool(key string, def ...bool) bool {
	if v, ok := c.Req.getParam[key]; ok {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	} else {
		if len(def) > 0 {
			return def[0]
		} else {
			return false
		}
	}
}

// GetAll 返回所有GET参数
// 如果key对应多个值，则都返回，只取第一个参数
// c.GetAll()
//   参数
//     void
//   返回
//     所有GET参数
func (c *Controller) GetAll() map[string]string {
	param := make(map[string]string)
	for k, v := range c.Req.getParam {
		param[k] = v[0]
	}
	return param
}

// PostString 根据key获取对应的POST参数
// 如果key对应多个值，则返回第一个值
// c.PostString("bb", "44")
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) PostString(key string, def ...string) string {
	if v, ok := c.Req.postParam[key]; ok {
		return strings.TrimSpace(v[0])
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// PostStrings 根据key获取对应的GET参数
// 如果key对应多个值，则都返回
// c.PostStrings("bb", []string{"33", "44"})
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) PostStrings(key string, def ...[]string) []string {
	if v, ok := c.Req.postParam[key]; ok {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return []string{}
}

// PostInt 返回int型
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) PostInt(key string, def ...int) (int, error) {
	v, _ := c.Req.postParam[key]
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}

	return strconv.Atoi(strings.TrimSpace(v[0]))
}

// PostBool 返回bool型
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) PostBool(key string, def ...bool) bool {
	if v, ok := c.Req.postParam[key]; ok {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	} else {
		if len(def) > 0 {
			return def[0]
		} else {
			return false
		}
	}
}

// PostAll 返回所有POST参数，只取第一个参数
// 如果key对应多个值，则都返回
// c.PostAll()
//   参数
//     void
//   返回
//     所有POST参数
func (c *Controller) PostAll() map[string]string {
	param := make(map[string]string)
	for k, v := range c.Req.postParam {
		param[k] = v[0]
	}
	return param
}

// ParamString 根据key获取对应的PATH参数
// c.ParamString("bb", "44")
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) ParamString(key string, def ...string) string {
	if v, ok := c.Req.pathParam[key]; ok {
		return strings.TrimSpace(v[0])
	}
	def = append(def, "")
	return def[0]
}

// ParamInt 返回int型
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) ParamInt(key string, def ...int) (int, error) {
	v, _ := c.Req.pathParam[key]
	if len(v) == 0 && len(def) > 0 {
		return def[0], nil
	}

	return strconv.Atoi(strings.TrimSpace(v[0]))
}

// ParamBool 返回bool型
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) ParamBool(key string, def ...bool) bool {
	if v, ok := c.Req.pathParam[key]; ok {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	} else {
		def = append(def, false)
		return def[0]
	}
}

// ParamAll 返回所有PATH参数
// 如果key对应多个值，则都返回，只取第一个参数
// c.ParamAll()
//   参数
//     key: key值
//     def: 默认值
//   返回
//     所有PATH参数
func (c *Controller) ParamAll() map[string]string {
	param := make(map[string]string)
	for k, v := range c.Req.pathParam {
		param[k] = v[0]
	}
	return param
}

// ReqString 根据key获取对应的post和get的参数
// 优先级为post > get
// 如果key对应多个值，则返回第一个值
// c.ReqString("bb", "44")
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) ReqString(key string, def ...string) string {
	// POST
	if v, ok := c.Req.postParam[key]; ok {
		return strings.TrimSpace(v[0])
	}

	// GET
	if v, ok := c.Req.getParam[key]; ok {
		return strings.TrimSpace(v[0])
	}

	// Default
	def = append(def, "")
	return def[0]
}

// ReqInt 返回int型
// 优先级为post > get
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) ReqInt(key string, def ...int) (int, error) {
	// POST
	if v, ok := c.Req.postParam[key]; ok && len(v) > 0 {
		return strconv.Atoi(strings.TrimSpace(v[0]))
	}

	// GET
	if v, ok := c.Req.getParam[key]; ok && len(v) > 0 {
		return strconv.Atoi(strings.TrimSpace(v[0]))
	}

	// Default
	def = append(def, 0)
	return def[0], nil
}

// ReqBool 返回bool型
// 优先级为post > get
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) ReqBool(key string, def ...bool) bool {
	// POST
	if v, ok := c.Req.postParam[key]; ok && len(v) > 0 {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	}

	// GET
	if v, ok := c.Req.getParam[key]; ok && len(v) > 0 {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	}

	// Default
	def = append(def, false)
	return def[0]
}

// ReqAll 返回POST和GET所有参数，只取第一个参数
// 如果GET和POST都有的参数，则保留POST
// 如果key对应多个值，则都返回
//   参数
//     void
//   返回
//     所有POST参数
func (c *Controller) ReqAll() map[string]string {
	param := make(map[string]string)

	// GET
	for k, v := range c.Req.getParam {
		param[k] = v[0]
	}

	// POST
	for k, v := range c.Req.postParam {
		param[k] = v[0]
	}
	return param
}

// VarString 根据key获取对应的所有类型的参数
// 优先级为path > post > get
// 如果key对应多个值，则返回第一个值
// c.VarString("bb", "44")
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) VarString(key string, def ...string) string {
	// PATH
	if v, ok := c.Req.pathParam[key]; ok {
		return strings.TrimSpace(v[0])
	}

	// POST
	if v, ok := c.Req.postParam[key]; ok {
		return strings.TrimSpace(v[0])
	}

	// GET
	if v, ok := c.Req.getParam[key]; ok {
		return strings.TrimSpace(v[0])
	}

	// Default
	def = append(def, "")
	return def[0]
}

// VarInt 返回int型
// 优先级为path > post > get
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) VarInt(key string, def ...int) (int, error) {
	// PATH
	if v, ok := c.Req.pathParam[key]; ok {
		return strconv.Atoi(strings.TrimSpace(v[0]))
	}

	// POST
	if v, ok := c.Req.postParam[key]; ok && len(v) > 0 {
		return strconv.Atoi(strings.TrimSpace(v[0]))
	}

	// GET
	if v, ok := c.Req.getParam[key]; ok && len(v) > 0 {
		return strconv.Atoi(strings.TrimSpace(v[0]))
	}

	// Default
	def = append(def, 0)
	return def[0], nil
}

// VarBool 返回bool型
// 优先级为path > post > get
// 如果key不存在，则返回默认值
//   参数
//     key: key值
//     def: 默认值
//   返回
//     value值
func (c *Controller) VarBool(key string, def ...bool) bool {
	// PATH
	if v, ok := c.Req.pathParam[key]; ok {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	}

	// POST
	if v, ok := c.Req.postParam[key]; ok && len(v) > 0 {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	}

	// GET
	if v, ok := c.Req.getParam[key]; ok && len(v) > 0 {
		switch strings.ToUpper(strings.TrimSpace(v[0])) {
		case "1", "T", "TRUE", "YES", "Y", "ON":
			return true
		default:
			return false
		}
	}

	// Default
	def = append(def, false)
	return def[0]
}

// GetBody 获取post提交的字符串，比如直接post提交一个xml串或json串
//   参数
//     void
//   返回
//     post提交的字符串，成功返回nil，失败返回错误信息
func (c *Controller) GetBody() (string, error) {
	if c.Req.r.Body == nil {
		return "", nil
	}

	safe := &io.LimitedReader{R: c.Req.r.Body, N: MaxMem}
	body, err := ioutil.ReadAll(safe)
	c.Req.r.Body.Close()

	return string(body), err
}

// GetFile 返回上传的文件信息
//   参数
//     key: 文件key值
//   返回
//     文件的信息，如果key对应多个，将返回第一个文件，失败时返回错误信息
func (c *Controller) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
	c.Req.r.ParseMultipartForm(MaxMem)
	return c.Req.r.FormFile(key)
}

// GetFiles 返回上传的多个文件信息
//   参数
//     key: 文件key值
//   返回
//     文件信息列表，失败时返回错误信息
func (c *Controller) GetFiles(key string) ([]*multipart.FileHeader, error) {
	c.Req.r.ParseMultipartForm(MaxMem)
	if files, ok := c.Req.r.MultipartForm.File[key]; ok {
		return files, nil
	}
	return nil, http.ErrMissingFile
}

// SaveFile 保存文件
//   参数
//     key: 文件key值
//     desPath: 生成的文件路径
//   返回
//     成功返回nil，失败返回错误信息
func (c *Controller) SaveFile(key, dstFile string) error {
	c.Req.r.ParseMultipartForm(MaxMem)
	file, _, err := c.Req.r.FormFile(key)
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

// Assign 设置模板变量
//   参数
//     key: key值
//     val: value值
//   返回
//     void
func (c *Controller) Assign(key string, val interface{}) {
	c.tplData[key] = val
}

// Display 显示模板文件
//   参数
//     tplFile: 模板文件
//   返回
//     void
func (c *Controller) Display(tplFile string) {
	err := GTemplate.ViewTemp.ExecuteTemplate(c.Rsp.w, tplFile, c.tplData)
	if err != nil {
		panic(err.Error())
	}
}

// SetCookie 设置cookie
//   参数
//     name:   cookie的key值
//     value:  cookie的value值
//     others: 其它参数，依次为下面几项
//         MxAge    int    设置过期时间，对应浏览器cookie的MaxAge属性
//         Path     string 路径
//         Domain   string 域名
//         Secure   bool   设置Secure属性
//         HttpOnly bool   设置httpOnly属性
//   返回
//     void
func (c *Controller) SetCookie(name, value string, others ...interface{}) {
	c.Rsp.SetCookie(name, value, others...)
}

// GetCookie 获取cookie
//   参数
//     name: cookie的key值
//   返回
//     cookie的value值
func (c *Controller) GetCookie(name string) string {
	return c.Req.GetCookie(name)
}

// GetUrlPath 返回不带参数的url
//   参数
//     void
//   返回
//     不带参数的url，如: /user/index
func (c *Controller) GetUrlPath() string {
	return c.Req.UrlPath()
}

// GetClientIp 返回客户端IP
// 如果代理为非le的，返回第一个代理IP，否则返回最后一个代理IP
// 为空时返回127.0.0.1
//   参数
//     void
//   返回
//     客户端IP
func (c *Controller) GetClientIp() string {
	return c.Req.ClientIp()
}

// GetClientPort 返回客户端端口号
//   参数
//     void
//   返回
//     客户端端口号
func (c *Controller) GetClientPort() string {
	return c.Req.ClientPort()
}

// GetUserAgent 返回请求的User-Agent数据
//   参数
//     void
//   返回
//     User-Agent数据
func (c *Controller) GetUserAgent() string {
	return c.Req.UserAgent()
}

// WriteString 写数据到页面
//   参数
//     data: 写到页面的数据
//   返回
//     void
func (c *Controller) WriteString(data string) {
	c.outBuf.WriteString(data)
}

// WriteJson 将数据转成json串输出到页面
//   参数
//     data: 要输出的数据
//     args: 其它参数，按顺序分别为
//         jsonp的callback函数
//         是否需要做json编码，开发模式默认为false，生产模式默认为true
//         是否要做格式化缩进，开发模式默认为true，生产模式默认为false
//   返回
//     void
func (c *Controller) WriteJson(data interface{}, args ...interface{}) {
	encode := true
	indent := false
	callback := ""
	var ok bool

	// 开发模式默认不编码，并进行缩进
	if AppCfg.RunMode == DEV {
		encode = false
		indent = true
	}

	n := len(args)

	// jsonp的callback函数
	if n > 0 {
		if callback, ok = args[0].(string); !ok {
			callback = ""
		}
	}

	// 是否需要做json编码
	if n > 1 {
		if encode, ok = args[1].(bool); !ok {
			encode = true
		}
	}

	// 是否要做格式化缩进
	if n > 2 {
		if indent, ok = args[2].(bool); !ok {
			indent = false
		}
	}

	var content []byte
	var err error
	if indent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}
	if err != nil {
		http.Error(c.Rsp.w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	if encode {
		content = []byte(utils.StrToJSON(content))
	}

	if callback == "" {
		c.outBuf.Write(content)
		return
	}

	callback = template.JSEscapeString(callback)
	//callbackContent := bytes.NewBufferString(" if(window." + callback + ")" + callback)
	callbackContent := bytes.NewBufferString(callback)
	callbackContent.WriteString("(")
	callbackContent.Write(content)
	callbackContent.WriteString(")")
	c.outBuf.WriteString(callbackContent.String())
}

// WriteJson 将数据转成json串输出到页面
//   参数
//     data: 要输出的数据
//     args: 是否要做格式化缩进，开发模式默认为true，生产模式默认为false
//   返回
//     void
func (c *Controller) WriteXml(data interface{}, args ...bool) {
	indent := false

	// 开发模式默认不编码，并进行缩进
	if AppCfg.RunMode == DEV {
		indent = true
	}

	// 是否要做格式化缩进
	if len(args) > 0 {
		indent = args[0]
	}

	var content []byte
	var err error
	if indent {
		content, err = xml.MarshalIndent(data, "", "  ")
	} else {
		content, err = xml.Marshal(data)
	}
	if err != nil {
		http.Error(c.Rsp.w, err.Error(), http.StatusInternalServerError) // 500
		return
	}
	c.outBuf.Write(content)
}

// WriteBinary 写二进行文件数据到页面
//   参数
//     data:        写到页面的数据
//     contentType: Content-Type，如：png图片为 image/png，jpg图片为 image/jpeg
//   返回
//     void
func (c *Controller) WriteBinary(data []byte, contentType string) {
	c.Rsp.Binary(data, contentType)
}

// Redirect 跳转到一个本地url
//   参数
//     urlStr: 跳转的url地址
//     status: http状态
//   返回
//
func (c *Controller) Redirect(urlStr string, status ...int) {
	code := 302
	if len(status) > 0 {
		code = status[0]
	}
	http.Redirect(c.Rsp.w, c.Req.r, urlStr, code)
}

// StartSession 返回SessionData
//   参数
//     void
//   返回
//     void
func (c *Controller) StartSession() {
	if GlobalSession == nil {
		return
	}
	if c.CurSession == nil {
		c.CurSession, _ = GlobalSession.SessStart(c.Rsp.w, c.Req.r)
	}
	return
}

// StartSession 返回SessionData
//   参数
//     void
//   返回
//     void
func (c *Controller) SessionId() string {
	if GlobalSession == nil {
		return ""
	}
	c.StartSession()

	return c.CurSession.Id()
}

// SetSession 设置Session
//   参数
//     key:   session的key值
//     value: session的Value值
//   返回
//     成功返回nil，失败返回错误信息
func (c *Controller) SetSession(key string, value interface{}) error {
	if GlobalSession == nil {
		return errors.New("session: GlobalSession is nil.")
	}
	c.StartSession()

	err := c.CurSession.Set(key, value)
	return err
}

// GetSession 返回Session
//   参数
//     key:   session的key值
//   返回
//     session的Value值
func (c *Controller) GetSession(key string) interface{} {
	if GlobalSession == nil {
		return ""
	}
	c.StartSession()

	return c.CurSession.Get(key)
}

// DelSession 删除key对应的session值
//   参数
//     key:   session的key值
//   返回
//     void
func (c *Controller) DelSession(key string) {
	c.StartSession()
	c.CurSession.Delete(key)
}

// DestroySession 清空数据和cookie
//   参数
//     void
//   返回
//     void
func (c *Controller) DestroySession() {
	c.StartSession()

	GlobalSession.SessDestroy(c.Rsp.w, c.Req.r)
	c.CurSession = nil
}

// GetHeader 返回Header项
//   参数
//     key: Header的key值
//   返回
//     Header项，如果key不存在，返回nil
func (c *Controller) GetHeader(key string) string {
	return c.Req.Header(key)
}

// Header 设置response的header
//   参数
//     key: Header的key值
//     val: Header的value值
//   返回
//     void
func (c *Controller) SetHeader(key, val string) {
	c.Rsp.Header(key, val)
}
