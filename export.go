package quicklyHttps

import (
	"net/http"
)

// Get is a shortcut for doing a GET request without making a new client.
func Get(url string, params, headers map[string]string) (*Response, error) {
	return NewClient().Get(url, params, headers)
}

// Get is a convenience helper for doing simple GET requests.
func (c *Client) Get(url string, params, headers map[string]string) (*Response, error) {
	return c.SetMethod(http.MethodGet).R().SetQueryParams(params).SetHeaders(headers).Execute(url)
}

// Head is a shortcut for doing a HEAD request without making a new client.
func Head(url string, params, headers map[string]string) (*Response, error) {
	return NewClient().Head(url, params, headers)

}

// Head is a convenience method for doing simple HEAD requests.
func (c *Client) Head(url string, params, headers map[string]string) (*Response, error) {
	return c.SetMethod(http.MethodHead).R().SetQueryParams(params).SetHeaders(headers).Execute(url)
}

func (c *Client) Post(url string, params, headers map[string]string) (*Response, error) {
	return c.SetMethod(http.MethodPost).R().SetQueryParams(params).SetHeaders(headers).Execute(url)
}

// PostForm is a shortcut to perform a POST with form data without creating a new client.
func PostForm(url string, data, headers map[string]string) (*Response, error) {
	return NewClient().PostForm(url, data, headers)
}

// PostForm is a convenience method for doing simple POST operations using pre-filled url.Values form data.
func (c *Client) PostForm(url string, data, headers map[string]string) (*Response, error) {
	return c.SetMethod(http.MethodPost).R().SetHeader("Content-Type", ContentTypeForm).SetFormParams(data).SetHeaders(headers).Execute(url)
}

// PostJSON is a shortcut to perform a POST with JSON data without creating a new client.
func PostJSON(url string, data any, headers map[string]string) (*Response, error) {
	return NewClient().PostJSON(url, data, headers)
}

// PostJSON is a convenience method for doing simple POST operations using JSON data.
func (c *Client) PostJSON(url string, data any, headers map[string]string) (*Response, error) {
	return c.SetMethod(http.MethodPost).R().SetBodyJSON(data).SetHeaders(headers).Execute(url)
}
