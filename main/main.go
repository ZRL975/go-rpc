package main

import (
	"context"
	go_rpc "go-rpc"
	"go-rpc/registry"
	"go-rpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	_ = http.Serve(l, nil)
}

func startServer(registryAddr string, wg *sync.WaitGroup, port string, ch chan *go_rpc.Server) {
	l, _ := net.Listen("tcp", port)
	server := go_rpc.NewServer()
	go func() { ch <- server }()
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(registry string) {
	d := xclient.NewRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "call", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {

	log.SetFlags(0)
	registryAddr := "http://localhost:9999/rpc/registry"
	var wg sync.WaitGroup
	var reg_wg sync.WaitGroup
	var foo Foo
	ch := make(chan *go_rpc.Server)
	go startRegistry(&reg_wg)

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg, ":8080", ch)
	go startServer(registryAddr, &wg, ":8081", ch)
	wg.Wait()
	time.Sleep(time.Second)
	go call(registryAddr)
	s1 := <-ch
	s2 := <-ch
	s1.Register(&foo)
	s2.Register(&foo)
	l, _ := net.Listen("tcp", ":9090")
	servers := []*go_rpc.Server{s1, s2}
	go_rpc.HandleHTTP(servers)
	_ = http.Serve(l, nil)
	//broadcast(registryAddr)
}
