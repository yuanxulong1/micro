package handler

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
	rmock "github.com/micro/go-micro/registry/mock"
)

func TestProxyHandler(t *testing.T) {
	r := rmock.NewRegistry()

	cmd.DefaultCmd = cmd.NewCmd(cmd.Registry(&r))

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	parts := strings.Split(l.Addr().String(), ":")

	var host string
	var port int

	host = parts[0]
	port, _ = strconv.Atoi(parts[1])

	s := &registry.Service{
		Name: "go.micro.api.test",
		Nodes: []*registry.Node{
			&registry.Node{
				Id:      "1",
				Address: host,
				Port:    port,
			},
		},
	}

	r.Register(s)
	defer r.Deregister(s)

	// setup the test handler
	m := http.NewServeMux()
	m.HandleFunc("/test/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`you got served`))
	})

	// start http test serve
	go http.Serve(l, m)

	// create new request and writer
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/test/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	// initialise the handler
	p := Proxy("go.micro.api", false)

	// execute the handler
	p.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Expected 200 response got %d %s", w.Code, w.Body.String())
	}

	if w.Body.String() != "you got served" {
		t.Fatal("Expected body: you got served. Got: %s", w.Body.String())
	}
}
