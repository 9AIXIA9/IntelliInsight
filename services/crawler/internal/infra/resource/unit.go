package resource

import (
	"crawler/internal/domain"
	"crawler/internal/infra/chromedpx"
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

func NewUnit(dataDir string, ip string, fingerPrint string, enableHeadless bool, browserPath string) domain.ResourceUnit {
	unit := &Unit{
		browser:        chromedpx.NewBrowser(dataDir, ip, fingerPrint, enableHeadless, browserPath),
		ipProxy:        ip,
		fingerPrint:    fingerPrint,
		dataDir:        dataDir,
		enableHeadless: enableHeadless,
		browserPath:    browserPath,
	}

	return unit
}

func (u *Unit) Refresh(dataDir string, ip string, fingerPrint string) {
	u.browser.Close()

	u.browser = chromedpx.NewBrowser(dataDir, ip, fingerPrint, u.enableHeadless, u.browserPath)
	u.dataDir = dataDir
	u.ipProxy = ip
	u.fingerPrint = fingerPrint
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
func (u *Unit) CheckHealth() bool {
	logx.Infof("检查资源单元健康")

	// 如果没有浏览器实例，直接返回不健康
	if u.browser == nil {
		logx.Errorf("资源单元中发生错误，没有浏览器实例")
		return false
	}

	// 尝试导航到空白页面
	err := u.browser.Navigate("about:blank")
	if err != nil {
		logx.Errorf("访问空白页面出错,错误：%v", err)
		return false
	}
	return true
}

func (u *Unit) Close() {
	u.browser.Close()
}
