// 字符串处理函数测试
//   变更历史
//     2017-03-30  lixiaoya  新建
package utils

import (
	"testing"
	"fmt"
)

// TestDelRepeat DelRepeat测试
func TestDelRepeat(t *testing.T) {
	s := "///aa///bb///cc///dd//"
	res := DelRepeat(s, '/')
	if res != "/aa/bb/cc/dd/" {
		t.Errorf("TestDelRepeat failed. Got %s, expected /aa/bb/cc/dd/.", res)
		return
	}

	s = "////"
	res = DelRepeat(s, '/')
	if res != "/" {
		t.Errorf("TestDelRepeat failed. Got %s, expected /.", res)
		return
	}
}

// TestReplace Replace测试
func TestReplace(t *testing.T) {
	s := "11m122m233m344m155m299"
	olds := []string{"m1", "m2", "m3"}
	news := []string{"+", "-", "="}

	res := Replace(s, olds, news, -1)
	if res != "11+22-33=44+55-99" {
		t.Errorf("TestDelRepeat failed. Got %s, expected 11+22-33=44+55-99.", res)
		return
	}
}

// TestSubstr Substr测试
func TestSubstr(t *testing.T) {
	s := "1234567890"
	sub := Substr(s, 0, 30)
	if sub != s {
		t.Errorf("Substr err, Got %s, expected %s", sub, s)
		return
	}

	sub = Substr(s, 0, 3)
	if sub != "123" {
		t.Errorf("Substr err, Got %s, expected %s", sub, "123")
		return
	}

	sub = Substr(s, 0, 10)
	if sub != s {
		t.Errorf("Substr err, Got %s, expected %s", sub, s)
		return
	}

	sub = Substr(s, 2, 5)
	if sub != "34567" {
		t.Errorf("Substr err, Got %s, expected %s", sub, "34567")
		return
	}

	sub = Substr(s, -3, 30)
	if sub != "890" {
		t.Errorf("Substr err, Got %s, expected %s", sub, "890")
		return
	}

	sub = Substr(s, -1, 10)
	if sub != "0" {
		t.Errorf("Substr err, Got %s, expected %s", sub, "0")
		return
	}

	sub = Substr(s, -100, 10)
	if sub != s {
		t.Errorf("Substr err, Got %s, expected %s", sub, s)
		return
	}

	sub = Substr(s, -100, 5)
	if sub != "12345" {
		t.Errorf("Substr err, Got %s, expected %s", sub, "12345")
		return
	}

	sub = Substr(s, -5, 2)
	if sub != "67" {
		t.Errorf("Substr err, Got %s, expected %s", sub, "67")
		return
	}
}

// TestEmpty Empty测试
func TestEmpty(t *testing.T) {
	str := "HelloWorld!"
	if Empty(str) {
		t.Error("Empty err, Got true, expected false")
		return
	}

	str = ""
	if !Empty(str) {
		t.Error("Empty err, Got false, expected true")
		return
	}
}

// TestGetSafeSql GetSafeSql测试
func TestGetSafeSql(t *testing.T) {
	str := "Select 11 seleCt 22 upDate 33 information_schema.columns 44 table_sChema 55 net user BBBSelect Update"
	dst := " 11  22  33  44  55  BBBSelect "
	str1 := GetSafeSql(str)
	if str1 != dst {
		t.Errorf("GetSafeSql err, Got %s, expected %s", str1, dst)
		return
	}
}

// TestHtml2Str Html2Str测试
func TestHtml2Str(t *testing.T) {
	html := `<html>hello</html>`
	str := Html2Str(html)
	if str != "hello" {
		t.Errorf("Html2str err, Got %s, expected %s", str, "hello")
		return
	}

	html = `hello<style>ssss</style>`
	str = Html2Str(html)
	if str != "hello" {
		t.Errorf("Html2str err, Got %s, expected %s", str, "hello")
		return
	}

	html = `hello<script>ssss</script>`
	str = Html2Str(html)
	if str != "hello" {
		t.Errorf("Html2str err, Got %s, expected %s", str, "hello")
		return
	}

	html = `111<br />222`
	str = Html2Str(html)
	if str != "111\n222" {
		t.Errorf("Html2str err, Got %s, expected %s", str, "111\n222")
		return
	}
}

// TestStrSplit StrSplit测试
func TestStrSplit(t *testing.T) {
	src := "abcdef"
	arr := StrSplit(src, 0)
	fmt.Println("0>>", arr)

	arr = StrSplit(src, 1)
	fmt.Println("1>>", arr)

	arr = StrSplit(src, 2)
	fmt.Println("2>>", arr)

	arr = StrSplit(src, 3)
	fmt.Println("3>>", arr)

	arr = StrSplit(src, 4)
	fmt.Println("4>>", arr)

	arr = StrSplit(src, 5)
	fmt.Println("5>>", arr)

	arr = StrSplit(src, 6)
	fmt.Println("6>>", arr)

	arr = StrSplit(src, 7)
	fmt.Println("6>>", arr)

	arr = StrSplit(src, 10)
	fmt.Println("10>>", arr)
}
