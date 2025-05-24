package site

import (
	"context"
	"crawler/internal/domain"
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

func (X *XHS) GetPosterSelector() string {
	return "span.username"
}

func (X *XHS) GetPostTagsSelector() string {
	return "div.note-content .tag"
}

func (X *XHS) GetLikeCountSelector() string {
	return ".engage-bar-style .like-wrapper .count"
}

func (X *XHS) GetCommentCountSelector() string {
	return ".engage-bar-style .chat-wrapper .count"
}

func (X *XHS) GetCollectCountSelector() string {
	return ".engage-bar-style .collect-wrapper .count"
}

func (X *XHS) GetImageURLSelector() string {
	return "div.img-container img"
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

func (X *XHS) GetCommentItemSelector() string {
	return "div.comment-item"
}

func (X *XHS) GetCommentIDSelector() string {
	return "id"
}

func (X *XHS) GetCommenterSelector() string {
	return ".author"
}

func (X *XHS) GetCommentTimeSelector() string {
	return ".info .date"
}

func (X *XHS) GetCommentLocationSelector() string {
	return ".info .date .location"
}

func (X *XHS) GetCommentContentSelector() string {
	//".content .note-text"
	return "span.note-text"
}

func (X *XHS) GetCommentLikeCountSelector() string {
	return ".like .count"
}

func (X *XHS) GetCommentReplyCountSelector() string {
	return ".reply .count"
}

func (X *XHS) ParsePostLinks(html string, filter domain.Filter, minLikes uint64) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析HTML文档失败: %w", err)
	}

	links := make([]string, 0)
	baseURL := "https://www.xiaohongshu.com"

	doc.Find(X.GetPostCardSelector()).Each(func(i int, s *goquery.Selection) {
		// 获取帖子点赞数
		likesText := s.Find(X.GetLikeCountSelector()).Text()
		likesNum := X.ParseNumber(likesText)

		// 只收集点赞数达到要求的帖子
		if likesNum < minLikes {
			logx.Debugf("点赞数量过少跳过帖子")
			return
		}

		href, exists := s.Find(X.GetPostLinkSelector()).Attr("href")
		if !exists {
			return
		}

		// 如果是相对路径，添加基础URL
		if !strings.HasPrefix(href, "http") {
			href = baseURL + href
		}

		// 去重过滤
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		exist, err := filter.ExistsCtx(ctx, []byte(href))
		if err != nil {
			logx.Errorf("检查帖子是否被爬取失败：%v", err)
			return
		}

		if !exist {
			links = append(links, href)

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := filter.AddCtx(ctx, []byte(href)); err != nil {
				logx.Errorf("添加帖子到过滤器失败：%v", err)
			}
		}

		logx.Debugf("存在帖子已跳过")
	})

	return links, nil
}
func (X *XHS) ParsePostPage(html string, includeImages bool) (*domain.Post, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	post := &domain.Post{}

	// 标题 - 处理转义字符
	title := strings.TrimSpace(doc.Find(X.GetPostTitleSelector()).Text())
	post.Title = strings.ReplaceAll(title, "\\n", "\n")

	// 内容 - 处理转义字符
	content := strings.TrimSpace(doc.Find(X.GetPostContentSelector()).Text())
	post.Content = strings.ReplaceAll(content, "\\n", "\n")

	// 作者
	post.Poster = strings.TrimSpace(doc.Find(X.GetPosterSelector()).First().Text())

	// 解析时间和地点
	timeAndLocationText := strings.TrimSpace(doc.Find(X.GetTimeAndLocationSelector()).Text())
	post.Time, post.Location = X.ParseTimeAndLocation(timeAndLocationText)

	// 统计数据
	post.LikeCount = X.ParseNumber(doc.Find(X.GetLikeCountSelector()).Text())
	post.CollectCount = X.ParseNumber(doc.Find(X.GetCollectCountSelector()).Text())
	post.CommentCount = X.ParseNumber(doc.Find(X.GetCommentCountSelector()).Text())

	// 标签 - 处理转义字符
	var tags []string
	doc.Find(X.GetPostTagsSelector()).Each(func(i int, s *goquery.Selection) {
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
		doc.Find(X.GetImageURLSelector()).Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				images = append(images, src)
			} else if dataSrc, exists := s.Attr("data-src"); exists {
				images = append(images, dataSrc)
			}
		})
		post.ImageURLs = images
	}

	return post, nil
}

func (X *XHS) ParseComments(html string, count uint64, minReplyCount uint64) ([]*domain.Comment, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var comments []*domain.Comment
	commentCount := 0

	doc.Find(X.GetCommentItemSelector()).Each(func(i int, s *goquery.Selection) {
		if uint64(commentCount) >= count {
			return
		}

		comment := &domain.Comment{}

		// 评论ID
		// id = comment-681e61ed00000000090169d6 class = comment-item
		if id, exists := s.Attr(X.GetCommentIDSelector()); exists {
			comment.ID = X.ParseCommentID(id)
		}

		// 评论内容
		comment.Content = strings.TrimSpace(s.Find(X.GetCommentContentSelector()).Text())

		// 评论作者
		comment.Commenter = strings.TrimSpace(s.Find(X.GetCommenterSelector()).Text())

		// 点赞数
		comment.LikeCount = X.ParseNumber(s.Find(X.GetCommentLikeCountSelector()).Text())

		// 回复数
		comment.ReplyCount = X.ParseNumber(s.Find(X.GetCommentReplyCountSelector()).Text())

		//  评论时间和地点 放在一起的
		timeAndLocationText := strings.TrimSpace(s.Find(X.GetCommentTimeSelector()).Text())
		comment.Time, comment.Location = X.ParseTimeAndLocation(timeAndLocationText)

		comments = append(comments, comment)
		commentCount++
	})

	return comments, nil
}

