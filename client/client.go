package main

import (
	"fmt"
	"github.com/gookit/ini/v2" // 引入 gookit/ini 包用于处理 INI 配置文件
	"net"
	"os"
	"time"
)

const (
	configFilePath = "./Client.ini"
)

// init 函数在 main 函数之前自动调用，用于初始化设置。
func init() {
	// 检查配置文件是否存在，如果不存在则创建
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if _, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			fmt.Printf("无法创建配置文件 'Client.ini': %s\n", err)
			os.Exit(1)
		}
	}
}

// readClientIniFile 函数用于读取 INI 配置文件中的值。
func readClientIniFile(text string) string {
	// 加载并检查 Client.ini 文件是否存在。
	err := ini.LoadExists(configFilePath)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}
	// 获取指定配置项的值并返回。
	value := ini.String(text)
	return value
}

// connect 函数用于与服务器建立连接，并返回连接对象。
func connect() net.Conn {
	// 从配置文件中读取服务器的 IP 地址和端口号。
	ipaddress := readClientIniFile("socket.ipaddress")
	port := readClientIniFile("socket.port")
	ipAndPort := fmt.Sprintf("%s:%s", ipaddress, port)

	// 如果 IP 地址或端口为空，则输出错误信息并退出。
	if ipaddress == "" || port == "" {
		fmt.Printf("IP 地址与端口为空，未检测到 ini 文件内的端口，无法开启连接...\n")
		os.Exit(1)
	}
	// 解析 TCP 地址。
	_, err := net.ResolveTCPAddr("tcp4", ipAndPort)
	if err != nil {
		// 如果解析 TCP 地址失败，输出错误信息并退出。
		fmt.Printf("连接服务器失败，请检查服务地址是否配置正确: %s\n", err)
		os.Exit(1)
	}

	// 创建 TCP 连接。
	conn, err := net.DialTimeout("tcp", ipAndPort, 5*time.Second)
	if err != nil {
		// 如果连接失败，输出错误信息并退出。
		fmt.Printf("连接服务器失败，请确定服务器是否开启: %s\n", err)
		os.Exit(1)
	}
	// 返回建立的连接。
	return conn
}

// sender 函数负责通过连接发送数据，并接收服务器的响应。
func sender(conn net.Conn) {
	defer conn.Close() // 函数结束时关闭连接。

	// 检查是否提供了要发送的数据。
	if len(os.Args) < 2 {
		fmt.Println("请提供要发送的数据")
		os.Exit(1)
	}

	// 获取要发送的消息。
	message := os.Args[1]
	// 发送消息。
	_, err := conn.Write([]byte(message))
	if err != nil {
		// 如果发送失败，输出错误信息并退出。
		fmt.Printf("发送数据失败: %s\n", err)
		os.Exit(1)
	}

	// 创建缓冲区以接收响应数据。
	buf := make([]byte, 10240)
	// 读取响应数据。
	n, err := conn.Read(buf)
	if err != nil {
		// 如果读取失败，输出错误信息并退出。
		fmt.Printf("接收服务器数据失败: %s\n", err)
		os.Exit(1)
	}

	// 打印接收到的服务器响应数据。
	fmt.Printf("Server 服务器返回已处理的数据: %s\n", string(buf[:n]))
}

// main 函数是程序的入口点。
func main() {
	conn := connect() // 建立与服务器的连接。
	sender(conn)      // 发送数据并接收服务器响应。
}

// 更新代码：重构代码结构，提取配置文件路径为常量；增加连接超时设置。
