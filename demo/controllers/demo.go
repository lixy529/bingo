// controller demo
//   变更历史
//     2017-02-10  lixiaoya  新建
package controllers

import (
	"fmt"
	"io"
	"github.com/lixy529/bingo"
	_ "github.com/lixy529/bingo/cache/memcache"
	_ "github.com/lixy529/bingo/cache/redis/redisc"
	_ "github.com/lixy529/bingo/cache/redis/redism"
	_ "github.com/lixy529/bingo/cache/redis/redisd"
	"github.com/lixy529/bingo/demo/models"
	"github.com/lixy529/bingo/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"io/ioutil"
)

type DemoController struct {
	BaseController
}

type Info struct {
	Name  string
	Email string
	Phone string
}

func (i *Info) ToString() string {
	return i.Name + "|" + i.Email + "|" + i.Phone
	//return "{\"Name\":\""+ i.Name + "\",\"Email\":\"" + i.Email + "\",\"Phone\":\"" + i.Phone + "\"}"
}

func (i *Info) ToObj(str string) {
	l := strings.Split(str, "|")
	if len(l) > 2 {
		i.Name = l[0]
		i.Email = l[1]
		i.Phone = l[2]
	}
}

func (c *DemoController) IndexAction() {
	data := Info{
		Name:  "lixiaoya",
		Email: "lixiaoya@le.com",
		Phone: "15811112222",
	}

	c.Assign("Title", "this a test page")
	c.Assign("Info", data)
	c.Assign("Time", time.Now())
	c.Assign("Footer", ">>>The page is end<<<")

	c.Display("demo/index.html")
}

func (c *DemoController) Demo1Action() {
	a := c.GetString("b")
	c.WriteString(a + "<br />")
	name := c.GetString("name", "sso")
	c.WriteString(name + "<br />")
	c.WriteString("Hell World!")
}

func (c *DemoController) Demo2Action() {
	v := c.PostString("mobile", "unknow")
	c.WriteString(v)
}

func (c *DemoController) Demo3Action() {
	c.Redirect("/demo/index")
	//c.Redirect("http://www.baidu.com", 302)
}

func (c *DemoController) Demo4Action() {
	panic("excption test")
}

func (c *DemoController) Demo5Action() {
	//c.SetCookie("uid", "u00001", "/", "localhost", false, true)
	c.SetCookie("uid", "u00001")

	c.SetCookie("name", "lixiaoya", 100, "/")
	name := c.GetCookie("name")
	c.WriteString(name)
}

func (c *DemoController) Demo6Action() {
	uid := c.GetCookie("uid")
	c.WriteString(uid)
}

// SessAction Session测试
func (c *DemoController) SessAction() {
	isSet := c.GetBool("is_set")
	isDestory := c.GetBool("is_destory")

	if isDestory {
		c.DestroySession()
	} else if isSet {
		v1 := c.GetString("v1", "v111")
		c.SetSession("k1", v1)

		v2 := c.GetString("v2", "v222")
		c.SetSession("k2", v2)

		info := Info{Name: "lixy", Email: "lixiaoya@le.com", Phone: "15811112222"}
		c.SetSession("info", info.ToString())

		c.WriteString(c.SessionId())
	} else {
		v1 := c.GetSession("k1")
		if v1 == nil {
			c.WriteString("k1 session 不存在<br />")
		} else {
			c.WriteString("v1 = " + v1.(string) + "<br />")
		}

		info := c.GetSession("info")
		if info == nil {
			c.WriteString("info session 不存在<br />")
		} else {
			ifo := Info{}
			ifo.ToObj(info.(string))
			name := ifo.Name
			email := ifo.Email
			phone := ifo.Phone
			c.WriteString("name = " + name + " ")
			c.WriteString("email = " + email + " ")
			c.WriteString("phone = " + phone + "<br />")
		}

		v2 := c.GetSession("k2")
		if v1 == nil {
			c.WriteString("k2 session 不存在<br />")
		} else {
			c.WriteString("v2 = " + v2.(string) + "<br />")
		}

		// 删除k1
		c.DelSession("k1")
		v1 = c.GetSession("k1")
		if v1 == nil {
			c.WriteString("k1 session 不存在<br />")
		} else {
			c.WriteString("v1 = " + v1.(string) + "<br />")
		}

		v2 = c.GetSession("k2")
		if v2 == nil {
			c.WriteString("k2 session 不存在<br />")
		} else {
			c.WriteString("v2 = " + v2.(string) + "<br />")
		}
	}
}

