# go-websocket

基于gorilla/websocket封装的websocket库，实现系统推送，群组推送，多客户端推送。

### 一、安装

```
go get -u github.com/lackone/go-websocket
```

### 二、启动示例

```
go run example/main.go
```

默认端口：8080

### 三、代码示例

```go
package main

import (
	"fmt"
	go_websocket "github.com/lackone/go-websocket"
	"log"
	"net/http"
)

func main() {
	manage := go_websocket.NewClientManage()
	go manage.Run()

	//ws请求回调
	go_websocket.WsClientHandler.Register("/test", func(client *go_websocket.Client, params interface{}) (go_websocket.IResponse, error) {
		fmt.Println(params)
		return go_websocket.NewOkClientRes(map[string]interface{}{
			"test": "test",
		}), nil
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client, err := go_websocket.Upgrade(manage, w, r)
		if err != nil {
			log.Fatalln(err)
		}

		res := go_websocket.NewOkClientRes(map[string]interface{}{
			"client_id": client.GetID(),
		})

		client.SendResponse(res)
	})

	//广播
	http.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.Broadcast(bytes)
	})

	//针对系统推送
	http.HandleFunc("/push_system", func(w http.ResponseWriter, r *http.Request) {
		systemId := r.FormValue("system_id")
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.SendSystemMsg(bytes, systemId)
	})

	//针对组推送
	http.HandleFunc("/push_group", func(w http.ResponseWriter, r *http.Request) {
		group := r.FormValue("group")
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.SendGroupMsg(bytes, group)
	})

	//针对客户端推送
	http.HandleFunc("/push_client", func(w http.ResponseWriter, r *http.Request) {
		clientId := r.FormValue("client_id")
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.SendClientMsg(bytes, clientId)
	})

	//添加组
	http.HandleFunc("/add_group", func(w http.ResponseWriter, r *http.Request) {
		clientId := r.FormValue("client_id")
		group := r.FormValue("group")

		c := manage.GetClientByID(clientId)

		manage.AddGroupsByClient(c, group)
	})

	//删除组
	http.HandleFunc("/del_group", func(w http.ResponseWriter, r *http.Request) {
		clientId := r.FormValue("client_id")
		group := r.FormValue("group")

		c := manage.GetClientByID(clientId)

		manage.RemoveGroupsByClient(c, group)
	})

	//列表数据
	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(manage.GetSystemList())
		fmt.Println(manage.GetGroupsList())
		fmt.Println(manage.GetClientList())
	})

	http.ListenAndServe(":8080", nil)
}

```