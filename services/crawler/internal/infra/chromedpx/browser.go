package chromedpx

import (
	"context"
	"crawler/internal/domain"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/zeromicro/go-zero/core/logx"
	"os"
	"sync"
)

type Browser struct {
	ctx         context.Context
	cancelFunc  context.CancelFunc
	fingerPrint string
	execPath    string
	mutex       sync.Mutex // 保护浏览器操作线程安全
}

func NewHealthyBrowser(dataDir string, ip string, fingerPrint string, enableHeadless bool, browserPath string) (domain.Browser, error) {
	parentCtx := context.Background()

	logx.Debugf("初始化浏览器: 数据目录=%s, 代理IP=%s", dataDir, ip)

	// 确保数据目录存在
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("创建浏览器数据目录失败: %w", err)
		}
	}

	opts := getOptions(dataDir, ip, fingerPrint, enableHeadless, browserPath)

	// 创建新的执行分配器
	ctx, cancel := chromedp.NewExecAllocator(parentCtx, opts...)

	// 创建浏览器上下文，添加日志记录
	browserCtx, browserCancel := chromedp.NewContext(
		ctx,
		chromedp.WithLogf(logx.Infof),
	)

	// 创建浏览器对象
	browser := &Browser{
		ctx: browserCtx,
		cancelFunc: func() {
			browserCancel() // 先取消浏览器上下文
			cancel()        // 再取消分配器上下文
			logx.Debugf("浏览器实例已关闭")
		},
		fingerPrint: fingerPrint,
		execPath:    browserPath,
		mutex:       sync.Mutex{},
	}

	if err := browser.CheckHealth(); err != nil {
		return nil, err
	}
	return browser, nil
}

func (b *Browser) CheckHealth() error {
	// 使用页面状态检查
	var result bool
	err := chromedp.Run(b.ctx, chromedp.Evaluate(`document.readyState === "complete"`, &result))
	if err != nil {
		b.Close()
		return fmt.Errorf("浏览器状态检查失败：%w", err)
	}
	return nil
}

func (b *Browser) Navigate(url string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return chromedp.Run(b.ctx, chromedp.Navigate(url))
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

	return os.WriteFile(path, buf, 0644)
}

func (b *Browser) Close() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.cancelFunc != nil {
		b.cancelFunc()
	}
}

func getOptions(dataDir string, ip string, fingerPrint string, enableHeadless bool, browserPath string) []chromedp.ExecAllocatorOption {
	// 基础选项
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		// 添加一些额外选项来提高稳定性
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
		// 添加数据目录
		chromedp.UserDataDir(dataDir),

		// 添加窗口尺寸参数
		chromedp.WindowSize(1080, 810),
		chromedp.Flag("window-size", "1080,810"),
	}

	// 添加代理选项
	if ip != "" && (ip == "http://localhost" || ip == "http://127.0.0.1") {
		opts = append(opts, chromedp.ProxyServer(ip))
	}

	// 设置浏览器路径
	if browserPath != "" {
		if _, err := os.Stat(browserPath); err == nil {
			opts = append(opts, chromedp.ExecPath(browserPath))
		} else {
			logx.Errorf("浏览器路径不存在: %s", browserPath)
		}
	}

	// 无头模式
	if enableHeadless {
		opts = append(opts, chromedp.Headless)
	}

	// 用户代理
	if fingerPrint != "" {
		opts = append(opts, chromedp.UserAgent(fingerPrint))
	}

	return opts
}
