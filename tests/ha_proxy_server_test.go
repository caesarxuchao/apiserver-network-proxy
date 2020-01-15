package tests

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"sigs.k8s.io/apiserver-network-proxy/pkg/agent/client"
)

type backend struct {
	proxy *httputil.ReverseProxy
}
type haServer struct {
	backends []backend
}

func (ha *haServer) loadbalance(w http.ResponseWriter, r *http.Request) {
	// i := rand.Intn(len(ha.backends))
	i := 0
	b := ha.backends[i]
	fmt.Printf("CHAO: headers are %v\n", r.Header)
	b.proxy.ServeHTTP(w, r)
}

func (ha *haServer) serve() {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", 8000),
		Handler: http.HandlerFunc(ha.loadbalance),
	}
	server.ListenAndServe()
}

func proxyWithHttp2Transport(url *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			fmt.Println("CHAO: http2 transport called")
			ta, err := net.ResolveTCPAddr(network, addr)
			if err != nil {
				return nil, err
			}
			return net.DialTCP(network, nil, ta)
		},
	}
	return proxy
}

func TestBasicHAProxyServer_GRPC(t *testing.T) {
	server := httptest.NewServer(newEchoServer("hello"))
	defer server.Close()

	stopCh := make(chan struct{})
	defer close(stopCh)

	proxy1, cleanup1, err := runGRPCProxyServer()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup1()
	url1, err := url.Parse("http://" + proxy1.agent)
	if err != nil {
		t.Fatal(err)
	}

	// proxy2, cleanup2, err := runGRPCProxyServer()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer cleanup2()
	// url2, err := url.Parse("http://" + proxy2.agent)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// proxy3, cleanup3, err := runGRPCProxyServer()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer cleanup3()
	// url3, err := url.Parse("http://" + proxy3.agent)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	ha := haServer{
		backends: []backend{
			{
				proxy: proxyWithHttp2Transport(url1),
			},
			// 		{
			// 			proxy: proxyWithHttp2Transport(url2),
			// 		},
			// 		{
			// 			proxy: proxyWithHttp2Transport(url3),
			// 		},
		},
	}
	go ha.serve()

	if err := runAgent(":8000", stopCh); err != nil {
		t.Fatal(err)
	}

	// TODO: Wait for agent to register on proxy server
	time.Sleep(5 * time.Second)
	return

	// run test client
	tunnel, err := client.CreateGrpcTunnel(proxy1.front, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	c := &http.Client{
		Transport: &http.Transport{
			Dial: tunnel.Dial,
		},
	}

	r, err := c.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != "hello" {
		t.Errorf("expect %v; got %v", "hello", string(data))
	}
}

func TestBasicNonHAProxyServer_GRPC(t *testing.T) {
	server := httptest.NewServer(newEchoServer("hello"))
	defer server.Close()

	stopCh := make(chan struct{})
	defer close(stopCh)

	proxy1, cleanup1, err := runGRPCProxyServer()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup1()
	url1, err := url.Parse("http://" + proxy1.agent)
	if err != nil {
		t.Fatal(err)
	}

	// proxy2, cleanup2, err := runGRPCProxyServer()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer cleanup2()
	// url2, err := url.Parse("http://" + proxy2.agent)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// proxy3, cleanup3, err := runGRPCProxyServer()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer cleanup3()
	// url3, err := url.Parse("http://" + proxy3.agent)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	ha := haServer{
		backends: []backend{
			{
				proxy: proxyWithHttp2Transport(url1),
			},
			// 		{
			// 			proxy: proxyWithHttp2Transport(url2),
			// 		},
			// 		{
			// 			proxy: proxyWithHttp2Transport(url3),
			// 		},
		},
	}
	go ha.serve()

	if err := runAgent(proxy1.agent, stopCh); err != nil {
		t.Fatal(err)
	}

	// TODO: Wait for agent to register on proxy server
	time.Sleep(5 * time.Second)
	return

	// run test client
	tunnel, err := client.CreateGrpcTunnel(proxy1.front, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	c := &http.Client{
		Transport: &http.Transport{
			Dial: tunnel.Dial,
		},
	}

	r, err := c.Get(server.URL)
	if err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != "hello" {
		t.Errorf("expect %v; got %v", "hello", string(data))
	}
}
