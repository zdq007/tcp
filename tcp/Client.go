package tcp2

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	addr              *net.TCPAddr
	onConnect         OnConnect
	session           *Session
	status            int32  //1心跳  2重连  0关闭
	heartDuration     int32  //心跳间隔时间
	heartPage         []byte //心跳数据包
	protocolGenerator ProtocolGenerator
}

func CreateClient() *Client {
	return NewClient(&DefaulJsonProtocolGenerator{}, 0, nil)
}
func NewClient(protocolGenerator ProtocolGenerator, keeptimer int32, heartpage []byte) *Client {
	if protocolGenerator == nil {
		protocolGenerator = &DefaulJsonProtocolGenerator{}
	}
	client := &Client{
		protocolGenerator: protocolGenerator,
		status:            0,
	}
	if keeptimer > 0 && heartpage != nil {
		client.SetKeepAlive(keeptimer, heartpage)
	}
	return client
}

//自定义协议生成器，连接前设置
func (self *Client) SetProtocolGenerator(protocolGenerator ProtocolGenerator) {
	self.protocolGenerator = protocolGenerator
}

//设置了保持客户端活动才支持重连和自动心跳发送机制，参数为发送间隔时间和心跳包内容
//需要保持活动要在连接前设置
func (self *Client) SetKeepAlive(second int32, data []byte) {
	self.heartDuration = second
	self.heartPage = data
	self.status = 1
}
func (self *Client) Connect(addr string, onConnect OnConnect) error {
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
	self.onConnect = onConnect
	self.addr = paddr
	return self.connect()
}

func (self *Client) connect() error {
	con, err := net.DialTCP("tcp", nil, self.addr)
	if err != nil {
		fmt.Println(err)
		if self.status == 1 {
			self.status = 2
			go self.aliveManager()
		}
		return err
	}
	if self.status == 1 {
		go self.aliveManager()
	}
	con.SetNoDelay(false)
	self.session = NewSession(self.protocolGenerator, con)
	self.onConnect(self.session)
	go func() {
		err := self.session.readLoop()
		if err != nil {
			self.status = 2
		}
	}()
	return nil
}

//心跳和重连共用的活动管理器线程，由status状态开关来作重连或发生心跳包
//连接关闭后状态置2开始重连，连接恢复后状态置1发送心跳
func (self *Client) aliveManager() {
	for self.status != 0 {
		if self.status == 1 {
			_, err := self.session.Write(self.heartPage)
			if err != nil {
				fmt.Println("发送心跳包异常:", err)
			}
			time.Sleep(time.Duration(self.heartDuration) * time.Second)

		} else if self.status == 2 {
			fmt.Println("连接异常，尝试重新连接")
			err := self.connect()
			if err == nil {
				self.status = 1
			}
		}
		time.Sleep(time.Second * 2)
	}
}

//手动关闭连接，手动关闭的不会重连
func (self *Client) Close() {
	self.status = 0
	if self.session != nil {
		self.session.Close()
	}
}
/**
 *@mark+ 同步发送
 */
func (self *Client) Write(data []byte) (n int, err error){
	return self.session.Write(data)
}
/**
 *@mark+ 同步发送
 */
func (self *Client) Send(data []byte) (n int, err error){
	return self.session.Write(data)
}
