package ipdb

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"
)

type IDCInfo struct {
	CountryName string `json:"country_name"`
	RegionName  string `json:"region_name"`
	CityName    string `json:"city_name"`
	OwnerDomain string `json:"owner_domain"`
	IspDomain   string `json:"isp_domain"`
	IDC         string `json:"idc"`
}

type IDC struct {
	reader *reader
	cache  *sync.Map
	mu     sync.RWMutex
}

func NewIDC(name string) (*IDC, error) {
	r, e := newReader(name, &IDCInfo{})
	if e != nil {
		return nil, fmt.Errorf("初始化IDC数据库失败: %v", e)
	}

	return &IDC{
		reader: r,
		cache:  &sync.Map{},
	}, nil
}

func (db *IDC) Reload(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := os.Stat(name); err != nil {
		return fmt.Errorf("数据库文件不存在: %v", err)
	}

	reader, err := newReader(name, &IDCInfo{})
	if err != nil {
		return fmt.Errorf("加载数据库失败: %v", err)
	}

	db.reader = reader
	db.ClearCache()

	return nil
}

func (db *IDC) Find(addr, language string) ([]string, error) {
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.find1(addr, language)
}

func (db *IDC) FindMap(addr, language string) (map[string]string, error) {
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

func (db *IDC) FindInfo(addr, language string) (*IDCInfo, error) {
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	cacheKey := addr + language
	if val, ok := db.cache.Load(cacheKey); ok {
		if info, ok := val.(*IDCInfo); ok {
			return info, nil
		}
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := db.reader.FindMap(addr, language)
	if err != nil {
		return nil, fmt.Errorf("查找IP信息失败: %v", err)
	}

	info := &IDCInfo{}
	val := reflect.ValueOf(info).Elem()

	for k, v := range data {
		field := val.FieldByName(db.reader.refType[k])
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		field.SetString(v)
	}

	db.cache.Store(cacheKey, info)

	return info, nil
}

func (db *IDC) ClearCache() {
	db.cache = &sync.Map{}
}

func (db *IDC) IsIPv4() bool {
	return db.reader.IsIPv4Support()
}

func (db *IDC) IsIPv6() bool {
	return db.reader.IsIPv6Support()
}

func (db *IDC) Languages() []string {
	return db.reader.Languages()
}

func (db *IDC) Fields() []string {
	return db.reader.meta.Fields
}

func (db *IDC) BuildTime() time.Time {
	return db.reader.Build()
}
