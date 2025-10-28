🧩 Project Description

A distributed key-value store implemented in Go using gRPC.
Three server nodes run on different ports and communicate with each other for replication.
When any node receives a Put(key, value) request, it updates its local map and replicates the change to all peers.

🔧 Features

Put(key, value) → Store or update a value across all nodes

Get(key) → Retrieve a value from any node

List() → Stream all key–value pairs stored locally

🚀 How to Run
1. Generate gRPC code (already generated)
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/kvstore.proto

2. Start 3 servers
go run server/main.go 5001 5002,5003
go run server/main.go 5002 5001,5003
go run server/main.go 5003 5001,5002

3. Start a client on any port (e.g., 5001)
go run client/main.go 5001


Run a command:

```bash
> put name surname
Stored and replicated.
```

4. Start another client for a different peer (e.g., 5002)
go run client/main.go 5002


Then check replication:

```bash
> list
name = surname
...
```