// HeaderAction Header测试
func (c *DemoController) HeaderAction() {
	c.SetHeader("token", "0aaaaaaaaaaaaaaaaaaaaa1")
	ae := c.GetHeader("Accept-Encoding")
	c.WriteString(ae + "<br />")
}

// JsonAction json串输出测试
func (c *DemoController) JsonAction() {
	c.WriteString("<pre />")
	info := &Info{
		Name:  "李四",
		Phone: "15812345678",
		Email: "lixiaoya@le.com",
	}

	c.WriteJson(info, "alert", true, false)
}

// XmlAction xml串输出测试
func (c *DemoController) XmlAction() {
	info := &Info{
		Name:  "李四",
		Phone: "15812345678",
		Email: "lixiaoya@le.com",
	}

	c.WriteXml(info, false)
}

// BodyAction 获取post字符串测试
func (c *DemoController) BodyAction() {
	body, err := c.GetBody()
	if err != nil {
		c.WriteString(err.Error())
		return
	}
	c.WriteString("body=" + body)
}

// UploadAction 上传文件测试
func (c *DemoController) UploadAction() {
	if c.Req.IsGet() {
		c.Display("demo/upload.html")
		return
	}

	// GetFile
	/*file, handler, err := c.GetFile("uploadfile")
	  if err != nil {
	      fmt.Println(err)
	      return
	  }
	  defer file.Close()
	  f, err := os.OpenFile("./"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	  if err != nil {
	      fmt.Println(err)
	      return
	  }
	  defer f.Close()
	  io.Copy(f, file)*/

	// SaveFile
	/*
	   err := c.SaveFile("uploadfile", "./1111.png")
	   if err != nil {
	       fmt.Println(err)
	   }
	*/

	// GetFiles
	files, err := c.GetFiles("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := range files {
		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		dst, err := os.Create("./" + files[i].Filename)
		defer dst.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		if _, err := io.Copy(dst, file); err != nil {
			fmt.Println(err)
			return
		}
	}
}

// memcCacheAction Memcache Cache测试
func (c *DemoController) CacheAction() {
	cacheName := strings.ToLower(c.GetString("name", ""))
	if cacheName == "" {
		c.WriteString("Please input name")
		return
	}
	m := &models.DemoModel{}
	val, err := m.CacheTest(cacheName)
	if err != nil {
		c.WriteString("Cache error," + err.Error() + "<br />")
		return
	}

	c.WriteString(val)
}

// dbTestAction db测试
func (c *DemoController) DbTestAction() {
	m := &models.DemoModel{}

	// 插入
	uid, err := m.InsertInfo("letv", 100, "北京市")
	if err != nil {
		c.WriteString(err.Error())
		return
	}

	// 查询
	res, err := m.GetInfo(int(uid))
	if err != nil {
		c.WriteString(err.Error())
		return
	}
	c.WriteString(res["name"] + ":" + res["addr"] + "<br />")

	// 更新
	cnt, err := m.UpdateInfo(int(uid), "乐视大厦")
	if err != nil {
		c.WriteString(err.Error())
		return
	}
	c.WriteString("update cnt = " + strconv.Itoa(int(cnt)) + "<br />")

	// 查询
	res, err = m.GetInfo(int(uid))
	if err != nil {
		c.WriteString(err.Error())
		return
	}
	c.WriteString(res["name"] + ":" + res["addr"] + "<br />")

	// 删除
	cnt, err = m.DeleteInfo(int(uid))
	if err != nil {
		c.WriteString(err.Error())
		return
	}
	c.WriteString("delete cnt = " + strconv.Itoa(int(cnt)) + "<br />")

	// 事务测试
	m.TxTest()
}

