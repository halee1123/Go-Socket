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

func init() {
	// 检查配置文件是否存在，如果不存在则创建
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		_, err = os.Create(configFilePath)
		if err != nil {
			log.Fatalf("无法创建配置文件 'Client.ini': %s\n", err)
		}
	}
}

func readClientIniFile(key string) (string, error) {
	// 加载并检查 Client.ini 文件是否存在。
	err := ini.LoadExists(configFilePath)
	if err != nil {
		return "", fmt.Errorf("加载配置文件失败: %v", err)
	}
	// 获取指定配置项的值。
	value := ini.String(key)
	if value == "" {
		return "", fmt.Errorf("配置项 %s 为空", key)
	}
	return value, nil
}

func connect() (net.Conn, error) {
	// 从配置文件中读取服务器的 IP 地址和端口号。
	ipaddress, err := readClientIniFile("socket.ipaddress")
	if err != nil {
		return nil, err
	}
	port, err := readClientIniFile("socket.port")
	if err != nil {
		return nil, err
	}
	ipAndPort := fmt.Sprintf("%s:%s", ipaddress, port)

	// 解析 TCP 地址并创建连接。
	conn, err := net.DialTimeout("tcp", ipAndPort, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败，请检查服务地址是否配置正确或确认服务器是否开启: %v", err)
	}
	return conn, nil
}

func sender(conn net.Conn, message string) error {
	defer conn.Close() // 函数结束时关闭连接。

	// 发送消息。
	_, err := conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("发送数据失败: %v", err)
	}

	// 创建缓冲区以接收响应数据。
	buf := make([]byte, 10240)
	// 读取响应数据。
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("接收服务器数据失败: %v", err)
	}

	// 打印接收到的服务器响应数据。
	fmt.Printf("服务器返回的数据: %s\n", string(buf[:n]))
	return nil
}

func main() {
	// 检查是否提供了要发送的数据。
	if len(os.Args) < 2 {
		log.Fatal("请提供要发送的数据，例如: go run client.go readIPaddress")
	}
	message := os.Args[1]

	conn, err := connect() // 建立与服务器的连接。
	if err != nil {
		log.Fatal(err)
	}

	// 发送数据并接收服务器响应。
	if err := sender(conn, message); err != nil {
		log.Fatal(err)
	}
}
