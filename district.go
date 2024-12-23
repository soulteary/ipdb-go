package ipdb

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"
)

type DistrictInfo struct {
	CountryName    string `json:"country_name"`
	RegionName     string `json:"region_name"`
	CityName       string `json:"city_name"`
	DistrictName   string `json:"district_name"`
	ChinaAdminCode string `json:"china_admin_code"`
	CoveringRadius string `json:"covering_radius"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
}

type District struct {
	reader *reader
	cache  *sync.Map    // 添加缓存
	mu     sync.RWMutex // 用于保护并发访问
}

func NewDistrict(name string) (*District, error) {
	r, e := newReader(name, &DistrictInfo{})
	if e != nil {
		return nil, fmt.Errorf("初始化District数据库失败: %v", e)
	}

	return &District{
		reader: r,
		cache:  &sync.Map{},
	}, nil
}

func (db *District) Reload(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := os.Stat(name); err != nil {
		return fmt.Errorf("数据库文件不存在: %v", err)
	}

	reader, err := newReader(name, &DistrictInfo{})
	if err != nil {
		return fmt.Errorf("加载数据库失败: %v", err)
	}

	db.reader = reader
	db.ClearCache() // 清理缓存

	return nil
}

// ClearCache 清理缓存
func (db *District) ClearCache() {
	db.cache = &sync.Map{}
}

func (db *District) Find(addr, language string) ([]string, error) {
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.find1(addr, language)
}

func (db *District) FindMap(addr, language string) (map[string]string, error) {
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := db.reader.find1(addr, language)
	if err != nil {
		return nil, fmt.Errorf("查找IP信息失败: %v", err)
	}

	info := make(map[string]string, len(db.reader.meta.Fields))
	for k, v := range data {
		info[db.reader.meta.Fields[k]] = v
	}

	return info, nil
}

func (db *District) FindInfo(addr, language string) (*DistrictInfo, error) {
	// 验证IP地址
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	// 检查缓存
	if val, ok := db.cache.Load(addr + language); ok {
		if info, ok := val.(*DistrictInfo); ok {
			return info, nil
		}
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := db.reader.FindMap(addr, language)
	if err != nil {
		return nil, fmt.Errorf("查找IP信息失败: %v", err)
	}

	info := &DistrictInfo{}
	val := reflect.ValueOf(info).Elem()

	for k, v := range data {
		field := val.FieldByName(db.reader.refType[k])
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		field.SetString(v)
	}

	// 存入缓存
	db.cache.Store(addr+language, info)

	return info, nil
}

func (db *District) IsIPv4() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.IsIPv4Support()
}

func (db *District) IsIPv6() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.IsIPv6Support()
}

func (db *District) Languages() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.Languages()
}

func (db *District) Fields() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.meta.Fields
}

func (db *District) BuildTime() time.Time {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.Build()
}
