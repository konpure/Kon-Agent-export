package processor

import (
	"log"
	"time"

	"github.com/konpure/Kon-Agent-export/pkg/protocol"
)

// ProcessedMetric 处理后的监控数据结构
type ProcessedMetric struct {
	AgentID   string              `json:"agent_id"`
	Timestamp time.Time           `json:"timestamp"`
	Name      string              `json:"name"`
	Value     float64             `json:"value"`
	Labels    map[string]string   `json:"labels"`
	Type      string              `json:"type"`
	RawType   protocol.MetricType `json:"-"`
	Payload   []byte              `json:"payload,omitempty"`
}

// Processor 数据处理接口
type Processor interface {
	ProcessBatchRequest(req *protocol.BatchMetricsRequest) ([]ProcessedMetric, error)
	ProcessSingleMetric(agentID string, metric *protocol.Metric) (*ProcessedMetric, error)
}

// DefaultProcessor 默认数据处理器
type DefaultProcessor struct{}

// NewDefaultProcessor 创建默认数据处理器
func NewDefaultProcessor() Processor {
	return &DefaultProcessor{}
}

// ProcessBatchRequest 处理批量监控数据请求
func (p *DefaultProcessor) ProcessBatchRequest(req *protocol.BatchMetricsRequest) ([]ProcessedMetric, error) {
	processedMetrics := make([]ProcessedMetric, 0, len(req.Metrics))

	// 处理每个监控数据
	for _, metric := range req.Metrics {
		processedMetric, err := p.ProcessSingleMetric(req.AgentId, metric)
		if err != nil {
			log.Printf("Failed to process metric: %v", err)
			continue
		}
		processedMetrics = append(processedMetrics, *processedMetric)
	}

	return processedMetrics, nil
}

// ProcessSingleMetric 处理单个监控数据
func (p *DefaultProcessor) ProcessSingleMetric(agentID string, metric *protocol.Metric) (*ProcessedMetric, error) {
	// 验证数据完整性
	if err := p.validateMetric(metric); err != nil {
		return nil, err
	}

	// 转换时间戳
	timestamp := time.Unix(0, metric.Timestamp*int64(time.Millisecond))

	// 转换指标类型
	typeStr := metric.Type.String()

	// 创建处理后的指标
	processedMetric := &ProcessedMetric{
		AgentID:   agentID,
		Timestamp: timestamp,
		Name:      metric.Name,
		Value:     metric.Value,
		Labels:    metric.Labels,
		Type:      typeStr,
		RawType:   metric.Type,
		Payload:   metric.Payload,
	}

	// 可以在这里添加额外的处理逻辑，如数据聚合、过滤等

	return processedMetric, nil
}

// validateMetric 验证监控数据完整性
func (p *DefaultProcessor) validateMetric(metric *protocol.Metric) error {
	// 检查必填字段
	if metric.Name == "" {
		return ErrEmptyMetricName
	}
	if metric.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}

	// 检查指标类型是否有效
	if metric.Type < protocol.MetricType_CPU_USAGE || metric.Type > protocol.MetricType_EBPF_RAW {
		return ErrInvalidMetricType
	}

	return nil
}

// 自定义错误类型
var (
	ErrEmptyMetricName   = &MetricError{"metric name is empty"}
	ErrInvalidTimestamp  = &MetricError{"invalid timestamp"}
	ErrInvalidMetricType = &MetricError{"invalid metric type"}
)

// MetricError 指标错误结构
type MetricError struct {
	Message string
}

// Error 实现error接口
func (e *MetricError) Error() string {
	return e.Message
}
