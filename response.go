package quicklyHttps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Response 封装了 HTTP 响应，提供了便捷的方法来处理响应。
type Response struct {
	*http.Response
	Err             error
	body            []byte
	bodyMutex       sync.Mutex
	rawRequest      *Request
	jsonMarshaler   func(v any) ([]byte, error)
	jsonUnmarshaler func(data []byte, v any) error
	receivedAt      time.Time
	error           interface{}
	result          interface{}
}

// Body 返回响应体的字节数组。
func (r *Response) Body() []byte {
	if r.Response == nil {
		return nil
	}
	r.bodyMutex.Lock()
	defer r.bodyMutex.Unlock()
	if r.body == nil && r.Response.Body != nil {
		var err error
		r.body, err = readBody(r.Response.Body)
		if err != nil {
			r.Err = err
			return nil
		}
	}
	return r.body
}

// String 返回响应体的字符串表示。
func (r *Response) String() string {
	body := r.Body()
	if body == nil {
		return ""
	}
	return string(body)
}

// readBody 读取并返回响应体。
func readBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// StatusCode 返回响应的状态码。
func (r *Response) StatusCode() int {
	if r.Response != nil {
		return r.Response.StatusCode
	}
	return 0
}

// Header 返回响应的头部信息。
func (r *Response) Header() http.Header {
	if r.Response != nil {
		return r.Response.Header
	}
	return http.Header{}
}

// JSON 解析响应体为 JSON。
func (r *Response) JSON(v interface{}) error {
	return r.jsonUnmarshaler(r.Body(), v)
}

// IsSuccess 检查响应是否表示成功的请求。
func (r *Response) IsSuccess() bool {
	return r.StatusCode() >= 200 && r.StatusCode() < 300
}

// IsClientError 检查响应是否表示客户端错误。
func (r *Response) IsClientError() bool {
	return r.StatusCode() >= 400 && r.StatusCode() < 500
}

// IsServerError 检查响应是否表示服务器错误。
func (r *Response) IsServerError() bool {
	return r.StatusCode() >= 500 && r.StatusCode() < 600
}

// SaveToFile 将响应体保存到指定文件。
func (r *Response) SaveToFile(filepath string) error {
	if r.body == nil {
		var err error
		r.body, err = readBody(r.Response.Body)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(filepath, r.body, 0644)
}

// ToBytesBuffer 返回响应体的字节缓冲区。
func (r *Response) ToBytesBuffer() *bytes.Buffer {
	return bytes.NewBuffer(r.Body())
}

// ToMap 将响应体解析为 map。
func (r *Response) ToMap() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := r.jsonUnmarshaler(r.Body(), &result)
	return result, err
}

// logResponse 记录响应信息
func (r *Response) logResponse() {
	logger := r.rawRequest.rawClient.logger()

	// 将 headers 和 cookies 转换为更易读的格式
	headers := make(map[string]string)
	for key, values := range r.Header() {
		headers[key] = strings.Join(values, ", ")
	}

	cookies := make([]string, len(r.Cookies()))
	for i, cookie := range r.Cookies() {
		cookies[i] = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
	}

	// 创建日志消息
	logMessage := map[string]interface{}{
		"status_code": r.StatusCode(),
		"status":      r.Status,
		"headers":     headers,
		"cookies":     cookies,
		"body":        r.String(),
	}

	// 记录日志
	logger.Info("Received response", logMessage)
}

// DetectEncoding 检测响应体的编码并转换为 UTF-8
func (r *Response) DetectEncoding() error {
	r.bodyMutex.Lock()
	defer r.bodyMutex.Unlock()
	body := r.Body()
	if !utf8.Valid(body) {
		// 假设响应体是 GBK 编码，进行转换
		decodedBody, err := ConvertGBKToUTF8(body)
		if err != nil {
			return fmt.Errorf("failed to convert body to UTF-8: %w", err)
		}
		r.body = decodedBody
	}
	return nil
}

// Gjson 解析响应体为 gjson.Result
func (r *Response) Gjson() gjson.Result {
	return gjson.ParseBytes(r.Body())
}

// GetCookies 获取响应的 Cookies
func (r *Response) GetCookies() []*http.Cookie {
	return r.Cookies()
}

// GetHeader 获取指定的响应头信息
func (r *Response) GetHeader(key string) string {
	return r.Header().Get(key)
}

// HasHeader 检查指定的响应头是否存在
func (r *Response) HasHeader(key string) bool {
	_, ok := r.Header()[key]
	return ok
}

// GetHeaderValues 获取指定的响应头的所有值
func (r *Response) GetHeaderValues(key string) []string {
	return r.Header()[key]
}

// PrettyPrint 以易读的格式打印响应体
func (r *Response) PrettyPrint() string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, r.Body(), "", "  ")
	if err != nil {
		return r.String()
	}
	return prettyJSON.String()
}
