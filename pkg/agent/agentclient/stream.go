package agentclient

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-network-proxy/proto/agent"
)

const (
	defaultInterval = 5 * time.Second
)

type ReconnectError struct {
	internalErr error
	errChan     <-chan error
}

func (e *ReconnectError) Error() string {
	return "transient error: " + e.internalErr.Error()
}

func (e *ReconnectError) Wait() error {
	return <-e.errChan
}

type RedialableAgentClient struct {
	cs *ClientSet // the clientset that includes this RedialableAgentClient.

	stream agent.AgentService_ConnectClient

	agentID     string
	serverID    string
	serverCount int

	// connect opts
	address       string
	opts          []grpc.DialOption
	conn          *grpc.ClientConn
	stopCh        chan struct{}
	reconnOngoing bool
	reconnWaiters []chan error

	// locks
	sendLock   sync.Mutex
	recvLock   sync.Mutex
	reconnLock sync.Mutex

	// Interval between every reconnect
	Interval time.Duration
}

func copyRedialableAgentClient(in RedialableAgentClient) RedialableAgentClient {
	out := in
	out.stopCh = make(chan struct{})
	out.reconnOngoing = false
	out.reconnWaiters = nil
	out.sendLock = sync.Mutex{}
	out.recvLock = sync.Mutex{}
	out.reconnLock = sync.Mutex{}
	return out
}

func NewRedialableAgentClient(address, agentID string, cs *ClientSet, opts ...grpc.DialOption) (*RedialableAgentClient, error) {
	c := &RedialableAgentClient{
		cs:       cs,
		address:  address,
		agentID:  agentID,
		opts:     opts,
		Interval: defaultInterval,
		stopCh:   make(chan struct{}),
	}
	serverID, err := c.Connect()
	if err != nil {
		return nil, err
	}
	c.serverID = serverID
	return c, nil
}

func (c *RedialableAgentClient) probe() {
	for {
		select {
		case <-c.stopCh:
			return
		case <-time.After(c.Interval):
			// health check
			if c.conn != nil && c.conn.GetState() == connectivity.Ready {
				continue
			} else {
				klog.Infof("Connection state %v", c.conn.GetState())
			}
		}

		klog.Info("probe failure: reconnect")
		if err := <-c.triggerReconnect(); err != nil {
			klog.Infof("probe reconnect failed: %v", err)
		}
	}
}

func (c *RedialableAgentClient) Send(pkt *agent.Packet) error {
	c.sendLock.Lock()
	defer c.sendLock.Unlock()

	if err := c.stream.Send(pkt); err != nil {
		if err == io.EOF {
			return err
		}
		return &ReconnectError{
			internalErr: err,
			errChan:     c.triggerReconnect(),
		}
	}

	return nil
}

func (c *RedialableAgentClient) RetrySend(pkt *agent.Packet) error {
	err := c.Send(pkt)
	if err == nil {
		return nil
	} else if err == io.EOF {
		return err
	}

	if err2, ok := err.(*ReconnectError); ok {
		err = err2.Wait()
	}
	if err != nil {
		return err
	}
	return c.RetrySend(pkt)
}

func (c *RedialableAgentClient) triggerReconnect() <-chan error {
	c.reconnLock.Lock()
	defer c.reconnLock.Unlock()

	errch := make(chan error)
	c.reconnWaiters = append(c.reconnWaiters, errch)

	if !c.reconnOngoing {
		go c.reconnect()
		c.reconnOngoing = true
	}

	return errch
}

func (c *RedialableAgentClient) doneReconnect(err error) {
	c.reconnLock.Lock()
	defer c.reconnLock.Unlock()

	for _, ch := range c.reconnWaiters {
		ch <- err
	}
	c.reconnOngoing = false
	c.reconnWaiters = nil
}

func (c *RedialableAgentClient) Recv() (*agent.Packet, error) {
	c.recvLock.Lock()
	defer c.recvLock.Unlock()

	var pkt *agent.Packet
	var err error

	if pkt, err = c.stream.Recv(); err != nil {
		if err == io.EOF {
			return pkt, err
		}
		return pkt, &ReconnectError{
			internalErr: err,
			errChan:     c.triggerReconnect(),
		}
	}

	return pkt, nil
}

