package tests

import (
	"net/http/httptest"
	"testing"
)

func TestBasicHAProxyServer_GRPC(t *testing.T) {
	server := httptest.NewServer(newEchoServer("hello"))
	defer server.Close()

	stopCh := make(chan struct{})
	defer close(stopCh)

	proxy, cleanup, err := runGRPCProxyServer()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if err := runAgent(proxy.agent, stopCh); err != nil {
		t.Fatal(err)
	}
}
