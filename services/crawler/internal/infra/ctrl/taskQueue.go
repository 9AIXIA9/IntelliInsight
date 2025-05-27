package ctrl

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/domain"
	"crawler/internal/infra/crawler"
	"crawler/internal/infra/site"
	"crawler/internal/infra/utils/snowflake"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

// TaskQueue 任务队列实现
type TaskQueue struct {
	repo         domain.Repository   // MongoDB存储仓库
	filter       domain.Filter       //过滤器
	resourcePool domain.ResourcePool //资源池
	crawler      domain.Crawler      //爬虫

	divideThreshold uint64

	mutex      sync.RWMutex      // 读写锁
	taskChan   chan *domain.Task // 任务通道(生产者到消费者)
	workerPool chan struct{}     // 工作池控制并发
	stopChan   chan struct{}     // 停止信号通道
	waitGroup  sync.WaitGroup    // 等待所有工作完成
	isRunning  bool              //运行状态
}

// NewTaskQueue 创建爬虫任务队列
func NewTaskQueue(conf *config.TaskQueue, repo domain.Repository, filter domain.Filter, resourcePool domain.ResourcePool) domain.TaskQueue {
	if repo == nil {
		logx.Severef("仓库不能为空")
	}

	tq := &TaskQueue{
		repo:            repo,
		filter:          filter,
		resourcePool:    resourcePool,
		crawler:         crawler.NewCrawler(),
		divideThreshold: 10,
		mutex:           sync.RWMutex{},
		taskChan:        make(chan *domain.Task, conf.MaxTaskCacheSize),
		workerPool:      make(chan struct{}, conf.MaxWorkers),
		stopChan:        make(chan struct{}),
		waitGroup:       sync.WaitGroup{},
		isRunning:       true,
	}

	// 启动消费者调度器
	threading.GoSafe(tq.Dispatcher)

	return tq
}

// AddTask 【生产者】添加新任务到队列
func (tq *TaskQueue) AddTask(task *domain.Task) error {
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := tq.repo.SaveTask(ctx, task); err != nil {
			logx.Errorf("保存任务失败：%v", err)
		}
	}()

	logx.Infof("收到任务：%v", task.Info.ID)

	if tq.NeedToDivide(task) {
		logx.Debugf("任务过大，进行分治：%v", task.Info.ID)
		subTasks := tq.Divide(task)

		// 发送子任务到任务通道（如果队列在运行）
		tq.mutex.Lock()
		if tq.isRunning {
			for _, subTask := range subTasks {
				tq.taskChan <- subTask
			}
			tq.mutex.Unlock()

			task.Info.Status = domain.StatusDivided.String()
			return nil
		}

		tq.mutex.Unlock()

		task.Info.Status = domain.StatusFailed.String()
		task.Info.Err = errors.New("任务队列未运行")
		return task.Info.Err
	}

	// 发送单个任务到任务通道（如果队列在运行）
	tq.mutex.Lock()
	if tq.isRunning {
		tq.taskChan <- task
		tq.mutex.Unlock()

		task.Info.Status = domain.StatusPending.String()

		return nil
	}

	tq.mutex.Unlock()
	task.Info.Status = domain.StatusFailed.String()
	task.Info.Err = errors.New("任务队列未运行")
	return task.Info.Err
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

				//分治子任务
				if task.Info.ParentID != "" || task.Info.Status == domain.StatusDivided.String() {
					tq.ProcessSubTask(task)
				} else {
					//正常任务
					tq.ProcessTask(task)
				}

			})

		case <-tq.stopChan:
			return
		}
	}
}

