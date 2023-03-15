package go_websocket

import (
	"encoding/json"
	"sync"
)

type ResponseFormatFunc func(c *Client, data []byte) (res IResponse, err error)
type RequestFormatFunc func(c *Client, data []byte) (req IRequest, err error)

type ClientManage struct {
	clients     map[string]*Client //所有客户端
	clientsLock sync.RWMutex       //客户端锁

	register   chan *Client //注册通道
	unregister chan *Client //退出通道

	broadcast chan []byte //广播通道

	groups      map[string]map[string]*Client //所有组客户端
	groupsLock  sync.RWMutex                  //组锁
	systems     map[string]map[string]*Client //所有系统客户端
	systemsLock sync.RWMutex                  //系统锁

	reqFormatFn RequestFormatFunc  //请求格式化方法
	resFormatFn ResponseFormatFunc //响应格式化方法
}

func NewClientManage() *ClientManage {
	return &ClientManage{
		clients:     make(map[string]*Client),
		clientsLock: sync.RWMutex{},
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		groups:      make(map[string]map[string]*Client),
		groupsLock:  sync.RWMutex{},
		systems:     make(map[string]map[string]*Client),
		systemsLock: sync.RWMutex{},
	}
}

// 注册
func (cm *ClientManage) Register(c *Client) {
	cm.register <- c
}

// 退出
func (cm *ClientManage) UnRegister(c *Client) {
	cm.unregister <- c
}

// 获取客户端
func (cm *ClientManage) GetClientByID(id string) *Client {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()

	if c, ok := cm.clients[id]; ok {
		return c
	}
	return nil
}

// 客户端列表
func (cm *ClientManage) GetClientList() []string {
	if len(cm.clients) <= 0 {
		return nil
	}

	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()

	list := make([]string, 0)
	for k, _ := range cm.clients {
		list = append(list, k)
	}
	return list
}

// 获取系统列表
func (cm *ClientManage) GetSystemList() map[string][]string {
	if len(cm.systems) <= 0 {
		return nil
	}

	cm.systemsLock.RLock()
	defer cm.systemsLock.RUnlock()

	list := make(map[string][]string)
	for k, system := range cm.systems {
		list[k] = make([]string, 0)
		for id, _ := range system {
			list[k] = append(list[k], id)
		}
	}
	return list
}

// 获取组列表
func (cm *ClientManage) GetGroupsList() map[string][]string {
	if len(cm.groups) <= 0 {
		return nil
	}

	cm.groupsLock.RLock()
	defer cm.groupsLock.RUnlock()

	list := make(map[string][]string)
	for k, group := range cm.groups {
		list[k] = make([]string, 0)
		for id, _ := range group {
			list[k] = append(list[k], id)
		}
	}
	return list
}

// 事件循环
func (cm *ClientManage) Run() {
	for {
		select {
		case client, ok := <-cm.register:
			if !ok {
				//通道关闭直接return
				return
			}
			cm.AddClient(client)
		case client, ok := <-cm.unregister:
			if !ok {
				//通道关闭直接return
				return
			}
			cm.RemoveClient(client)
		case msg, ok := <-cm.broadcast:
			if !ok {
				//通道关闭直接return
				return
			}
			cm.Broadcast(msg)
		}
	}
}

// 全局广播
func (cm *ClientManage) Broadcast(msg []byte) {
	if len(cm.clients) <= 0 {
		return
	}
	for _, c := range cm.clients {
		c.SendMsg(msg)
	}
}

// 给组发消息
func (cm *ClientManage) SendGroupMsg(msg []byte, groups ...string) {
	if len(groups) <= 0 {
		return
	}
	for _, g := range groups {
		if _, ok := cm.groups[g]; ok {
			for _, c := range cm.groups[g] {
				c.SendMsg(msg)
			}
		}
	}
}

