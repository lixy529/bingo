// 带包名的自动路由测试
package api

import "github.com/lixy529/bingo/demo/controllers"

type UserController struct {
	controllers.BaseController
}

func (c *UserController) IndexAction() {
	c.WriteString("user package")
}

func (c *UserController) GetAction() {
	c.WriteString("get package")
}

func (c *UserController) SetAction() {
	c.WriteString("set package")
}
