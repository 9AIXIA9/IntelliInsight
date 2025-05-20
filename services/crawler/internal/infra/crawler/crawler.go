package crawler

import (
	"crawler/internal/domain"
	"crawler/proto"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/encoding/simplifiedchinese"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Crawler 小红书爬虫结构体
type Crawler struct {
}

// NewCrawler 创建爬虫实例
func NewCrawler() domain.Crawler {
	return &Crawler{}
}

func (c *Crawler) CollectPostLinks(resource domain.ResourceUnit, site domain.Site, keyword string, count int32, minLikes int32) (postLinks []string, err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Crawler) CollectPostDetail(resource domain.ResourceUnit, site domain.Site, postURL string, opts ...*domain.CollectPostDetailOption) (*proto.PostItem, error) {
	//TODO implement me
	panic("implement me")
}

// normalizeKeyword 规范化关键词，将关键词编码为小红书搜索URL可接受的格式
func normalizeKeyword(keyword string) string {
	// URL编码关键词（UTF-8）
	keywordTempCode := url.QueryEscape(keyword)
	// 将UTF-8转换为GB2312编码
	gb2312Bytes, err := simplifiedchinese.HZGB2312.NewEncoder().Bytes([]byte(keywordTempCode))
	if err != nil {
		logx.Errorf("编码转换错误: %v", err)
		return keywordTempCode
	}
	keywordEncode := url.QueryEscape(string(gb2312Bytes))
	return keywordEncode
}

// parseTime 解析帖子发布时间
func parseTime(timeStr string) (int64, error) {
	now := time.Now()

	// 处理"今天"、"昨天"等相对日期
	if strings.Contains(timeStr, "今天") {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return today.Unix(), nil
	} else if strings.Contains(timeStr, "昨天") {
		yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
		return yesterday.Unix(), nil
	} else if strings.Contains(timeStr, "前") && strings.Contains(timeStr, "天") {
		// 处理"N天前"
		re := regexp.MustCompile(`(\d+)天前`)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) > 1 {
			days, _ := strconv.Atoi(matches[1])
			pastDate := time.Date(now.Year(), now.Month(), now.Day()-days, 0, 0, 0, 0, now.Location())
			return pastDate.Unix(), nil
		}
	} else if strings.Contains(timeStr, "小时前") {
		// 处理"N小时前"
		re := regexp.MustCompile(`(\d+)小时前`)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) > 1 {
			hours, _ := strconv.Atoi(matches[1])
			pastTime := now.Add(time.Duration(-hours) * time.Hour)
			return pastTime.Unix(), nil
		}
	} else if strings.Contains(timeStr, "-") {
		// 处理MM-DD格式
		re := regexp.MustCompile(`(\d{2})-(\d{2})`)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) > 2 {
			month, _ := strconv.Atoi(matches[1])
			day, _ := strconv.Atoi(matches[2])

			// 假设是今年的日期
			publishTime := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location())

			// 如果计算出的时间在未来，可能是去年的日期
			if publishTime.After(now) {
				publishTime = time.Date(now.Year()-1, time.Month(month), day, 0, 0, 0, 0, now.Location())
			}

			return publishTime.Unix(), nil
		}
	}

	// 无法解析，返回当前时间
	return now.Unix(), fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// parseNumber 解析数量
func parseNumber(numStr string) int {
	// 处理带"万"的数字
	if strings.Contains(numStr, "万") {
		numStr = strings.ReplaceAll(numStr, "万", "")
		num, err := strconv.ParseFloat(numStr, 64)
		if err == nil {
			return int(num * 10000)
		}
	}

	// 处理带"w"的数字
	if strings.Contains(numStr, "w") {
		numStr = strings.ReplaceAll(numStr, "w", "")
		num, err := strconv.ParseFloat(numStr, 64)
		if err == nil {
			return int(num * 10000)
		}
	}

	// 处理普通数字
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(numStr, -1)
	if len(matches) > 0 {
		num, err := strconv.Atoi(matches[0])
		if err == nil {
			return num
		}
	}

	return 0
}
