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

package agentserver

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc/metadata"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-network-proxy/proto/agent"
)

// ProxyClientConnection...
type ProxyClientConnection struct {
	Mode      string
	Grpc      agent.ProxyService_ProxyServer
	HTTP      net.Conn
	connected chan struct{}
	connectID int64
}

func (c *ProxyClientConnection) send(pkt *agent.Packet) error {
	if c.Mode == "grpc" {
		stream := c.Grpc
		return stream.Send(pkt)
	} else if c.Mode == "http-connect" {
		if pkt.Type == agent.PacketType_CLOSE_RSP {
			return c.HTTP.Close()
		} else if pkt.Type == agent.PacketType_DATA {
			_, err := c.HTTP.Write(pkt.GetData().Data)
			return err
		} else if pkt.Type == agent.PacketType_DIAL_RSP {
			return nil
		} else {
			return fmt.Errorf("attempt to send via unrecognized connection type %v", pkt.Type)
		}
	} else {
		return fmt.Errorf("attempt to send via unrecognized connection mode %q", c.Mode)
	}
}

// ProxyServer
type ProxyServer struct {
	mu sync.Mutex //protects the following
	// A map between agentID and its grpc connections.
	// For a given agent, ProxyServer prefers backends[agentID][0] to send
	// traffic, because backends[agentID][1:] are more likely to be closed
	// by the agent to deduplicate connections to the same server.
	backends map[string][]agent.AgentService_ConnectServer
	agentIDs []string

	// A map between agentID and the ID of the connections handled by this agent.
	mu2             sync.Mutex // protects the following
	connIDToAgentID map[int64]string

	// connID track
	Frontends   map[int64]*ProxyClientConnection
	PendingDial map[int64]*ProxyClientConnection

	serverID    string // unique ID of this server
	serverCount int    // Number of proxy server instances, should be 1 unless it is a HA server.
}

var _ agent.AgentServiceServer = &ProxyServer{}

var _ agent.ProxyServiceServer = &ProxyServer{}

func (s *ProxyServer) addBackend(agentID string, conn agent.AgentService_ConnectServer) {
	klog.Infof("register Backend %v for agentID %s", conn, agentID)
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.backends {
		if k == agentID {
			s.backends[k] = append(s.backends[k], conn)
			return
		}
	}
	s.backends[agentID] = []agent.AgentService_ConnectServer{conn}
	s.agentIDs = append(s.agentIDs, agentID)
}

func (s *ProxyServer) removeBackend(agentID string, conn agent.AgentService_ConnectServer) {
	klog.Infof("remove Backend %v for agentID %s", conn, agentID)
	s.mu.Lock()
	defer s.mu.Unlock()
	backends, ok := s.backends[agentID]
	if !ok {
		klog.Warningf("can't find agentID %s in the backends", agentID)
		return
	}
	for i, c := range backends {
		if c == conn {
			s.backends[agentID] = append(s.backends[agentID][:i], s.backends[agentID][i+1:]...)
			if len(s.backends[agentID]) == 0 {
				delete(s.backends, agentID)
				for i := range s.agentIDs {
					if s.agentIDs[i] == agentID {
						s.agentIDs[i] = s.agentIDs[len(s.agentIDs)-1]
						s.agentIDs = s.agentIDs[:len(s.agentIDs)-1]
						break
					}
				}
			}
			return
		}
	}
	klog.Warningf("can't find conn %v for agentID %s in the backends", conn, agentID)
}

func (s *ProxyServer) randomBackend() (agent.AgentService_ConnectServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.backends) == 0 {
		return nil, fmt.Errorf("No backend available")
	}
	agentID := s.agentIDs[rand.Intn(len(s.agentIDs))]
	return s.backends[agentID][0], nil
}

// func (s *ProxyServer) backend(connID int64) (agent.AgentService_ConnectServer, error) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	agentID, ok := s.toAgentID(connID)
// 	if !ok {
// 		return nil, fmt.Errorf("cannot find corresponding agent for connection %d", connID)
// 	}
//
// 	backends := s.Backends[agentID]
// 	if len(backends) == 0 {
// 		return nil, fmt.Errorf("cannot find corresponding agent for connection %d", connID)
// 	}
// 	return backends[0], nil
// }
//
// func (s *ProxyServer) addConnID(agentID string, connID int64) {
// 	s.mu2.Lock()
// 	defer s.mu2.Unlock()
// 	s.connIDToAgentID[connID] = agentID
// }
//
// func (s *ProxyServer) removeConnID(agentID string, connID int64) {
// 	s.mu2.Lock()
// 	defer s.mu2.Unlock()
// 	delete(s.connIDToAgentID, connID)
// }
//
// func (s *ProxyServer) toAgentID(connID int64) (string, bool) {
// 	s.mu2.Lock()
// 	defer s.mu2.Unlock()
// 	agentID, ok := s.connIDToAgentID[connID]
// 	return agentID, ok
// }

