package latency

import (
	"upbitBnServer/internal/infra/observe/metric/metricIndex"
	"upbitBnServer/internal/resource/resourceEnum"
)

type LatencyType uint8

const (
	RECEIVE       LatencyType = iota // 接收延迟
	PROCESS_READ                     // 读取延迟
	PROCESS_PARSE                    // 解析延迟
	PROCESS_TOTAL                    // 处理延迟
)

type Http struct {
	latencyKey   string
	latencyType  LatencyType
	resourceType resourceEnum.ResourceType
}

func NewHttpMonitor(latencyKey string, latencyType LatencyType, resourceType resourceEnum.ResourceType) *Http {
	return &Http{
		latencyKey:   latencyKey,
		latencyType:  latencyType,
		resourceType: resourceType,
	}
}

func (s *Http) Record(symbolName string, ts float64) {
	switch s.latencyType {
	case RECEIVE, PROCESS_TOTAL, PROCESS_READ:
		metricIndex.ObserveReceiveLatencyUS(s.resourceType.String(), symbolName, s.latencyKey, ts)
	case PROCESS_PARSE:
		metricIndex.ObserveProcessLatencyUS(s.resourceType.String(), symbolName, s.latencyKey, ts)
	}
}
