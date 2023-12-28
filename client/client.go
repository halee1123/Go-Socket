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
		log.Fatalf("无法创建配置文件 'Client.ini': %s\n", err)
	}
}

func readClientIniFile(text string) string {
	// 加载并检查 Client.ini 文件是否存在。
	err := ini.LoadExists(configFilePath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v\n", err)
	}
	// 获取指定配置项的值并返回。
	value := ini.String(text)
	if value == "" {
		log.Fatalf("配置项 %s 为空\n", text)
	}
	return value
}

func connect() net.Conn {
	// 从配置文件中读取服务器的 IP 地址和端口号。
	ipaddress := readClientIniFile("socket.ipaddress")
	port := readClientIniFile("socket.port")
	ipAndPort := fmt.Sprintf("%s:%s", ipaddress, port)

	// 解析 TCP 地址。
	_, err := net.ResolveTCPAddr("tcp4", ipAndPort)
	if err != nil {
		log.Fatalf("连接服务器失败，请检查服务地址是否配置正确: %s\n", err)
	}

	// 创建 TCP 连接。
	conn, err := net.DialTimeout("tcp", ipAndPort, 5*time.Second)
	if err != nil {
		log.Fatalf("连接服务器失败，请确定服务器是否开启: %s\n", err)
	}
	return conn
}

func sender(conn net.Conn) {
	defer conn.Close() // 函数结束时关闭连接。

	// 检查是否提供了要发送的数据。
	if len(os.Args) < 2 {
		log.Fatal("请提供要发送的数据")
	}

	// 获取要发送的消息。
	message := os.Args[1]
	// 发送消息。
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatalf("发送数据失败: %s\n", err)
	}

	// 创建缓冲区以接收响应数据。
	buf := make([]byte, 10240)
	// 读取响应数据。
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("接收服务器数据失败: %s\n", err)
	}

	// 打印接收到的服务器响应数据。
	fmt.Printf("Server 服务器返回已处理的数据: %s\n", string(buf[:n]))
}

func main() {
	conn := connect() // 建立与服务器的连接。
	sender(conn)      // 发送数据并接收服务器响应。
}
