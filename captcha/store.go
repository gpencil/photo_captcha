package captcha

import (
	"sync"
	"time"
)

// CaptchaData 验证码数据结构
type CaptchaData struct {
	ID        string
	PositionX int // 缺口X坐标
	PositionY int // 缺口Y坐标
	CreatedAt time.Time
}

// Store 验证码存储接口
type Store interface {
	Set(id string, data *CaptchaData)
	Get(id string) (*CaptchaData, bool)
	Delete(id string)
	CleanExpired()
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	mu       sync.RWMutex
	data     map[string]*CaptchaData
	ttl      time.Duration
	stopChan chan struct{}
}

// NewMemoryStore 创建新的内存存储
func NewMemoryStore(ttl time.Duration) *MemoryStore {
	store := &MemoryStore{
		data:     make(map[string]*CaptchaData),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}

	// 启动清理过期数据的协程
	go store.cleanupLoop()

	return store
}

// Set 存储验证码数据
func (m *MemoryStore) Set(id string, data *CaptchaData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data.CreatedAt = time.Now()
	m.data[id] = data
}

// Get 获取验证码数据
func (m *MemoryStore) Get(id string) (*CaptchaData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[id]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(data.CreatedAt) > m.ttl {
		return nil, false
	}

	return data, true
}

// Delete 删除验证码数据
func (m *MemoryStore) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, id)
}

// CleanExpired 清理所有过期数据
func (m *MemoryStore) CleanExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, data := range m.data {
		if now.Sub(data.CreatedAt) > m.ttl {
			delete(m.data, id)
		}
	}
}

// cleanupLoop 定期清理过期数据
func (m *MemoryStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.CleanExpired()
		case <-m.stopChan:
			return
		}
	}
}

// Stop 停止存储
func (m *MemoryStore) Stop() {
	close(m.stopChan)
}

// 默认存储实例，5分钟过期
var defaultStore = NewMemoryStore(5 * time.Minute)

// Set 使用默认存储存储数据
func Set(id string, data *CaptchaData) {
	defaultStore.Set(id, data)
}

// Get 使用默认存储获取数据
func Get(id string) (*CaptchaData, bool) {
	return defaultStore.Get(id)
}

// Delete 使用默认存储删除数据
func Delete(id string) {
	defaultStore.Delete(id)
}
