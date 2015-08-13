package tcp2

import (
	"bytes"
	"fmt"
)
const (
	//json协议包接收buf大小
	JSON_RECV_BUF_LEN       = 10 * 1024
	//最大包长度 10M
	JSON_MAX_BUF      int64 = 1024 * 1024 * 1024 * 10
	//json协议包客户端缓存buf长度
	JSON_CLIENT_BUF         = 2 * bytes.MinRead
)
type ProtoJson struct {
	buf     *bytes.Buffer
	session *Session
}

func NewProtoJson(session *Session) *ProtoJson {
	return &ProtoJson{
		buf: bytes.NewBuffer(make([]byte, 0, JSON_CLIENT_BUF)),
		session :session,
	}
}

//读取数据
func (self *ProtoJson) read() error{
	readbuf := make([]byte, JSON_RECV_BUF_LEN)
	conn := self.session.conn
	//clientAddr := client.conn.RemoteAddr()
	for {
		n, err := conn.Read(readbuf)
		if err != nil {
			if !self.session.isClosed {
				//fmt.Println(clientAddr, "连接异常!", err)
				//关闭并释放资源，否则服务器会有CLOSE_WAIT出现，客户端会员 FIN_WAIT2
				self.session.Close()
				if err.Error() == "EOF" {
					self.session.onClose()
				} else {
					self.session.onError(err)
				}
			}
			return err
		}
		if n <= 0 {
			fmt.Println("客户端关闭连接")
			if !self.session.isClosed {
				self.session.Close()
				self.session.onClose()
			}
			return err
		}
		self.splitPackage(readbuf[:n])
	}
}

//拆包
func (self *ProtoJson) splitPackage(buf []byte)(err error){
	k := 0
	for i, byteval := range buf {
		//1: i > 0 && buf[i-1] == 13
		//2: i==0 && client.Allbuf.Len() > 0 && client.Allbuf.Bytes()[client.Allbuf.Len()-1] == 13
		if byteval == 10 {
			if i > 0 && buf[i-1] == 13 {
				//读取到包结束符号
				self.buf.Write(buf[k : i-1])
				self.session.onData(self.buf.Bytes())
				self.buf.Truncate(0)
				k = i + 1
			} else if i == 0 && self.buf.Len() > 0 && self.buf.Bytes()[self.buf.Len()-1] == 13 {
				//读取到包结束符号
				self.session.onData(self.buf.Bytes()[:self.buf.Len()-1])
				self.buf.Truncate(0)
				k = i + 1
			}
		}
	}
	//检测是否还有半截包
	if k < len(buf) {
		self.buf.Write(buf[k:])
		//包过大
		if int64(self.buf.Len()) > JSON_MAX_BUF {
			self.session.Close()
			self.session.onError(TOO_LAGER)
		}
		//fmt.Println("半截包内容:", string(client.Allbuf.Bytes()))
	}
	return
}
