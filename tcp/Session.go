package tcp2

import (
	"net"
	"strings"
	"time"
)
type OnConnect func(session *Session)
type OnData func([]byte)
type OnClose func()
type OnError func(error)

/**
 *@mark+ 	回话类,采用了基于事件触发的机制，封装了三种事件对象onData，onClose，onError
 *@author 	zdq
 *@time  	2015-8-12
 */
type Session struct {
	conn     *net.TCPConn
	protocol Protocol
	isClosed bool
	attrs    map[interface{}]interface{} //绑定属性
	intime     int64                  //连接时间
	onData   OnData
	onClose  OnClose
	onError  OnError
}

func NewSession(protocolGenerator ProtocolGenerator, conn *net.TCPConn) *Session {
	conn.SetNoDelay(false)
	session := &Session{
		conn:     conn,
		isClosed: false,
		attrs:    make(map[interface{}]interface{},2),
		intime:   time.Now().Unix(),
		onData:   func([]byte) {},
		onClose:  func() {},
		onError:  func(error) {},
	}
	session.protocol = protocolGenerator.New(session)
	return session
}
/**
 *@mark+  获取会话对应的协议
 *@return 协议接口对象，上层转换成相应协议对象
 */
func (self *Session) Proto() Protocol{
	return self.protocol
}
/**
 *@mark+  绑定数据事件
 *@param  OnData 数据回调函数，通过参数data []byte 将数据返回给用户
 */
func (self *Session) OnData(fn OnData) {
	if fn != nil {
		self.onData = fn
	}
}

/**
 *@mark+  绑定连接关闭事件
 *@param  OnClose 数据回调函数
 */
func (self *Session) OnClose(fn OnClose) {
	if fn != nil {
		self.onClose = fn
	}
}

/**
 *@mark+  绑定错误事件
 *@param  OnError 数据回调函数，通过参数error 将错误对象返回给用户
 */
func (self *Session) OnError(fn OnError) {
	if fn != nil {
		self.onError = fn
	}
}

/**
 *@mark+  关闭连接，go的net库在对方连接关闭后，服务器方也要发送关闭
 */
func (self *Session) Close() {
	self.isClosed = true
	self.conn.Close()
}

/**
 *@mark+  绑定事件监听
 *@param  key 监听事件类型，有以下事件类型，error，data,close
 */
func (self *Session) On(key string, event interface{}) {
	if event == nil {
		return
	}
	switch strings.ToLower(key) {
	case "error":
		if fn, ok := event.(OnError); ok {
			self.onError = fn
		}
	case "data":
		if fn, ok := event.(OnData); ok {
			self.onData = fn
		}
	case "close":
		if fn, ok := event.(OnClose); ok {
			self.onClose = fn
		}
	}
}
func (self *Session) Set(key interface{}, val interface{}) {
	self.attrs[key] = val
}
func (self *Session) Get(key interface{}) interface{} {
	return self.attrs[key]
}
func (self *Session) GetString(key interface{}) string {
	v := self.attrs[key]
	if str, ok := v.(string); ok {
		return str
	} else {
		return ""
	}
}
func (self *Session) GetInt(key interface{}) int {
	v := self.attrs[key]
	if val, ok := v.(int); ok {
		return val
	} else {
		return 0
	}
}
func (self *Session) GetInt64(key interface{}) int64 {
	v := self.attrs[key]
	if val, ok := v.(int64); ok {
		return val
	} else {
		return 0
	}
}
func (self *Session) GetUint64(key interface{}) uint64 {
	v := self.attrs[key]
	if val, ok := v.(uint64); ok {
		return val
	} else {
		return 0
	}
}
func (self *Session) Del(key interface{}) {
	delete(self.attrs, key)
}
func (self *Session) RemoteAddr() string {
	return self.conn.RemoteAddr().String()
}
func (self *Session) IP() string {
	arr := strings.Split(self.conn.RemoteAddr().String(), ":")
	return arr[0]
}
func (self *Session) SetTimeout(sec int32) {
	self.conn.SetReadDeadline(time.Now().Add(time.Duration(sec) * time.Second))
}
/**
 *@mark  循环读取数据，委托给协议处理
 */
func (self *Session) readLoop() error{
	return self.protocol.read()
}
/**
 *@mark+ 同步发送,调用底层连接直接发送，数据包要求已经实现了协议
 */
func (self *Session) Write(data []byte) (n int, err error){
	return self.conn.Write(data)
}
/**
 *@mark+ 同步发送，调用封装协议发送，自动包装协议
 */
func (self *Session) Send(data []byte,params ...interface{}) (n int, err error){
	return self.protocol.write(data,params...)
}

/**
 *@mark+ 异步发送，交由一个发送协程来处理,调用底层连接直接发送，数据包要求已经实现了协议
 */
func (self *Session) AsynWrite(data []byte) {

}

/**
 *@mark+ 异步发送，交由一个发送协程来处理，调用封装协议发送，自动包装协议
 */
func (self *Session) AsynSend(data []byte) {

}
