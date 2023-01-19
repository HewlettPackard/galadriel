package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
)

func main() {
	var serverAddr string
	var wlapiAddr string

	flag.StringVar(&serverAddr, "server", "localhost:8080", "host:port of the server")
	flag.StringVar(&wlapiAddr, "workloadapi", "unix:///tmp/spire-agent/public/api.sock", "workload endpoint socket URI, starting with tcp:// or unix:// scheme")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.Println("Starting up...")
	log.Println("Server Address:", serverAddr)

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
	creds := grpccredentials.MTLSClientCredentials(source, source, tlsconfig.AuthorizeAny())

	// Establish gRPC connection to the server
	client, err := grpc.DialContext(timeoutCtx, serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Make calls to the greeter server
	greeterClient := helloworld.NewGreeterClient(client)

	const interval = time.Second * 10
	log.Printf("Issuing requests every %s...", interval)
	for {
		issueRequest(ctx, greeterClient)
		time.Sleep(interval)
	}
}

func issueRequest(ctx context.Context, c helloworld.GreeterClient) {
	p := new(peer.Peer)
	resp, err := c.SayHello(ctx, &helloworld.HelloRequest{
		Name: "SPIFFE community",
	}, grpc.Peer(p))
	if err != nil {
		log.Printf("Failed to say hello: %v", err)
		return
	}

	serverID := "UNKNOWN-SERVER"
	if peerID, ok := grpccredentials.PeerIDFromPeer(p); ok {
		serverID = peerID.String()
	}

	log.Printf("%s said %q", serverID, resp.Message)
}
