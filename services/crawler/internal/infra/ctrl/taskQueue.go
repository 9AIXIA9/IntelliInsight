package ctrl

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/domain"
	"crawler/internal/infra/crawler"
	"crawler/internal/infra/site"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

// TaskQueue 任务队列实现
type TaskQueue struct {
	// 依赖项
	repo         domain.Repository
	filter       domain.Filter
	resourcePool domain.ResourcePool
	crawler      domain.Crawler

	// 配置项
	divideThreshold uint64

	// 并发控制
	mutex      sync.RWMutex
	taskChan   chan *domain.Task
	workerPool chan struct{}
	stopChan   chan struct{}
	waitGroup  sync.WaitGroup
	isRunning  bool
}

// NewTaskQueue 创建爬虫任务队列
func NewTaskQueue(conf *config.TaskQueue, repo domain.Repository, filter domain.Filter, resourcePool domain.ResourcePool) domain.TaskQueue {
	if repo == nil {
		logx.Severef("仓库不能为空")
	}

	tq := &TaskQueue{
		// 依赖项初始化
		repo:         repo,
		filter:       filter,
		resourcePool: resourcePool,
		crawler:      crawler.NewCrawler(),

		// 配置项初始化
		divideThreshold: conf.DivideThreshold,

		// 并发控制初始化
		mutex:      sync.RWMutex{},
		taskChan:   make(chan *domain.Task, conf.MaxTaskCacheSize),
		workerPool: make(chan struct{}, conf.MaxWorkers),
		stopChan:   make(chan struct{}),
		waitGroup:  sync.WaitGroup{},
		isRunning:  true,
	}

	// 启动消费者调度器
	threading.GoSafe(tq.Dispatcher)

	return tq
}

// AddTask 添加新任务到队列
func (tq *TaskQueue) AddTask(task *domain.Task) error {
	// 任务完成后保存
	defer tq.persistTask(task)

	logx.Infof("收到任务：%v", task.ID)

	// 检查是否需要分治
	if tq.NeedToDivide(task) {
		return tq.handleDivideTask(task)
	}

	// 发送常规任务
	return tq.sendTaskToQueue(task)
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

// 持久化保存任务
func (tq *TaskQueue) persistTask(task *domain.Task) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := tq.repo.SaveTask(ctx, task); err != nil {
		logx.Errorf("保存任务失败：%v", err)
	}
}

// 处理需要分治的任务
func (tq *TaskQueue) handleDivideTask(task *domain.Task) error {
	logx.Debugf("任务过大，进行分治：%v", task.ID)
	subTasks := tq.DivideTask(task)

	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if !tq.isRunning {
		task.Status = domain.StatusFailed
		task.Err = errors.New("任务队列未运行")
		return task.Err
	}

	// 发送子任务到任务通道
	for _, subTask := range subTasks {
		tq.taskChan <- subTask
	}

	task.Status = domain.StatusDivided
	return nil
}

// 发送任务到队列
func (tq *TaskQueue) sendTaskToQueue(task *domain.Task) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if !tq.isRunning {
		task.Status = domain.StatusFailed
		task.Err = errors.New("任务队列未运行")
		return task.Err
	}

	tq.taskChan <- task
	task.Status = domain.StatusPending
	return nil
}

// Dispatcher 调度器 - 负责分配任务给工作协程
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
					<-tq.workerPool // 释放worker槽位
					tq.waitGroup.Done()
				}()

				task.Status = domain.StatusRunning

				if task.ParentID != "" || task.Status == domain.StatusDivided {
					tq.processSubTask(task)
				} else {
					tq.ProcessTask(task)
				}
			})

		case <-tq.stopChan:
			return
		}
	}
}

// ProcessTask 处理单个任务
func (tq *TaskQueue) ProcessTask(task *domain.Task) {
	task.StartTime = time.Now()

	defer func() {
		tq.finalizeTask(task)
	}()

	// 1. 转换站点配置
	s, err := tq.prepareSite(task)
	if err != nil {
		task.Err = err
		return
	}

	// 2. 获取爬虫资源
	resource, err := tq.acquireResource(task)
	if err != nil {
		task.Err = err
		return
	}

	if resource == nil {
		task.Err = errors.New("获取资源为空")
		logx.Errorf("获取资源为空，但是未报错")
		return
	}

	defer tq.resourcePool.Put(resource)

	// 3. 爬取帖子链接
	links, err := tq.collectLinks(resource, s, task)
	if err != nil {
		task.Err = err
		return
	}

	// 4. 爬取帖子详情
	posts := tq.collectPosts(resource, s, links, task)

	// 5. 保存爬取结果
	tq.saveCrawlResults(task, posts)
}

// processSubTask 处理子任务
func (tq *TaskQueue) processSubTask(task *domain.Task) {
	tq.ProcessTask(task)

	if err := tq.UpdateParentTaskProgress(task); err != nil {
		logx.Errorf("完成子任务：%v - %v, 但更新父任务进度失败: %v", task.ID, task.ParentID, err)
	}
}

