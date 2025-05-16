//go:generate mockgen -source=$GOFILE -destination=./mock/taskqueue_mock.go -package=queue
package queue

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/threading"
	"sync"
	"time"

	"crawler/proto"
)

// 任务队列相关常量
const (
	// DefaultMaxWorkers DefaultWorkQueueSize 工作池相关配置
	DefaultMaxWorkers    = 5
	DefaultWorkQueueSize = 100

	// DefaultItemsPerPage SimulatedCrawlDelay 爬虫相关常量
	DefaultItemsPerPage = 10
	SimulatedCrawlDelay = 500 * time.Millisecond

	// MsgTaskPending MsgTaskRunning MsgTaskUnknown MsgTaskCompleted MsgTaskRunning 状态消息
	MsgTaskPending   = "任务等待中"
	MsgTaskRunning   = "任务执行中"
	MsgTaskCompleted = "任务已完成"
	MsgTaskFailed    = "任务执行失败"
	MsgTaskUnknown   = "未知状态"
)

// TaskQueue 定义任务队列接口
type TaskQueue interface {
	// AddTask GetTaskStatus GetTaskData 核心操作方法
	AddTask(request *proto.CrawlRequest) string                                  // 添加新任务并返回任务ID
	GetTaskStatus(taskID string) (*proto.StatusResponse, error)                  // 获取任务进度状态
	GetTaskData(taskID string, offset, limit int32) (*proto.DataResponse, error) // 获取任务结果数据

	// AddTaskResult UpdateProgress 任务数据更新方法
	AddTaskResult(taskID string, items []*proto.PostItem) error           // 添加爬取结果到任务
	UpdateProgress(taskID string, progress float32, itemsCollected int32) // 更新任务进度

	// Start Stop 队列控制方法
	Start() // 启动任务队列处理器
	Stop()  // 停止任务队列处理器
}

// Task 表示一个爬虫任务
type Task struct {
	ID             string                      // 唯一标识
	Request        *proto.CrawlRequest         // 原始请求参数
	Status         proto.StatusResponse_Status // 当前状态
	Progress       float32                     // 进度百分比
	ItemsCollected int32                       // 已收集项目数
	StartTime      time.Time                   // 开始时间
	Results        []*proto.PostItem           // 爬取结果
}

// taskQueue 任务队列实现
type taskQueue struct {
	tasks      map[string]*Task // 任务存储
	mutex      sync.RWMutex     // 读写锁
	workQueue  chan *Task       // 待处理任务队列
	workerPool chan struct{}    // 工作池控制并发
	maxWorkers int              // 最大并发数
	isRunning  bool             // 队列运行状态
	stopChan   chan struct{}    // 停止信号通道
	waitGroup  sync.WaitGroup   // 等待所有工作完成
}

// NewTaskQueue 创建任务队列实例
func NewTaskQueue() TaskQueue {
	tq := &taskQueue{
		tasks:      make(map[string]*Task),
		workQueue:  make(chan *Task, DefaultWorkQueueSize),
		workerPool: make(chan struct{}, DefaultMaxWorkers),
		maxWorkers: DefaultMaxWorkers,
		isRunning:  true,
		stopChan:   make(chan struct{}),
	}

	// 自动启动调度器
	threading.GoSafe(tq.dispatcher)

	return tq
}

// Start 启动任务队列系统
// 用于在系统停止后重新启动
func (tq *taskQueue) Start() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if !tq.isRunning {
		tq.isRunning = true
		tq.stopChan = make(chan struct{})
		threading.GoSafe(tq.dispatcher)
	}
}

// Stop 优雅停止任务队列系统
// 等待所有已开始的任务完成
func (tq *taskQueue) Stop() {
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
}

// dispatcher 任务调度器
// 负责从工作队列获取任务并分配给可用工作者
func (tq *taskQueue) dispatcher() {
	for {
		select {
		case task := <-tq.workQueue:
			// 获取worker槽位
			tq.workerPool <- struct{}{}
			tq.waitGroup.Add(1)

			// 启动worker处理任务
			go func(t *Task) {
				defer func() {
					// 释放worker槽位
					<-tq.workerPool
					tq.waitGroup.Done()
				}()

				tq.executeTask(t)
			}(task)

		case <-tq.stopChan:
			return
		}
	}
}

// executeTask 执行单个爬虫任务
// 管理任务状态更新和执行爬虫流程
func (tq *taskQueue) executeTask(task *Task) {
	// 更新状态为运行中
	tq.updateTaskStatus(task.ID, proto.StatusResponse_RUNNING, 0.0, 0)

	// 执行爬虫任务
	success := executeWebCrawling(task, tq)

	// 根据执行结果更新任务状态
	if success {
		tq.updateTaskStatus(task.ID, proto.StatusResponse_COMPLETED, 100.0, int32(len(task.Results)))
	} else {
		tq.updateTaskStatus(task.ID, proto.StatusResponse_FAILED, task.Progress, task.ItemsCollected)
	}
}

