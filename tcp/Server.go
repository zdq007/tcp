package tcp2

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

/**
 *@mark+ 	tcp服务器封装，简化了tcp的调用流程，并封装了简单的协议buf，实现了数据流的拆分和封装
 *@author 	zdq
 *@time  	2015-8-12
 */
type Server struct {
	listener          *net.TCPListener
	sessions          []*Session
	protocolGenerator ProtocolGenerator
	onConnect         OnConnect
}

var (
	addrFormatError = errors.New("Address format error, you should enter this type of address (ip:port).")
)

type Listener func(err error, server *Server)

/**
 *@mark+   创建服务
 *@param  onConnect 连接处理函数
 *@return 服务对象
 */
func CreateServer(onConnect OnConnect) *Server {
	var protocolGenerator ProtocolGenerator
	protocolGenerator = &DefaulJsonProtocolGenerator{}
	return NewServer(protocolGenerator, onConnect)
}

/**
 *@mark+   创建服务
 *@param  protocol  协议
 *@param  onConnect 连接处理函数
 *@return 服务对象
 */
func NewServer(protocolGenerator ProtocolGenerator, onConnect OnConnect) *Server {
	if protocolGenerator == nil {
		protocolGenerator = &DefaulJsonProtocolGenerator{}
	}
	if onConnect == nil {
		onConnect = func(session *Session) {}
	}
	return &Server{protocolGenerator: protocolGenerator, onConnect: onConnect}
}

/**
 *@mark+   开启服务监听
 *@param  addr 		连接地址，示例：127.0.0.1:80  *:80 :80
 *@param  listener 	监听回调函数
 *@return 错误对象
 */
func (self *Server) Listen(addr string, listener Listener) error {
	ipAndPort := strings.Split(addr, ":")
	ipstr := "*"
	var port int
	if len(ipAndPort) < 2 {
		return addrFormatError
	} else {
		var err error
		if len(ipAndPort[0]) > 0 {
			ipstr = ipAndPort[0]
		}
		port, err = strconv.Atoi(ipAndPort[1])
		if err != nil {
			return addrFormatError
		}
	}
	paddr := &net.TCPAddr{IP: net.ParseIP(ipstr), Port: port}
	l, err := net.ListenTCP("tcp", paddr)
	listener(err, self)
	if err == nil {
		self.listener = l
		return self.accept()
	} else {
		return err
	}
}

/**
 *@mark-  循环堵塞接收连接
 */
func (self *Server) accept() error {
	for {
		conn, err := self.listener.AcceptTCP()
		if err != nil {
			return err
		}
		session := NewSession(self.protocolGenerator, conn)
		self.onConnect(session)
		go session.readLoop()
	}
}
