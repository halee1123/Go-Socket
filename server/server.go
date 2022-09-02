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

// init函数
func init() {
    // 获取当前路径
    str, _ := os.Getwd()
    // 在当前路径下创建cLIent.ini文件
    var filePath = str + "/Server.ini"

    // ini文件路径
    _, err := os.Stat(filePath)

    if err == nil {
        return
        //fmt.Printf(" 当前路径:%s/%s 文件存在\n", str, err)
    }
    if os.IsNotExist(err) {


        _, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, os.ModePerm)
        if err != nil {
            return
        }
    }
}

//处理client传过来的数据
func handleConnection(conn net.Conn) {  
  
    buffer := make([]byte, 2048)
    for {
        n, err := conn.Read(buffer)
        if err != nil {  
            Log(conn.RemoteAddr().String()," 数据已处理,退出: ", err)
            return  
        }

        // 获取Client 传来的数据参数
        content := strings.TrimSpace(string(buffer[:n]))

        // shell脚本拼接, 脚本以命令形式执行 ..
        strScripy := "./shell " + content

        // 执行shell脚本命令
        out, err := exec.Command("/bin/sh", "-c", strScripy).Output()
        if err != nil {
            // log.Fatal(err)
            return
        }
        // 打印脚本执行后的内容...
        fmt.Printf("Client传过来的消息: %s", out)

        //Log(conn.RemoteAddr().String(), "Client传过来的消息:", string(out))
    }
}

// Log 日志
func Log(v ...interface{}) {  
    log.Println(v...)  
}

// CheckError 连接判断
func CheckError(err error) {  
    if err != nil {
        _, err := fmt.Fprintf(os.Stderr, "无法连接: %s", err.Error())
        if err != nil {
            return 
        }
        os.Exit(1)  
    }  
}

// ReadServeriniFile // 读取ini文件
func ReadServeriniFile(Text string) string {
    // 获取当前路径
    //str, _ := os.Getwd()
    //var filePath = str + "./Server.ini"

    err := ini.LoadExists("./Server.ini")
    if err != nil {
        panic(err)
    }
    value := ini.String(Text)
    //fmt.Println(value)
    return value
}

// server 开启程序
func serverconn(){

    // 获取ini文件数据参数
    ipaddress := ReadServeriniFile("socket.ipaddress")
    port := ReadServeriniFile("socket.port")
    ipAndPort := ipaddress + ":" + port

    // 判断ip和端口是否为空
    if ipaddress == "" && port == "" {
        fmt.Printf("ip地址与端口为空,ini文件未写入,无法开启...\n")
    } else {
        //建立socket，监听端口
        netListen, err := net.Listen("tcp", ipAndPort)
        CheckError(err)
        defer func(netListen net.Listener) {
            err := netListen.Close()
            if err != nil {

            }
        }(netListen)

        Log(ipAndPort, "等待客户连接...")
        for {
            conn, err := netListen.Accept()
            if err != nil {
                continue
            }
            Log(conn.RemoteAddr().String(), "客户端连接成功...")
            go handleConnection(conn)
        }
    }
}

// 执行程序
func main() {
    // 执行Server程序
    serverconn()
}