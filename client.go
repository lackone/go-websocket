package go_websocket

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Client struct {
	id           string              //客户端ID
	conn         *websocket.Conn     //ws连接
	clientManage *ClientManage       //管理客户端
	systemId     string              //系统ID，该客户端属于哪个系统
	groups       map[string]struct{} //组，该客户端加入的组
	groupsLock   sync.RWMutex        //组锁
	send         chan []byte         //发送消息通道
}

func NewClient(id string, systemId string, conn *websocket.Conn, clientMange *ClientManage) *Client {
	return &Client{
		id:           id,
		conn:         conn,
		clientManage: clientMange,
		systemId:     systemId,
		groups:       make(map[string]struct{}),
		groupsLock:   sync.RWMutex{},
		send:         make(chan []byte, 256),
	}
}

// 客户端ID
func (c *Client) GetID() string {
	return c.id
}

// 系统ID
func (c *Client) GetSystemId() string {
	return c.systemId
}

// 所有组
func (c *Client) GetGroups() []string {
	c.groupsLock.RLock()
	defer c.groupsLock.RUnlock()
	list := make([]string, 0)
	if len(c.groups) > 0 {
		for k := range c.groups {
			list = append(list, k)
		}
	}
	return list
}

// 加入组
func (c *Client) AddGroup(groups ...string) {
	if len(groups) <= 0 {
		return
	}
	c.groupsLock.Lock()
	defer c.groupsLock.Unlock()
	for _, g := range groups {
		c.groups[g] = struct{}{}
	}
}

// 删除组
func (c *Client) DelGroup(groups ...string) {
	if len(groups) <= 0 {
		return
	}
	c.groupsLock.Lock()
	defer c.groupsLock.Unlock()
	for _, g := range groups {
		delete(c.groups, g)
	}
}

// 发送消息
func (c *Client) SendMsg(msg []byte) error {
	defer func() {
		if err := recover(); err != nil {
			Log.Error(context.Background(), "SendMsg Panic", err)
		}
	}()

	res, err := c.clientManage.resFormatFn(c, msg)
	if err != nil {
		Log.Error(context.Background(), "resFormatFn Error", err)
		return err
	}

	return c.SendResponse(res)
}

// 发送响应
func (c *Client) SendResponse(res IResponse) error {
	defer func() {
		if err := recover(); err != nil {
			Log.Error(context.Background(), "SendResponse Panic", err)
		}
	}()

	bytes, err := res.GetBytes()
	if err != nil {
		Log.Error(context.Background(), "GetBytes Error", err)
		return err
	}

	select {
	case c.send <- bytes:
	}

	return nil
}

// 读循环
func (c *Client) ReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			Log.Error(context.Background(), "ReadLoop Panic", err)
		}
	}()

	defer func() {
		c.clientManage.UnRegister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(ReadLimit)
	c.conn.SetReadDeadline(time.Now().Add(ReadDeadline))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(ReadDeadline))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Log.Error(context.Background(), "ReadMessage Error", err)
			}
			return
		}

		if c.clientManage.reqFormatFn == nil {
			c.clientManage.reqFormatFn = c.clientManage.DefaultRequestFormatFunc()
		}

		err = c.ProcessMessage(msg)
		if err != nil {
			Log.Error(context.Background(), "ProcessMessage Error", err)
		}
	}
}

// 处理消息
func (c *Client) ProcessMessage(msg []byte) error {
	req, err := c.clientManage.reqFormatFn(c, msg)
	if err != nil {
		return err
	}

	handler, ok := WsClientHandler.GetHandler(req.GetUrl())
	if !ok {
		return errors.New(req.GetUrl() + " handler not found")
	}

	res, err := handler(c, req.GetParams())
	if err != nil {
		return err
	}

	return c.SendResponse(res)
}

// 写循环
func (c *Client) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			Log.Error(context.Background(), "WriteLoop Panic", err)
		}
	}()

	//定时器，定时发送心跳包
	ticker := time.NewTicker(HeartbeatInterval)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if c.clientManage.resFormatFn == nil {
				c.clientManage.resFormatFn = c.clientManage.DefaultResponseFormatFunc()
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteDeadline))

			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(PingMessage)); err != nil {
				return
			}
		}
	}
}
