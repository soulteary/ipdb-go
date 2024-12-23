package ipdb

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

// RiskInfo 存储IP风险信息
type RiskInfo struct {
	Score       int    `json:"score"`        // 风险分数
	Behavior    string `json:"behavior"`     // 行为类型
	CountryCode string `json:"country_code"` // 国家代码
}

// Risk 风险数据库结构
type Risk struct {
	reader *reader
	cache  *sync.Map    // 缓存
	mu     sync.RWMutex // 并发保护
}

// NewRisk 创建新的风险数据库实例
func NewRisk(filename string) (*Risk, error) {
	if filename == "" {
		return nil, errors.New("数据库文件名不能为空")
	}

	reader, err := newReader(filename, &RiskInfo{})
	if err != nil {
		return nil, fmt.Errorf("初始化风险数据库失败: %v", err)
	}

	return &Risk{
		reader: reader,
		cache:  &sync.Map{},
	}, nil
}

// FindInfo 查询IP地址的风险信息
func (r *Risk) FindInfo(addr string) (*RiskInfo, error) {
	// 验证IP地址
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	// 检查缓存
	if val, ok := r.cache.Load(addr); ok {
		if info, ok := val.(*RiskInfo); ok {
			return info, nil
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// 查询数据
	data, err := r.reader.FindMap(addr, "CN")
	if err != nil {
		return nil, fmt.Errorf("查询风险信息失败: %v", err)
	}

	info := &RiskInfo{}

	// 解析数据
	if v, ok := data["score"]; ok {
		score, err := strconv.Atoi(v)
		if err == nil {
			info.Score = score
		}
	}
	if v, ok := data["behavior"]; ok {
		info.Behavior = v
	}
	if v, ok := data["country_code"]; ok {
		info.CountryCode = v
	}

	// 写入缓存
	r.cache.Store(addr, info)

	return info, nil
}

// ClearCache 清理缓存
func (r *Risk) ClearCache() {
	r.cache = &sync.Map{}
}

// Reload 重新加载数据库
func (r *Risk) Reload(filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reader, err := newReader(filename, &RiskInfo{})
	if err != nil {
		return fmt.Errorf("重新加载数据库失败: %v", err)
	}

	r.reader = reader
	r.ClearCache()

	return nil
}

// IsIPv4 检查是否支持IPv4
func (r *Risk) IsIPv4() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reader.IsIPv4Support()
}

// IsIPv6 检查是否支持IPv6
func (r *Risk) IsIPv6() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reader.IsIPv6Support()
}
