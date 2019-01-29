package lang

import (
	"testing"
)

// TestLang test language package.
func TestLang(t *testing.T) {
	root := "./testdata/"
	l, err := NewLang(root)
	if err != nil {
		t.Errorf("New lang failed, err: %s", err.Error())
		return
	}

	name,_ := l.ReadLang("zh-cn", "name")
	if name != "李四" {
		t.Errorf("ReadLang err, got [%s], expected [李四]", name)
		return
	}

	name = l.String("en-us", "name")
	if name != "Nick" {
		t.Errorf("ReadLang err, got [%s], expected [nick]", name)
		return
	}

	region,_ := l.ReadLang("zh-cn", "region")
	if region != "" {
		t.Errorf("ReadLang err, got [%s], expected [中国北京]", region)
		return
	}

	region = l.String("zh-cn", "region")
	if region != "" {
		t.Errorf("String err, got [%s], expected [中国北京]", region)
		return
	}

	errs := l.Map("en-us", "err")
	if err != nil {
		t.Errorf("Map err")
		return
	} else if errs["1001"] != "err1" || errs["1002"] != "err2" {
		t.Errorf("Map err, got [%s]-[%s], expected [err1][err2]", errs["1001"], errs["1002"])
		return
	}

	err1001 := l.String("en-us", "err:1001")
	if err1001 != "err1" {
		t.Errorf("String err, got [%s], expected [err1]", err1001)
		return
	}

	host := l.String("en-us", "email:send:host")
	if host != "send" {
		t.Errorf("String err, got [%s], expected [send]", host)
		return
	}

	user := l.String("zh-cn", "email:sohu:user")
	if user != "管理员" {
		t.Errorf("String err, got [%s], expected [管理员]", user)
		return
	}
}
