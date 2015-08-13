# tcp
This is a go language of the TCP Library of the packaging, the realization of the two agreements

Server Demo

默认是基于json的与\r\n 分隔的协议

tcp.CreateServer(func(session *tcp.Session) {
	session.OnData(func(data[] byte) {
		fmt.Println("接收到数据: ",string(data))
	})
}).Listen(":8020", func(err error, server *tcp.Server) {
	if err != nil {
		fmt.Println("创建服务器失败!", err)
	} else {
		fmt.Println("服务器创建成功!")
	}
})


Client Demo

json数据包与\r\n 分隔

data:=[]byte(`{"command":"test"}`+"\r\n")
client:=tcp.NewClient(nil,50,[]byte(`{"command":"Ping"}`+"\r\n"))
client.Connect("1127.0.0.1:8020",func(session *tcp.Session){
	session.OnData(func(data []byte){
		fmt.Println("接收到数据: ",string(data))
	})	
})