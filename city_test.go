package ipdb_test

import (
	"fmt"
	"testing"

	"github.com/ipipdotnet/ipdb-go"
	"github.com/stretchr/testify/assert"
)

var db *ipdb.City

func init() {
	var err error
	db, err = ipdb.NewCity("city.free.ipdb")
	if err != nil {
		panic(err)
	}
}

func TestNewCity(t *testing.T) {
	tests := []struct {
		name    string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "valid database",
			dbPath:  "city.free.ipdb",
			wantErr: false,
		},
		{
			name:    "invalid database path",
			dbPath:  "not_exists.ipdb",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := ipdb.NewCity(tt.dbPath)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, db)
			assert.NotEmpty(t, db.BuildTime())
			assert.NotEmpty(t, db.Fields())
		})
	}
}

func TestCity_Find(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		language  string
		wantErr   bool
		checkFunc func(*testing.T, []string, error)
	}{
		{
			name:     "valid ip",
			ip:       "1.1.1.1",
			language: "CN",
			wantErr:  false,
			checkFunc: func(t *testing.T, result []string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			},
		},
		{
			name:     "invalid ip",
			ip:       "invalid.ip",
			language: "CN",
			wantErr:  true,
			checkFunc: func(t *testing.T, result []string, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := db.Find(tt.ip, tt.language)
			tt.checkFunc(t, result, err)
		})
	}
}

func TestCity_FindInfo(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		language  string
		wantErr   bool
		checkFunc func(*testing.T, *ipdb.CityInfo, error)
	}{
		{
			name:     "valid ip info",
			ip:       "123.123.123.123",
			language: "CN",
			wantErr:  false,
			checkFunc: func(t *testing.T, info *ipdb.CityInfo, err error) {
				fmt.Println(info)
				assert.NoError(t, err)
				assert.NotNil(t, info)

				// assert.NotEmpty(t, info.Route)
				// assert.NotEmpty(t, info.ASN)
				// assert.NotNil(t, info.ASNInfo)
			},
		},
		{
			name:     "invalid ip info",
			ip:       "invalid.ip",
			language: "CN",
			wantErr:  true,
			checkFunc: func(t *testing.T, info *ipdb.CityInfo, err error) {
				assert.Error(t, err)
				assert.Nil(t, info)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := db.FindInfo(tt.ip, tt.language)
			tt.checkFunc(t, info, err)
		})
	}
}

// 保留原有的基准测试
func BenchmarkCity_Find(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db.Find("118.28.1.1", "CN")
	}
}

func BenchmarkCity_FindMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db.FindMap("118.28.1.1", "CN")
	}
}

func BenchmarkCity_FindInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db.FindInfo("118.28.1.1", "CN")
	}
}
