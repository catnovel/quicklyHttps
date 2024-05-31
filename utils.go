package quicklyHttps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type HandleRequestResult func(rawRequest *http.Request) *http.Request
type HandleResponseResult func(rawRequest *Request, response *Response)

const (
	defaultHeaderAuthorizationKey = "Authorization"
	defaultAuthScheme             = "Bearer"
	retryMax                      = 5
	ContentTypeJson               = "application/json"
	ContentTypeForm               = "application/x-www-form-urlencoded"
	ContentTypeXml                = "application/xml"
	ContentTypeStream             = "application/octet-stream"
	ContentTypeText               = "text/plain"
	ContentTypeHtml               = "text/html"
	ContentTypeMultipart          = "multipart/form-data"
)

// LeveledLogger 接口定义了分级日志记录的方法
type LeveledLogger interface {
	Error(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	WithContext(ctx context.Context) LeveledLogger
}

// standardLogger 是实现 LeveledLogger 接口的默认日志记录器
type standardLogger struct {
	ctx  context.Context
	file *os.File
	*log.Logger
}

// newStandardLogger 创建一个新的标准日志记录器
func newStandardLogger() *standardLogger {
	return &standardLogger{
		file:   os.Stderr,
		ctx:    context.Background(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (l *standardLogger) SetOutput(w io.Writer) {
	l.Logger.SetOutput(w)
}

// Error 实现 LeveledLogger 的 Error 方法
func (l *standardLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Printf("[ERROR] "+msg, keysAndValues...)
}

// Info 实现 LeveledLogger 的 Info 方法
func (l *standardLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Printf("[INFO] "+msg, keysAndValues...)
}

// Debug 实现 LeveledLogger 的 Debug 方法
func (l *standardLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.Printf("[DEBUG] "+msg, keysAndValues...)
}

// Warn 实现 LeveledLogger 的 Warn 方法
func (l *standardLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.Printf("[WARN] "+msg, keysAndValues...)
}
func (l *standardLogger) WithContext(ctx context.Context) LeveledLogger {
	l.ctx = ctx
	return l
}

// IsStringEmpty method tells whether given string is empty or not
func IsStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// parseCookies 将 Cookie 字符串解析为 *http.Cookie 切片
func parseCookies(cookies string) []*http.Cookie {
	var result []*http.Cookie
	for _, cookie := range strings.Split(cookies, ";") {
		cookie = strings.TrimSpace(cookie)
		if cookie == "" {
			continue
		}
		cookieArr := strings.SplitN(cookie, "=", 2)
		if len(cookieArr) == 2 {
			result = append(result, &http.Cookie{Name: cookieArr[0], Value: cookieArr[1]})
		}
	}
	return result
}

// marshalJSON marshals the input data to a JSON string.
func marshalJSON(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		if json.Valid([]byte(v)) {
			return v, nil
		}
		return "", fmt.Errorf("invalid JSON string")
	default:
		jsonString, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(jsonString), nil
	}
}

// ConvertGBKToUTF8 将 GBK 编码的字节数组转换为 UTF-8 编码
func ConvertGBKToUTF8(gbkData []byte) ([]byte, error) {
	reader := transform.NewReader(
		io.NopCloser(bytes.NewReader(gbkData)),
		simplifiedchinese.GBK.NewDecoder(),
	)
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to convert GBK to UTF-8: %w", err)
	}
	return utf8Data, nil
}

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	hostnameAndPort := strings.Split(host, ":")
	if len(hostnameAndPort) > 1 && hostnameAndPort[1] == "" {
		return hostnameAndPort[0]
	}
	return host
}

type User struct {
	Username, Password string
}
