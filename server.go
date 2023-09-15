package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	ID        string
	Port      int
	Onlinemap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(id string, port int) *Server {
	server := &Server{
		ID:        id,
		Port:      port,
		Onlinemap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Messge channel的协程给在线用户的chann发消息
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		//将msg发送给全部在线user
		this.mapLock.Lock()
		for _, cli := range this.Onlinemap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendmsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendmsg
}

func (this *Server) Handler(conn net.Conn) {
	//创建用户
	user := NewUser(conn, this)

	//上线业务
	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1])

			//用户针对msg的消息进行处理
			user.DoMessage(msg)

			//用户的任意消息代表用户是活跃的
			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
			//当前用户活跃，重置定时器
			//不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 30):
			//已经超时，将当前的user强制关闭
			user.SendMsg("你被踢了")

			//销毁用户的资源
			close(user.C)

			//关闭连接
			conn.Close()

			//退出当前的handler
			return
		}
	}

}

func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", this.ID, this.Port))
	if err != nil {
		fmt.Println("net.listen error:", err)
		return
	}
	//defer close
	defer listener.Close()

	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
		}
		//do handler
		go this.Handler(conn)
	}

}
