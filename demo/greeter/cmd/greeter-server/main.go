package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	var bindAddr string
	var wlapiAddr string

	flag.StringVar(&bindAddr, "bind", "localhost:8080", "host:port of the server")
	flag.StringVar(&wlapiAddr, "workloadapi", "unix:///tmp/spire-agent/public/api.sock", "workload endpoint socket URI, starting with tcp:// or unix:// scheme")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.Println("Starting up...")

	ctx := context.Background()

	// Get credentials from the Workload API
	log.Println("Calling Workload API...")
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	opts := workloadapi.WithClientOptions(workloadapi.WithAddr(wlapiAddr))
	source, err := workloadapi.NewX509Source(timeoutCtx, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()
	creds := grpccredentials.MTLSServerCredentials(source, source, tlsconfig.AuthorizeAny())

	// Listens for requests
	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(grpc.Creds(creds))
	helloworld.RegisterGreeterServer(server, greeter{})

	log.Println("Serving on", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}

}

type greeter struct {
	helloworld.UnimplementedGreeterServer
}

func (greeter) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	clientID := "UNKNOWN-CLIENT"
	if peerID, ok := grpccredentials.PeerIDFromContext(ctx); ok {
		clientID = peerID.String()
	}

	log.Printf("%s has requested that I say hello to %q...", clientID, req.Name)
	return &helloworld.HelloReply{
		Message: fmt.Sprintf("On behalf of %s, hello %s!", clientID, req.Name),
	}, nil
}