// NewProxyServer creates a new ProxyServer instance
func NewProxyServer(serverID string, serverCount int) *ProxyServer {
	return &ProxyServer{
		Frontends:       make(map[int64]*ProxyClientConnection),
		PendingDial:     make(map[int64]*ProxyClientConnection),
		serverID:        serverID,
		serverCount:     serverCount,
		backends:        make(map[string][]agent.AgentService_ConnectServer),
		connIDToAgentID: make(map[int64]string),
	}
}

// Proxy handles incoming streams from gRPC frontend.
func (s *ProxyServer) Proxy(stream agent.ProxyService_ProxyServer) error {
	klog.Info("proxy request from client")

	recvCh := make(chan *agent.Packet, 10)
	stopCh := make(chan error)

	go s.serveRecvFrontend(stream, recvCh)

	defer func() {
		close(recvCh)
	}()

	// Start goroutine to receive packets from frontend and push to recvCh
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(stopCh)
				return
			}
			if err != nil {
				klog.Warningf(">>> Stream read from frontend error: %v", err)
				close(stopCh)
				return
			}

			recvCh <- in
		}
	}()

	return <-stopCh
}

func (s *ProxyServer) serveRecvFrontend(stream agent.ProxyService_ProxyServer, recvCh <-chan *agent.Packet) {
	klog.Info("start serving frontend stream")

	var firstConnID int64
	// The first packet should a DIAL_REQ, we will randomly get a backend
	// from s.backends then.
	var backend agent.AgentService_ConnectServer
	var err error

	for pkt := range recvCh {
		switch pkt.Type {
		case agent.PacketType_DIAL_REQ:
			klog.Info(">>> Received DIAL_REQ")
			// TODO: if we track what agent has historically served
			// the address, then we can use the same agent to send
			// the new DialRequest, to save the agent from creating
			// a connection to the address.
			backend, err = s.randomBackend()
			if err != nil {
				klog.Errorf(">>> failed to get a backend: %v", err)
				continue
			}
			klog.Infof(">>> randomly got backend %v", backend)
			if err := backend.Send(pkt); err != nil {
				klog.Warningf(">>> DIAL_REQ to Backend failed: %v", err)
			}
			s.PendingDial[pkt.GetDialRequest().Random] = &ProxyClientConnection{
				Mode:      "grpc",
				Grpc:      stream,
				connected: make(chan struct{}),
			}
			klog.Info(">>> DIAL_REQ sent to backend") // got this. but backend didn't receive anything.

		case agent.PacketType_CLOSE_REQ:
			connID := pkt.GetCloseRequest().ConnectID
			klog.Infof(">>> Received CLOSE_REQ(id=%d)", connID)
			if backend == nil {
				klog.Errorf("backend has not been initialized for connID %d. Client should send a Dial Request first.", connID)
				continue
			}
			if err := backend.Send(pkt); err != nil {
				// TODO: retry with other backends connecting to this agent.
				klog.Warningf(">>> CLOSE_REQ to Backend failed: %v", err)
			}
			klog.Info(">>> CLOSE_REQ sent to backend")

		case agent.PacketType_DATA:
			connID := pkt.GetData().ConnectID
			if firstConnID == 0 {
				firstConnID = connID
			} else if firstConnID != connID {
				klog.Warningf(">>> Data(id=%d) doesn't match first connection id %d", firstConnID, connID)
			}
			klog.Infof(">>> Received DATA(id=%d)", connID)

			if backend == nil {
				klog.Errorf("backend has not been initialized for connID %d. Client should send a Dial Request first.", connID)
				continue
			}
			if err := backend.Send(pkt); err != nil {
				// TODO: retry with other backends connecting to this agent.
				klog.Warningf(">>> DATA to Backend failed: %v", err)
			}
			klog.Info(">>> DATA sent to backend")

		default:
			klog.Infof(">>> Ignore %v packet coming from frontend", pkt.Type)
		}
	}

	klog.Infof(">>> Close streaming (id=%d)", firstConnID)

	pkt := &agent.Packet{
		Type: agent.PacketType_CLOSE_REQ,
		Payload: &agent.Packet_CloseRequest{
			CloseRequest: &agent.CloseRequest{
				ConnectID: firstConnID,
			},
		},
	}

	if backend == nil {
		klog.Errorf("backend has not been initialized for connID %d. Client should send a Dial Request first.", firstConnID)
		return
	}
	if err := backend.Send(pkt); err != nil {
		klog.Warningf(">>> CLOSE_REQ to Backend failed: %v", err)
	}
}

