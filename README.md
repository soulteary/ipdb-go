# ipdb-go

IPIP.net officially supported IP database ipdb format parsing library

# Installing

```bash
go get github.com/soulteary/ipdb-go
```

# Code Example

## 支持IPDB格式
<pre>
<code>
package main

import (
	"github.com/soulteary/ipdb-go"
	"fmt"
	"log"
)

func main() {
	db, err := ipdb.NewCity("/path/to/city.ipv4.ipdb")
	if err != nil {
		log.Fatal(err)
	}

	db.Reload("/path/to/city.ipv4.ipdb") // 更新 ipdb 文件后可调用 Reload 方法重新加载内容

	fmt.Println(db.IsIPv4()) // check database support ip type
	fmt.Println(db.IsIPv6()) // check database support ip type
	fmt.Println(db.BuildTime()) // database build time
	fmt.Println(db.Languages()) // database support language
	fmt.Println(db.Fields()) // database support fields

	fmt.Println(db.FindInfo("2001:250:200::", "CN")) // return CityInfo
	fmt.Println(db.Find("1.1.1.1", "CN")) // return []string
	fmt.Println(db.FindMap("118.28.8.8", "CN")) // return map[string]string
	fmt.Println(db.FindInfo("127.0.0.1", "CN")) // return CityInfo

	fmt.Println()
}
</code>
</pre>
## 数据字段说明
<pre>
region_name  : 省名字 
city_name    : 城市名字 
owner_domain : 所有者  
isp_domain  : 运营商 
</pre>
