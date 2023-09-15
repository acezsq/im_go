package main

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

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

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}

// 监听当前user的channel的go协程，一但有消息发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