// 给系统发消息
func (cm *ClientManage) SendSystemMsg(msg []byte, systemIds ...string) {
	if len(systemIds) <= 0 {
		return
	}
	for _, s := range systemIds {
		if _, ok := cm.systems[s]; ok {
			for _, c := range cm.systems[s] {
				c.SendMsg(msg)
			}
		}
	}
}

// 给多个客户端发消息
func (cm *ClientManage) SendClientMsg(msg []byte, clientIds ...string) {
	if len(clientIds) <= 0 {
		return
	}
	for _, id := range clientIds {
		c := cm.GetClientByID(id)
		if c != nil {
			c.SendMsg(msg)
		}
	}
}

// 添加客户端
func (cm *ClientManage) AddClient(c *Client) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	if _, ok := cm.clients[c.GetID()]; ok {
		return
	}

	cm.clients[c.GetID()] = c

	//添加进系统
	cm.AddSystemIdByClient(c, c.GetSystemId())

	//添加进组
	cm.AddGroupsByClient(c, c.GetGroups()...)
}

// 给客户端添加系统
func (cm *ClientManage) AddSystemIdByClient(c *Client, systemId string) {
	if len(systemId) <= 0 {
		return
	}

	cm.systemsLock.Lock()
	defer cm.systemsLock.Unlock()

	if _, ok := cm.systems[systemId]; !ok {
		cm.systems[systemId] = make(map[string]*Client)
	}

	cm.systems[systemId][c.GetID()] = c
}

// 给客户端添加组
func (cm *ClientManage) AddGroupsByClient(c *Client, groups ...string) {
	if len(groups) <= 0 {
		return
	}

	cm.groupsLock.Lock()
	defer cm.groupsLock.Unlock()

	for _, g := range groups {
		if _, ok := cm.groups[g]; !ok {
			cm.groups[g] = make(map[string]*Client)
		}
		cm.groups[g][c.GetID()] = c

		c.AddGroup(g)
	}
}

// 删除客户端
func (cm *ClientManage) RemoveClient(c *Client) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	delete(cm.clients, c.GetID())

	close(c.send)

	//删除系统
	cm.RemoveSystemIdByClient(c, c.GetSystemId())

	//删除组
	cm.RemoveGroupsByClient(c, c.GetGroups()...)
}

// 给客户端删除系统
func (cm *ClientManage) RemoveSystemIdByClient(c *Client, systemId string) {
	if len(systemId) <= 0 {
		return
	}

	cm.systemsLock.Lock()
	defer cm.systemsLock.Unlock()

	if _, ok := cm.systems[systemId]; !ok {
		return
	}

	delete(cm.systems[systemId], c.GetID())
}

// 给客户端删除组
func (cm *ClientManage) RemoveGroupsByClient(c *Client, groups ...string) {
	if len(groups) <= 0 {
		return
	}

	cm.groupsLock.Lock()
	defer cm.groupsLock.Unlock()

	for _, g := range groups {
		if _, ok := cm.groups[g]; !ok {
			continue
		}
		delete(cm.groups[g], c.GetID())

		c.DelGroup(g)
	}
}

// 设置响应格式化方法
func (cm *ClientManage) SetResponseFormatFunc(fn ResponseFormatFunc) {
	cm.resFormatFn = fn
}

// 默认的响应格式化
func (cm *ClientManage) DefaultResponseFormatFunc() ResponseFormatFunc {
	return func(c *Client, data []byte) (IResponse, error) {
		res := &ClientResponse{}
		if err := json.Unmarshal(data, res); err != nil {
			return nil, err
		}
		return res, nil
	}
}

// 设置请求格式化方法
func (cm *ClientManage) SetRequestFormatFunc(fn RequestFormatFunc) {
	cm.reqFormatFn = fn
}

// 默认的请求格式化
func (cm *ClientManage) DefaultRequestFormatFunc() RequestFormatFunc {
	return func(c *Client, data []byte) (IRequest, error) {
		req := &ClientRequest{}
		if err := json.Unmarshal(data, req); err != nil {
			return nil, err
		}
		return req, nil
	}
}
