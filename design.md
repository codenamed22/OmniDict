design.md – Core Concepts & Notes


---

What is Consistent Hashing?

Consistent Hashing is a way to assign keys to nodes using a circular hash space (0 to MAX).

Each node and key is hashed to a position on this circle (called the "ring").

A key is stored in the first node found clockwise from its position.

When a node joins or leaves, only nearby keys need to move — not all.

This minimizes key remapping and supports dynamic scaling.


Simple Example:
Think of the hash ring as a clock. Nodes are placed at numbers on the clock. If a key hashes to position 3, and there’s a node at 4, that node stores the key. If a new node joins at 3.5, it takes some keys from node 4 — but not all.


---

What is a Node and a Shard?

A node is a single instance (like a server) that participates in storing and managing data.

A shard is a logical subset of data. Instead of all nodes storing everything, data is split into shards and distributed.

One node can hold one or more shards.

Sharding helps scale the system by dividing the storage and load.


Child Analogy:
If the full dictionary is split into A–M and N–Z, each part is a shard. A node is like a student holding one or both parts.


---

What is Raft and Why Do We Use It?

Raft is a consensus algorithm used to keep multiple nodes in sync.

It ensures that all nodes holding the same shard agree on updates, even if some fail.

It handles:

Leader election

Log replication (Put operations)

Fault recovery



Why Raft:

Makes the key-value store fault-tolerant.

Ensures strong consistency across replicas of the same shard.



---

How Does gRPC Communication Work?

gRPC is a high-performance communication protocol built on top of Protocol Buffers (protobuf).

It defines the messages and services in a .proto file.

The server implements services like Put, Get, and Join.

The client sends a request to the server using generated Go code from the .proto file.

gRPC handles:

Serialization of data (via protobuf)

Network transport

Function call abstraction (looks like calling a function even though it’s over the network)



Use in this project:

Nodes in the OmniDict cluster use gRPC to talk to each other and to the client.



---

What is CLI and how do CLI commands works?

CLI (Command-Line Interface) is a text-based interface that allows users to interact with a program by typing commands into a terminal/console.

The OmniDict CLI uses Go’s Cobra library to define modular, self-contained commands like put, get, delete, and ttl for managing key-value data. 

Each command maps to a gRPC function, allowing users to interact with the distributed store from the terminal as if calling local functions. 

Cobra also generates intuitive --help pages and supports argument parsing, making the CLI scriptable and easy to extend.

Although it is limited to terminal interaction it allows for clean command organization, cross-platform utility and quick prototyping.

The CLI commands used and their functions are :-
  
  delete   -   Delete a key-value pair
  exists   -   Check if key exists
  expire   -   Set expiration time for a key
  flush    -   Clear all keys
  get      -   Retrieve a value by key
  help     -   Help about any command
  keys     -   List keys matching pattern
  put      -   Store a key-value pair
  ttl      -   Get time to live for a key
  update   -   Update an existing key
