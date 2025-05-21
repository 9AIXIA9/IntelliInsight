package crawler

import (
	"crawler/internal/domain"
	"crawler/internal/infra/utils"
	"crawler/proto"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxTimes = 1000
)

// Crawler 小红书爬虫结构体
type Crawler struct {
}

// NewCrawler 创建爬虫实例
func NewCrawler() domain.Crawler {
	return &Crawler{}
}

func (c *Crawler) CollectPostLinks(browser domain.Browser, site domain.Site, keyword string, count int32, minLikes int32) ([]string, error) {
	//导航到搜索页面
	searchURL := site.GetSearchURL(keyword)

	err := browser.Navigate(searchURL)
	if err != nil {
		return nil, err
	}

	//检查是否需要登录
	content := browser.GetPageContent()
	if site.NeedsLogin(content) {
		logx.Infof("%v需要登录", site.GetName())

		if err = site.Login(); err != nil {
			return nil, fmt.Errorf("登录站点%s,出错：%v", site.GetName(), err)
		}

		logx.Infof("登录成功%v", site.GetName())
	}

	links := make([]string, 0, count)

	for i := 0; i < maxTimes; i++ {
		content = browser.GetPageContent()
		//开始之前的数量
		start := len(links)

		rawLinks, err := site.ParsePostLinks(content, minLikes)
		if err != nil {
			return nil, err
		}

		for _, link := range rawLinks {
			if !utils.Contains(links, link) {
				links = append(links, link)
			}

			if int32(len(links)) >= count {
				return links[:count], nil
			}
		}

		//没有增多 -> 到底
		if len(links) == start {
			return links, nil
		}

		logx.Infof("本页收集完毕，前往下页")
		err = browser.ScrollPage()
		if err != nil {
			return nil, err
		}

		utils.DelayRandomly(3000)
	}

	return nil, errors.New("次数过多")
}

func (c *Crawler) CollectPostDetail(browser domain.Browser, site domain.Site, postURL string, opts ...*domain.CollectPostDetailOption) (*proto.PostItem, error) {
	// 处理默认选项
	var opt *domain.CollectPostDetailOption
	if len(opts) > 0 {
		opt = opts[0].WithDefault()
	} else {
		opt = (&domain.CollectPostDetailOption{}).WithDefault()
	}

	err := browser.Navigate(postURL)
	if err != nil {
		return nil, err
	}

	content := browser.GetPageContent()
	if site.NeedsLogin(content) {
		logx.Infof("%v需要登录", site.GetName())

		if err = site.Login(); err != nil {
			return nil, fmt.Errorf("登录站点%s,出错：%v", site.GetName(), err)
		}

		logx.Infof("登录成功%v", site.GetName())
	}

	//基础信息
	post, err := site.ParsePostDetail(content, opt.IncludeImages)
	if err != nil {
		return nil, err
	}

	//获取评论及回复
	if opt.IncludeComments {
		comments, err := site.ParseComments(content, opt.CommentsPerPost, opt.RepliesPerComment)
		if err != nil {
			logx.Errorf("获取评论错误：%v", err)
			return post, nil
		}
		post.Comments = comments
	}

	return post, nil
}
