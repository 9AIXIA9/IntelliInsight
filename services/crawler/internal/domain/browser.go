package domain

import "errors"

// Browser 浏览器接口
type Browser interface {
	Navigate(url string) error
	GetPageContent() string
	FindElement(selector string) (Element, error)
	FindAllElements(selector string) ([]Element, error)
	ScrollPage() error
	ExecuteJS(script string, result interface{}) error
	Screenshot(path string) error
	CheckHealth() error
	Close()
}

// Element 页面元素接口
type Element interface {
	Click() error
	Input(text string) error
	GetText() string
	GetAttribute(name string) string
}

var ErrElementNotFound = errors.New("元素未发现")
