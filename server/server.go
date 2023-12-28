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

// 将允许执行的命令放入map中，更具可扩展性
var allowedCommands = map[string]bool{
	"getpath":       true, // 示例命令
	"readIPaddress": true,
	"ViewOnline":    true,
	// 添加其他允许执行的命令
}

// init 函数在 main 之前自动执行，用于程序的初始化
func init() {
	// 获取当前工作目录路径
	str, _ := os.Getwd()
	// 构造配置文件 Server.ini 的完整路径
	filePath := fmt.Sprintf("%s/Server.ini", str)

	// 检查配置文件是否存在，如果不存在则创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			log.Fatalf("无法创建配置文件 'Server.ini': %s\n", err)
		}
	}
}

// handleConnection 处理从客户端接收到的每一个连接请求
func handleConnection(conn net.Conn) {
	defer conn.Close() // 确保连接在处理完成后关闭

	buffer := make([]byte, 2048)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), "服务器接收到的数据处理完成，客户端已退出: ", err)
			return
		}

		content := strings.TrimSpace(string(buffer[:n]))
		args := strings.Fields(content)
		if len(args) == 0 {
			continue
		}

		// 检查命令是否在白名单中
		if allowed, ok := allowedCommands[args[0]]; ok && allowed {
			// 构造 shell 命令
			strScript := "./shell " + content
			// 执行 shell 命令
			out, err := exec.Command("/bin/bash", "-c", strScript).Output()
			if err != nil {
				Log("执行命令失败: ", err)
				conn.Write([]byte("执行命令出错\n"))
				return
			}
			// 发送命令执行的输出结果给客户端
			conn.Write(out)
		} else {
			msg := fmt.Sprintf("命令 %s 不允许执行\n", args[0])
			conn.Write([]byte(msg))
			Log("未授权的命令尝试: ", args[0])
		}
	}
}

// Log 函数用于记录日志信息
func Log(v ...interface{}) {
	log.Println(v...)
}

// CheckError 函数用于检查并处理错误
func CheckError(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
	}
}

// ReadServeriniFile 函数用于从 Server.ini 配置文件中读取指定的配置项
func ReadServeriniFile(key string) string {
	// 加载配置文件
	err := ini.LoadExists("./Server.ini")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	// 读取指定配置项的值
	value := ini.String(key)
	return value
}

// serverconn 函数用于启动 TCP 服务器并等待客户端连接
func serverconn() {
	ipaddress := ReadServeriniFile("socket.ipaddress")
	port := ReadServeriniFile("socket.port")
	ipAndPort := ipaddress + ":" + port

	if ipaddress == "" || port == "" {
		log.Fatalln("IP 地址或端口为空，无法启动服务器。")
	}

	netListen, err := net.Listen("tcp", ipAndPort)
	CheckError(err)
	defer netListen.Close()

	Log("服务器启动，正在等待客户端连接于: ", ipAndPort)

	for {
		conn, err := netListen.Accept()
		if err != nil {
			Log("客户端连接失败: ", err)
			continue
		}
		Log(conn.RemoteAddr().String(), " 客户端连接成功")
		go handleConnection(conn)
	}
}

// main 函数是程序的入口点
func main() {
	serverconn()
}
