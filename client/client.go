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
)

func init() {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if _, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			log.Fatalf("无法创建配置文件 'Client.ini': %s\n", err)
		}
	}
}

func readClientIniFile(text string) string {
	err := ini.LoadExists(configFilePath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v\n", err)
	}
	value := ini.String(text)
	if value == "" {
		log.Fatalf("无法获取配置项 '%s' 的值\n", text)
	}
	return value
}

func connect() net.Conn {
	ipaddress := readClientIniFile("socket.ipaddress")
	port := readClientIniFile("socket.port")
	ipAndPort := fmt.Sprintf("%s:%s", ipaddress, port)

	if ipaddress == "" || port == "" {
		log.Fatalf("IP 地址与端口为空，未检测到 ini 文件内的端口，无法开启连接...\n")
	}

	_, err := net.ResolveTCPAddr("tcp4", ipAndPort)
	if err != nil {
		log.Fatalf("连接服务器失败，请检查服务地址是否配置正确: %s\n", err)
	}

	conn, err := net.DialTimeout("tcp", ipAndPort, 5*time.Second)
	if err != nil {
		log.Fatalf("连接服务器失败，请确定服务器是否开启: %s\n", err)
	}

	return conn
}

func sender(conn net.Conn) {
	defer conn.Close()

	if len(os.Args) < 2 {
		log.Fatal("请提供要发送的数据\n")
	}

	message := os.Args[1]
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatalf("发送数据失败: %s\n", err)
	}

	buf := make([]byte, 10240)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("接收服务器数据失败: %s\n", err)
	}

	fmt.Printf("Server 服务器返回已处理的数据: %s\n", string(buf[:n]))
}

func main() {
	conn := connect()
	sender(conn)
}
