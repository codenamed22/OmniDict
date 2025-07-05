# OmniDict

OmniDict is a **distributed key-value store** built to demonstrate core distributed systems concepts using **Raft consensus**, **Gossip-based cluster management**, and **gRPC-powered communication**. It supports a range of key operations like replication, leader election, TTL-based expiry, and real-time fault-tolerant transactions.

---

## 📌 Features

- ✅ Raft-based consensus for strong consistency
- 🔁 Gossip protocol for decentralized cluster membership and heartbeat checks
- ⚙️ gRPC-powered client-server communication
- 🔐 TTL (Time to Live), key expiry, and update operations
- 💥 Cluster transactions and fault tolerance
- 📦 Easily extensible and modular codebase

---

## 📁 Project Structure

```
OmniDict/
├── .vscode/             # VS Code workspace settings
├── client/              # Client command implementations
├── cluster/             # Cluster management logic
├── cmd/                 # CLI command bindings
├── proto/               # gRPC proto definitions
├── raftstore/           # Raft FSM and state machine
├── ring/                # Gossip ring and node discovery
├── server/              # gRPC server logic
├── store/               # Internal key-value store logic
├── .gitignore
├── design.md            # Design documentation
├── docker-compose.yml
├── DockerFile
├── go.mod / go.sum      # Go dependencies
├── main.go              # Entry point
├── omnidict             # Compiled binary
└── README.md
```

---

## ⚙️ Installation

### 1. Generate Go Proto Files

```bash
protoc --go_out=./proto --go-grpc_out=./proto proto/kv.proto
protoc --go_out=./proto --go-grpc_out=./proto proto/ring.proto
```

### 2. Build the Binary

```bash
go build -o omnidict main.go
```

---

## 🚀 Usage

### 1. Run Multiple Servers (Simulate Cluster)

Open **separate terminals**, and run:

```bash
./omnidict server
```

### 2. Run Client Commands

```bash
./omnidict put key value
./omnidict get key
./omnidict update key value
./omnidict delete key
./omnidict exists key
./omnidict ttl key
./omnidict expire key <time-in-seconds>
./omnidict flush
./omnidict keys
./omnidict <command> -h or --help
```

💡 If `./omnidict` doesn’t work, use:
```bash
go run main.go <command> <args>
```

---

## 🧪 Test Instructions

- Build using the `go build` step above.
- Launch multiple servers to test cluster behavior.
- Use commands in parallel to observe Raft-based leader election and replication.
- Gossip-based health checks and node joins are logged in the terminal.
- All commands support basic validation. Extendable CLI with Cobra is in place.

---

## 👨‍💻 Contributors

- [**Atique Ahmad**](https://github.com/codenamed22) — Core engineer, Raft + Gossip design  
- [**Jiya**](https://github.com/jiya-username) — gRPC + client integration  
- [**Ronit**](https://github.com/ronit-username) — Proto design + testing automation  

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

## 🤝 Contribution Guidelines

We welcome pull requests, feature ideas, and issue reports. Please make sure to follow the code style and include meaningful commit messages.

---

## 📬 Contact

For any issues, ideas, or bugs, feel free to open an [Issue](https://github.com/codenamed22/OmniDict/issues) or reach out to contributors directly.