// 时间函数
//   变更历史
//     2017-03-01  lixiaoya  新建
package utils

import (
	"time"
)

// CurTime 当前时间
//   参数
//     fmtStr: 返回当前时间的格式串，默认为yyyy-mm-dd hh:mi:ss
//   返回
//     当前时间
func CurTime(fmtStr ...string) string {
	str := "2006-01-02 15:04:05"
	if len(fmtStr) > 0 {
		str = fmtStr[0]
	}

	return time.Now().Format(str)
}

// AddDate 获取当前时间增加或减少给定年、月、日的时间
//   参数
//     years:  年份
//     months: 月份
//     days:   天数
//   返回
//     增加或减少的时间
func AddDate(years, months, days int) time.Time {
	nTime := time.Now()
	return nTime.AddDate(years, months, days)
}

// StrToTimeStamp 日期字符串转时间戳
//   参数
//     strTime: 日期字符串
//     fmtTime: 日期的格式，默认为2006-01-02 15:04:05
//   返回
//     时间戳
func StrToTimeStamp(timeStr string, timeFmt ...string) int64 {
	tmpFmt := "2006-01-02 15:04:05"
	if len(timeFmt) > 0 {
		tmpFmt = timeFmt[0]
	}
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(tmpFmt, timeStr, loc)
	return theTime.Unix()
}

// TimeStampToStr 时间戳转日期字符串
//   参数
//     args: 可以传如下参数
//       时间戳: 小于等于0时取当前时间，类型为int64
//       日期的格式: 默认为2006-01-02 15:04:05，类型为string
//   返回
//     日期字符串
func TimeStampToStr(args ...interface{}) string {
	argc := len(args)
	var timeStamp int64
	var ok bool
	timeFmt := "2006-01-02 15:04:05"
	if argc == 0 {
		return time.Now().Format(timeFmt)
	}
	if argc > 0 {
		timeStamp, ok = args[0].(int64)
		if !ok {
			return ""
		}
	}
	if argc > 1 {
		timeFmt, ok = args[1].(string)
		if !ok {
			return ""
		}
	}

	return time.Unix(timeStamp, 0).Format(timeFmt)
}

// TomrrowRest 获取从现在到明天凌晨的剩余时间，单位秒
//   参数
//     void
//   返回
//     到明天凌晨的剩余时间
func TomrrowRest() int64 {
	tom := AddDate(0, 0, 1)
	tomStr, _ := time.ParseInLocation("2006-01-02", tom.Format("2006-01-02"), time.Local)

	return tomStr.Unix() - time.Now().Unix()
}
