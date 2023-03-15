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

	http.HandleFunc("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.Broadcast(bytes)
	})

	http.HandleFunc("/push_system", func(w http.ResponseWriter, r *http.Request) {
		systemId := r.FormValue("system_id")
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.SendSystemMsg(bytes, systemId)
	})

	http.HandleFunc("/push_group", func(w http.ResponseWriter, r *http.Request) {
		group := r.FormValue("group")
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.SendGroupMsg(bytes, group)
	})

	http.HandleFunc("/push_client", func(w http.ResponseWriter, r *http.Request) {
		clientId := r.FormValue("client_id")
		msg := r.FormValue("msg")

		bytes, _ := go_websocket.NewOkClientRes(map[string]interface{}{
			"msg": msg,
		}).GetBytes()

		manage.SendClientMsg(bytes, clientId)
	})

	http.HandleFunc("/add_group", func(w http.ResponseWriter, r *http.Request) {
		clientId := r.FormValue("client_id")
		group := r.FormValue("group")

		c := manage.GetClientByID(clientId)

		manage.AddGroupsByClient(c, group)
	})

	http.HandleFunc("/del_group", func(w http.ResponseWriter, r *http.Request) {
		clientId := r.FormValue("client_id")
		group := r.FormValue("group")

		c := manage.GetClientByID(clientId)

		manage.RemoveGroupsByClient(c, group)
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(manage.GetSystemList())
		fmt.Println(manage.GetGroupsList())
		fmt.Println(manage.GetClientList())
	})

	http.ListenAndServe(":8080", nil)
}
