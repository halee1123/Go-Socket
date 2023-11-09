package main

import (
	"fmt"
	"github.com/gookit/ini/v2" // 引入gookit/ini包用于处理INI配置文件
	"net"
	"os"
)

// init函数在main函数之前自动调用，用于初始化设置。
func init() {
	// 获取当前程序运行的目录路径。
	str, _ := os.Getwd()

	// 构造Client.ini文件的完整路径。
	var filePath = str + "/Client.ini"

	// 检查Client.ini文件是否存在。
	_, err := os.Stat(filePath)

	// 如果文件已存在，则不执行任何操作。
	if err == nil {
		return
	}
	// 如果文件不存在，则尝试创建该文件。
	if os.IsNotExist(err) {
		// 尝试创建配置文件，设置文件权限为0666。
		_, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			// 如果创建文件失败，输出错误信息并退出程序。
			fmt.Printf("无法创建配置文件 'Client.ini': %s\n", err)
			os.Exit(1)
		}
	}
}

// ReadClieniniFile函数用于读取INI配置文件中的值。
func ReadClieniniFile(Text string) string {
	// 加载并检查Client.ini文件是否存在。
	err := ini.LoadExists("./Client.ini")
	if err != nil {
		// 如果加载失败，抛出panic异常。
		panic(err)
	}
	// 获取指定配置项的值并返回。
	value := ini.String(Text)
	return value
}

// connect函数用于与服务器建立连接，并返回连接对象。
func connect() net.Conn {
	// 从配置文件中读取服务器的IP地址和端口号。
	ipaddress := ReadClieniniFile("socket.ipaddress")
	port := ReadClieniniFile("socket.port")
	ipAndPort := ipaddress + ":" + port

	// 如果IP地址或端口为空，则输出错误信息并退出。
	if ipaddress == "" && port == "" {
		fmt.Printf("IP地址与端口为空，未检测到ini文件内的端口，无法开启连接...\n")
		os.Exit(1)
	}
	// 解析TCP地址。
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ipAndPort)
	if err != nil {
		// 如果解析TCP地址失败，输出错误信息并退出。
		fmt.Printf("连接服务器失败，请检查服务地址是否配置正确: %s\n", err)
		os.Exit(1)
	}

	// 创建TCP连接。
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		// 如果连接失败，输出错误信息并退出。
		fmt.Printf("连接服务器失败，请确定服务器是否开启: %s\n", err)
		os.Exit(1)
	}
	// 返回建立的连接。
	return conn
}

// sender函数负责通过连接发送数据，并接收服务器的响应。
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
	fmt.Printf("Server服务器返回已处理的数据: %s\n", string(buf[:n]))
}

// main函数是程序的入口点。
func main() {
	conn := connect() // 建立与服务器的连接。
	sender(conn)      // 发送数据并接收服务器响应。
}
