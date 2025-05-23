package site

import (
	"crawler/proto"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zeromicro/go-zero/core/logx"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type XHS struct{}

func (X *XHS) GetName() string {
	return "小红书"
}

func (X *XHS) GetSearchURL(keyword string) string {
	return fmt.Sprintf("https://www.xiaohongshu.com/search_result?keyword=%s", url.QueryEscape(keyword))
}

func (X *XHS) GetPostCardSelector() string {
	return "section.note-item"
}

func (X *XHS) GetPostLinkSelector() string {
	return "a.cover.mask.ld"
}

func (X *XHS) GetPostTitleSelector() string {
	return ".note-content .title"
}

func (X *XHS) GetBaseURL() string {
	return "https://www.xiaohongshu.com/explore"
}

func (X *XHS) GetPostContentSelector() string {
	return ".note-content .note-text"
}

func (X *XHS) GetAuthorSelector() string {
	return "span.username"
}

func (X *XHS) GetLikesSelector() string {
	return "span.like-wrapper .count"
}

func (X *XHS) GetChatsSelector() string {
	return "span.chat-wrapper .count"
}

func (X *XHS) GetCollectsSelector() string {
	return "span.collect-wrapper .count"
}

func (X *XHS) GetImagesSelector() string {
	return "div.img-container img"
}

func (X *XHS) GetTagsSelector() string {
	return "div.note-content .desc"
}

func (X *XHS) GetTimeSelector() string {
	return "div.note-content .date"
}

func (X *XHS) GetLocationSelector() string {
	return "div.note-content .date"
}

func (X *XHS) GetTimeAndLocationSelector() string {
	return "div.note-content .date"
}

func (X *XHS) ParsePostLinks(html string, minLikes int32) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	links := make([]string, 0)
	baseURL := "https://www.xiaohongshu.com"

	doc.Find(X.GetPostCardSelector()).Each(func(i int, s *goquery.Selection) {
		// 获取帖子点赞数
		likesText := s.Find(X.GetLikesSelector()).Text()
		likes := X.ParseNumber(likesText)

		// 只收集点赞数达到要求的帖子
		if likes >= minLikes {
			if href, exists := s.Find(X.GetPostLinkSelector()).Attr("href"); exists {
				// 如果是相对路径，添加基础URL
				if !strings.HasPrefix(href, "http") {
					href = baseURL + href
				}
				links = append(links, href)
			}
		}
	})

	return links, nil
}

func (X *XHS) ParsePostDetail(html string, includeImages bool) (*proto.PostItem, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	post := &proto.PostItem{}

	// 标题 - 处理转义字符
	title := strings.TrimSpace(doc.Find(X.GetPostTitleSelector()).Text())
	post.PostTitle = strings.ReplaceAll(title, "\\n", "\n")

	// 内容 - 处理转义字符
	content := strings.TrimSpace(doc.Find(X.GetPostContentSelector()).Text())
	post.PostContent = strings.ReplaceAll(content, "\\n", "\n")

	// 作者
	post.PostAuthor = strings.TrimSpace(doc.Find(X.GetAuthorSelector()).First().Text())

	// 解析时间和地点
	timeAndLocationText := strings.TrimSpace(doc.Find(X.GetTimeAndLocationSelector()).Text())
	logx.Debugf("时间和位置的文本：%v", timeAndLocationText)
	post.PostTime, post.PostLocation = X.ParseTimeAndLocation(timeAndLocationText)
	logx.Debugf("提取的时间和位置：%v + %v", time.Unix(post.PostTime, 0), post.PostLocation)

	// 统计数据
	post.PostLikes = X.ParseNumber(doc.Find(X.GetLikesSelector()).Text())
	post.PostCollects = X.ParseNumber(doc.Find(X.GetCollectsSelector()).Text())
	post.PostChats = X.ParseNumber(doc.Find(X.GetChatsSelector()).Text())

	// 标签 - 处理转义字符
	var tags []string
	doc.Find(X.GetTagsSelector()).Each(func(i int, s *goquery.Selection) {
		tag := strings.TrimSpace(s.Text())
		tag = strings.ReplaceAll(tag, "\\n", "\n")
		if tag != "" && !strings.Contains(tag, "#") {
			tag = "#" + tag
		}
		tags = append(tags, tag)
	})
	post.Tags = tags

	// 图片URL列表（如果需要）
	if includeImages {
		var images []string
		doc.Find(X.GetImagesSelector()).Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				images = append(images, src)
			} else if dataSrc, exists := s.Attr("data-src"); exists {
				images = append(images, dataSrc)
			}
		})
		post.Images = images
	}

	return post, nil
}

