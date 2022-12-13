package epoll

import (
	"github.com/jiangshuai341/zbus/lockfreequeue"
	"github.com/jiangshuai341/zbus/logger"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const _EPOLLET = 0x80000000
const MaxAsyncTasksOnceLoop = 256

const (
	readEvents      = syscall.EPOLLPRI | syscall.EPOLLIN | _EPOLLET
	writeEvents     = syscall.EPOLLOUT | _EPOLLET
	readWriteEvents = readEvents | writeEvents
)

var log = logger.GetLogger("epoller")

type Epoller struct {
	fd int //epoll fd

	asyncTack
}

type TaskFunc func(...any)

type Task struct {
	Run TaskFunc
	Arg []any
}

type AsyncTaskQueue interface {
	Enqueue(*Task)
	Dequeue() *Task
	IsEmpty() bool
}

type TaskQueue struct {
	q *lockfreequeue.LockFreeList
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		q: lockfreequeue.NewLockFreeList(),
	}
}

func (que *TaskQueue) Enqueue(task *Task) {
	que.q.Enqueue(task)
}
func (que *TaskQueue) Dequeue() *Task {
	return que.q.Dequeue().(*Task)
}
func (que *TaskQueue) IsEmpty() bool {
	return que.q.IsEmpty()
}

var taskPool = sync.Pool{New: func() any { return new(Task) }}

type asyncTack struct {
	isWeak    int32
	triggerFD int            //trigger handle task queue
	normal    AsyncTaskQueue // queue with low priority
	urgent    AsyncTaskQueue // queue with high priority
}

// getTask gets a cached Task from pool.
func (a *asyncTack) getTask() *Task {
	return taskPool.Get().(*Task)
}

// putTask puts the trashy Task back in pool.
func (a *asyncTack) putTask(task *Task) {
	task.Run, task.Arg = nil, nil
	taskPool.Put(task)
}

func OpenEpoller() (poller *Epoller, err error) {
	poller = &Epoller{}
	poller.fd, err = syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		poller = nil
		err = os.NewSyscallError("epoll_create1", err)
		return
	}
	r, _, e := syscall.Syscall(syscall.SYS_EVENTFD2, uintptr(0), uintptr(syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC), 0)
	if e != 0 {
		err = e
	}
	poller.triggerFD = int(r)

	if err != nil {
		_ = poller.Close()
		poller = nil
		err = os.NewSyscallError("asyncTack triggerFD Create", err)
		return
	}
	err = poller.AddRead(poller.triggerFD)
	if err != nil {
		_ = poller.Close()
		poller = nil
		return
	}
	poller.asyncTack.normal = NewTaskQueue()
	poller.asyncTack.urgent = NewTaskQueue()
	return
}

func (p *Epoller) Close() error {
	if err := os.NewSyscallError("close epollFD", syscall.Close(p.fd)); err != nil {
		return err
	}
	return os.NewSyscallError("close triggerFD", syscall.Close(p.triggerFD))
}

func (p *Epoller) Delete(fd int) error {
	return os.NewSyscallError("epoll_ctl del", syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil))
}

var (
	u uint64 = 1
	b        = (*(*[8]byte)(unsafe.Pointer(&u)))[:]
)

func (p *Epoller) AppendUrgentTask(fn TaskFunc, arg []any) (err error) {
	var a uint64 = 1
	task := p.getTask()
	task.Run, task.Arg = fn, arg
	p.urgent.Enqueue(task)
	if atomic.CompareAndSwapInt32(&p.isWeak, 0, 1) {
		if _, err = syscall.Write(p.triggerFD, (*(*[8]byte)(unsafe.Pointer(&a)))[:]); err == syscall.EAGAIN {
			err = nil
		}
	}
	return os.NewSyscallError("Write", err)
}
func (p *Epoller) AppendTask(fn TaskFunc, arg ...any) (err error) {
	task := p.getTask()
	task.Run, task.Arg = fn, arg
	p.normal.Enqueue(task)
	if atomic.CompareAndSwapInt32(&p.isWeak, 0, 1) {
		if _, err = syscall.Write(p.triggerFD, b); err == syscall.EAGAIN {
			err = nil
		}
	}
	return os.NewSyscallError("Write", err)
}

//if eventHandling&syscall.EPOLLRDHUP != 0 {
//	//对端关闭
//	continue
//}
//if eventHandling&syscall.EPOLLOUT != 0 {
//	//Socket发送缓冲区状态 写满 -> 可写
//}
//if eventHandling&syscall.EPOLLIN != 0 {
//	//Socket接收缓冲区状态 空 -> 可读
//}

func (p *Epoller) Epolling(callback func(fd int, ev uint32)) (err error) {
	var maxEvents = make([]syscall.EpollEvent, 1024)
	var eventNums int

	var fdHandling int32
	var eventHandling uint32

	var triggerReadBuf = make([]byte, 8)
	var doTask bool

	var currentTask *Task
	for {
		switch eventNums, err = syscall.EpollWait(p.fd, maxEvents, -1); err {
		case nil:
		case syscall.EAGAIN, syscall.EINTR:
			continue
		default:
			return
		}

		for i := 0; i < eventNums; i++ {
			fdHandling = maxEvents[i].Fd
			eventHandling = maxEvents[i].Events

			if fdHandling == int32(p.triggerFD) {
				_, _ = syscall.Read(p.triggerFD, triggerReadBuf)
				doTask = true
			}
			callback(int(fdHandling), eventHandling)
		}

		if doTask {
			doTask = false
			for currentTask = p.urgent.Dequeue(); currentTask != nil; currentTask = p.urgent.Dequeue() {
				currentTask.Run(currentTask.Arg...)
				p.putTask(currentTask)
			}
			for i := 0; i < MaxAsyncTasksOnceLoop; i++ {
				if currentTask = p.normal.Dequeue(); currentTask == nil {
					break
				}
				currentTask.Run(currentTask.Arg...)
				p.putTask(currentTask)
			}
			atomic.StoreInt32(&p.isWeak, 0)
			//这个间隙 其他线程是有可能写入任务的需要重新检查
			if (!p.normal.IsEmpty() || !p.urgent.IsEmpty()) &&
				atomic.CompareAndSwapInt32(&p.isWeak, 0, 1) {
				//有任务没做完 先占位 后触发eventfd
				switch _, err = syscall.Write(p.triggerFD, b); err {
				case nil, syscall.EAGAIN:
				default:
					//写失败(触发失败) 强置 doTask = true
					doTask = true
				}
			}
		}
	}
}

func (p *Epoller) AddRead(fd int) error {
	return os.NewSyscallError("epoll_ctl add",
		syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readEvents}))
}

func (p *Epoller) AddWrite(fd int) error {
	return os.NewSyscallError("epoll_ctl add",
		syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: writeEvents}))
}

func (p *Epoller) AddReadWrite(fd int) error {
	return os.NewSyscallError("epoll_ctl add",
		syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readWriteEvents}))
}

func (p *Epoller) ModRead(fd int) error {
	return os.NewSyscallError("epoll_ctl mod",
		syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: readEvents}))
}

func (p *Epoller) ModWrite(fd int) error {
	return os.NewSyscallError("epoll_ctl mod",
		syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{Fd: int32(fd), Events: writeEvents}))
}

func (p *Epoller) ModReadWrite(fd int32) error {
	return os.NewSyscallError("epoll_ctl mod",
		syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, int(fd), &syscall.EpollEvent{Fd: fd, Events: readWriteEvents}))
}
