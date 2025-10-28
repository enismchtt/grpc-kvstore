// grpc-kvstore/client/main.go
package main

import (
	"bufio"
	"context"
	"fmt"
	pb "grpc-kvstore/proto"
	"log"
	"os"
	"strings"
	"time"
	"io"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)


func main() {

	// setting up connection to server on specified
	connection_port := os.Args[1]
	conn , err := grpc.NewClient(fmt.Sprintf("localhost:%s",connection_port) , grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ERROR: Client couldn't connect : %v", err)
	}
	defer conn.Close()

	c := pb.NewKVStoreServiceClient(conn)


	// CLI session 
	sc := bufio.NewScanner(os.Stdin)

	fmt.Println("hi")
	fmt.Println("Commands: put/get/list/exit") 

	for  {
		fmt.Print(">")

		if !sc.Scan() {
			return
		}

		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]
		
		switch command{
		case "exit" :
			return
		
		case "put" :
			if len(parts) != 3{
				fmt.Println("Usage: put <key> <value>")
			}else{
				key , value := parts[1] , parts[2]

				ctx, cancel := context.WithTimeout(context.Background(), 4* time.Second)
				defer cancel()

				r, err := c.Put(ctx , &pb.PutRequest{ Key: key , Value : value})
				if err != nil {
					log.Fatalf("Put request failed : %v", err)
				}
				log.Printf("%s",r.GetResponse())
			}

		case "get" :
			if len(parts) != 2{
				fmt.Println("Usage: get <key>")
			}else{
				key := parts[1]
				ctx, cancel := context.WithTimeout(context.Background(), 4* time.Second)
				defer cancel()
				
				r, _ := c.Get(ctx , &pb.GetRequest{ Key: key})  // if key doesn't exists it will print empty string
				
				log.Printf("%s = %s",key,r.GetValue())

			}
		case "list" :
			ctx, cancel := context.WithTimeout(context.Background(), 4* time.Second)
			defer cancel()
				
			stream, err := c.List(ctx , &emptypb.Empty{}) // how should I pass empty request ?
			if err != nil {
				log.Fatalf("List request failed : %v", err)
			}
			for {
				msg, err := stream.Recv()
				if err == io.EOF {
					// server finished sending
					break
				}
				if err != nil {
					// if the 4s deadline hit: errors.Is(err, context.DeadlineExceeded)
					log.Fatalf("List() recv failed: %v", err)
				}

				log.Printf("%s = %s", msg.GetKey(), msg.GetValue())
			}
		}

	}


}