package quicklyHttps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	urlpkg "net/url"
	"strings"
	"time"
)

// Request 封装了 HTTP 请求及其相关数据
type Request struct {
	*http.Request
	ctx         context.Context
	method      string
	GetBody     func() (io.ReadCloser, error)
	startedAt   time.Time
	body        string
	urlPoint    string
	Header      http.Header
	cookies     []*http.Cookie
	queryParams map[string]string
	formParams  url.Values
	rawClient   *Client
}

// logRequest 记录请求信息
func (r *Request) logRequest() {
	logger := r.rawClient.logger()
	// 将 headers 和 cookies 转换为更易读的格式
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = strings.Join(values, ", ")
	}

	cookies := make([]string, len(r.cookies))
	for i, cookie := range r.cookies {
		cookies[i] = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
	}
	// 创建日志消息
	logMessage := map[string]interface{}{
		"method":       r.Request.Method,
		"url":          r.Request.URL.String(),
		"headers":      headers,
		"cookies":      cookies,
		"query_params": r.queryParams,
		"form_params":  r.formParams,
		"body":         r.body,
	}

	// 记录日志
	logger.Error("Performing request", logMessage)
}

// SetMethod 设置请求方法
func (r *Request) SetMethod(method string) *Request {
	r.method = method
	return r
}

// SetHeader 设置单个请求头
func (r *Request) SetHeader(key, value string) *Request {
	r.Header.Set(key, value)
	return r
}

// SetHeaders 设置多个请求头
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	return r
}

// AddHeader 添加请求头
func (r *Request) AddHeader(key, value string) *Request {
	r.Header.Add(key, value)
	return r
}

// DelHeader 删除请求头
func (r *Request) DelHeader(key string) *Request {
	r.Header.Del(key)
	return r
}

// GetHeader 获取请求头
func (r *Request) GetHeader(key string) string {
	return r.Header.Get(key)
}

// SetCookie 设置 Cookie
func (r *Request) SetCookie(cookies string) *Request {
	r.cookies = append(r.cookies, parseCookies(cookies)...)
	return r
}

// SetCookiesRaw 设置原始 Cookie 切片
func (r *Request) SetCookiesRaw(cookies []*http.Cookie) *Request {
	r.cookies = append(r.cookies, cookies...)
	return r
}

// SetCookieRaw 设置单个原始 Cookie
func (r *Request) SetCookieRaw(cookie *http.Cookie) *Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

// SetFormParams 设置多个表单参数
func (r *Request) SetFormParams(params map[string]string) *Request {
	for key, value := range params {
		r.formParams.Set(key, value)
	}
	return r
}

// SetFormParam 设置单个表单参数
func (r *Request) SetFormParam(key, value string) *Request {
	r.formParams.Set(key, value)
	return r
}

// SetQueryParams 设置多个查询参数
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for key, value := range params {
		r.queryParams[key] = value
	}
	return r
}

// SetQueryParam 设置单个查询参数
func (r *Request) SetQueryParam(key, value string) *Request {
	r.queryParams[key] = value
	return r
}

// DelQueryParam 删除查询参数
func (r *Request) DelQueryParam(key string) *Request {
	delete(r.queryParams, key)
	return r
}

func (r *Request) SetBody(body string) *Request {
	r.body = body
	return r
}

func (r *Request) SetBodyJSON(data any) *Request {
	switch body := data.(type) {
	case string:
		if isJSON(body) {
			r.body = body
		} else {
			r.rawClient.logger().Error("invalid JSON string", "body", body)
		}
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			r.rawClient.logger().Error("failed to marshal JSON", "error", err)
		} else {
			r.body = string(jsonData)
		}
	}
	r.SetHeader("Content-Type", ContentTypeJson)
	return r
}

// isJSON 判断字符串是否为 JSON 格式
func isJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

// SetBodyBytes 设置请求体为字节数组
func (r *Request) SetBodyBytes(body []byte) *Request {
	r.body = string(body)
	return r
}

// prepareRequestBody 准备请求体
func (r *Request) prepareRequestBody() *bytes.Reader {
	if len(r.formParams) > 0 {
		return bytes.NewReader([]byte(r.formParams.Encode()))
	}
	return bytes.NewReader([]byte(r.body))
}

// prepareRequestURL 准备请求 URL
func (r *Request) prepareRequestURL() string {
	urlPath := strings.TrimPrefix(r.urlPoint, "/")
	if len(r.queryParams) > 0 {
		queryParams := url.Values{}
		for key, value := range r.queryParams {
			queryParams.Add(key, value)
		}
		urlPath += "?" + queryParams.Encode()
	}
	return urlPath
}

func (r *Request) newRequest() (*http.Request, error) {
	u, err := urlpkg.Parse(fmt.Sprintf("%s/%s", r.rawClient.BaseURL, r.prepareRequestURL()))
	if err != nil {
		return nil, err
	}
	u.Host = removeEmptyPort(u.Host)

	var reqBody io.ReadCloser
	var contentLength int64
	if r.GetBody != nil {
		reqBody, err = r.GetBody()
		if err != nil {
			return nil, err
		}
	} else {
		prepareBody := r.prepareRequestBody()
		contentLength = int64(prepareBody.Len())
		reqBody = io.NopCloser(prepareBody)
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte(r.body))), nil
		}
	}

	if r.method == "" {
		return nil, fmt.Errorf("HTTP method is not set")
	}
	if r.ctx == nil {
		r.ctx = context.Background()
	}
	req := &http.Request{
		Method:        r.method,
		Header:        r.Header.Clone(),
		URL:           u,
		Host:          u.Host,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: contentLength,
		Body:          reqBody,
		GetBody:       r.GetBody,
	}
	req = req.WithContext(r.ctx)
	for _, cookie := range r.cookies {
		req.AddCookie(cookie)
	}

	if r.rawClient.UserInfo != nil { // takes precedence
		r.Request.SetBasicAuth(r.rawClient.UserInfo.Username, r.rawClient.UserInfo.Password)
	} else {
		if !IsStringEmpty(r.rawClient.BasicAuthToken) {
			r.SetHeader(r.rawClient.HeaderAuthorizationKey, r.rawClient.AuthScheme+" "+r.rawClient.BasicAuthToken)
		}
	}
	return req, nil
}

func (r *Request) SetContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

// Execute 执行请求并返回响应
func (r *Request) Execute(urlPath string) (*Response, error) {
	r.urlPoint = strings.TrimPrefix(urlPath, "/")
	request, err := r.newRequest()
	if err != nil {
		r.rawClient.logger().Error("failed to build HTTP request", "error", err)
		return nil, err
	}
	if r.rawClient.handleRequestResultFunc != nil {
		request = r.rawClient.handleRequestResultFunc(request)
	}
	r.Request = request
	for i := 0; i < r.rawClient.RetryMax; i++ {
		response, ok := r.Do()
		if ok == nil && response.Response != nil {
			return response, nil
		}
	}
	return nil, fmt.Errorf("failed to execute request")
}
