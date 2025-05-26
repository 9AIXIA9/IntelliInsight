package crawler

import (
	"crawler/internal/domain"
	"crawler/internal/infra/utils"
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

func (c *Crawler) CollectPostLinks(browser domain.Browser, filter domain.Filter, site domain.Site, keyword string, count uint64, minLikes uint64) ([]string, error) {
	//导航到搜索页面
	searchURL := site.GetSearchURL(keyword)

	err := browser.Navigate(searchURL)
	if err != nil {
		return nil, err
	}

	// 在页面导航后添加随机延时
	utils.DelayRandomly(10000)

	//检查是否需要登录
	content := browser.GetPageContent()
	require, err := site.RequireLogin(content)
	if err != nil {
		return nil, err
	}

	if require {
		logx.Infof("%v需要登录", site.GetName())

		if err = site.Login(browser); err != nil {
			return nil, fmt.Errorf("登录站点%s,出错：%w", site.GetName(), err)
		}

		logx.Infof("登录成功%v", site.GetName())
		// 登录成功后添加随机延时
		utils.DelayRandomly(5000)
	}

	links := make([]string, 0, count)

	for i := 0; i < maxTimes; i++ {
		logx.Debugf("第%v次在页面中收集帖子", i)
		content = browser.GetPageContent()

		rawLinks, err := site.ParsePostLinks(content, filter, minLikes)
		if err != nil {
			return nil, err
		}

		for _, link := range rawLinks {
			logx.Debugf("第%v个帖子", i)
			if !utils.Contains(links, link) {
				links = append(links, link)
			}

			if uint64(len(links)) >= count {
				return links[:count], nil
			}
		}

		logx.Debugf("本页收集完毕，正在下翻")
		err = browser.ScrollPage()
		if err != nil {
			return nil, err
		}

		utils.DelayRandomly(8000)
	}

	return nil, errors.New("次数过多")
}

func (c *Crawler) CollectPostDetail(browser domain.Browser, site domain.Site, postLink string, opts ...*domain.CollectPostDetailOption) (*domain.Post, error) {
	// 处理默认选项
	var opt *domain.CollectPostDetailOption
	if len(opts) > 0 {
		opt = opts[0].WithDefault()
	} else {
		opt = (&domain.CollectPostDetailOption{}).WithDefault()
	}

	err := browser.Navigate(postLink)
	if err != nil {
		return nil, err
	}

	// 在页面导航后添加随机延时
	utils.DelayRandomly(9000)

	content := browser.GetPageContent()
	require, err := site.RequireLogin(content)
	if err != nil {
		return nil, err
	}

	if require {
		logx.Infof("%v需要登录", site.GetName())

		if err = site.Login(browser); err != nil {
			return nil, fmt.Errorf("登录站点%s,出错：%w", site.GetName(), err)
		}

		logx.Infof("登录成功%v", site.GetName())
		// 登录成功后添加随机延时关闭资源池被取消
		utils.DelayRandomly(10000)
	}

	//基础信息
	post, err := site.ParsePostPage(content, opt.IncludeImages)
	if err != nil {
		return nil, err
	}

	post.Link = postLink

	//获取评论及回复
	if opt.IncludeComments {
		comments, err := site.ParseComments(content, opt.CommentsPerPost, opt.MinLikes)
		if err != nil {
			logx.Errorf("获取评论错误：%v", err)
			return post, nil
		}
		post.Comments = comments
	}

	post.ID = site.ParsePostIDFromURL(postLink)

	logx.Debugf("获取到帖子内容:%v", post)

	return post, nil
}