func (X *XHS) ParseComments(html string, count int32, repliesCount int32) ([]*proto.Comment, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var comments []*proto.Comment
	commentCount := 0

	doc.Find("div.comment-item").Each(func(i int, s *goquery.Selection) {
		if int32(commentCount) >= count {
			return
		}

		comment := &proto.Comment{}

		// 评论ID
		if id, exists := s.Attr("id"); exists {
			comment.CommentId = id
		}

		// 评论内容
		comment.CommentContent = strings.TrimSpace(s.Find(".content .note-text").Text())

		// 评论作者
		comment.CommentAuthor = strings.TrimSpace(s.Find(".author").Text())

		// 点赞数
		comment.CommentLikes = X.ParseNumber(s.Find("div.like .count").Text())

		// 回复数
		comment.CommentReplies = X.ParseNumber(s.Find("div.reply .count").Text())

		// 评论时间
		timeText := strings.TrimSpace(s.Find(".date").Text())
		if commentTime, err := X.ParseTime(timeText); err == nil {
			comment.CommentTime = commentTime
		}

		// 评论位置
		comment.CommentLocation = strings.TrimSpace(s.Find(".location").Text())

		// 解析回复（如果有）
		if repliesCount > 0 {
			var replies []*proto.Comment
			replyCounter := 0

			s.Find(".reply-container .reply-item").Each(func(j int, r *goquery.Selection) {
				if int32(replyCounter) >= repliesCount {
					return
				}

				reply := &proto.Comment{}

				// 回复ID
				if id, exists := r.Attr("id"); exists {
					reply.CommentId = id
				}

				// 回复内容
				reply.CommentContent = strings.TrimSpace(r.Find(".content .note-text").Text())

				// 回复作者
				reply.CommentAuthor = strings.TrimSpace(r.Find(".author").Text())

				// 回复点赞
				reply.CommentLikes = X.ParseNumber(r.Find("div.like .count").Text())

				// 回复时间
				replyTimeText := strings.TrimSpace(r.Find(".date").Text())
				if replyTime, err := X.ParseTime(replyTimeText); err == nil {
					reply.CommentTime = replyTime
				}

				replies = append(replies, reply)
				replyCounter++
			})

			comment.Replies = replies
		}

		comments = append(comments, comment)
		commentCount++
	})

	return comments, nil
}

func (X *XHS) ParseNumber(numStr string) int32 {
	// 清理字符串，只保留数字
	re := regexp.MustCompile(`\d+`)
	matches := re.FindStringSubmatch(numStr)
	if len(matches) > 0 {
		num, err := strconv.Atoi(matches[0])
		if err == nil {
			return int32(num)
		}
	}

	// 检查是否包含"万"
	if strings.Contains(numStr, "万") {
		re := regexp.MustCompile(`(\d+(\.\d+)?)万`)
		matches := re.FindStringSubmatch(numStr)
		if len(matches) >= 2 {
			num, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				return int32(num * 10000)
			}
		}
	}

	return 0
}

