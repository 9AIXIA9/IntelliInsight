package ctrl

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/control"
	"crawler/internal/domain"
	"crawler/internal/infra/crawler"
	"crawler/internal/infra/site"
	"crawler/internal/infra/utils"
	"errors"
	"fmt"
	"sync"
	"time"

	"crawler/proto"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

// TaskQueue 任务队列实现
type TaskQueue struct {
	repo         domain.Repository   // MongoDB存储仓库
	resourcePool domain.ResourcePool //资源池
	crawler      domain.Crawler
	mutex        sync.RWMutex      // 读写锁
	taskChan     chan *domain.Task // 任务通道(生产者到消费者)
	workerPool   chan struct{}     // 工作池控制并发
	stopChan     chan struct{}     // 停止信号通道
	waitGroup    sync.WaitGroup    // 等待所有工作完成
	isRunning    bool              //运行状态
}

// NewTaskQueue 创建爬虫任务队列
func NewTaskQueue(conf *config.TaskQueue, repo domain.Repository, resourcePool domain.ResourcePool) domain.TaskQueue {
	if repo == nil {
		control.LogSeveref("仓库不能为空")
	}

	tq := &TaskQueue{
		repo:         repo,
		resourcePool: resourcePool,
		crawler:      crawler.NewCrawler(),
		mutex:        sync.RWMutex{},
		taskChan:     make(chan *domain.Task, conf.MaxTaskCacheSize),
		workerPool:   make(chan struct{}, conf.MaxWorkers),
		stopChan:     make(chan struct{}),
		waitGroup:    sync.WaitGroup{},
		isRunning:    true,
	}

	// 启动消费者调度器
	threading.GoSafe(tq.Dispatcher)

	return tq
}

// Dispatcher 【消费者】调度器 - 负责分配任务给工作协程
func (tq *TaskQueue) Dispatcher() {
	for {
		select {
		case task := <-tq.taskChan:
			// 获取worker槽位（限制并发数）
			tq.workerPool <- struct{}{}
			tq.waitGroup.Add(1)

			// 启动worker处理任务
			threading.GoSafe(func() {
				defer func() {
					// 释放worker槽位
					<-tq.workerPool
					tq.waitGroup.Done()
				}()

				tq.ProcessTask(task)
			})

		case <-tq.stopChan:
			return
		}
	}
}

// ProcessTask 处理单个任务的逻辑
func (tq *TaskQueue) ProcessTask(task *domain.Task) {
	defer func() {
		task.EndTime = time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := tq.repo.SaveTasks(ctx, task); err != nil {
			logx.Errorf("保存任务到MongoDB失败: %v", err)
		}
	}()

	s, err := site.Convert(task.Request.Site)
	if err != nil {
		task.Status = domain.StatusFailed.String()
		task.Err = errors.New("无法获取资源实例")
		logx.Error(task.Err)
		return
	}

	// 从池获取资源
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resource, err := tq.resourcePool.Get(ctx)
	if err != nil {
		task.Status = domain.StatusFailed.String()
		task.Err = errors.New("无法获取资源实例")
		logx.Error(task.Err)
		return
	}
	// 确保使用完后归还浏览器
	defer tq.resourcePool.Put(resource)

	startTime := time.Now()

	// 搜索关键词，获取帖子链接列表
	logx.Infof("开始收集到帖子链接")

	links, err := tq.crawler.CollectPostLinks(resource.Browser(), s, task.Request.Keyword, task.Request.PostCount, task.Request.MinLikes)
	if err != nil {
		task.Status = domain.StatusFailed.String()
		task.Err = fmt.Errorf("收集帖子链接失败：%w", err)
		logx.Error(task.Err)
		return
	}

	logx.Infof("成功收集到 %d 个帖子链接", len(links))

	posts := make([]*proto.PostItem, 0, len(links))

	for i, link := range links {
		logx.Infof("开始爬取第 %d/%d 个帖子: %s", i+1, len(links), link)

		post, err := tq.crawler.CollectPostDetail(resource.Browser(), s, link, &domain.CollectPostDetailOption{
			IncludeComments:   task.Request.IncludeComments,
			IncludeImages:     task.Request.IncludeImages,
			CommentsPerPost:   task.Request.CommentsPerPost,
			RepliesPerComment: task.Request.RepliesPerComment,
		})
		if err != nil {
			logx.Errorf("收集%v帖子时出错：%v", link, err)
			continue
		}

		posts = append(posts, post)
		task.PostsCollected++

		logx.Infof("成功爬取帖子: %s, 标题: %s", link, post.PostTitle)
	}

	elapsedTime := time.Since(startTime)
	logx.Infof("爬虫任务完成，共爬取 %d 个帖子，耗时: %v", len(links), elapsedTime)

	// 保存爬取结果到MongoDB
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tq.repo.SavePosts(ctx, task.ID, posts); err != nil {
		logx.Errorf("保存爬取结果到MongoDB失败: %v", err)
	}

	task.Status = domain.StatusCompleted.String()
}

// AddTask 【生产者】添加新任务到队列
func (tq *TaskQueue) AddTask(request *proto.CrawlRequest) (string, error) {
	// 创建新任务
	taskID := utils.GenerateID()
	now := time.Now()
	logx.Infof("收到任务：%v", taskID)

	task := &domain.Task{
		ID:             taskID,
		Request:        request,
		PostsCollected: 0,
		StartTime:      now,
		EndTime:        time.Time{},
		Status:         domain.StatusPending.String(),
	}

	// 发送任务到任务通道（如果队列在运行）
	tq.mutex.Lock()
	if tq.isRunning {
		tq.taskChan <- task
		tq.mutex.Unlock()

		task.Status = domain.StatusRunning.String()
		return taskID, nil
	}

	tq.mutex.Unlock()
	task.Status = domain.StatusFailed.String()
	task.Err = errors.New("任务队列未运行")
	logx.Error(task.Err)
	return taskID, task.Err
}

// Stop 停止任务队列处理
func (tq *TaskQueue) Stop(ctx context.Context) {
	tq.mutex.Lock()
	if !tq.isRunning {
		tq.mutex.Unlock()
		return
	}

	tq.isRunning = false
	close(tq.stopChan)
	tq.mutex.Unlock()

	// 等待所有工作完成
	tq.waitGroup.Wait()

	// 关闭资源池
	tq.resourcePool.Close(ctx)
}