func (s *ProxyServer) serveSend(stream agent.ProxyService_ProxyServer, sendCh <-chan *agent.Packet) {
	klog.Info("start serve send ...")
	for pkt := range sendCh {
		err := stream.Send(pkt)
		if err != nil {
			klog.Warningf("stream write error: %v", err)
		}
	}
}

func agentID(stream agent.AgentService_ConnectServer) (string, error) {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return "", fmt.Errorf("failed to get context")
	}
	agentIDs := md.Get("agentID")
	if len(agentIDs) != 1 {
		return "", fmt.Errorf("expected one agent ID in the context, got %v", agentIDs)
	}
	return agentIDs[0], nil
}

// Connect is for agent to connect to ProxyServer as next hop
func (s *ProxyServer) Connect(stream agent.AgentService_ConnectServer) error {
	klog.Info("connect request from Backend")
	agentID, err := agentID(stream)
	if err != nil {
		return err
	}
	s.addBackend(agentID, stream)
	defer s.removeBackend(agentID, stream)

	header := metadata.Pairs("serverID", s.serverID, "serverCount", strconv.Itoa(s.serverCount))
	if err := stream.SendHeader(header); err != nil {
		return err
	}

	recvCh := make(chan *agent.Packet, 10)
	stopCh := make(chan error)

	go s.serveRecvBackend(agentID, stream, recvCh)

	defer func() {
		close(recvCh)
	}()

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(stopCh)
				return
			}
			if err != nil {
				klog.Warningf("stream read error: %v", err)
				close(stopCh)
				return
			}

			recvCh <- in
		}
	}()

	return <-stopCh
}

// route the packet back to the correct client
func (s *ProxyServer) serveRecvBackend(agentID string, stream agent.AgentService_ConnectServer, recvCh <-chan *agent.Packet) {
	var firstConnID int64

	for pkt := range recvCh {
		switch pkt.Type {
		case agent.PacketType_DIAL_RSP:
			resp := pkt.GetDialResponse()
			firstConnID = resp.ConnectID
			klog.Infof("<<< Received DIAL_RSP(rand=%d, id=%d)", resp.Random, resp.ConnectID)
			// s.addConnID(agentID, firstConnID)

			if client, ok := s.PendingDial[resp.Random]; !ok {
				klog.Warning("<<< DialResp not recognized; dropped")
			} else {
				err := client.send(pkt)
				delete(s.PendingDial, resp.Random)
				if err != nil {
					klog.Warningf("<<< DIAL_RSP send to client stream error: %v", err)
				} else {
					client.connectID = resp.ConnectID
					s.Frontends[resp.ConnectID] = client
					close(client.connected)
				}
			}

		case agent.PacketType_DATA:
			resp := pkt.GetData()
			klog.Infof("<<< Received DATA(id=%d)", resp.ConnectID)
			if client, ok := s.Frontends[resp.ConnectID]; ok {
				if err := client.send(pkt); err != nil {
					klog.Warningf("<<< DATA send to client stream error: %v", err)
				} else {
					klog.Infof("<<< DATA sent to frontend")
				}
			}

		case agent.PacketType_CLOSE_RSP:
			resp := pkt.GetCloseResponse()
			klog.Infof("<<< Received CLOSE_RSP(id=%d)", resp.ConnectID)
			if client, ok := s.Frontends[resp.ConnectID]; ok {
				if err := client.send(pkt); err != nil {
					// Normal when frontend closes it.
					klog.Warningf("<<< CLOSE_RSP send to client stream error: %v", err)
				} else {
					klog.Infof("<<< CLOSE_RSP sent to frontend")
				}
			}
			// s.removeConnID(agentID, resp.ConnectID)

		default:
			klog.Warningf("<<< Unrecognized packet %+v", pkt)
		}
	}

	klog.Infof("<<< Close streaming (id=%d)", firstConnID)
}
