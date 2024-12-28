package main

import (
	"fmt"
	"github.com/gookit/ini/v2" // 引入 gookit/ini 包用于处理 INI 配置文件
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	configFilePath = "./Server.ini" // 配置文件路径常量
)

// 全局配置缓存
var config = make(map[string]string)

// 缓冲池，用于复用缓冲区，减少内存分配和释放的开销
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

// 令牌桶结构
type TokenBucket struct {
	capacity     int        // 令牌桶的容量
	tokens       int        // 当前令牌数
	tokensPerSec int        // 每秒生成的令牌数
	lastRefill   time.Time  // 上次填充令牌的时间
	mutex        sync.Mutex // 锁，防止并发访问令牌桶
}

var tokenBucket TokenBucket

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
			Log("无法创建配置文件 'Server.ini', 请检查权限: ", err)
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

	// 读取令牌桶配置
	tokenBucket.capacity = ini.Int("socket.token_capacity")     // 令牌桶容量
	tokenBucket.tokensPerSec = ini.Int("socket.tokens_per_sec") // 每秒生成的令牌数
	tokenBucket.tokens = tokenBucket.capacity                   // 初始化令牌数为最大容量
	tokenBucket.lastRefill = time.Now()                         // 设置令牌的初始填充时间
}

// Log 函数用于记录日志信息，线程安全
func Log(v ...interface{}) {
	logMutex.Lock() // 获取锁，保证日志的线程安全
	defer logMutex.Unlock()
	log.Println(v...) // 将日志信息写入文件
	fmt.Println(v...) // 在终端也输出日志
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
		logFile.Close() // 确保程序退出时关闭日志文件
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

// TokenBucket 操作：每秒刷新一次令牌，处理并检查是否可以获取令牌
func (tb *TokenBucket) Take() bool {
	tb.mutex.Lock() // 获取锁，防止并发修改令牌桶
	defer tb.mutex.Unlock()

	// 每秒刷新一次令牌
	now := time.Now()
	duration := now.Sub(tb.lastRefill)
	if duration > time.Second {
		// 每秒生成的令牌
		refillTokens := int(duration.Seconds()) * tb.tokensPerSec
		tb.tokens += refillTokens
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity // 如果令牌数超过容量，限制为最大容量
		}
		tb.lastRefill = now

		// 打印令牌桶的当前状态到终端
		fmt.Printf("令牌桶已刷新，当前令牌数: %d/%d\n", tb.tokens, tb.capacity)
		Log("令牌桶已刷新，当前令牌数: ", tb.tokens, " 最大容量: ", tb.capacity)
	}

	// 如果令牌桶中有令牌，则返回 true
	if tb.tokens > 0 {
		tb.tokens-- // 获取一个令牌
		Log("取出一个令牌，当前令牌数: ", tb.tokens)
		return true
	}
	return false // 如果没有令牌，返回 false
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 从缓冲池获取缓冲区
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer) // 使用完毕后归还缓冲区

	for {
		// 等待令牌并处理客户端请求
		if !tokenBucket.Take() {
			Log("令牌桶为空，等待令牌...")
			timeout := time.After(10 * time.Second) // 最大等待时间 10 秒
			select {
			case <-timeout:
				Log("等待令牌超时，放弃当前请求")
				return
			case <-time.After(1 * time.Second): // 每秒检查一次令牌
				// 等待令牌生成
				continue
			}
		}

		// 读取客户端发送的数据
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// 客户端正常关闭连接
				Log(conn.RemoteAddr().String(), " 客户端连接关闭\n")
			} else {
				// 其他错误
				Log(conn.RemoteAddr().String(), " 客户端连接失败: \n", err)
			}
			return
		}

		// 处理客户端数据
		content := strings.TrimSpace(string(buffer[:n]))
		args := strings.Fields(content)
		if len(args) == 0 {
			continue
		}

		// 执行命令
		Log("接收到命令: ", args[0])
		cmd := exec.Command("./shell", args...)
		out, err := cmd.Output()
		if err != nil {
			Log("执行命令失败: ", err)
			conn.Write([]byte("执行命令出错\n"))
			return
		}
		// 发送命令执行的输出结果给客户端
		conn.Write(out)
		Log(args[0], "命令处理完成")
	}
}

// handleConnectionWithLimit 包装后的连接处理函数，增加最大连接数限制
func handleConnectionWithLimit(conn net.Conn) {
	maxConnections <- struct{}{}        // 占用一个连接槽位
	defer func() { <-maxConnections }() // 释放槽位
	handleConnection(conn)
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

	// 检查配置的 IP 地址和端口
	if ipaddress == "" || port == "" {
		Log("IP 地址或端口为空，无法启动服务器。")
		return
	}

	// 监听指定 IP 地址和端口
	netListen, err := net.Listen("tcp", ipAndPort)
	if err != nil {
		Log("无法监听端口: ", err)
		return
	}
	defer netListen.Close()

	Log("服务器程序已运行  IP: ", ipaddress, "端口: ", port, "等待连接")

	for {
		// 接受客户端连接请求
		conn, err := netListen.Accept()
		if err != nil {
			Log("客户端连接失败:", err)
			continue
		}
		Log(conn.RemoteAddr().String(), " 客户端连接成功")
		// 启动一个新的协程来处理客户端连接，确保服务器不被阻塞
		go handleConnectionWithLimit(conn)
	}
}

// main 主函数
func main() {
	defer closeLog() // 确保程序退出时关闭日志文件

	// 启动服务器监听连接
	serverconn()
}
