# OmniDict


## Fault-Tolerant Distributed Key-Value Store

**Background:**
Modern applications demand low-latency, highly available key–value storage. You’ll build a simplified “mini-Redis Cluster” that shards data and tolerates node failures.

**Your Mission:**
Implement a **sharded & replicated in-memory K/V system** that

1. **Distributes** keys evenly across N nodes via consistent hashing (≤ 25 % key movement on rebalance).
2. **Replicates** each shard with Raft to guarantee strong consistency and leader election.
3. **Survives** node crashes: as long as a quorum lives, reads and writes succeed.

**MVP Requirements:**

* A library or service exposing gRPC (or HTTP) APIs:

  ```protobuf
  rpc Put(PutRequest) returns (PutResponse);
  rpc Get(GetRequest) returns (GetResponse);
  rpc Join(NodeInfo) returns (JoinAck);
  ```
* A consistent-hash ring that assigns keys → shards, handling new/joining nodes with minimal remapping.
* A per-shard Raft group (using etcd/HashiCorp Raft) for log-replication and leader-based writes.
* A test CLI script that spins up 3–5 nodes, performs puts/gets, kills a node, and verifies no data loss.

**Evaluation Criteria:**

* **Correctness under failure:** no lost writes, linearizable reads.
* **Rebalance behavior:** observe ≤ 25 % key movement when adding/removing nodes.
* **Code clarity:** modular separation between ring logic, Raft integration, and API layer.
* **Stretch work (bonus):** gossip-based membership, client-side caching with invalidation, Prometheus metrics.
