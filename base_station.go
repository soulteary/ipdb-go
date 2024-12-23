package ipdb

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// IPInfo 定义通用IP信息接口
type IPInfo interface {
	GetCountryName() string
	GetRegionName() string
	GetCityName() string
}

// BaseStationInfo 存储基站信息的结构体
type BaseStationInfo struct {
	CountryName string `json:"country_name"`
	RegionName  string `json:"region_name"`
	CityName    string `json:"city_name"`
	OwnerDomain string `json:"owner_domain"`
	IspDomain   string `json:"isp_domain"`
	BaseStation string `json:"base_station"`
}

// 让 BaseStationInfo 实现 IPInfo 接口
func (b *BaseStationInfo) GetCountryName() string {
	return b.CountryName
}

func (b *BaseStationInfo) GetRegionName() string {
	return b.RegionName
}

func (b *BaseStationInfo) GetCityName() string {
	return b.CityName
}

// BaseStation 基站数据库结构体
type BaseStation struct {
	reader *reader
	mu     sync.RWMutex // 保护并发访问
	cache  *sync.Map    // 添加缓存
}

// NewBaseStation 创建新的基站数据库实例
func NewBaseStation(name string) (*BaseStation, error) {
	r, e := newReader(name, &BaseStationInfo{})
	if e != nil {
		return nil, e
	}

	return &BaseStation{
		reader: r,
		cache:  &sync.Map{},
	}, nil
}

// Reload 重新加载数据库文件
func (db *BaseStation) Reload(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := os.Stat(name)
	if err != nil {
		return err
	}

	reader, err := newReader(name, &BaseStationInfo{})
	if err != nil {
		return err
	}

	db.reader = reader
	return nil
}

// Find 查找IP地址对应的基站信息(字符串切片形式)
func (db *BaseStation) Find(addr, language string) ([]string, error) {
	if net.ParseIP(addr) == nil {
		return nil, ErrInvalidIP
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	result, err := db.reader.find1(addr, language)
	if err != nil {
		return nil, fmt.Errorf("查找IP信息失败: %v", err)
	}

	return result, nil
}

// FindMap 查找IP地址对应的基站信息(Map形式)
func (db *BaseStation) FindMap(addr, language string) (map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := db.reader.find1(addr, language)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string, len(db.reader.meta.Fields))
	for k, v := range data {
		info[db.reader.meta.Fields[k]] = v
	}

	return info, nil
}

// FindInfo 查找IP地址对应的基站信息(结构体形式)
func (db *BaseStation) FindInfo(addr, language string) (*BaseStationInfo, error) {
	// 先查缓存
	if info, ok := db.cache.Load(addr + language); ok {
		return info.(*BaseStationInfo), nil
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := db.reader.FindMap(addr, language)
	if err != nil {
		return nil, err
	}

	info := &BaseStationInfo{
		CountryName: data["country_name"],
		RegionName:  data["region_name"],
		CityName:    data["city_name"],
		OwnerDomain: data["owner_domain"],
		IspDomain:   data["isp_domain"],
		BaseStation: data["base_station"],
	}

	// 写入缓存
	db.cache.Store(addr+language, info)

	return info, nil
}

// IsIPv4 检查是否支持IPv4
func (db *BaseStation) IsIPv4() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.IsIPv4Support()
}

// IsIPv6 检查是否支持IPv6
func (db *BaseStation) IsIPv6() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.IsIPv6Support()
}

// Languages 返回支持的语言列表
func (db *BaseStation) Languages() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.Languages()
}

// Fields 返回支持的字段列表
func (db *BaseStation) Fields() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.meta.Fields
}

// BuildTime 返回数据库构建时间
func (db *BaseStation) BuildTime() time.Time {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.Build()
}

type BatchResult struct {
	IP    string
	Info  *BaseStationInfo
	Error error
}

func (db *BaseStation) BatchFind(addrs []string, language string) []BatchResult {
	results := make([]BatchResult, len(addrs))
	for i, addr := range addrs {
		info, err := db.FindInfo(addr, language)
		results[i] = BatchResult{
			IP:    addr,
			Info:  info,
			Error: err,
		}
	}
	return results
}

var (
	ErrInvalidIP = errors.New("无效的IP地址")
	ErrNotFound  = errors.New("未找到IP信息")
)