// 准备站点配置
func (tq *TaskQueue) prepareSite(task *domain.Task) (domain.Site, error) {
	s, err := site.Convert(task.Site)
	if err != nil {
		return nil, fmt.Errorf("转换站点错误: %w", err)
	}
	return s, nil
}

// 获取爬虫资源
func (tq *TaskQueue) acquireResource(task *domain.Task) (domain.ResourceUnit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resource, err := tq.resourcePool.Get(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// 超时获取不到资源则重新进入队列等待
			task.Status = domain.StatusPending
			if err = tq.AddTask(task); err != nil {
				return nil, fmt.Errorf("获取资源失败且无法重新加入队列：%w", task.Err)
			}
			return nil, err
		}
		return nil, fmt.Errorf("无法获取资源实例: %w", err)
	}

	// 资源非空检查
	if resource == nil {
		task.Status = domain.StatusPending
		task.Err = errors.New("获取到空资源")
		if err = tq.AddTask(task); err != nil {
			return nil, fmt.Errorf("获取到空资源且无法重新加入队列")
		}
		return nil, task.Err
	}

	return resource, nil
}

// 收集帖子链接
func (tq *TaskQueue) collectLinks(resource domain.ResourceUnit, s domain.Site, task *domain.Task) ([]string, error) {
	logx.Infof("开始收集帖子链接")
	links, err := tq.crawler.CollectPostLinks(
		resource.Browser(),
		tq.filter,
		s,
		task.Keyword,
		task.PostCount,
		task.MinLikes,
	)

	if err != nil {
		return nil, fmt.Errorf("收集帖子链接失败：%w", err)
	}

	logx.Debugf("成功收集到 %d 个帖子链接", len(links))
	return links, nil
}

// 收集帖子详情
func (tq *TaskQueue) collectPosts(resource domain.ResourceUnit, s domain.Site, links []string, task *domain.Task) []*domain.Post {
	posts := make([]*domain.Post, 0, len(links))

	for i, link := range links {
		logx.Debugf("开始爬取第 %d/%d 个帖子: %s", i+1, len(links), link)

		post, err := tq.crawler.CollectPostDetail(resource.Browser(), s, link, &domain.CollectPostDetailOption{
			IncludeComments: task.IncludeComments,
			IncludeImages:   task.IncludeImages,
			CommentsPerPost: task.CommentsPerPost,
			MinLikes:        task.MinLikes,
		})

		if err != nil {
			logx.Errorf("收集%v帖子时出错：%v", link, err)
			continue
		}

		posts = append(posts, post)
		task.PostsCollected++

		logx.Infof("成功爬取帖子: %s, 标题: %s", link, post.Title)
	}

	return posts
}

// 保存爬取结果
func (tq *TaskQueue) saveCrawlResults(task *domain.Task, posts []*domain.Post) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskID := task.ID
	if task.ParentID != "" {
		taskID = task.ParentID
	}

	if err := tq.repo.SavePostsAndComments(ctx, taskID, posts); err != nil {
		logx.Errorf("保存爬取帖子到MongoDB失败: %v", err)
	}
}

// finalizeTask 完成任务时的处理
func (tq *TaskQueue) finalizeTask(task *domain.Task) {
	if task.Status == domain.StatusPending {
		logx.Infof("任务%v重新加入队列", task.ID)
		return
	}

	task.EndTime = time.Now()
	if task.Err != nil {
		task.Status = domain.StatusFailed
	} else {
		task.Status = domain.StatusCompleted
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := tq.repo.SaveTask(ctx, task); err != nil {
		logx.Errorf("保存任务失败: %v", err)
	}
}

// NeedToDivide 是否需要分治
func (tq *TaskQueue) NeedToDivide(task *domain.Task) bool {
	return task.PostCount > tq.divideThreshold
}

// DivideTask 分治任务 - 均匀分配方式
func (tq *TaskQueue) DivideTask(task *domain.Task) []*domain.Task {
	//todo 看不懂
	//  计算需要多少个子任务（向上取整，确保每个子任务大小不超过阈值）
	n := (task.PostCount + tq.divideThreshold - 1) / tq.divideThreshold

	// 计算基本大小（向下取整）
	baseSize := task.PostCount / n

	// 计算有多少剩余的需要+1分配
	remainder := task.PostCount % n

	subTasks := make([]*domain.Task, 0, n)

	// 创建子任务，前remainder个任务大小为 baseSize +1，其余为 baseSize
	for i := uint64(0); i < n; i++ {
		taskSize := baseSize
		if i < remainder {
			taskSize++
		}
		subTasks = append(subTasks, createSubTask(task, taskSize))
	}

	logx.Debugf("task:%v 分治为%v个大小为%v、%v+1的子任务", task.ID, n, baseSize, baseSize)

	return subTasks
}

// UpdateParentTaskProgress 更新父任务进度
func (tq *TaskQueue) UpdateParentTaskProgress(task *domain.Task) error {
	// TODO: 实现更新父任务进度的逻辑
	return nil
}
