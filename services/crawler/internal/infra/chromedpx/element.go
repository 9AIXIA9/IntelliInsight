package chromedpx

import (
	"github.com/chromedp/chromedp"
)

// Element 实现domain.Element接口
type Element struct {
	browser  *Browser
	selector string
}

func (e *Element) Click() error {
	return chromedp.Run(e.browser.ctx, chromedp.Click(e.selector))
}

func (e *Element) Input(text string) error {
	return chromedp.Run(e.browser.ctx,
		chromedp.Clear(e.selector),
		chromedp.SendKeys(e.selector, text),
	)
}

func (e *Element) GetText() string {
	var text string
	err := chromedp.Run(e.browser.ctx, chromedp.Text(e.selector, &text))
	if err != nil {
		return ""
	}
	return text
}

func (e *Element) GetAttribute(name string) string {
	var value string
	err := chromedp.Run(e.browser.ctx, chromedp.AttributeValue(e.selector, name, &value, nil))
	if err != nil {
		return ""
	}
	return value
}