// Connect makes the grpc dial to the proxy server. It returns the serverID
// it connects to.
func (c *RedialableAgentClient) Connect() (string, error) {
	var err error
	c.serverID, c.serverCount, c.conn, c.stream, err = c.tryConnect()
	return c.serverID, err
}

// The goal is to make the chance that client's Connect rpc call has never hit
// the wanted server after "retries" times to be lower than 10^-2.
func retryLimit(serverCount int) (retries int) {
	switch serverCount {
	case 1:
		return 3 // to overcome transient errors
	case 2:
		return 3 + 7
	case 3:
		return 3 + 12
	case 4:
		return 3 + 17
	case 5:
		return 3 + 21
	default:
		// we don't expect HA server with more than 5 instances.
		return 3 + 21
	}
}

func (c *RedialableAgentClient) reconnect() {
	klog.Info("start to reconnect...")

	var err error
	var retry, limit int

	limit = retryLimit(c.serverCount)
	for retry < limit {
		serverID, _, err, conn, stream := c.tryConnect()
		switch {
		case err != nil:
			retry++
			klog.Infof("Failed to connect to proxy server, retry %d in %v: %v", retry, c.Interval, err)
			time.Sleep(c.Interval)
		case err == nil && serverID == c.serverID:
			klog.Info("reconnected to %s", serverID)
			c.conn = conn
			c.stream = stream
			c.doneReconnect(nil)
			return

		case err == nil && serverID != c.serverID && c.cs.HasID(serverID):
			// reset the connection
			err := conn.Close()
			if err != nil {
				klog.Infof("failed to close connection to %s: %v", serverID, err)
			}
			retry++
			klog.Infof("Trying to reconnect to proxy server %s, got connected to proxy server %s, for which there is already a connection, retry %d in %v", c.serverID, serverID, retry, c.Interval)
			time.Sleep(c.Interval)
		case err == nil && serverID != c.serverID && !c.cs.HasID(serverID):
			// create a new client
			cc := copyRedialableAgentClient(*c)
			cc.stream = stream
			cc.conn = conn
			ac := &AgentClient{
				connContext: make(map[int64]*connContext),
				stream:      &cc,
				serverID:    cc.serverID,
			}
			err := c.cs.AddClient(serverID, ac)
			if err != nil {
				klog.Infof("failed to add client for %s: %v", serverID, err)
			}
			go ac.Serve()
			retry++
			klog.Infof("Trying to reconnect to proxy server %s, got connected to proxy server %s. We will add this connection to the client set, but keep retrying connecting to proxy server %s, retry %d in %v", c.serverID, serverID, c.serverID, retry, c.Interval)
			time.Sleep(c.Interval)
		}
	}

	c.cs.RemoveClient(c.serverID)
	close(c.stopCh)
	c.doneReconnect(fmt.Errorf("Failed to connect to proxy server: %v", err))
}

func serverCount(stream agent.AgentService_ConnectClient) (int, error) {
	md, err := stream.Header()
	if err != nil {
		return 0, err
	}
	scounts := md.Get("serverCount")
	if len(scounts) == 0 {
		return 0, fmt.Errorf("missing server count")
	}
	scount := scounts[0]
	return strconv.Atoi(scount)
}

func serverID(stream agent.AgentService_ConnectClient) (string, error) {
	md, err := stream.Header()
	if err != nil {
		return "", err
	}
	sids := md.Get("serverID")
	if len(sids) != 1 {
		return "", fmt.Errorf("expected one server ID in the context, got %v", sids)
	}
	return sids[0], nil
}

// tryConnect makes the grpc dial to the proxy server. It returns the serverID
// it connects to, and the number of servers (1 if server is non-HA). It also
// updates c.stream.
func (c *RedialableAgentClient) tryConnect() (string, int, *grpc.ClientConn, agent.AgentService_ConnectClient, error) {
	var err error

	conn, err := grpc.Dial(c.address, c.opts...)
	if err != nil {
		return "", 0, nil, nil, err
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "agentID", c.agentID)
	stream, err := agent.NewAgentServiceClient(conn).Connect(ctx)
	if err != nil {
		return "", 0, nil, nil, err
	}
	sid, err := serverID(stream)
	if err != nil {
		return "", 0, nil, nil, err
	}
	count, err := serverCount(stream)
	if err != nil {
		return "", 0, nil, nil, err
	}
	return sid, count, conn, stream, err
}

func (c *RedialableAgentClient) Close() {
	c.conn.Close()
}
