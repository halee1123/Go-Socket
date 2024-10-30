//package main
//
//import (
//	"fmt"
//	"github.com/gookit/ini/v2" // 引入 gookit/ini 包用于处理 INI 配置文件
//	"log"
//	"net"
//	"os"
//	"os/exec"
//	"strings"
//)
//
//const (
//	configFilePath = "./Server.ini" // 配置文件路径常量
//)
//
//// 允许执行的命令列表，提高可扩展性
//var allowedCommands = map[string]bool{
//	"getpath":       true, // 示例命令
//	"readIPaddress": true,
//	"ViewOnline":    true,
//}
//
//// init 函数在 main 之前自动执行，用于程序的初始化
//func init() {
//	// 获取当前工作目录路径
//	str, _ := os.Getwd()
//	// 构造配置文件 Server.ini 的完整路径
//	filePath := fmt.Sprintf("%s/%s", str, configFilePath)
//
//	// 检查配置文件是否存在，如果不存在则创建
//	if _, err := os.Stat(filePath); os.IsNotExist(err) {
//		if _, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
//			log.Fatalf("无法创建配置文件 'Server.ini': %s\n", err)
//		}
//	}
//}
//
//// handleConnection 处理从客户端接收到的每一个连接请求
//func handleConnection(conn net.Conn) {
//	defer conn.Close()
//
//	buffer := make([]byte, 2048)
//	for {
//		// 读取客户端发送的数据
//		n, err := conn.Read(buffer)
//		if err != nil {
//			Log(conn.RemoteAddr().String(), "服务器接收到的数据处理完成，客户端已退出: ", err)
//			return
//		}
//
//		// 处理读取到的数据
//		content := strings.TrimSpace(string(buffer[:n]))
//		args := strings.Fields(content)
//		if len(args) == 0 {
//			continue
//		}
//
//		// 检查命令是否在白名单中
//		if allowed, ok := allowedCommands[args[0]]; ok && allowed {
//			// 构造 shell 命令
//			strScript := "./shell " + content
//			// 执行 shell 命令
//			out, err := exec.Command("/bin/bash", "-c", strScript).Output()
//			if err != nil {
//				Log("执行命令失败: ", err)
//				conn.Write([]byte("执行命令出错\n"))
//				return
//			}
//			// 发送命令执行的输出结果给客户端
//			conn.Write(out)
//		} else {
//			// 发送未授权命令的消息给客户端
//			msg := fmt.Sprintf("命令 %s 不允许执行\n", args[0])
//			conn.Write([]byte(msg))
//			Log("未授权的命令尝试: ", args[0])
//		}
//	}
//}
//
//// Log 函数用于记录日志信息
//func Log(v ...interface{}) {
//	log.Println(v...)
//}
//
//// CheckError 函数用于检查并处理错误
//func CheckError(err error) {
//	if err != nil {
//		log.Fatalf("Fatal error: %s", err.Error())
//	}
//}
//
//// ReadIniFile 函数用于从指定的 INI 配置文件中读取指定的配置项
//func ReadIniFile(filePath, key string) (string, error) {
//	// 加载配置文件
//	err := ini.LoadExists(filePath)
//	if err != nil {
//		return "", fmt.Errorf("加载配置文件失败: %v", err)
//	}
//	// 读取指定配置项的值
//	value := ini.String(key)
//	return value, nil
//}
//
//// serverconn 函数用于启动 TCP 服务器并等待客户端连接
//func serverconn() {
//	// 从配置文件中读取服务器的 IP 地址和端口号
//	ipaddress, err := ReadIniFile(configFilePath, "socket.ipaddress")
//	port, err := ReadIniFile(configFilePath, "socket.port")
//	if err != nil {
//		Log("读取配置失败:", err)
//		return
//	}
//	ipAndPort := ipaddress + ":" + port
//
//	if ipaddress == "" || port == "" {
//		Log("IP 地址或端口为空，无法启动服务器。")
//		return
//	}
//
//	// 监听指定 IP 地址和端口
//	netListen, err := net.Listen("tcp", ipAndPort)
//	CheckError(err)
//	defer netListen.Close()
//
//	Log("服务器启动，正在等待客户端连接于:", ipAndPort)
//
//	for {
//		// 接受客户端连接请求
//		conn, err := netListen.Accept()
//		if err != nil {
//			Log("客户端连接失败:", err)
//			continue
//		}
//		Log(conn.RemoteAddr().String(), " 客户端连接成功")
//		go handleConnection(conn)
//	}
//}
//
//// main 函数是程序的入口点
//func main() {
//	serverconn()
//}

package main

import (
	"fmt"
	"github.com/gookit/ini/v2" // 引入 gookit/ini 包用于处理 INI 配置文件
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
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

// init 函数在 main 之前自动执行，用于程序的初始化
func init() {
	// 获取当前工作目录路径
	str, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取当前工作目录: %s\n", err)
	}
	// 构造配置文件 Server.ini 的完整路径
	filePath := fmt.Sprintf("%s/%s", str, configFilePath)

	// 检查配置文件是否存在，如果不存在则创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			log.Fatalf("无法创建配置文件 'Server.ini': %s\n", err)
		}
	}

	// 加载配置文件内容
	if err := ini.LoadExists(filePath); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	// 将配置内容存入全局 map
	config["ipaddress"] = ini.String("socket.ipaddress")
	config["port"] = ini.String("socket.port")
}

// handleConnection 处理从客户端接收到的每一个连接请求
func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 2048)
	for {
		// 读取客户端发送的数据
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), "服务器接收到的数据处理完成，客户端已退出: ", err)
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

// Log 函数用于记录日志信息
func Log(v ...interface{}) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println(v...)
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
		go handleConnection(conn)
	}
}

// main 函数是程序的入口点
func main() {
	serverconn()
}
