package main

import (
	"fmt"
	"github.com/gookit/ini/v2"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
)

// 全局变量定义允许的命令白名单
var allowedCommands = map[string]bool{
	"getpath": true, // 示例命令
	"ls":      true,
	"ls -all": true,
	// 在这里添加其他允许执行的命令
}

// init函数在main之前自动执行，用于程序的初始化
func init() {
	// 获取当前工作目录路径
	str, _ := os.Getwd()
	// 拼接出配置文件Server.ini的完整路径
	var filePath = str + "/Server.ini"

	// 检查配置文件是否存在
	_, err := os.Stat(filePath)
	// 如果不存在，创建一个新的Server.ini文件
	if os.IsNotExist(err) {
		_, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			// 如果创建文件失败，则记录错误并退出
			log.Fatalf("无法创建配置文件: %v", err)
		}
	}
}

// handleConnection处理从客户端接收到的每一个连接请求
func handleConnection(conn net.Conn) {
	// 创建一个缓冲区用来存放从客户端接收到的数据
	buffer := make([]byte, 2048)
	for {
		// 读取客户端发送的数据
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), "服务器接收数据处理完成, 客户端已退出: ", err)
			return
		}

		// 处理接收到的数据，删除可能的空格
		content := strings.TrimSpace(string(buffer[:n]))
		// 拆分命令和参数
		args := strings.Fields(content)
		// 如果没有命令被发送，跳过本次循环
		if len(args) == 0 {
			continue
		}

		// 检查命令是否在白名单中
		if allowed, ok := allowedCommands[args[0]]; ok && allowed {
			// 如果命令在白名单中，构造shell命令
			strScript := "./shell " + content
			// 执行构造的shell命令
			out, err := exec.Command("/bin/bash", "-c", strScript).Output()
			if err != nil {
				Log("执行命令失败: ", err)
				conn.Write([]byte("执行命令出错\n"))
				return
			}
			// 发送命令执行的输出结果给客户端
			conn.Write(out)
		} else {
			// 如果命令不在白名单中，告知客户端命令执行不被允许
			msg := fmt.Sprintf("命令 %s 不允许执行\n", args[0])
			conn.Write([]byte(msg))
			Log("未授权的命令尝试: ", args[0])
		}
	}
}

// Log函数用于记录日志信息
func Log(v ...interface{}) {
	log.Println(v...)
}

// CheckError函数用于检查并处理错误
func CheckError(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
	}
}

// ReadServeriniFile函数用于从Server.ini配置文件中读取指定的配置项
func ReadServeriniFile(key string) string {
	// 加载配置文件
	err := ini.LoadExists("./Server.ini")
	if err != nil {
		// 如果加载失败，则记录错误并退出
		log.Fatalf("加载配置文件失败: %v", err)
	}
	// 读取指定配置项的值
	value := ini.String(key)
	return value
}

// serverconn函数用于启动TCP服务器并等待客户端连接
func serverconn() {
	// 从配置文件读取IP地址和端口号
	ipaddress := ReadServeriniFile("socket.ipaddress")
	port := ReadServeriniFile("socket.port")
	ipAndPort := ipaddress + ":" + port

	// 如果IP地址或端口为空，说明配置文件可能未正确设置，服务器无法启动
	if ipaddress == "" || port == "" {
		log.Fatalln("IP地址或端口为空，无法启动服务器。")
	}

	// 创建TCP监听器，开始监听配置文件中指定的IP地址和端口
	netListen, err := net.Listen("tcp", ipAndPort)
	CheckError(err)         // 检查是否有错误发生，如果有，则记录并退出
	defer netListen.Close() // 在函数返回之前，确保监听器被关闭

	Log("服务器启动，正在等待客户端连接于: ", ipAndPort)
	// 无限循环，等待客户端的连接
	for {
		conn, err := netListen.Accept() // 接受新的客户端连接
		if err != nil {
			Log("客户端连接失败: ", err)
			continue // 如果有错误发生，记录错误然后等待下一个连接
		}
		Log(conn.RemoteAddr().String(), " 客户端连接成功")
		go handleConnection(conn) // 使用goroutine来处理连接，以便同时处理多个连接
	} // }

}

// main函数是程序的入口点
func main() {
	serverconn() // 调用serverconn函数启动服务器
}
