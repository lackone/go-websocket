package go_websocket

import (
	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/websocket"
	"net/http"
)

var (
	snowflakeNode *snowflake.Node
	wsUpgrader    = &websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func init() {
	ip, err := GetExternalIP()
	if err != nil {
		panic(err)
	}
	n := InetAtoN(ip)
	snowflakeNode, err = snowflake.NewNode(n % 1023)
	if err != nil {
		panic(err)
	}
}

func Upgrade(clientManage *ClientManage, w http.ResponseWriter, r *http.Request) (*Client, error) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	systemId := r.FormValue("system_id")
	group := r.FormValue("group")

	if systemId == "" {
		systemId = RemoteIp(r)
	}

	//生成客户端ID
	clientId := GenerateClientId()

	//创建客户端
	wsClient := NewClient(clientId, systemId, conn, clientManage)

	if len(group) > 0 {
		clientManage.AddGroupsByClient(wsClient, group)
	}

	//添加客户端
	clientManage.Register(wsClient)

	go wsClient.ReadLoop()
	go wsClient.WriteLoop()

	return wsClient, nil
}

func GenerateClientId() string {
	return snowflakeNode.Generate().String()
}