func (X *XHS) ParsePostIDFromURL(url string) string {
	//https://www.xiaohongshu.com/search_result/676425e6000000000b0164ab?xsec_token=
	re := regexp.MustCompile(`/search_result/([^?]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func (X *XHS) NeedsLogin(html string) bool {
	return strings.Contains(html, "登录") &&
		strings.Contains(html, "注册") &&
		!strings.Contains(html, "退出登录")
}

func (X *XHS) Login() error {
	// 实际项目中需要实现登录逻辑
	// 可能涉及到扫码登录、账号密码登录等
	// 这里简化处理，返回未实现错误
	return errors.New("登录功能未实现")
}

// 预编译的正则表达式
// 格式：
//
//	编辑于 6天前 上海
//	6天前 浙江
//	05-14 湖北
//	2024-05-05
//	昨天 16:27 四川
//	5小时前 四川
var (
	// 时间相关正则
	dayAgoRegex        = regexp.MustCompile(`(\d+)天前`)
	hourAgoRegex       = regexp.MustCompile(`(\d+)小时前`)
	monthDayRegex      = regexp.MustCompile(`(\d{2})-(\d{2})`)
	yearMonthDayRegex  = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	todayTimeRegex     = regexp.MustCompile(`今天\s+(\d{1,2}):(\d{2})`)
	yesterdayTimeRegex = regexp.MustCompile(`昨天\s+(\d{1,2}):(\d{2})`)
)

// ParseTimeAndLocation 小红书的location和time放在一起的所以可以一起解析
func (X *XHS) ParseTimeAndLocation(str string) (timestamp int64, location string) {
	// 清理字符串
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "编辑于")
	str = strings.TrimSpace(str)

	// 优先检查年月日格式(YYYY-MM-DD)，因为这是完整日期格式
	if matches := yearMonthDayRegex.FindStringSubmatch(str); len(matches) == 4 {
		year, errY := strconv.Atoi(matches[1])
		month, errM := strconv.Atoi(matches[2])
		day, errD := strconv.Atoi(matches[3])
		if errY == nil && errM == nil && errD == nil {
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			timestamp = date.Unix()

			// 移除日期部分，剩余的才是位置
			locationStr := yearMonthDayRegex.ReplaceAllString(str, "")
			location = strings.TrimSpace(locationStr)
			return
		}
	}

	// 其他时间格式
	patterns := []*regexp.Regexp{
		dayAgoRegex,        // 6天前
		hourAgoRegex,       // 5小时前
		monthDayRegex,      // 05-14
		todayTimeRegex,     // 今天 16:27
		yesterdayTimeRegex, // 昨天 16:27
	}

	// 尝试匹配其他时间格式
	for _, re := range patterns {
		matches := re.FindStringSubmatch(str)
		if len(matches) >= 2 {
			// 从原字符串中提取时间部分
			timeStr := matches[0]
			// 将匹配到的时间字符串转换为Unix时间戳
			parsedTime, err := X.ParseTime(timeStr)
			if err == nil {
				timestamp = parsedTime
			}

			// 从原字符串中移除时间部分，剩余的就是位置
			locationStr := re.ReplaceAllString(str, "")
			location = strings.TrimSpace(locationStr)
			return
		}
	}

	// 未匹配到任何已知时间格式，可能是其他未知格式或只有位置信息
	// 此时根据空格分割，假设最后一部分是位置
	parts := strings.Fields(str)
	if len(parts) > 0 {
		location = parts[len(parts)-1]
		if len(parts) > 1 {
			// 尝试解析除最后一部分外的内容作为时间
			timeStr := strings.Join(parts[:len(parts)-1], " ")
			parsedTime, err := X.ParseTime(timeStr)
			if err == nil {
				timestamp = parsedTime
			}
		}
	}

	return
}

func (X *XHS) ParseTime(timeStr string) (int64, error) {
	now := time.Now()

	// 移除可能存在的"编辑于"前缀
	timeStr = strings.TrimPrefix(strings.TrimSpace(timeStr), "编辑于")
	timeStr = strings.TrimSpace(timeStr)

	// 处理"今天 HH:MM"格式
	if matches := todayTimeRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])
		if errH == nil && errM == nil {
			today := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
			return today.Unix(), nil
		}
	}

	// 处理"昨天 HH:MM"格式
	if matches := yesterdayTimeRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, errH := strconv.Atoi(matches[1])
		minute, errM := strconv.Atoi(matches[2])
		if errH == nil && errM == nil {
			yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, hour, minute, 0, 0, now.Location())
			return yesterday.Unix(), nil
		}
	}

	// 处理"X天前"格式
	if matches := dayAgoRegex.FindStringSubmatch(timeStr); len(matches) == 2 {
		days, err := strconv.Atoi(matches[1])
		if err == nil {
			daysAgo := time.Date(now.Year(), now.Month(), now.Day()-days, 0, 0, 0, 0, now.Location())
			return daysAgo.Unix(), nil
		}
	}

	// 处理"X小时前"格式
	if matches := hourAgoRegex.FindStringSubmatch(timeStr); len(matches) == 2 {
		hours, err := strconv.Atoi(matches[1])
		if err == nil {
			hoursAgo := now.Add(time.Duration(-hours) * time.Hour)
			return hoursAgo.Unix(), nil
		}
	}

	// 处理年月日格式 (YYYY-MM-DD)
	if matches := yearMonthDayRegex.FindStringSubmatch(timeStr); len(matches) == 4 {
		year, errY := strconv.Atoi(matches[1])
		month, errM := strconv.Atoi(matches[2])
		day, errD := strconv.Atoi(matches[3])
		if errY == nil && errM == nil && errD == nil {
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location())
			return date.Unix(), nil
		}
	}

	// 处理月日格式 (MM-DD)
	if matches := monthDayRegex.FindStringSubmatch(timeStr); len(matches) == 3 {
		month, errM := strconv.Atoi(matches[1])
		day, errD := strconv.Atoi(matches[2])
		if errM == nil && errD == nil {
			// 假设是当年
			date := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location())
			// 如果日期在未来，则可能是去年的日期
			if date.After(now) {
				date = time.Date(now.Year()-1, time.Month(month), day, 0, 0, 0, 0, now.Location())
			}
			return date.Unix(), nil
		}
	}

	return 0, fmt.Errorf("无法解析时间格式: %s", timeStr)
}

func (X *XHS) ParseLocation(locationStr string) string {
	// 清理字符串
	locationStr = strings.TrimSpace(locationStr)
	locationStr = strings.TrimPrefix(locationStr, "编辑于")
	locationStr = strings.TrimSpace(locationStr)

	// 处理各种日期格式模式并去除它们
	patterns := []*regexp.Regexp{
		dayAgoRegex,
		hourAgoRegex,
		monthDayRegex,
		yearMonthDayRegex,
		todayTimeRegex,
		yesterdayTimeRegex,
	}

	for _, re := range patterns {
		locationStr = re.ReplaceAllString(locationStr, "")
		locationStr = strings.TrimSpace(locationStr)
	}

	return locationStr
}
