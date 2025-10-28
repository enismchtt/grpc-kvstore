// grpc-kvstore/server/main.go
package main

import (
	"context"
	"fmt"
	pb "grpc-kvstore/proto"
	"sync"
	"log"
	"net"
	"os"
	"strings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"time"
)


/*
type Peer struct {
	addr string
    conn *grpc.ClientConn
    cli  pb.KVStoreServiceClient
}
*/

type gRPCServer struct {
	pb.UnimplementedKVStoreServiceServer
	kvStore map[string]string
	mu sync.Mutex
	peers []pb.KVStoreServiceClient
}




func buildServer( peerPorts []string) (*gRPCServer , error) {

	s := &gRPCServer{ kvStore : make(map[string]string) }  // initializing

	for _ , peer := range(peerPorts) {
		
		conn , err := grpc.NewClient("localhost:"+peer , grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			return nil, fmt.Errorf("connect peer %s: %w", peer, err)
		}

		s.peers = append(s.peers, pb.NewKVStoreServiceClient(conn)) // adding connected peer gRPC services
	}

	return s , nil
}




func (s *gRPCServer) Put (ctx context.Context, in *pb.PutRequest) (*pb.AckPutResponse, error) {
	// handle distubted server replicate with new protobuffer inner call 
	s.mu.Lock()
	key := in.GetKey()
	value := in.GetValue()
	if key == "" || value == "" {
		return &pb.AckPutResponse{} , fmt.Errorf("key or value is empty")
	}

	// local update
	s.kvStore[key] = value

	s.mu.Unlock() // beause it will be used for reading 

	// replicate to others
	for _ , c := range s.peers {
		pctx , cancel := context.WithTimeout(ctx , 2*time.Second) // If peer server doesn't responds in 2 second the request and replication operation will be canceled.Counting started right here
		_, err := c.Replicate(pctx, &pb.PutRequest{Key: key, Value: value})
		cancel() // finish earcly  right after replication done 

		if err != nil {
        return &pb.AckPutResponse{ Response:" Peer Replication Failed."} , err
    	}
	} 
	
	return &pb.AckPutResponse{Response :"Stored and replicated."} , nil

}


func (s *gRPCServer) Replicate(ctx context.Context, in *pb.PutRequest) (*pb.AckPutResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, value := in.GetKey(), in.GetValue()

	s.kvStore[key] = value
	return &pb.AckPutResponse{Response: "replicated succesfully"}, nil
}





func (s *gRPCServer) Get(ctx context.Context,in *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := in.GetKey()
	value := s.kvStore[key]

	return &pb.GetResponse{Value : value} , nil
}


func (s *gRPCServer) List(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.KeyValueResponse]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

    for k, v := range s.kvStore {
    if err := stream.Send(&pb.KeyValueResponse{Key: k, Value: v}); err != nil {
        return err
    }
	}
	return nil
}



func main() {

	args := os.Args
	main_port := args[1]
	peerPorts := strings.Split(args[2],",")
	
	// create listener on main port
	lis , err := net.Listen("tcp",fmt.Sprintf(":%s",main_port))
	if err!=nil {
		log.Fatalf("Error on listening.")
	}

	// create gRPC server
	s := grpc.NewServer()

	// build server with connection to peers 
	server , err := buildServer(peerPorts)

	if err != nil {
		log.Fatalf("Error on building server : %v", err)
	}

	pb.RegisterKVStoreServiceServer(s , server)
	log.Printf("Server listening on port %v",lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}


}
