# GoSocket

### Centos 7 + socket + shell

#
## 简介:

#### socket 连接后,client端给server端发送shell命令

#### server接收到client传来的消息会将此消息传给shell脚本调用.

#### shell 脚本可自定义...

#

#### 配置:

Server端:
创建一个Server.ini文件,或者执行一下server程序,目录下会自动创建ini文件

#

#### Server.ini文件配置:
[socket]

ipaddress = 127.0.0.1

port = 8000

#

#### Client端:
创建一个Client.ini文件,或者执行一下Client程序,目录下会自动创建ini文件


#### Client.ini文件配置:

[socket]

ipaddress = 127.0.0.1  (与Server端ip一致)

port = 8000 (与Server端的端口一致)

#

### 执行 Server端:

#### go run server

#### server执行 client传来的shell命令



#

### 执行 Client端: 以命令形式传参执行

#### go run client 这里传你的参数


#

### 测试实例:

#### server.go

go run server.go 

时间:[ 2022-09-02 14:51:01 ]: 2022/09/02 14:51:01 127.0.0.1:8000 等待客户连接...

#

#### client.go

go run client.go hello,我是client

Server连接成功...

向server发送数据成功...


#

### server端接收到的client消息

Client传过来的消息: 我是server shell 脚本,我已经接收到了client传来的消息:hello,我是client

2022/09/02 14:52:27 127.0.0.1:57493  数据已处理,退出:  EOF

#

#### 如果无法执行: 

#### go get github.com/gookit/ini/v2 


