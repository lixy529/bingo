// 时间函数测试
//   变更历史
//     2017-03-01  lixiaoya  新建
package utils

import (
	"fmt"
	"testing"
)

// TestCurTime CurTime测试
func TestCurTime(t *testing.T) {
	cur := CurTime()
	fmt.Println(cur)
}

// TestAddDate AddDate测试
func TestAddDate(t *testing.T) {
	dt := AddDate(-1, -1, -1)
	dd := dt.Format("20060102")
	fmt.Println(dd)
}

// TestStrToTimeStamp StrToTimeStamp测试
func TestStrToTimeStamp(t *testing.T) {
	strTime := "2017-07-06 16:27:28"
	st := StrToTimeStamp(strTime)
	if st != 1499329648 {
		t.Errorf("StrToTimeStamp err, Got %d, expected 1499329648", st)
		return
	}

	strTime = "20170706162728"
	fmtTime := "20060102150405"
	st = StrToTimeStamp(strTime, fmtTime)
	if st != 1499329648 {
		t.Errorf("StrToTimeStamp err, Got %d, expected 1499329648", st)
		return
	}
}

// TestTimeStampToStr TimeStampToStr测试
func TestTimeStampToStr(t *testing.T) {
	var timeStamp int64 = 1499329648
	strTime := TimeStampToStr(timeStamp)
	if strTime != "2017-07-06 16:27:28" {
		t.Errorf("TimeStampToStr err, Got %s, expected 2017-07-06 16:27:28", strTime)
		return
	}

	fmtTime := "20060102150405"
	strTime = TimeStampToStr(timeStamp, fmtTime)
	if strTime != "20170706162728" {
		t.Errorf("TimeStampToStr err, Got %s, expected 20170706162728", strTime)
		return
	}
}

// TestTomrrowRest TomrrowRest测试
func TestTomrrowRest(t *testing.T) {
	dd := TomrrowRest()
	fmt.Println(dd)
}
