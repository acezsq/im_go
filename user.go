package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//创建用户

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user的channel的go协程
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (this *User) Online() {
	//用户上线将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.Onlinemap[this.Name] = this
	this.server.mapLock.Unlock()

	//向在线用户广播
	this.server.BroadCast(this, "已上线")
}

// 用户的下线业务
func (this *User) Offline() {
	//用户下线将用户从OnlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.Onlinemap, this.Name)
	this.server.mapLock.Unlock()

	//向在线用户广播
	this.server.BroadCast(this, "下线")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n"))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.Onlinemap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]
		// newName := msg[7:]
		//判断这个名字是否已经存在
		_, ok := this.server.Onlinemap[newName]
		if ok {
			this.SendMsg("当前用户名被使用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.Onlinemap, this.Name)
			this.server.Onlinemap[newName] = this
			this.server.mapLock.Unlock()
			this.Name = newName
			this.SendMsg("您已经更新用户名：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式： to|张三|消息内容
		//1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用\"to|张三|你好啊\"格式。\n")
			return
		}
		//2 根据用户名得到对应的user对象
		remoteUser, ok := this.server.Onlinemap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}
		//3 获取消息内容，通过对方的user对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说：" + content)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前user的channel的go协程，一但有消息发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
