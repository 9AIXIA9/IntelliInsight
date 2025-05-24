package site

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/zeromicro/go-zero/core/logx"
)

// 格式：
//       注意：所有的位置都可能没有 也都可能有
//		编辑于 6天前 上海
//		6天前 浙江
//		05-14 湖北
//		2024-05-05
//		昨天 16:27 四川
//		5小时前 四川
//	    8分钟前 广东
//	    刚刚 江西
//      还有种情况可能日期和位置为贴在一起 比如 05-08浙江

// 预编译的正则表达式
var (
	dayAgoRegex        = regexp.MustCompile(`(\d+)天前`)
	hourAgoRegex       = regexp.MustCompile(`(\d+)小时前`)
	minuteAgoRegex     = regexp.MustCompile(`(\d+)分钟前`)
	justNowRegex       = regexp.MustCompile(`刚刚`)
	monthDayRegex      = regexp.MustCompile(`(\d{2})-(\d{2})`)
	monthDayLocRegex   = regexp.MustCompile(`(\d{2})-(\d{2})([^\s\d-]+)`) // 匹配如 "05-08浙江" 的格式
	yearMonthDayRegex  = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	todayTimeRegex     = regexp.MustCompile(`今天\s*(\d{1,2}):(\d{2})`)
	yesterdayTimeRegex = regexp.MustCompile(`昨天\s*(\d{1,2}):(\d{2})`)
	timeWithColonRegex = regexp.MustCompile(`(\d{1,2}):(\d{2})`) // 匹配时间格式 HH:MM
)

// ParseLocation 解析小红书中的位置信息
func (X *XHS) ParseLocation(locationStr string) string {
	original := locationStr
	locationStr = strings.TrimSpace(locationStr)

	// 如果字符串为空，直接返回空字符串
	if locationStr == "" {
		return ""
	}

	// 清理可能包含的时间信息
	// 移除"编辑于"前缀
	locationStr = strings.TrimPrefix(locationStr, "编辑于")
	locationStr = strings.TrimSpace(locationStr)

	// 移除可能的时间表达式，只保留地名部分
	patterns := []string{
		`\d+天前`,
		`\d+小时前`,
		`\d+分钟前`,
		`刚刚`,
		`\d{4}-\d{2}-\d{2}`,
		`今天\s*\d{1,2}:\d{2}`,
		`昨天\s*\d{1,2}:\d{2}`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		locationStr = re.ReplaceAllString(locationStr, "")
		locationStr = strings.TrimSpace(locationStr)
	}

	// 处理特殊情况：MM-DD直接跟地点的格式，如"05-08浙江"
	if matches := monthDayLocRegex.FindStringSubmatch(locationStr); len(matches) == 4 {
		locationStr = matches[3]
	} else {
		// 处理普通的MM-DD格式，移除日期部分
		locationStr = monthDayRegex.ReplaceAllString(locationStr, "")
		locationStr = strings.TrimSpace(locationStr)
	}

	// 移除时间格式 HH:MM
	locationStr = timeWithColonRegex.ReplaceAllString(locationStr, "")
	locationStr = strings.TrimSpace(locationStr)

	logx.Debugf("位置解析: %q => %q", original, locationStr)
	return locationStr
}

