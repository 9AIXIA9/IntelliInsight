package resource

import (
	"crawler/internal/domain"
	"crawler/internal/infra/chromedpx"
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
	return &Unit{
		browser:        chromedpx.NewBrowser(dataDir, ip, fingerPrint, enableHeadless, browserPath),
		ipProxy:        ip,
		fingerPrint:    fingerPrint,
		dataDir:        dataDir,
		enableHeadless: enableHeadless,
		browserPath:    browserPath,
	}
}

func (u *Unit) Update(dataDir string, ip string, fingerPrint string) {
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
