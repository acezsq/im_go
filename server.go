package main

import (
	"fmt"
	"net"
)

type Server struct {
	ID   string
	Port int
}

func NewServer(id string, port int) *Server {
	server := &Server{
		ID:   id,
		Port: port,
	}
	return server
}

func (this *Server) Handler(conn net.Conn) {
	//
	fmt.Println("连接创建成功")
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