// SrvTestAction 服务器超时测试
func (c *DemoController) SrvTestAction() {
	pid := os.Getpid()
	sPid := fmt.Sprintf("pid = %d", pid)

	//c.WriteString(sPid + "\n")
	log.Println(sPid)
	time.Sleep(10 * time.Second)
	c.WriteString("test end...")
}

// LogTestAction 日志测试
func (c *DemoController) LogTestAction() {
	bingo.Glogger.Debug("<182>" + "sso" + "[10001]:" + "usrlog|" + "LevelDebug")
	bingo.Glogger.Info("<182>" + "sso" + "[10002]:" + "usrlog|" + "LevelInfo")
	bingo.Glogger.Warn("<182>" + "sso" + "[10003]:" + "usrlog|" + "LevelWarn")
	bingo.Glogger.Error("<182>" + "sso" + "[10004]:" + "usrlog|" + "LevelError")
	bingo.Glogger.Fatal("<182>" + "sso" + "[10005]:" + "usrlog|" + "LevelFatal")

	bingo.Glogger.Debugf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20001, "usrlog", utils.CurTime(), "DEBUG", "LevelDebug")
	bingo.Glogger.Infof("<182>sso[%d]:%s|%s\t[%s]\t%s", 20002, "usrlog", utils.CurTime(), "INFO", "LevelInfo")
	bingo.Glogger.Warnf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20003, "usrlog", utils.CurTime(), "WARN", "LevelWarn")
	bingo.Glogger.Errorf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20004, "usrlog", utils.CurTime(), "ERROR", "LevelError")
	bingo.Glogger.Fatalf("<182>sso[%d]:%s|%s\t[%s]\t%s", 20005, "usrlog", utils.CurTime(), "FATAL", "LevelFatal")
}

// PressAction 一个简单的压力测试
func (c *DemoController) PressAction() {
	m := &models.DemoModel{}
	m.Sort()
}

// ParamAction 路径带参数测试
func (c *DemoController) ParamAction() {
	id, _ := c.ParamInt("id")
	c.WriteString(strconv.Itoa(id) + "<br />")

	name := c.ParamString("name", "lixy")
	c.WriteString(name + "<br />")

	isa := c.ParamBool("isa")
	if isa {
		c.WriteString("yes" + "<br />")
	} else {
		c.WriteString("no" + "<br />")
	}

	all := c.ParamAll()
	for k, v := range all {
		c.WriteString(k + "=" + v + "<br />")
	}

	c.WriteString("==========\n")
	id, _ = c.VarInt("id1")
	c.WriteString(strconv.Itoa(id) + "<br />")

	name = c.VarString("name1", "lixy")
	c.WriteString(name + "\n")

	isa = c.VarBool("isa1")
	if isa {
		c.WriteString("yes" + "<br />")
	} else {
		c.WriteString("no" + "<br />")
	}
}

// LangAction 语言包
func (c *DemoController) LangAction() {
	name := bingo.GLang.String("zh-cn", "name")
	addr := bingo.GLang.String("zh-cn", "addr")
	c.WriteString(name + ": " + addr + "<br />")

	name = bingo.GLang.String("en-us", "name")
	addr = bingo.GLang.String("en-us", "addr")
	c.WriteString(name + ": " + addr + "<br />")

	errs := bingo.GLang.Map("en-us", "err")
	for k, v := range errs {
		c.WriteString(k + ": " + v + "<br />")
	}
}

// BinAction 写二进制文件测试
func (c *DemoController) BinAction() {
	f := "/tmp/qq.png"
	t := "image/png"

	b, err := ioutil.ReadFile(f)
	if err != nil {
		c.WriteString(err.Error())
		return
	}

	c.WriteBinary(b, t)
}

// TmAction 终端测试
func (c *DemoController) TmAction() {
	userAgent := c.GetUserAgent()
	c.WriteString(userAgent)
	c.WriteString("<br />")
	tType, osType := utils.GetTerminal(userAgent)
	c.WriteString(tType + "-" + osType)
}
