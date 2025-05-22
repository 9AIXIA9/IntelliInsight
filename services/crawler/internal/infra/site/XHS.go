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
	return "div.note-content .title"
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
	return ".date"
}

func (X *XHS) GetLocationSelector() string {
	return ".location"
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
	post.Title = strings.ReplaceAll(title, "\\n", "\n")

	// 内容 - 处理转义字符
	content := strings.TrimSpace(doc.Find(X.GetPostContentSelector()).Text())
	post.Content = strings.ReplaceAll(content, "\\n", "\n")

	// 作者
	post.Author = strings.TrimSpace(doc.Find(X.GetAuthorSelector()).First().Text())

	// 时间和地点
	timeText := strings.TrimSpace(doc.Find(X.GetTimeSelector()).Text())
	locationText := strings.TrimSpace(doc.Find(X.GetLocationSelector()).Text())

	// 解析时间
	logx.Infof("timeText:%v", timeText)
	publishTime, err := X.ParseTime(timeText)
	if err == nil {
		post.PublishTime = publishTime
	}
	logx.Infof("publishTime:%v", time.Unix(publishTime, 0))
	post.Location = locationText

	// 统计数据
	post.Likes = X.ParseNumber(doc.Find(X.GetLikesSelector()).Text())
	post.Collects = X.ParseNumber(doc.Find(X.GetCollectsSelector()).Text())
	post.Chats = X.ParseNumber(doc.Find(X.GetChatsSelector()).Text())

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
		comment.Content = strings.TrimSpace(s.Find(".content .note-text").Text())

		// 评论作者
		comment.Author = strings.TrimSpace(s.Find(".author").Text())

		// 点赞数
		comment.Likes = X.ParseNumber(s.Find("div.like .count").Text())

		// 回复数
		comment.RepliesCount = X.ParseNumber(s.Find("div.reply .count").Text())

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
				reply.Content = strings.TrimSpace(r.Find(".content .note-text").Text())

				// 回复作者
				reply.Author = strings.TrimSpace(r.Find(".author").Text())

				// 回复点赞
				reply.Likes = X.ParseNumber(r.Find("div.like .count").Text())

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

func (X *XHS) ParseTime(timeStr string) (int64, error) {
	//编辑于 6天前 上海
	//6天前 浙江
	//05-14 湖北
	//2024-05-05
	//昨天 16:27 四川
	//5小时前 四川
	now := time.Now()

	// 移除可能存在的"编辑于"前缀
	timeStr = strings.TrimPrefix(strings.TrimSpace(timeStr), "编辑于")
	timeStr = strings.TrimSpace(timeStr)

	// 处理相对时间格式：今天、昨天
	if strings.Contains(timeStr, "今天") {
		parts := strings.Split(timeStr, " ")
		if len(parts) >= 2 {
			timeOnly := parts[1]
			hourMin := strings.Split(timeOnly, ":")
			if len(hourMin) == 2 {
				hour, _ := strconv.Atoi(hourMin[0])
				minute, _ := strconv.Atoi(hourMin[1])
				today := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
				return today.Unix(), nil
			}
		}
	}

	if strings.Contains(timeStr, "昨天") {
		parts := strings.Split(timeStr, " ")
		if len(parts) >= 2 {
			timeOnly := parts[1]
			hourMin := strings.Split(timeOnly, ":")
			if len(hourMin) == 2 {
				hour, _ := strconv.Atoi(hourMin[0])
				minute, _ := strconv.Atoi(hourMin[1])
				yesterday := time.Date(now.Year(), now.Month(), now.Day()-1, hour, minute, 0, 0, now.Location())
				return yesterday.Unix(), nil
			}
		}
	}

	// 处理"X天前"格式
	if strings.Contains(timeStr, "天前") {
		re := regexp.MustCompile(`(\d+)天前`)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) == 2 {
			days, _ := strconv.Atoi(matches[1])
			daysAgo := time.Date(now.Year(), now.Month(), now.Day()-days, 0, 0, 0, 0, now.Location())
			return daysAgo.Unix(), nil
		}
	}

	// 处理"X小时前"格式
	if strings.Contains(timeStr, "小时前") {
		re := regexp.MustCompile(`(\d+)小时前`)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) == 2 {
			hours, _ := strconv.Atoi(matches[1])
			hoursAgo := now.Add(time.Duration(-hours) * time.Hour)
			return hoursAgo.Unix(), nil
		}
	}

	// 处理月日格式 (MM-DD)
	re := regexp.MustCompile(`(\d{2})-(\d{2})`)
	matches := re.FindStringSubmatch(timeStr)
	if len(matches) == 3 {
		month, _ := strconv.Atoi(matches[1])
		day, _ := strconv.Atoi(matches[2])
		// 假设是当年
		date := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location())
		// 如果日期在未来，则可能是去年的日期
		if date.After(now) {
			date = time.Date(now.Year()-1, time.Month(month), day, 0, 0, 0, 0, now.Location())
		}
		return date.Unix(), nil
	}

	// 处理年月日格式 (YYYY-MM-DD)
	re = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	matches = re.FindStringSubmatch(timeStr)
	if len(matches) == 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, now.Location())
		return date.Unix(), nil
	}

	return 0, fmt.Errorf("无法解析时间格式: %s", timeStr)
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

func (X *XHS) FormatPostID(url string) string {
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
