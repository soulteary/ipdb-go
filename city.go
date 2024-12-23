package ipdb

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

// CityInfo is City Database Content
type CityInfo struct {
	CountryName       string `json:"country_name"`
	RegionName        string `json:"region_name"`
	CityName          string `json:"city_name"`
	DistrictName      string `json:"district_name"`
	OwnerDomain       string `json:"owner_domain"`
	IspDomain         string `json:"isp_domain"`
	Latitude          string `json:"latitude"`
	Longitude         string `json:"longitude"`
	Timezone          string `json:"timezone"`
	UtcOffset         string `json:"utc_offset"`
	ChinaRegionCode   string `json:"china_region_code"`
	ChinaCityCode     string `json:"china_city_code"`
	ChinaDistrictCode string `json:"china_district_code"`
	ChinaAdminCode    string `json:"china_admin_code"`
	IddCode           string `json:"idd_code"`
	CountryCode       string `json:"country_code"`
	ContinentCode     string `json:"continent_code"`
	IDC               string `json:"idc"`
	BaseStation       string `json:"base_station"`
	CountryCode3      string `json:"country_code3"`
	EuropeanUnion     string `json:"european_union"`
	CurrencyCode      string `json:"currency_code"`
	CurrencyName      string `json:"currency_name"`
	Anycast           string `json:"anycast"`

	Line string `json:"line"`

	DistrictInfo DistrictInfo `json:"district_info"`

	Route   string    `json:"route"`
	ASN     string    `json:"asn"`
	ASNInfo []ASNInfo `json:"asn_info"`

	AreaCode string `json:"area_code"`

	UsageType string `json:"usage_type"`
}

type ASNInfo struct {
	ASN      int    `json:"asn"`
	Registry string `json:"reg"`
	Country  string `json:"cc"`
	Net      string `json:"net"`
	Org      string `json:"org"`
	Type     string `json:"type"`
	Domain   string `json:"domain"`
}

// City struct
type City struct {
	reader *reader
	cache  *sync.Map    // 添加缓存
	mu     sync.RWMutex // 用于保护并发访问
}

// NewCity initialize
func NewCity(name string) (*City, error) {
	r, e := newReader(name, &CityInfo{})
	if e != nil {
		return nil, fmt.Errorf("初始化City数据库失败: %v", e)
	}

	return &City{
		reader: r,
		cache:  &sync.Map{},
	}, nil
}

// NewCityFromBytes initialize from bytes
func NewCityFromBytes(bs []byte) (*City, error) {
	r, e := newReaderFromBytes(bs, &CityInfo{})
	if e != nil {
		return nil, fmt.Errorf("从字节数据初始化City数据库失败: %v", e)
	}

	return &City{
		reader: r,
		cache:  &sync.Map{},
	}, nil
}

// Reload the database
func (db *City) Reload(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := os.Stat(name); err != nil {
		return fmt.Errorf("数据库文件不存在: %v", err)
	}

	reader, err := newReader(name, &CityInfo{})
	if err != nil {
		return fmt.Errorf("加载数据库失败: %v", err)
	}

	db.reader = reader
	db.ClearCache() // 清理缓存

	return nil
}

// ClearCache clears the internal cache
func (db *City) ClearCache() {
	db.cache = &sync.Map{}
}

// validateIP validates IP address format
func validateIP(addr string) error {
	if net.ParseIP(addr) == nil {
		return fmt.Errorf("无效的IP地址格式: %s", addr)
	}
	return nil
}

// FindInfo query with addr
func (db *City) FindInfo(addr, language string) (*CityInfo, error) {
	// 验证IP地址
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	// 检查缓存
	if val, ok := db.cache.Load(addr + language); ok {
		if info, ok := val.(*CityInfo); ok {
			return info, nil
		}
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := db.reader.FindMap(addr, language)
	if err != nil {
		return nil, fmt.Errorf("查找IP信息失败: %v", err)
	}

	info := &CityInfo{}

	// 使用反射优化
	val := reflect.ValueOf(info).Elem()
	for k, v := range data {
		field := val.FieldByName(db.reader.refType[k])
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		switch field.Type().String() {
		case "[]ipdb.ASNInfo":
			var asnList []ASNInfo
			if err := json.Unmarshal([]byte(v), &asnList); err == nil {
				field.Set(reflect.ValueOf(asnList))
			}
		case "ipdb.DistrictInfo":
			var dist DistrictInfo
			if err := json.Unmarshal([]byte(v), &dist); err == nil {
				field.Set(reflect.ValueOf(dist))
			}
		default:
			field.SetString(v)
		}
	}

	// 存入缓存
	db.cache.Store(addr+language, info)

	return info, nil
}

// Find query with addr
func (db *City) Find(addr, language string) ([]string, error) {
	if err := validateIP(addr); err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.reader.find1(addr, language)
}

// FindMap query with addr
func (db *City) FindMap(addr, language string) (map[string]string, error) {
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

// IsIPv4 whether support ipv4
func (db *City) IsIPv4() bool {
	return db.reader.IsIPv4Support()
}

// IsIPv6 whether support ipv6
func (db *City) IsIPv6() bool {
	return db.reader.IsIPv6Support()
}

// Languages return support languages
func (db *City) Languages() []string {
	return db.reader.Languages()
}

// Fields return support fields
func (db *City) Fields() []string {
	return db.reader.meta.Fields
}

// BuildTime return database build Time
func (db *City) BuildTime() time.Time {
	return db.reader.Build()
}
