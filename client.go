package quicklyHttps

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"strings"
	"sync"
	"time"
)

// Client 用于发出 HTTP 请求，添加了额外的功能
type Client struct {
	*http.Client                                                   // HTTP 客户端
	BasicAuthToken          string                                 // 基本认证令牌
	HeaderAuthorizationKey  string                                 // 认证头部键
	AuthScheme              string                                 // 认证方案
	Method                  string                                 // 请求方法
	BaseURL                 string                                 // 请求的基础 URL
	Timeout                 time.Duration                          // 请求超时
	Logger                  LeveledLogger                          // 日志记录器
	RetryMax                int                                    // 最大重试次数
	Cookies                 []*http.Cookie                         // 每个请求都要发送的 cookie
	Header                  http.Header                            // 每个请求都要发送的头部
	QueryParams             map[string]string                      // 请求的查询参数
	Body                    string                                 // 请求的主体内容
	FormParams              urlpkg.Values                          // 表单参数
	Debug                   bool                                   // 是否启用调试模式
	loggerInit              sync.Once                              // 用于初始化日志记录器
	UserInfo                *User                                  // 用户信息, 用于请求认证
	handleRequestResultFunc HandleRequestResult                    // 处理请求结果的函数
	jsonMarshal             func(v interface{}) ([]byte, error)    // JSON 编码器
	jsonUnmarshal           func(data []byte, v interface{}) error // JSON 解码器
	xmlMarshal              func(v interface{}) ([]byte, error)    // XML 编码器
	xmlUnmarshal            func(data []byte, v interface{}) error // XML 解码器
}

// NewClient 使用默认设置创建一个新的 Client
func NewClient() *Client {
	c := &Client{
		RetryMax:       retryMax,
		AuthScheme:     defaultAuthScheme,
		BasicAuthToken: defaultHeaderAuthorizationKey,
		Header:         make(http.Header),
		Cookies:        make([]*http.Cookie, 0),
		Logger:         newStandardLogger(),
		QueryParams:    make(map[string]string),
		FormParams:     make(urlpkg.Values),
		Timeout:        30 * time.Second,
		jsonMarshal:    json.Marshal,
		jsonUnmarshal:  json.Unmarshal,
		xmlMarshal:     xml.Marshal,
		xmlUnmarshal:   xml.Unmarshal,
	}
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c.Client = &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
	if c.Client.Transport == nil {
		c.Client.Transport = createTransport(nil)
	}
	return c
}

// SetProxyURL 设置代理服务器 URL
func (c *Client) SetProxyURL(proxy string) *Client {
	proxyURL, ok := urlpkg.Parse(proxy)
	if ok != nil {
		c.logger().Error("invalid proxy URL", "error", ok)
	} else {
		c.Client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
	}
	return c
}

func (c *Client) SetBasicAuth(username, password string) *Client {
	c.UserInfo = &User{Username: username, Password: password}
	return c
}
func (c *Client) SetBasicAuthToken(token string) *Client {
	c.BasicAuthToken = token
	return c
}
func (c *Client) SetAuthScheme(scheme string) *Client {
	c.AuthScheme = scheme
	return c
}

// GetCookies get cookies from the underlying `http.Client`'s `CookieJar`.
func (c *Client) GetCookies(url string) ([]*http.Cookie, error) {
	if c.Client.Jar == nil {
		return nil, errors.New("cookie jar is not enabled")
	}
	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}
	return c.Client.Jar.Cookies(u), nil
}

func (c *Client) ClearCookies() *Client {
	c.Cookies = nil
	return c
}

func (c *Client) SetHandleRequestResultFunc(f HandleRequestResult) *Client {
	if f != nil {
		c.handleRequestResultFunc = f
	}
	return c
}

// SetCheckRedirect 设置重定向函数
func (c *Client) SetCheckRedirect(f func(req *http.Request, via []*http.Request) error) *Client {
	c.Client.CheckRedirect = f
	return c
}

// SetDebug 启用或禁用调试模式
func (c *Client) SetDebug(debug bool) *Client {
	c.Debug = debug
	return c
}

// SetUserAgent 设置 User-Agent 头
func (c *Client) SetUserAgent(userAgent string) *Client {
	return c.SetHeader("User-Agent", userAgent)
}

// SetRetryMax 设置最大重试次数
func (c *Client) SetRetryMax(retryMax int) *Client {
	if retryMax < 0 {
		retryMax = 1
	} else {
		c.RetryMax = retryMax
	}
	return c
}

