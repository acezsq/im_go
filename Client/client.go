package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前的client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}
	client.conn = conn
	//返回对象
	return client
}

// 处理server回应的，直接显示到标准输出
func (client *Client) DealResponse() {
	//一旦conn有数据，就直接copy到标准输出，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
	//和上面的代码等价
	// for {
	// 	buf := make()
	// 	client.conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>>>请输入合法范围的数字<<<<<<<<<<")
		return false
	}

}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn error:", err)
		return
	}

}

// 私聊模式
func (client *Client) privateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println("请输入聊天对象的[用户名](输入exit退出)")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>>请输入消息内容(输入exit退出)")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			//如果消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>>请输入消息内容(输入exit退出)")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println("请输入聊天对象的[用户名](输入exit退出)")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) publicChat() {
	//提示用户输入信息
	var chatMsg string
	fmt.Println(">>>>>>请输入聊天内容(exit表示退出)")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送给服务器

		//如果消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>>请输入聊天内容(exit表示退出)")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名:")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			client.publicChat()
		case 2:
			//私聊模式
			client.privateChat()
		case 3:
			//跟新用户名
			client.UpdateName()

		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认为127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认为8888)")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>>链接服务器失败...")
		return
	}

	//处理服务器的会执信息
	go client.DealResponse()

	fmt.Println(">>>>>>>>>服务器连接成功...")

	//启动客户端的业务
	client.Run()
}
