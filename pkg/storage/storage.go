package storage

import (
	"github.com/konpure/Kon-Agent-export/pkg/processor"
	"log"
	"sync"
	"time"
)

type Storage interface {
	SaveMetrics(metrics []processor.ProcessedMetric) error
	GetMetricsByAgentID(agentID string, limit int) ([]processor.ProcessedMetric, error)
	GetMetricsByType(metricType string, limit int) ([]processor.ProcessedMetric, error)
	GetLatestMetrics(limit int) ([]processor.ProcessedMetric, error)
	GetMetricsByTimeRange(start, end time.Time, limit int) ([]processor.ProcessedMetric, error)
	CleanExpired()
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	mu         sync.RWMutex
	metrics    []processor.ProcessedMetric
	maxSize    int
	expireTime time.Duration
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage(maxSize int, expireTime time.Duration) Storage {
	storage := &MemoryStorage{
		metrics:    make([]processor.ProcessedMetric, 0, maxSize),
		maxSize:    maxSize,
		expireTime: expireTime,
	}

	// 启动定时清理过期数据的goroutine
	go storage.startCleanupTimer()

	return storage
}

// SaveMetrics 保存监控数据
func (s *MemoryStorage) SaveMetrics(metrics []processor.ProcessedMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 添加新数据
	s.metrics = append(s.metrics, metrics...)

	// 限制存储大小
	if len(s.metrics) > s.maxSize {
		// 计算需要删除的数量
		deleteCount := len(s.metrics) - s.maxSize
		// 删除最旧的数据
		s.metrics = s.metrics[deleteCount:]
	}

	log.Printf("Saved %d metrics, total: %d", len(metrics), len(s.metrics))
	return nil
}

// GetMetricsByAgentID 按Agent ID获取监控数据
func (s *MemoryStorage) GetMetricsByAgentID(agentID string, limit int) ([]processor.ProcessedMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]processor.ProcessedMetric, 0, limit)

	// 从最新的数据开始遍历
	for i := len(s.metrics) - 1; i >= 0 && len(result) < limit; i-- {
		if s.metrics[i].AgentID == agentID {
			result = append(result, s.metrics[i])
		}
	}

	return result, nil
}

// GetMetricsByType 按指标类型获取监控数据
func (s *MemoryStorage) GetMetricsByType(metricType string, limit int) ([]processor.ProcessedMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]processor.ProcessedMetric, 0, limit)

	// 从最新的数据开始遍历
	for i := len(s.metrics) - 1; i >= 0 && len(result) < limit; i-- {
		if s.metrics[i].Type == metricType {
			result = append(result, s.metrics[i])
		}
	}

	return result, nil
}

// GetLatestMetrics 获取最新的监控数据
func (s *MemoryStorage) GetLatestMetrics(limit int) ([]processor.ProcessedMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 确保limit不超过实际数据量
	if limit > len(s.metrics) {
		limit = len(s.metrics)
	}

	// 获取最新的limit条数据
	startIdx := len(s.metrics) - limit
	result := s.metrics[startIdx:]

	return result, nil
}

// GetMetricsByTimeRange 按时间范围获取监控数据
func (s *MemoryStorage) GetMetricsByTimeRange(start, end time.Time, limit int) ([]processor.ProcessedMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]processor.ProcessedMetric, 0, limit)

	// 从最新的数据开始遍历
	for i := len(s.metrics) - 1; i >= 0 && len(result) < limit; i-- {
		if (s.metrics[i].Timestamp.After(start) || s.metrics[i].Timestamp.Equal(start)) &&
			(s.metrics[i].Timestamp.Before(end) || s.metrics[i].Timestamp.Equal(end)) {
			result = append(result, s.metrics[i])
		}
	}

	return result, nil
}

// CleanExpired 清理过期数据
func (s *MemoryStorage) CleanExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiredTime := now.Add(-s.expireTime)

	// 找到第一个未过期的索引
	firstValidIdx := 0
	for i, metric := range s.metrics {
		if metric.Timestamp.After(expiredTime) {
			firstValidIdx = i
			break
		}
	}

	// 删除过期数据
	if firstValidIdx > 0 {
		log.Printf("Cleaned %d expired metrics", firstValidIdx)
		s.metrics = s.metrics[firstValidIdx:]
	}
}

// startCleanupTimer 启动定时清理计时器
func (s *MemoryStorage) startCleanupTimer() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CleanExpired()
		}
	}
}
