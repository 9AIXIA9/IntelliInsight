package chromedpx

import (
	"context"
	"crawler/internal/domain"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/zeromicro/go-zero/core/logx"
	"io/ioutil"
	"sync"
	"time"
)

type Browser struct {
	ctx         context.Context
	cancelFunc  context.CancelFunc
	fingerPrint string
	execPath    string
	mutex       sync.Mutex // 保护浏览器操作线程安全
}

func NewBrowser(dataDir string, ip string, fingerPrint string, enableHeadless bool, browserPath string) domain.Browser {
	logx.Infof("初始化浏览器: 数据目录=%s, 代理IP=%s", dataDir, ip)

	opts := getOptions(dataDir, ip, fingerPrint, enableHeadless, browserPath)

	// 创建有超时的上下文
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// 创建浏览器上下文并保存取消函数
	ctx, cancel := chromedp.NewContext(allocCtx)

	browser := &Browser{
		ctx: ctx,
		cancelFunc: func() {
			cancel()      // 先取消浏览器上下文
			allocCancel() // 再取消分配器上下文
		},
		fingerPrint: fingerPrint,
		execPath:    browserPath,
		mutex:       sync.Mutex{},
	}

	if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
		logx.Errorf("浏览器初始化失败: %v", err)
		browser.Close()
		return nil
	}

	return browser
}

func getOptions(dataDir string, ip string, fingerPrint string, enableHeadless bool, browserPath string) []chromedp.ExecAllocatorOption {
	// 设置Chrome选项
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.UserDataDir(dataDir),
	}

	// 仅当IP不是localhost或127.0.0.1时才设置代理
	if ip != "" && ip != "localhost" && ip != "127.0.0.1" {
		logx.Infof("使用代理: %s", ip)
		opts = append(opts, chromedp.ProxyServer(ip))
	}

	if browserPath != "" {
		opts = append(opts, chromedp.ExecPath(browserPath))
	}

	if enableHeadless {
		opts = append(opts, chromedp.Headless)
	}

	if fingerPrint != "" {
		opts = append(opts, chromedp.UserAgent(fingerPrint))
	}

	return opts
}

func (b *Browser) Navigate(url string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	timeoutCtx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
	defer cancel()

	return chromedp.Run(timeoutCtx, chromedp.Navigate(url))
}

func (b *Browser) GetPageContent() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var html string
	err := chromedp.Run(b.ctx, chromedp.OuterHTML("html", &html))
	if err != nil {
		logx.Errorf("获取页面内容失败: %v", err)
		return ""
	}
	return html
}

func (b *Browser) FindElement(selector string) (domain.Element, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var exists bool
	err := chromedp.Run(b.ctx, chromedp.EvaluateAsDevTools(`document.querySelector("`+selector+`") !== null`, &exists))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, domain.ErrElementNotFound
	}

	return &Element{
		browser:  b,
		selector: selector,
	}, nil
}
func (b *Browser) FindAllElements(selector string) ([]domain.Element, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 首先检查元素是否存在并获取元素数量
	var count int
	err := chromedp.Run(b.ctx, chromedp.EvaluateAsDevTools(`
		document.querySelectorAll("`+selector+`").length
	`, &count))

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return []domain.Element{}, nil
	}

	// 创建元素列表
	elements := make([]domain.Element, count)
	for i := 0; i < count; i++ {
		// 为每个元素创建选择器 (使用 :nth-child 索引从 1 开始)
		indexStr := fmt.Sprintf("%d", i+1)
		indexSelector := selector + ":nth-child(" + indexStr + ")"
		elements[i] = &Element{
			browser:  b,
			selector: indexSelector,
		}
	}

	return elements, nil
}

func (b *Browser) ScrollPage() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return chromedp.Run(b.ctx, chromedp.Evaluate(`
		window.scrollTo({
			top: document.body.scrollHeight,
			behavior: 'smooth'
		});
	`, nil))
}

func (b *Browser) ExecuteJS(script string, result interface{}) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return chromedp.Run(b.ctx, chromedp.Evaluate(script, result))
}

func (b *Browser) Screenshot(path string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var buf []byte
	if err := chromedp.Run(b.ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		return err
	}

	return ioutil.WriteFile(path, buf, 0644)
}

func (b *Browser) Close() {
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
}

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
