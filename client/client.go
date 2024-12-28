package main

import (
	"fmt"
	"github.com/gookit/ini/v2"
	"log"
	"net"
	"os"
	"time"
)

const (
	configFilePath = "./Client.ini"
	logFilePath    = "./client_error.log" // 错误日志文件路径
)

var (
	ipaddress string
	port      string
	timeout   time.Duration
	logFile   *os.File
)

// 初始化函数
func init() {
	// 设置日志文件
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开日志文件: %v", err)
	}
	log.SetOutput(logFile) // 将日志输出到文件

	// 加载配置文件并缓存
	err = ini.LoadExists(configFilePath)
	if err != nil {
		logAndPrintError(fmt.Sprintf("加载配置文件失败, 请检查配置文件: %v", err))
	}

	// 从配置文件中读取服务器的 IP 地址和端口号
	ipaddress = ini.String("socket.ipaddress")
	port = ini.String("socket.port")
	timeout = time.Duration(ini.Int("socket.timeout", 5)) * time.Second // 默认 5 秒超时

	// 检查配置是否有效
	if ipaddress == "" || port == "" {
		logAndPrintError("配置文件中 IP 地址或端口缺失, 请检测 IP 与端口是否正确")
	}

	// 确保配置文件存在
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		_, err = os.Create(configFilePath)
		if err != nil {
			logAndPrintError(fmt.Sprintf("无法创建配置文件 'Client.ini': %s\n", err))
		}
	}
}

// logAndPrintError 记录错误日志并在终端打印
func logAndPrintError(msg string) {
	log.Println(msg) // 将错误记录到日志文件
	fmt.Println(msg) // 在终端打印错误
}

// connect 建立与服务器的连接
func connect() (net.Conn, error) {
	// 连接服务器
	ipAndPort := fmt.Sprintf("%s:%s", ipaddress, port)
	var conn net.Conn
	var err error

	// 重试 3 次，如果连接失败，则退出
	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", ipAndPort, timeout)
		if err == nil {
			return conn, nil
		}
		logAndPrintError(fmt.Sprintf("连接失败，重试 %d/3: %v", i+1, err))
		time.Sleep(2 * time.Second)
	}

	// 如果重试 3 次仍然失败，返回错误
	return nil, fmt.Errorf("连接服务器失败，请检查服务地址是否配置正确或确认服务器是否开启: %v", err)
}

// sender 发送消息并接收响应
func sender(conn net.Conn, message string) error {
	defer conn.Close() // 函数结束时关闭连接。

	// 发送消息
	_, err := conn.Write([]byte(message))
	if err != nil {
		logAndPrintError(fmt.Sprintf("发送数据失败: %v", err))
		return err
	}

	// 创建缓冲区以接收响应数据
	buf := make([]byte, 1024)
	// 读取响应数据
	n, err := conn.Read(buf)
	if err != nil {
		logAndPrintError(fmt.Sprintf("接收服务器数据失败: %v", err))
		return err
	}

	// 打印接收到的服务器响应数据
	fmt.Printf("服务器返回的数据: %s\n", string(buf[:n]))
	return nil
}

// 使用flag包处理命令行参数
func parseArgs() (string, error) {
	if len(os.Args) < 2 {
		return "", fmt.Errorf("请提供要发送的数据，例如: go run client.go readIPaddress")
	}
	return os.Args[1], nil
}

func main() {
	// 解析命令行参数
	message, err := parseArgs()
	if err != nil {
		logAndPrintError(err.Error())
		return
	}

	// 建立与服务器的连接
	conn, err := connect()
	if err != nil {
		logAndPrintError(err.Error())
		return
	}

	// 发送数据并接收服务器响应
	if err := sender(conn, message); err != nil {
		logAndPrintError(err.Error())
	}
}