// AddTask 添加新任务到队列
// 返回唯一任务ID供后续查询使用
func (tq *taskQueue) AddTask(request *proto.CrawlRequest) string {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	taskID := generateTaskID()
	task := &Task{
		ID:             taskID,
		Request:        request,
		Status:         proto.StatusResponse_PENDING,
		Progress:       0.0,
		ItemsCollected: 0,
		StartTime:      time.Now(),
		Results:        make([]*proto.PostItem, 0),
	}

	tq.tasks[taskID] = task

	// 将任务添加到工作队列
	if tq.isRunning {
		tq.workQueue <- task
	}

	return taskID
}

// GetTaskStatus 获取当前任务状态信息
// 返回标准proto状态响应
func (tq *taskQueue) GetTaskStatus(taskID string) (*proto.StatusResponse, error) {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务 %s 不存在", taskID)
	}

	return &proto.StatusResponse{
		Status:         task.Status,
		Progress:       task.Progress,
		ItemsCollected: task.ItemsCollected,
		Message:        getStatusMessage(task.Status),
	}, nil
}

// UpdateProgress 更新任务进度
// 由爬虫实现调用以报告当前状态
func (tq *taskQueue) UpdateProgress(taskID string, progress float32, itemsCollected int32) {
	tq.updateTaskStatus(taskID, proto.StatusResponse_RUNNING, progress, itemsCollected)
}

// GetTaskData 获取任务爬取的数据结果
// 支持分页查询，仅当任务完成时可用
func (tq *taskQueue) GetTaskData(taskID string, offset, limit int32) (*proto.DataResponse, error) {
	tq.mutex.RLock()
	defer tq.mutex.RUnlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("任务 %s 不存在", taskID)
	}

	if task.Status != proto.StatusResponse_COMPLETED {
		return nil, fmt.Errorf("任务 %s 尚未完成，当前状态: %v", taskID, task.Status)
	}

	// 计算分页
	totalCount := int32(len(task.Results))
	endIndex := offset + limit
	if endIndex > totalCount {
		endIndex = totalCount
	}

	var items []*proto.PostItem
	if offset < totalCount {
		items = task.Results[offset:endIndex]
	} else {
		items = []*proto.PostItem{}
	}

	return &proto.DataResponse{
		Items:      items,
		TotalCount: totalCount,
		HasMore:    endIndex < totalCount,
	}, nil
}

// updateTaskStatus 更新任务状态（内部方法）
func (tq *taskQueue) updateTaskStatus(taskID string, status proto.StatusResponse_Status,
	progress float32, itemsCollected int32) {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if task, exists := tq.tasks[taskID]; exists {
		task.Status = status
		task.Progress = progress
		task.ItemsCollected = itemsCollected
	}
}

// AddTaskResult 添加爬取结果到任务
// 由爬虫模块调用以保存获取的数据
func (tq *taskQueue) AddTaskResult(taskID string, items []*proto.PostItem) error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	task, exists := tq.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务 %s 不存在", taskID)
	}

	task.Results = append(task.Results, items...)
	task.ItemsCollected = int32(len(task.Results))

	return nil
}

// generateTaskID 生成唯一任务ID
func generateTaskID() string {
	//todo return fmt.Sprintf("%d", time.Now().UnixMicro())
	return fmt.Sprintf("%d", time.Now().UnixMicro())
}

// getStatusMessage 获取状态描述信息
func getStatusMessage(status proto.StatusResponse_Status) string {
	switch status {
	case proto.StatusResponse_PENDING:
		return MsgTaskPending
	case proto.StatusResponse_RUNNING:
		return MsgTaskRunning
	case proto.StatusResponse_COMPLETED:
		return MsgTaskCompleted
	case proto.StatusResponse_FAILED:
		return MsgTaskFailed
	default:
		return MsgTaskUnknown
	}
}

// executeWebCrawling 执行具体的爬虫逻辑
// 根据任务参数爬取数据并更新进度
func executeWebCrawling(task *Task, queue *taskQueue) bool {
	totalPages := task.Request.PageCount
	progress := float32(0.0) // 提前定义progress变量

	for page := 1; page <= int(totalPages); page++ {
		// 实际爬虫实现中:
		// 1. 发送HTTP请求获取页面内容
		// 2. 解析HTML提取数据
		// 3. 将提取的内容转换为PostItem对象
		items := crawlPage(task.Request.Keyword, page)

		// 保存本页爬取的数据
		if len(items) > 0 {
			if err := queue.AddTaskResult(task.ID, items); err != nil {
				task.Status = proto.StatusResponse_FAILED
				task.Progress = progress
				queue.updateTaskStatus(task.ID, task.Status, task.Progress, task.ItemsCollected)
				return false
			}
		}

		// 更新当前进度
		progress = float32(page) / float32(totalPages) * 100.0
		queue.UpdateProgress(task.ID, progress, int32(len(task.Results)))

		// 检查是否需要停止
		if !queue.isRunning {
			return false
		}
	}
	return true
}
