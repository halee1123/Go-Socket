package main

import (
	"fmt"
	"github.com/gookit/ini/v2" // 引入 gookit/ini 包用于处理 INI 配置文件
	"log"
	"net"
	"os"
	"time"
)

const (
	configFilePath = "./Client.ini"
)

var ipaddress string
var port string
var timeout time.Duration

func init() {
	// 加载配置文件并缓存
	err := ini.LoadExists(configFilePath)
	if err != nil {
		log.Fatalf("加载配置文件失败,请检查配置文件: %v", err)
	}

	// 从配置文件中读取服务器的 IP 地址和端口号
	ipaddress = ini.String("socket.ipaddress")
	port = ini.String("socket.port")
	timeout = time.Duration(ini.Int("socket.timeout", 5)) * time.Second // 默认 5 秒超时

	// 检查配置是否有效
	if ipaddress == "" || port == "" {
		log.Fatalf("配置文件中 IP 地址或端口缺失,请检测IP与端口是否正确")
	}

	// 确保配置文件存在
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		_, err = os.Create(configFilePath)
		if err != nil {
			log.Fatalf("无法创建配置文件 'Client.ini': %s\n", err)
		}
	}
}

func connect() (net.Conn, error) {
	// 连接服务器
	ipAndPort := fmt.Sprintf("%s:%s", ipaddress, port)
	conn, err := net.DialTimeout("tcp", ipAndPort, timeout)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败，请检查服务地址是否配置正确或确认服务器是否开启: %v", err)
	}
	return conn, nil
}

func sender(conn net.Conn, message string) error {
	defer conn.Close() // 函数结束时关闭连接。

	// 发送消息
	_, err := conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("发送数据失败: %v", err)
	}

	// 创建缓冲区以接收响应数据
	buf := make([]byte, 10240)
	// 读取响应数据
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("接收服务器数据失败: %v", err)
	}

	// 打印接收到的服务器响应数据
	fmt.Printf("服务器返回的数据: %s\n", string(buf[:n]))
	return nil
}

func main() {
	// 检查是否提供了要发送的数据
	if len(os.Args) < 2 {
		log.Fatal("请提供要发送的数据，例如: go run client.go readIPaddress")
	}
	message := os.Args[1]

	// 建立与服务器的连接
	conn, err := connect()
	if err != nil {
		log.Fatal(err)
	}

	// 发送数据并接收服务器响应
	if err := sender(conn, message); err != nil {
		log.Fatal(err)
	}
}
