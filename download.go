package ipdb

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Download 结构体用于处理文件下载
type Download struct {
	URL        *url.URL
	Progress   float64
	httpClient *http.Client
}

// Progress 用于跟踪下载进度的回调函数类型
type ProgressFunc func(current, total int64)

// NewDownload 创建新的下载实例
func NewDownload(httpUrl string) (*Download, error) {
	v, err := url.Parse(httpUrl)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %v", err)
	}

	return &Download{
		URL: v,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SaveToFile 将URL指向的文件下载到指定路径
func (dl *Download) SaveToFile(fn string, progress ProgressFunc) error {
	// 创建上下文用于超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", dl.URL.String(), nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 发送请求
	resp, err := dl.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	// 获取文件大小
	fileSize := resp.ContentLength

	// 创建进度跟踪器
	counter := &WriteCounter{
		Total:    fileSize,
		Progress: progress,
	}

	// 复制数据到文件
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// WriteCounter 用于跟踪写入进度
type WriteCounter struct {
	Current  int64
	Total    int64
	Progress ProgressFunc
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Current += int64(n)
	if wc.Progress != nil {
		wc.Progress(wc.Current, wc.Total)
	}
	return n, nil
}
