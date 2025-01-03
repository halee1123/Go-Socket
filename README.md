# GoSocket

## 简介

GoSocket 是一款基于 Linux、Socket 和 Shell 的简单通信工具。通过 Socket 连接，客户端可以向服务器发送 Shell 命令，服务器将命令传递给 Shell 脚本进行处理，并将结果返回给客户端。为了限制并发连接，系统采用了令牌桶算法来控制每秒处理的请求数，确保资源不会被过度占用。

## 配置

### Server 端

在 Server 端，您可以创建一个 `Server.ini` 文件，也可以直接执行以下 server 程序，系统将自动创建 ini 文件。

#### Server.ini 文件配置:

```ini
[socket]

# Server IP地址
ipaddress = 127.0.0.1

# Server 端口
port = 8001

# 自定义令牌数量
token_capacity = 10

# 每秒生成的令牌数量
tokens_per_sec = 2

ini文件中,可自定义令牌数量与每秒生成的令牌数量。

Client 端
在 Client 端，同样可以创建一个 Client.ini 文件，或者执行 Client 程序，系统将自动创建 ini 文件。

Client.ini 文件配置:
[socket]
ipaddress = 127.0.0.1  # 与 Server 端 ip 一致
port = 8000  # 与 Server 端的端口一致


## 使用方法:

执行 Server 端: go run server


执行 Client 端: go run client (这里输入你所放行的shell命令)
比如: go run client readIPaddress


## 测试实例
server.go:

##执行server: go run server.go

2023/12/28 16:25:23 服务器启动，正在等待客户端连接于:  127.0.0.1:8001
2023/12/28 16:25:42 127.0.0.1:50085  客户端连接成功
2023/12/28 16:25:42 127.0.0.1:50085 服务器接收到的数据处理完成，客户端已退出:  EOF


client.go
##执行client:  go run client.go readIPaddress

Server 服务器返回已处理的数据:
以下为Server接收到的:[ readIPaddress ]参数所执行的代码;

127.0.0.1 192.168.31.119 169.254.218.112


## 注意事项
如果执行时无法找到依赖，请执行以下命令：
go get github.com/gookit/ini/v2



