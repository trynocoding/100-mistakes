package main

import (
	"errors"
	"fmt"
	"time"
)

// Server 表示一个服务器配置
type Server struct {
	Host     string
	Port     int
	Timeout  time.Duration
	MaxConn  int
}

// Option 是一个函数类型，用于配置 Server
type Option func(*Server) *Server

// OptionWithValidate 是一个返回 error 的选项函数类型
type OptionWithValidate func(*Server) error

// ============ 基本选项函数 ============

// WithHost 设置服务器主机地址
func WithHost(host string) Option {
	return func(s *Server) *Server {
		s.Host = host
		return s
	}
}

// WithPort 设置服务器端口
func WithPort(port int) Option {
	return func(s *Server) *Server {
		s.Port = port
		return s
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(s *Server) *Server {
		s.Timeout = timeout
		return s
	}
}

// WithMaxConn 设置最大连接数
func WithMaxConn(maxConn int) Option {
	return func(s *Server) *Server {
		s.MaxConn = maxConn
		return s
	}
}

// NewServer 创建一个新的服务器配置
// 使用可变参数 options 来接收可选配置
func NewServer(options ...Option) *Server {
	// 设置默认值
	server := &Server{
		Host:    "localhost",
		Port:    8080,
		Timeout: 30 * time.Second,
		MaxConn: 100,
	}

	// 应用所有选项
	for _, option := range options {
		option(server)
	}

	return server
}

// ============ 带验证的选项函数 ============

// validatePort 验证端口号
func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("端口号必须在 1-65535 之间，当前值: %d", port)
	}
	return nil
}

// validateHost 验证主机地址
func validateHost(host string) error {
	if host == "" {
		return errors.New("主机地址不能为空")
	}
	return nil
}

// WithPortValid 创建一个带验证的端口选项
func WithPortValid(port int) OptionWithValidate {
	return func(s *Server) error {
		if err := validatePort(port); err != nil {
			return err
		}
		s.Port = port
		return nil
	}
}

// WithHostValid 创建一个带验证的主机选项
func WithHostValid(host string) OptionWithValidate {
	return func(s *Server) error {
		if err := validateHost(host); err != nil {
			return err
		}
		s.Host = host
		return nil
	}
}

// NewServerWithValidation 创建一个带验证的服务器配置
func NewServerWithValidation(options ...OptionWithValidate) (*Server, error) {
	server := &Server{
		Host:    "localhost",
		Port:    8080,
		Timeout: 30 * time.Second,
		MaxConn: 100,
	}

	// 应用所有选项并收集错误
	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("配置服务器失败: %w", err)
		}
	}

	return server, nil
}

func main() {
	fmt.Println("=== 函数式选项模式演示 ===\n")

	// 示例 1: 使用默认值
	server1 := NewServer()
	fmt.Printf("示例 1 - 默认配置:\n%+v\n\n", server1)

	// 示例 2: 使用部分选项
	server2 := NewServer(
		WithHost("192.168.1.1"),
		WithPort(9090),
	)
	fmt.Printf("示例 2 - 自定义主机和端口:\n%+v\n\n", server2)

	// 示例 3: 使用所有选项
	server3 := NewServer(
		WithHost("example.com"),
		WithPort(443),
		WithTimeout(60*time.Second),
		WithMaxConn(500),
	)
	fmt.Printf("示例 3 - 完整自定义配置:\n%+v\n\n", server3)

	// 示例 4: 链式调用演示
	server4 := NewServer(
		WithHost("api.example.com"),
		WithPort(8080),
		WithTimeout(45*time.Second),
		WithMaxConn(200),
	)
	fmt.Printf("示例 4 - 链式配置结果:\n")
	fmt.Printf("  主机: %s\n", server4.Host)
	fmt.Printf("  端口: %d\n", server4.Port)
	fmt.Printf("  超时: %v\n", server4.Timeout)
	fmt.Printf("  最大连接数: %d\n\n", server4.MaxConn)

	// 示例 5: 带验证 - 成功情况
	fmt.Println("示例 5 - 带验证的配置:")
	server5, err := NewServerWithValidation(
		WithHostValid("valid-host.com"),
		WithPortValid(3000),
	)
	if err != nil {
		fmt.Printf("验证错误: %v\n", err)
	} else {
		fmt.Printf("验证成功: %+v\n\n", server5)
	}

	// 示例 6: 带验证 - 端口无效
	fmt.Println("示例 6 - 无效端口验证:")
	_, err = NewServerWithValidation(
		WithPortValid(70000), // 无效端口
	)
	if err != nil {
		fmt.Printf("验证错误: %v\n\n", err)
	}

	// 示例 7: 带验证 - 主机为空
	fmt.Println("示例 7 - 空主机验证:")
	_, err = NewServerWithValidation(
		WithHostValid(""), // 空主机
	)
	if err != nil {
		fmt.Printf("验证错误: %v\n\n", err)
	}

	// 示例 8: 组合验证错误
	fmt.Println("示例 8 - 多个验证错误:")
	_, err = NewServerWithValidation(
		WithHostValid(""),
		WithPortValid(99999),
	)
	if err != nil {
		fmt.Printf("组合验证错误: %v\n", err)
	}
}