// ProcessTask 处理单个任务的逻辑
func (tq *TaskQueue) ProcessTask(task *domain.Task) {
	//状态处理
	task.Info.StartTime = time.Now()
	defer func() {
		task.Info.EndTime = time.Now()
		if task.Info.Err != nil {
			task.Info.Status = domain.StatusFailed.String()
		} else {
			task.Info.Status = domain.StatusCompleted.String()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := tq.repo.SaveTask(ctx, task); err != nil {
			logx.Errorf("保存任务失败:%v", err)
		}
	}()

	s, err := site.Convert(task.Request.Site)
	if err != nil {
		task.Info.Err = fmt.Errorf("转换站点错误:%w", err)
		logx.Error(task.Info.Err)
		return
	}

	// 从池获取资源
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resource, err := tq.resourcePool.Get(ctx)
	if err != nil {
		task.Info.Err = fmt.Errorf("无法获取资源实例:%w", err)
		logx.Error(task.Info.Err)
		return
	}
	// 确保使用完后归还浏览器
	defer tq.resourcePool.Put(resource)

	startTime := time.Now()

	// 搜索关键词，获取帖子链接列表
	logx.Infof("开始收集到帖子链接")

	links, err := tq.crawler.CollectPostLinks(resource.Browser(), tq.filter, s, task.Request.Keyword, task.Request.PostCount, task.Request.MinLikes)
	if err != nil {
		task.Info.Err = fmt.Errorf("收集帖子链接失败：%w", err)
		logx.Error(task.Info.Err)
		return
	}

	logx.Debugf("成功收集到 %d 个帖子链接", len(links))

	posts := make([]*domain.Post, 0, len(links))

	for i, link := range links {
		logx.Debugf("开始爬取第 %d/%d 个帖子: %s", i+1, len(links), link)

		post, err := tq.crawler.CollectPostDetail(resource.Browser(), s, link, &domain.CollectPostDetailOption{
			IncludeComments: task.Request.IncludeComments,
			IncludeImages:   task.Request.IncludeImages,
			CommentsPerPost: task.Request.CommentsPerPost,
			MinLikes:        task.Request.MinLikes,
		})
		if err != nil {
			logx.Errorf("收集%v帖子时出错：%v", link, err)
			continue
		}

		posts = append(posts, post)
		task.Info.PostsCollected++

		logx.Infof("成功爬取帖子: %s, 标题: %s", link, post.Title)
	}

	elapsedTime := time.Since(startTime)
	logx.Infof("爬虫任务完成，共爬取 %d 个帖子，耗时: %v", len(links), elapsedTime)

	// 保存爬取结果到MongoDB
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tq.repo.SavePostsAndComments(ctx, task.Info.ID, posts); err != nil {
		logx.Errorf("保存爬取帖子到MongoDB失败: %v", err)
	}
}

// ProcessSubTask 处理单个分治任务的逻辑
func (tq *TaskQueue) ProcessSubTask(SubTask *domain.Task) {
	tq.ProcessTask(SubTask)

	if err := tq.UpdateProgress(SubTask); err != nil {
		logx.Errorf("完成子任务：%v - %v,但更新任务错误:%v", SubTask.Info.ID, SubTask.Info.ParentID, err)
	}
}

func (tq *TaskQueue) NeedToDivide(task *domain.Task) bool {
	return task.Request.PostCount > tq.divideThreshold
}

func (tq *TaskQueue) Divide(task *domain.Task) (subTasks []*domain.Task) {
	n := task.Request.PostCount / tq.divideThreshold
	if remaining := task.Request.PostCount % tq.divideThreshold; remaining == 0 {
		subTasks = make([]*domain.Task, 0, n)
	} else {
		subTasks = make([]*domain.Task, 0, n+1)

		subTasks = append(subTasks, &domain.Task{
			Info: &domain.TaskInfo{
				ID:             snowflake.GenerateID(),
				ParentID:       task.Info.ID,
				Status:         domain.StatusDivided.String(),
				PostsCollected: 0,
				Err:            nil,
				StartTime:      time.Time{},
				EndTime:        time.Time{},
			},
			Request: &domain.TaskRequest{
				Site:            task.Request.Site,
				Keyword:         task.Request.Keyword,
				PostCount:       remaining,
				MinLikes:        task.Request.MinLikes,
				CommentMinLikes: task.Request.CommentMinLikes,
				CommentsPerPost: task.Request.CommentsPerPost,
				IncludeComments: task.Request.IncludeComments,
				IncludeImages:   task.Request.IncludeImages,
			},
		})
	}

	for i := 0; i < int(n); i++ {
		subTasks = append(subTasks, &domain.Task{
			Info: &domain.TaskInfo{
				ID:             snowflake.GenerateID(),
				ParentID:       task.Info.ID,
				Status:         domain.StatusDivided.String(),
				PostsCollected: 0,
				Err:            nil,
				StartTime:      time.Time{},
				EndTime:        time.Time{},
			},
			Request: &domain.TaskRequest{
				Site:            task.Request.Site,
				Keyword:         task.Request.Keyword,
				PostCount:       tq.divideThreshold,
				MinLikes:        task.Request.MinLikes,
				CommentMinLikes: task.Request.CommentMinLikes,
				CommentsPerPost: task.Request.CommentsPerPost,
				IncludeComments: task.Request.IncludeComments,
				IncludeImages:   task.Request.IncludeImages,
			},
		})
	}

	return subTasks
}

func (tq *TaskQueue) UpdateProgress(task *domain.Task) error {
	return nil
}
