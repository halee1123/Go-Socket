package main

import (
    "fmt"
    "github.com/gookit/ini/v2"
    "net"
    "os"
)

// init函数,自动调用
func init() {

    // 获取当前路径
    str, _ := os.Getwd()

    // 在当前路径下创建cLIent.ini文件
    var filePath = str + "/Client.ini"

    // 当前ini文件路径
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

// ReadClieniniFile // 读取ini文件
func ReadClieniniFile(Text string) string {

   err := ini.LoadExists("./Client.ini")
   if err != nil {
       panic(err)
   }
   value := ini.String(Text)
   //fmt.Println(value)
   return value
}

// 向server发送数据
func sender(conn net.Conn) {

    var strText []string

    for _, v := range os.Args {
        //fmt.Println( v)
        strText = append(strText, fmt.Sprintf("%v", v))
    }


    // 发送数据
    _, err := conn.Write([]byte(fmt.Sprintf(strText[1])))
    if err != nil {
        return
    }

    // 接收服务端返回的数据
    buf := [10240]byte{}
    serverMsg, err := conn.Read(buf[:]) // 服务端返回的信息
    if err != nil {
       fmt.Println("recv failed err:", err)
       return
    }

    // 打印Server端返回的数据
    fmt.Printf("Server端返回的数据:%s \n", string(buf[:serverMsg]))

    //fmt.Println("执行完毕,client程序退出...")

    // 退出程序
    os.Exit(0)

}

// 连接server
func connect() {

    // 获取ini文件数据参数
    ipaddress := ReadClieniniFile("socket.ipaddress")
    port := ReadClieniniFile("socket.port")
    ipAndPort := ipaddress + ":" + port

    // 判断ip和端口是否为空
    if ipaddress == "" && port == "" {
        fmt.Printf("ip地址与端口为空,ini文件未写入,无法开启...\n")
    } else {
        tcpAddr, err := net.ResolveTCPAddr("tcp4", ipAndPort)
        if err != nil {
            _, err := fmt.Fprintf(os.Stderr, "连接server失败: %s", err.Error())
            if err != nil {
                return
            }
            os.Exit(1)
        }

        conn, err := net.DialTCP("tcp", nil, tcpAddr)
        if err != nil {
            _, err := fmt.Fprintf(os.Stderr, "连接server失败,请确定Server是否开启: %s", err.Error())
            if err != nil {
                return
            }
            os.Exit(1)
        }

        //fmt.Println("Server连接成功...")

        // 执行数据发送
        sender(conn)
    }
}

// 主函数 执行
func main() {
    // 执行连接程序
      connect()
}
