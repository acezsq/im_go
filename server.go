package main

import (
	"fmt"
	"io"
	"net"
	"sync"
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
	user := NewUser(conn)
	//将用户加入map中
	this.mapLock.Lock()
	this.Onlinemap[user.Name] = user
	this.mapLock.Unlock()

	//向在线用户广播
	this.BroadCast(user, "已上线")

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			msg := string(buf[:n-1])

			//将得到的消息进行广播
			this.BroadCast(user, msg)
		}
	}()

	select {}

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