// SetBaseURL 设置基础 URL
func (c *Client) SetBaseURL(baseURL string) *Client {
	c.BaseURL = strings.TrimSuffix(baseURL, "/")
	return c
}

// SetHeader 设置单个请求头部
func (c *Client) SetHeader(key, value string) *Client {
	c.Header.Set(key, value)
	return c
}

// SetHeaders 设置多个请求头部
func (c *Client) SetHeaders(headers map[string]string) *Client {
	for key, value := range headers {
		c.SetHeader(key, value)
	}
	return c
}

// SetBody 设置请求体
func (c *Client) SetBody(body string) *Client {
	c.Body = body
	return c
}

// SetCookie 解析并设置 cookie 字符串
func (c *Client) SetCookie(cookies string) *Client {
	for _, cookie := range parseCookies(cookies) {
		c.Cookies = append(c.Cookies, cookie)
	}
	return c
}

// SetCookiesRaw 设置原始 cookie 切片
func (c *Client) SetCookiesRaw(cookies []*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cookies...)
	return c
}

// SetCookieRaw 增加单个原始 cookie
func (c *Client) SetCookieRaw(cookie *http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cookie)
	return c
}

// SetQueryParams 设置多个查询参数
func (c *Client) SetQueryParams(params map[string]string) *Client {
	for key, value := range params {
		c.SetQueryParam(key, value)
	}
	return c
}

// SetQueryParam 设置单个查询参数
func (c *Client) SetQueryParam(key, value string) *Client {
	c.QueryParams[key] = value
	return c
}

// SetFormParams 设置多个表单参数
func (c *Client) SetFormParams(params map[string]string) *Client {
	for key, value := range params {
		c.SetFormParam(key, value)
	}
	return c
}

// SetFormParam 设置单个表单参数
func (c *Client) SetFormParam(key, value string) *Client {
	c.FormParams.Set(key, value)
	return c
}

func (c *Client) R() *Request {
	if c.Method == "" {
		c.Method = http.MethodGet
	}
	return &Request{
		rawClient:   c,
		method:      c.Method,
		body:        c.Body,
		Header:      c.Header.Clone(),
		startedAt:   time.Now(),
		queryParams: copyMap(c.QueryParams),
		formParams:  copyValues(c.FormParams),
		cookies:     append([]*http.Cookie{}, c.Cookies...),
	}
}

// copyMap 用于复制 map[string]string
func copyMap(original map[string]string) map[string]string {
	c := make(map[string]string, len(original))
	for key, value := range original {
		c[key] = value
	}
	return c
}

// copyValues 用于复制 url.Values
func copyValues(original urlpkg.Values) urlpkg.Values {
	c := make(urlpkg.Values, len(original))
	for key, values := range original {
		c[key] = append([]string(nil), values...)
	}
	return c
}

// SetMethod 设置请求方法
func (c *Client) SetMethod(method string) *Client {
	c.Method = method
	return c
}

func (r *Request) Do() (*Response, error) {
	if r.rawClient.Timeout > 0 {
		r.rawClient.Client.Timeout = r.rawClient.Timeout
	}
	response, err := r.rawClient.Client.Do(r.Request)
	if err != nil {
		r.rawClient.logger().Error("request failed", "error", err)
		r.logRequest()
		return nil, err
	}
	do := &Response{
		rawRequest:      r,
		Response:        response,
		jsonUnmarshaler: json.Unmarshal,
		jsonMarshaler:   json.Marshal,
		receivedAt:      time.Now(),
	}
	defer func() {
		if do.rawRequest.rawClient.Debug {
			do.rawRequest.logRequest()
			do.logResponse()
		}
	}()
	return do, nil
}

// logger 返回日志记录器实例
func (c *Client) logger() LeveledLogger {
	c.loggerInit.Do(func() {
		if c.Logger == nil {
			c.Logger = newStandardLogger()
		} else {
			switch c.Logger.(type) {
			case LeveledLogger:
				// ok
			default:
				panic(fmt.Sprintf("invalid logger type passed, must be LeveledLogger, was %T", c.Logger))
			}
		}
	})
	return c.Logger
}

// SetBodyJSON 将请求体设置为 JSON 对象
func (c *Client) SetBodyJSON(data interface{}) *Client {
	jsonString, err := marshalJSON(data)
	if err != nil {
		c.logger().Error("failed to marshal JSON", "error", err)
		return c
	}
	c.Body = jsonString
	c.SetHeader("Content-Type", ContentTypeJson)
	return c
}

// SetTimeout 设置请求超时
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.Timeout = timeout
	return c
}
