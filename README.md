# QuicklyHttps

QuicklyHttps is a versatile and easy-to-use HTTP client library for Go that simplifies the process of making HTTP requests. It offers enhanced features such as logging, retry mechanism, and simplified API for common HTTP methods.

## Introduction

QuicklyHttps addresses the common challenges developers face when working with HTTP requests in Go. The motivation behind creating QuicklyHttps was to provide an all-in-one solution that integrates useful functionalities like logging, retry logic, and easy-to-use interfaces for making HTTP requests. This library is designed to be both robust and flexible, allowing developers to handle HTTP communications effortlessly.

With QuicklyHttps, you can easily configure and send HTTP requests, handle responses, and manage cookies and headers. The library also supports JSON and form data submissions, making it an ideal choice for building RESTful APIs and web clients.

### Key Features

- **Enhanced Logging**: Integrated support for leveled logging to track request and response details.
- **Retry Mechanism**: Configurable retry logic to handle transient network issues.
- **Easy Configuration**: Simple methods to set up headers, query parameters, and form data.
- **JSON and Form Data Handling**: Convenient methods to send and receive JSON and form data.
- **Proxy Support**: Easily configure proxy settings for HTTP requests.
- **Debug Mode**: Enable debug mode to log detailed request and response information.

## Installation and Deployment

To install QuicklyHttps, you need to have Go installed on your machine. You can install QuicklyHttps using the following command:

```sh
go get github.com/catnovel/quicklyHttps
```

Next, import the package in your Go code:

```go
import "github.com/catnovel/quicklyHttps"
```

## Usage

### Creating a Client and Making Requests

Here’s how you can create a client and make a simple GET request:

```go
client := quicklyHttps.NewClient()
response, err := client.Get("https://api.example.com/data", nil, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.String())
```

### Setting Headers and Query Parameters

You can set headers and query parameters for your requests:

```go
client := quicklyHttps.NewClient()
response, err := client.SetHeader("Authorization", "Bearer token").SetQueryParam("key", "value").Get("https://api.example.com/data", nil, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.String())
```

### Posting Form Data

To post form data, you can use the `PostForm` method:

```go
formData := map[string]string{
    "username": "testuser",
    "password": "password123",
}
response, err := client.PostForm("https://api.example.com/login", formData, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.String())
```

### Posting JSON Data

For posting JSON data, use the `PostJSON` method:

```go
jsonData := map[string]interface{}{
    "title": "New Post",
    "body": "This is the content of the new post.",
}
response, err := client.PostJSON("https://api.example.com/posts", jsonData, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.String())
```

## Running Tests

To run tests for QuicklyHttps, use the following command:

```sh
go test ./...
```

This command will run all the test cases and display the results. Ensure you have written comprehensive test cases to cover different scenarios and edge cases.

## Dependencies

QuicklyHttps relies on the following dependencies:

- `golang.org/x/text/encoding/simplifiedchinese` for encoding conversions.
- `github.com/tidwall/gjson` for JSON parsing.

Ensure these dependencies are included in your `go.mod` file.

## Development Status and Contributions

Currently, QuicklyHttps is in active development. We plan to add more features and improve the existing functionalities. Contributions are welcome! If you find any issues or want to contribute, please create a pull request or open an issue on GitHub.
 