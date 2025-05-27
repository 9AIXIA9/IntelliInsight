package resource

import (
	"crawler/internal/domain"
	"crawler/internal/infra/chromedpx"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

type Unit struct {
	browser        domain.Browser
	ipProxy        string
	fingerPrint    string
	dataDir        string
	enableHeadless bool
	browserPath    string
}

func NewHealthyUnit(dataDir string, ip string, fingerPrint string, enableHeadless bool, browserPath string) (domain.ResourceUnit, error) {
	browser, err := chromedpx.NewHealthyBrowser(dataDir, ip, fingerPrint, enableHeadless, browserPath)
	if err != nil {
		return nil, fmt.Errorf("新建浏览器错误：%v", err)
	}

	return &Unit{
		browser:        browser,
		ipProxy:        ip,
		fingerPrint:    fingerPrint,
		dataDir:        dataDir,
		enableHeadless: enableHeadless,
		browserPath:    browserPath,
	}, nil
}

func (u *Unit) Refresh(dataDir string, ip string, fingerPrint string) (domain.ResourceUnit, error) {
	if u.browser != nil {
		u.browser.Close()
	}

	var err error
	u.dataDir = dataDir
	u.ipProxy = ip
	u.fingerPrint = fingerPrint
	if u.browser, err = chromedpx.NewHealthyBrowser(dataDir, ip, fingerPrint, u.enableHeadless, u.browserPath); err != nil {
		return nil, err
	} else {
		return u, nil
	}
}

func (u *Unit) Browser() domain.Browser {
	return u.browser
}

func (u *Unit) IP() string {
	return u.ipProxy
}

func (u *Unit) Fingerprint() string {
	return u.fingerPrint
}

func (u *Unit) DataDir() string {
	return u.dataDir
}

// CheckHealth 检查资源单元是否健康
func (u *Unit) CheckHealth() error {
	logx.Infof("检查资源单元健康")

	// 如果没有浏览器实例，直接返回不健康
	if u.browser == nil {
		return errors.New("资源单元中发生错误，没有浏览器实例")
	}
	return u.browser.CheckHealth()
}

func (u *Unit) Close() {
	if u.browser != nil {
		u.browser.Close()
	}
}