// ParseTimeAndLocation 解析小红书中时间和位置信息
func (X *XHS) ParseTimeAndLocation(str string) (timex time.Time, location string) {
	original := str
	logx.Debugf("原始时间和位置文本: %q", str)

	// 清理字符串
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "编辑于")
	str = strings.TrimSpace(str)

	if str == "" {
		logx.Debug("输入为空，返回空时间和位置")
		return time.Time{}, ""
	}

	// 检查"刚刚"的情况
	if justNowRegex.MatchString(str) {
		timex = time.Now()
		location = X.ParseLocation(strings.Replace(str, "刚刚", "", 1))
		logx.Debugf("匹配'刚刚'格式 => 时间: %v, 位置: %q", timex.Format("2006-01-02 15:04:05"), location)
		return
	}

	// 检查"X分钟前"的情况
	if matches := minuteAgoRegex.FindStringSubmatch(str); len(matches) == 2 {
		minutes, err := strconv.Atoi(matches[1])
		if err == nil {
			timex = time.Now().Add(time.Duration(-minutes) * time.Minute)
			location = X.ParseLocation(minuteAgoRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'%s分钟前'格式 => 时间: %v, 位置: %q", matches[1], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查"X小时前"的情况
	if matches := hourAgoRegex.FindStringSubmatch(str); len(matches) == 2 {
		hours, err := strconv.Atoi(matches[1])
		if err == nil {
			timex = time.Now().Add(time.Duration(-hours) * time.Hour)
			location = X.ParseLocation(hourAgoRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'%s小时前'格式 => 时间: %v, 位置: %q", matches[1], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查"X天前"的情况
	if matches := dayAgoRegex.FindStringSubmatch(str); len(matches) == 2 {
		days, err := strconv.Atoi(matches[1])
		if err == nil {
			now := time.Now()
			// 使用AddDate替代手动计算日期
			timex = now.AddDate(0, 0, -days)
			location = X.ParseLocation(dayAgoRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'%s天前'格式 => 时间: %v, 位置: %q", matches[1], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查MM-DD直接跟地点的情况，如"05-08浙江"
	if matches := monthDayLocRegex.FindStringSubmatch(str); len(matches) == 4 {
		month, errM := strconv.Atoi(matches[1])
		day, errD := strconv.Atoi(matches[2])
		location = matches[3]

		if errM == nil && errD == nil {
			now := time.Now()
			// 假设是当年
			timex = time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local)
			// 如果日期在未来，则可能是去年的日期
			if timex.After(now) {
				timex = time.Date(now.Year()-1, time.Month(month), day, 0, 0, 0, 0, time.Local)
			}
			logx.Debugf("匹配'MM-DD位置'格式(%s-%s%s) => 时间: %v, 位置: %q", matches[1], matches[2], matches[3], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查年月日格式(YYYY-MM-DD)
	if matches := yearMonthDayRegex.FindStringSubmatch(str); len(matches) == 4 {
		year, errY := strconv.Atoi(matches[1])
		month, errM := strconv.Atoi(matches[2])
		day, errD := strconv.Atoi(matches[3])

		if errY == nil && errM == nil && errD == nil {
			timex = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			location = X.ParseLocation(yearMonthDayRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'YYYY-MM-DD'格式(%s-%s-%s) => 时间: %v, 位置: %q", matches[1], matches[2], matches[3], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查"今天 HH:MM"格式
	if matches := todayTimeRegex.FindStringSubmatch(str); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])

		if errH == nil && errM == nil {
			now := time.Now()
			timex = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
			location = X.ParseLocation(todayTimeRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'今天 HH:MM'格式(%s:%s) => 时间: %v, 位置: %q", matches[1], matches[2], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查"昨天 HH:MM"格式
	if matches := yesterdayTimeRegex.FindStringSubmatch(str); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])

		if errH == nil && errM == nil {
			now := time.Now()
			timex = time.Date(now.Year(), now.Month(), now.Day()-1, hour, minute, 0, 0, time.Local)
			location = X.ParseLocation(yesterdayTimeRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'昨天 HH:MM'格式(%s:%s) => 时间: %v, 位置: %q", matches[1], matches[2], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查月日格式(MM-DD)
	if matches := monthDayRegex.FindStringSubmatch(str); len(matches) == 3 {
		month, errM := strconv.Atoi(matches[1])
		day, errD := strconv.Atoi(matches[2])

		if errM == nil && errD == nil {
			now := time.Now()
			// 假设是当年
			timex = time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local)

			// 如果日期在未来，则可能是去年的日期
			if timex.After(now) {
				timex = time.Date(now.Year()-1, time.Month(month), day, 0, 0, 0, 0, time.Local)
			}

			// 移除日期部分，剩余的是位置
			location = X.ParseLocation(monthDayRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'MM-DD'格式(%s-%s) => 时间: %v, 位置: %q", matches[1], matches[2], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 检查时间格式(HH:MM)，通常与其他文本混合
	if matches := timeWithColonRegex.FindStringSubmatch(str); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])

		if errH == nil && errM == nil {
			now := time.Now()
			timex = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
			location = X.ParseLocation(timeWithColonRegex.ReplaceAllString(str, ""))
			logx.Debugf("匹配'HH:MM'格式(%s:%s) => 时间: %v, 位置: %q", matches[1], matches[2], timex.Format("2006-01-02 15:04:05"), location)
			return
		}
	}

	// 未能识别任何时间格式，可能整个字符串就是位置
	// 或者是未知的时间格式
	parts := strings.Fields(str)
	if len(parts) > 0 {
		// 尝试识别最后一个部分是否是地名（简单判断：不包含数字）
		lastPart := parts[len(parts)-1]
		if !containsDigit(lastPart) {
			location = lastPart
			timeStr := strings.Join(parts[:len(parts)-1], " ")
			if timeStr != "" {
				// 尝试解析剩余部分作为时间
				if parsedTime, err := X.ParseTime(timeStr); err == nil {
					timex = parsedTime
					logx.Debugf("分离时间和位置 => 时间: %v, 位置: %q", timex.Format("2006-01-02 15:04:05"), location)
					return
				}
			}
		} else {
			// 可能整个字符串都是时间描述
			if parsedTime, err := X.ParseTime(str); err == nil {
				timex = parsedTime
				logx.Debugf("整体解析为时间 => 时间: %v, 位置: %q", timex.Format("2006-01-02 15:04:05"), location)
				return
			}
		}
	}

	// 如果前面都无法解析，尝试将整个字符串作为位置处理
	if location == "" && !containsDigit(str) {
		location = str
		logx.Debugf("整体解析为位置 => 位置: %q", location)
	}

	logx.Debugf("解析结果 - 原始:%q => 时间:%v, 位置:%q", original, timex.Format("2006-01-02 15:04:05"), location)
	return
}

// ParseTime 解析各种格式的时间字符串
func (X *XHS) ParseTime(timeStr string) (time.Time, error) {
	original := timeStr
	now := time.Now()
	timeStr = strings.TrimSpace(timeStr)

	if timeStr == "" {
		return time.Time{}, fmt.Errorf("空时间字符串")
	}

	// 处理"刚刚"
	if justNowRegex.MatchString(timeStr) {
		logx.Debugf("时间解析: %q => 刚刚(当前时间)", original)
		return now, nil
	}

	// 处理"X分钟前"
	if matches := minuteAgoRegex.FindStringSubmatch(timeStr); len(matches) == 2 {
		minutes, err := strconv.Atoi(matches[1])
		if err == nil {
			t := now.Add(time.Duration(-minutes) * time.Minute)
			logx.Debugf("时间解析: %q => %d分钟前(%v)", original, minutes, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理"X小时前"
	if matches := hourAgoRegex.FindStringSubmatch(timeStr); len(matches) == 2 {
		hours, err := strconv.Atoi(matches[1])
		if err == nil {
			t := now.Add(time.Duration(-hours) * time.Hour)
			logx.Debugf("时间解析: %q => %d小时前(%v)", original, hours, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理"X天前"
	if matches := dayAgoRegex.FindStringSubmatch(timeStr); len(matches) == 2 {
		days, err := strconv.Atoi(matches[1])
		if err == nil {
			t := now.AddDate(0, 0, -days)
			logx.Debugf("时间解析: %q => %d天前(%v)", original, days, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理"今天 HH:MM"
	if matches := todayTimeRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])
		if errH == nil && errM == nil {
			t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
			logx.Debugf("时间解析: %q => 今天%d:%d(%v)", original, hour, minute, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理"昨天 HH:MM"
	if matches := yesterdayTimeRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])
		if errH == nil && errM == nil {
			t := time.Date(now.Year(), now.Month(), now.Day()-1, hour, minute, 0, 0, time.Local)
			logx.Debugf("时间解析: %q => 昨天%d:%d(%v)", original, hour, minute, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理年月日格式 (YYYY-MM-DD)
	if matches := yearMonthDayRegex.FindStringSubmatch(timeStr); len(matches) == 4 {
		year, errY := strconv.Atoi(matches[1])
		month, errM := strconv.Atoi(matches[2])
		day, errD := strconv.Atoi(matches[3])
		if errY == nil && errM == nil && errD == nil {
			t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			logx.Debugf("时间解析: %q => %d-%02d-%02d(%v)", original, year, month, day, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理月日格式 (MM-DD)
	if matches := monthDayRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		month, errM := strconv.Atoi(matches[1])
		day, errD := strconv.Atoi(matches[2])
		if errM == nil && errD == nil {
			// 假设是当年
			t := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local)
			// 如果日期在未来，则可能是去年的日期
			if t.After(now) {
				t = time.Date(now.Year()-1, time.Month(month), day, 0, 0, 0, 0, time.Local)
			}
			logx.Debugf("时间解析: %q => %02d-%02d(%v)", original, month, day, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	// 处理时间格式 (HH:MM)
	if matches := timeWithColonRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])
		if errH == nil && errM == nil {
			t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)
			logx.Debugf("时间解析: %q => %d:%02d(%v)", original, hour, minute, t.Format("2006-01-02 15:04:05"))
			return t, nil
		}
	}

	logx.Debugf("无法解析时间格式: %q", original)
	return time.Time{}, fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// containsDigit 检查字符串是否包含数字
func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
