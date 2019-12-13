/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agentclient

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"k8s.io/klog"
)

const (
	syncInterval = 5 * time.Second
)

// ClientSet consists of clients connected to each instance of an HA proxy server.
type ClientSet struct {
	mu      sync.Mutex              //protects the following
	clients map[string]*AgentClient // map between sererID and the client
	// connects to this server.

	agentID     string // ID of this agent
	address     string // proxy server address. Assuming HA proxy server
	serverCount int    // number of proxy server instances, should be 1
	// unless it is an HA server. Initialized when the ClientSet creates
	// the first client.
	dialOption grpc.DialOption
}

func (cs *ClientSet) clientsCount() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return len(cs.clients)
}

func (cs *ClientSet) hasIDLocked(serverID string) bool {
	for k, _ := range cs.clients {
		if k == serverID {
			return true
		}
	}
	return false
}

func (cs *ClientSet) HasID(serverID string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.hasIDLocked(serverID)
}

func (cs *ClientSet) addClientLocked(serverID string, c *AgentClient) error {
	if cs.hasIDLocked(serverID) {
		return fmt.Errorf("client for proxy server %s already exists", serverID)
	}
	cs.clients[serverID] = c
	return nil

}

func (cs *ClientSet) AddClient(serverID string, c *AgentClient) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.addClientLocked(serverID, c)
}

func (cs *ClientSet) RemoveClient(serverID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.clients, serverID)
}

func NewAgentClientSet(address, agentID string, dialOption grpc.DialOption) *ClientSet {
	return &ClientSet{
		clients:    make(map[string]*AgentClient),
		agentID:    agentID,
		address:    address,
		dialOption: dialOption,
	}
}

// CreateClients creates one agent client for each server instance.
func (cs *ClientSet) CreateClients() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	c, err := cs.newAgentClient()
	if err != nil {
		return err
	}
	cs.serverCount = c.serverCount
	klog.Infof("added client connecting to proxy server %s", c.serverID)
	klog.Infof("server count is %d", cs.serverCount)
	cs.clients[c.serverID] = c

	for len(cs.clients) != cs.serverCount {
		c, err := cs.newAgentClient()
		if err != nil {
			return err
		}
		if err := cs.addClientLocked(c.serverID, c); err != nil {
			klog.Infof("closing connection: %v", err)
			c.Close()
			continue
		}
		klog.Infof("added client connecting to proxy server %s", c.serverID)
	}
	return nil
}

func (cs *ClientSet) newAgentClient() (*AgentClient, error) {
	return newAgentClient(cs.address, cs.agentID, cs, cs.dialOption)
}

// sync makes sure that #clients >= #proxy servers
func (cs *ClientSet) sync() {
	for {
		time.Sleep(syncInterval)
		cc := cs.clientsCount()
		if cc < cs.serverCount {
			c, err := cs.newAgentClient()
			if err != nil {
				klog.Error(err)
				continue
			}
			if err := cs.AddClient(c.serverID, c); err != nil {
				klog.Infof("closing connection: %v", err)
				c.Close()
				continue
			}
			klog.Infof("sync added client connecting to proxy server %s", c.serverID)
			c.Serve()
		}
	}
}

func (cs *ClientSet) Serve() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for _, c := range cs.clients {
		go c.Serve()
	}
	go cs.sync()
}
