package utils

import (
	"sync"
)

// Broadcaster 实现了一个简单的字符串消息广播器，支持历史记录和 SSE 同步
type Broadcaster struct {
	clients    map[chan string]bool
	register   chan chan string
	unregister chan chan string
	messages   chan string
	history    []string // 存储最近的 50 条日志
	mu         sync.Mutex
}

var GlobalBroadcaster *Broadcaster

func init() {
	GlobalBroadcaster = NewBroadcaster()
	go GlobalBroadcaster.run()
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients:    make(map[chan string]bool),
		register:   make(chan chan string),
		unregister: make(chan chan string),
		messages:   make(chan string, 100),
		history:    make([]string, 0, 50),
	}
}

func (b *Broadcaster) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.mu.Unlock()
		case message := <-b.messages:
			b.mu.Lock()
			// 更新历史记录
			b.history = append(b.history, message)
			if len(b.history) > 50 {
				b.history = b.history[1:]
			}
			// 广播给所有在线客户端
			for client := range b.clients {
				select {
				case client <- message:
				default:
					// 客户端读取太慢则跳过，防止阻塞整个系统
				}
			}
			b.mu.Unlock()
		}
	}
}

// Subscribe 注册一个新客户端
func (b *Broadcaster) Subscribe() chan string {
	client := make(chan string, 10)
	b.register <- client
	return client
}

// Unsubscribe 注销客户端
func (b *Broadcaster) Unsubscribe(client chan string) {
	b.unregister <- client
}

// GetRecent 获取最近的历史日志
func (b *Broadcaster) GetRecent() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	res := make([]string, len(b.history))
	copy(res, b.history)
	return res
}

// ClearRecent 清空最近的历史日志
func (b *Broadcaster) ClearRecent() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = make([]string, 0, 50)
}

// Broadcast 发送广播消息（所有模块通过此方法输出实时日志）
func (b *Broadcaster) Broadcast(message string) {
	select {
	case b.messages <- message:
	default:
		// 队列满时忽略，防止极端高频日志影响系统稳定性
	}
}
