# IPDB (Go version)

[![IPDB Database API Document](https://godoc.org/github.com/soulteary/ipdb-go?status.svg)](https://godoc.org/github.com/soulteary/ipdb-go)

IPIP.net 官方支持的 IP 数据库 ipdb 格式解析库

## 安装方法

```bash
go get github.com/soulteary/ipdb-go
```

## 代码示例

```go
package main

import (
	"fmt"
	"log"

	"github.com/soulteary/ipdb-go"
)

func main() {
	// 创建城市数据库实例
	db, err := ipdb.NewCity("/path/to/city.ipv4.ipdb")
	if err != nil {
		log.Fatal(err)
	}
	// 支持热更新，更新 ipdb 文件后调用 Reload 方法
	db.Reload("/path/to/city.ipv4.ipdb")
	// 数据库元信息查询
	fmt.Println(db.IsIPv4())    // 检查是否支持 IPv4
	fmt.Println(db.IsIPv6())    // 检查是否支持 IPv6
	fmt.Println(db.BuildTime()) // 数据库构建时间
	fmt.Println(db.Languages()) // 支持的语言列表
	fmt.Println(db.Fields())    // 支持的字段列表
	// IP 查询示例
	fmt.Println(db.FindInfo("2001:250:200::", "CN")) // 返回 CityInfo 结构体
	fmt.Println(db.Find("1.1.1.1", "CN"))            // 返回字符串数组
	fmt.Println(db.FindMap("118.28.8.8", "CN"))      // 返回字符串映射
	fmt.Println(db.FindInfo("127.0.0.1", "CN"))      // 返回 CityInfo 结构体
}
```

### 返回结果字段说明

| 字段名 | 说明 |
|--------|------|
| country_name | 国家名称 |
| region_name | 省份名称 |
| city_name | 城市名称 |
| owner_domain | 所有者 |
| isp_domain | 运营商 |
| latitude | 纬度 |
| longitude | 经度 |
| timezone | 时区 |
| utc_offset | UTC 时区 |
| china_admin_code | 中国行政区划代码 |
| idd_code | 国家电话号码前缀 |
| country_code | 国家二位代码 |
| continent_code | 大洲代码 |
| idc | IDC / VPN |
| base_station | 基站 / WIFI |
| country_code3 | 国家三位代码 |
| european_union | 是否为欧盟成员国（1:是 0:否）|
| currency_code | 当前国家货币代码 |
| currency_name | 当前国家货币名称 |
| anycast | ANYCAST |

## 支持的查询方法

- `FindInfo(ip, language)`: 返回结构化的 CityInfo 对象
- `Find(ip, language)`: 返回字符串数组
- `FindMap(ip, language)`: 返回字符串映射

## 注意事项

1. 支持 IPv4 和 IPv6 地址
2. 支持多语言查询
3. 数据库文件支持热更新
4. 所有查询方法都是线程安全的

## 许可证

该项目采用 MIT 许可证
