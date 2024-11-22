package main

import (
	"fmt"
	"github.com/gookit/ini/v2" // 引入 gookit/ini 包用于处理 INI 配置文件
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	configFilePath = "./Server.ini" // 配置文件路径常量
)

// 允许执行的命令列表，提高可扩展性
var allowedCommands = map[string]bool{
	"getpath":       true, // 示例命令
	"readIPaddress": true,
	"ViewOnline":    true,
}

// 全局配置缓存
var config = make(map[string]string)

// 缓冲池用于复用缓冲区，减少内存分配和释放的开销
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 2048) // 默认缓冲区大小
	},
}

// 最大连接数限制，使用信号量控制并发量
var maxConnections = make(chan struct{}, 100) // 最大允许 100 个并发连接

// 日志管理相关变量
var (
	logFile  *os.File
	logMutex sync.Mutex // 防止多线程日志写入竞争
)

// init 函数在 main 之前自动执行，用于程序的初始化
func init() {
	// 初始化日志
	initLog()

	// 获取当前工作目录路径
	str, err := os.Getwd()
	if err != nil {
		Log("无法获取当前工作目录: ", err)
		os.Exit(1)
	}
	// 构造配置文件 Server.ini 的完整路径
	filePath := fmt.Sprintf("%s/%s", str, configFilePath)

	// 检查配置文件是否存在，如果不存在则创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			Log("无法创建配置文件 'Server.ini': ", err)
			os.Exit(1)
		}
	}

	// 加载配置文件内容
	if err := ini.LoadExists(filePath); err != nil {
		Log("加载配置文件失败: ", err)
		os.Exit(1)
	}
	// 将配置内容存入全局 map
	config["ipaddress"] = ini.String("socket.ipaddress")
	config["port"] = ini.String("socket.port")
}

// handleConnection 处理从客户端接收到的每一个连接请求
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 从缓冲池获取缓冲区
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer) // 使用完毕后归还缓冲区

	for {
		// 读取客户端发送的数据
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), " 服务器接收到的数据处理完成，客户端已退出: ", err)
			return
		}

		// 处理读取到的数据
		content := strings.TrimSpace(string(buffer[:n]))
		args := strings.Fields(content)
		if len(args) == 0 {
			continue
		}

		// 检查命令是否在白名单中
		if allowed, ok := allowedCommands[args[0]]; ok && allowed {
			// 构造 shell 命令并执行
			cmd := exec.Command("./shell", args...)
			out, err := cmd.Output()
			if err != nil {
				Log("执行命令失败: ", err)
				conn.Write([]byte("执行命令出错\n"))
				return
			}
			// 发送命令执行的输出结果给客户端
			conn.Write(out)
		} else {
			// 发送未授权命令的消息给客户端
			msg := fmt.Sprintf("命令 %s 不允许执行\n", args[0])
			conn.Write([]byte(msg))
			Log("未授权的命令尝试: ", args[0])
		}
	}
}

// handleConnectionWithLimit 包装后的连接处理函数，增加最大连接数限制
func handleConnectionWithLimit(conn net.Conn) {
	maxConnections <- struct{}{}        // 占用一个连接槽位
	defer func() { <-maxConnections }() // 释放槽位
	handleConnection(conn)
}

// Log 函数用于记录日志信息，线程安全
func Log(v ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	log.Println(v...)
}

// initLog 初始化日志文件
func initLog() {
	var err error
	logFile, err = os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开日志文件: %s", err)
	}
	log.SetOutput(logFile) // 将日志输出重定向到文件
}

// closeLog 关闭日志文件
func closeLog() {
	if logFile != nil {
		logFile.Close()
	}
}

// CheckError 函数用于检查并处理错误
func CheckError(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
	}
}

// ReadIniFile 函数用于从全局配置中读取指定的配置项
func ReadIniFile(key string) (string, error) {
	value, exists := config[key]
	if !exists || value == "" {
		return "", fmt.Errorf("无法找到配置项: %s", key)
	}
	return value, nil
}

// serverconn 函数用于启动 TCP 服务器并等待客户端连接
func serverconn() {
	// 从全局配置中读取服务器的 IP 地址和端口号
	ipaddress, err := ReadIniFile("ipaddress")
	port, err := ReadIniFile("port")
	if err != nil {
		Log("读取配置失败:", err)
		return
	}
	ipAndPort := ipaddress + ":" + port

	if ipaddress == "" || port == "" {
		Log("IP 地址或端口为空，无法启动服务器。")
		return
	}

	// 监听指定 IP 地址和端口
	netListen, err := net.Listen("tcp", ipAndPort)
	CheckError(err)
	defer netListen.Close()

	Log("服务器启动，正在等待客户端连接于:", ipAndPort)

	for {
		// 接受客户端连接请求
		conn, err := netListen.Accept()
		if err != nil {
			Log("客户端连接失败:", err)
			continue
		}
		Log(conn.RemoteAddr().String(), " 客户端连接成功")
		go handleConnectionWithLimit(conn) // 使用限制版本的连接处理
	}
}

// main 函数是程序的入口点
func main() {
	defer closeLog() // 确保程序退出时关闭日志文件
	serverconn()
}