func (X *XHS) ParsePostIDFromURL(url string) string {
	// 示例URL: https://www.xiaohongshu.com/explore/681dcb4700000000230155f6?xsec_token=AB61wxogAE0aGV4XXz3xRU7L08DSBxqe14KCT6N8SLzDA=&xsec_source=pc_cfeed

	// 分离查询参数部分
	parts := strings.Split(url, "?")
	pathPart := parts[0]

	// 分离路径部分，获取最后一段作为ID
	pathSegments := strings.Split(pathPart, "/")
	if len(pathSegments) > 0 {
		postID := pathSegments[len(pathSegments)-1]
		return postID
	}

	return ""
}

func (X *XHS) ParseCommentID(commentStr string) string {
	// 提取格式为 "comment-XXXXXXXX" 的评论ID中的数字部分
	if strings.Contains(commentStr, "comment-") {
		parts := strings.Split(commentStr, "comment-")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return commentStr
}

//格式：
//   5210 数字
//   2.7万 24.2万 含汉字
//   赞 回复 收藏 点赞 评论 ——> 0
//   带"千"字或"k/K"的数字（如"3.5千"或"3.5k"）
//   带"亿"字的数字（如"1.2亿"）
//   带加号的数字（如"999+"）
//   识别"w"作为"万"的替代表示

func (X *XHS) ParseNumber(numStr string) uint64 {
	// 去除空白字符
	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		logx.Debugf("解析数字: 空字符串 -> 0")
		return 0
	}

	// 处理特定文本标签
	specialLabels := []string{"赞", "回复", "收藏", "点赞", "评论"}
	for _, label := range specialLabels {
		if numStr == label {
			logx.Debugf("解析数字: %s -> 0 (特定标签)", numStr)
			return 0
		}
	}
	var result uint64

	// 检查是否包含"万"或"w"
	if strings.Contains(numStr, "万") || strings.Contains(numStr, "w") {
		// 移除所有空格
		numStr = strings.ReplaceAll(numStr, " ", "")
		re := regexp.MustCompile(`(\d+(\.\d+)?)[万w]`)
		matches := re.FindStringSubmatch(numStr)
		if len(matches) >= 2 {
			num, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result = uint64(num * 10000)
				logx.Debugf("解析数字: %s -> %d (万格式)", numStr, result)
				return result
			}
		}
	}

	// 检查是否包含"亿"
	if strings.Contains(numStr, "亿") {
		numStr = strings.ReplaceAll(numStr, " ", "")
		re := regexp.MustCompile(`(\d+(\.\d+)?)亿`)
		matches := re.FindStringSubmatch(numStr)
		if len(matches) >= 2 {
			num, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result = uint64(num * 100000000)
				logx.Debugf("解析数字: %s -> %d (亿格式)", numStr, result)
				return result
			}
		}
	}

	// 检查是否包含"千"或"k"
	if strings.Contains(numStr, "千") || strings.Contains(numStr, "k") || strings.Contains(numStr, "K") {
		numStr = strings.ReplaceAll(numStr, " ", "")
		re := regexp.MustCompile(`(\d+(\.\d+)?)[千kK]`)
		matches := re.FindStringSubmatch(numStr)
		if len(matches) >= 2 {
			num, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				result = uint64(num * 1000)
				logx.Debugf("解析数字: %s -> %d (千格式)", numStr, result)
				return result
			}
		}
	}

	// 检查是否带有加号，如"999+"
	if strings.Contains(numStr, "+") {
		numStr = strings.ReplaceAll(numStr, " ", "")
		re := regexp.MustCompile(`(\d+)\+`)
		matches := re.FindStringSubmatch(numStr)
		if len(matches) >= 2 {
			num, err := strconv.Atoi(matches[1])
			if err == nil {
				result = uint64(num)
				logx.Debugf("解析数字: %s -> %d (加号格式)", numStr, result)
				return result
			}
		}
	}

	// 如果以上都不匹配，尝试将其作为纯数字解析
	re := regexp.MustCompile(`\d+`)
	matches := re.FindStringSubmatch(numStr)
	if len(matches) > 0 {
		num, err := strconv.Atoi(matches[0])
		if err == nil {
			result = uint64(num)
			logx.Debugf("解析数字: %s -> %d (纯数字格式)", numStr, result)
			return result
		}
	}

	logx.Debugf("解析数字: %s -> 0 (解析失败)", numStr)
	return 0
}

func (X *XHS) RequireLogin(html string) bool {
	return strings.Contains(html, "登录") &&
		strings.Contains(html, "注册") &&
		!strings.Contains(html, "退出登录")
}

func (X *XHS) Login() error {
	// todo: 扫码登录 或 手机验证码登录
	return errors.New("登录功能未实现")
}
