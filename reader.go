package ipdb

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	IPv4 = 0x01
	IPv6 = 0x02
)

var (
	ErrFileSize          = errors.New("IP数据库文件大小错误")
	ErrMetaData          = errors.New("IP数据库元数据错误")
	ErrReadFull          = errors.New("IP数据库读取错误")
	ErrDatabase          = errors.New("数据库错误")
	ErrIPFormat          = errors.New("IP地址格式错误")
	ErrNoSupportLanguage = errors.New("不支持该语言")
	ErrNoSupportIPv4     = errors.New("不支持IPv4")
	ErrNoSupportIPv6     = errors.New("不支持IPv6")
	ErrDataNotExists     = errors.New("数据不存在")
)

type MetaData struct {
	Build     int64          `json:"build"`
	IPVersion uint16         `json:"ip_version"`
	Languages map[string]int `json:"languages"`
	NodeCount int            `json:"node_count"`
	TotalSize int            `json:"total_size"`
	Fields    []string       `json:"fields"`
}

type reader struct {
	sync.RWMutex
	fileSize  int
	nodeCount int
	v4offset  int

	meta MetaData
	data []byte

	refType map[string]string
	cache   sync.Map
}

func newReader(name string, obj interface{}) (*reader, error) {
	var err error
	var fileInfo os.FileInfo
	fileInfo, err = os.Stat(name)
	if err != nil {
		return nil, err
	}
	fileSize := int(fileInfo.Size())
	if fileSize < 4 {
		return nil, ErrFileSize
	}
	body, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, ErrReadFull
	}

	return initBytes(body, fileSize, obj)
}

func newReaderFromBytes(body []byte, obj interface{}) (*reader, error) {
	if len(body) < 4 {
		return nil, ErrFileSize
	}
	return initBytes(body, len(body), obj)
}

func initBytes(body []byte, fileSize int, obj interface{}) (*reader, error) {
	var meta MetaData
	metaLength := int(binary.BigEndian.Uint32(body[0:4]))
	if fileSize < (4 + metaLength) {
		return nil, ErrFileSize
	}
	if err := json.Unmarshal(body[4:4+metaLength], &meta); err != nil {
		return nil, err
	}
	if len(meta.Languages) == 0 || len(meta.Fields) == 0 {
		return nil, ErrMetaData
	}
	if fileSize != (4 + metaLength + meta.TotalSize) {
		return nil, ErrFileSize
	}

	var dm map[string]string
	if obj != nil {
		t := reflect.TypeOf(obj).Elem()
		dm = make(map[string]string, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			k := t.Field(i).Tag.Get("json")
			dm[k] = t.Field(i).Name
		}
	}

	db := &reader{
		fileSize:  fileSize,
		nodeCount: meta.NodeCount,

		meta:    meta,
		refType: dm,

		data: body[4+metaLength:],
	}

	if db.v4offset == 0 {
		node := 0
		for i := 0; i < 96 && node < db.nodeCount; i++ {
			if i >= 80 {
				node = db.readNode(node, 1)
			} else {
				node = db.readNode(node, 0)
			}
		}
		db.v4offset = node
	}

	return db, nil
}

func (db *reader) Find(addr, language string) ([]string, error) {
	db.RLock()
	defer db.RUnlock()
	return db.find1(addr, language)
}

func (db *reader) FindMap(addr, language string) (map[string]string, error) {
	db.RLock()
	defer db.RUnlock()

	if val, ok := db.cache.Load(addr + language); ok {
		return val.(map[string]string), nil
	}

	data, err := db.find1(addr, language)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string, len(db.meta.Fields))
	for k, v := range data {
		info[db.meta.Fields[k]] = v
	}

	db.cache.Store(addr+language, info)

	return info, nil
}

func (db *reader) find0(addr string) ([]byte, error) {
	var err error
	var node int
	ipv := net.ParseIP(addr)
	if ip := ipv.To4(); ip != nil {
		if !db.IsIPv4Support() {
			return nil, ErrNoSupportIPv4
		}

		node, err = db.search(ip, 32)
	} else if ip := ipv.To16(); ip != nil {
		if !db.IsIPv6Support() {
			return nil, ErrNoSupportIPv6
		}

		node, err = db.search(ip, 128)
	} else {
		return nil, ErrIPFormat
	}

	if err != nil || node < 0 {
		return nil, err
	}

	return db.resolve(node)
}

func (db *reader) find1(addr, language string) ([]string, error) {
	off, ok := db.meta.Languages[language]
	if !ok {
		return nil, ErrNoSupportLanguage
	}

	body, err := db.find0(addr)
	if err != nil {
		return nil, err
	}

	str := string(body)
	tmp := strings.Split(str, "\t")

	if (off + len(db.meta.Fields)) > len(tmp) {
		return nil, ErrDatabase
	}

	return tmp[off : off+len(db.meta.Fields)], nil
}

func (db *reader) search(ip net.IP, bitCount int) (int, error) {
	node := 0
	if bitCount == 32 {
		node = db.v4offset
	}

	for i := 0; i < bitCount && node < db.nodeCount; i++ {
		node = db.readNode(node, int((ip[i>>3]>>(7-(i&7)))&1))
	}

	if node > db.nodeCount {
		return node, nil
	}

	return -1, ErrDataNotExists
}

func (db *reader) readNode(node, index int) int {
	off := node*8 + index*4
	return int(binary.BigEndian.Uint32(db.data[off : off+4]))
}

func (db *reader) resolve(node int) ([]byte, error) {
	resolved := node - db.nodeCount + db.nodeCount*8
	if resolved >= db.fileSize {
		return nil, ErrDatabase
	}

	size := int(binary.BigEndian.Uint16(db.data[resolved : resolved+2]))
	if (resolved + 2 + size) > len(db.data) {
		return nil, ErrDatabase
	}
	bytes := db.data[resolved+2 : resolved+2+size]

	return bytes, nil
}

func (db *reader) IsIPv4Support() bool {
	return (int(db.meta.IPVersion) & IPv4) == IPv4
}

func (db *reader) IsIPv6Support() bool {
	return (int(db.meta.IPVersion) & IPv6) == IPv6
}

func (db *reader) Build() time.Time {
	return time.Unix(db.meta.Build, 0).In(time.UTC)
}

func (db *reader) Languages() []string {
	ls := make([]string, 0, len(db.meta.Languages))
	for k := range db.meta.Languages {
		ls = append(ls, k)
	}
	return ls
}

func (db *reader) ClearCache() {
	db.cache = sync.Map{}
}
