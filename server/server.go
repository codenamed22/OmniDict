package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"omnidict/gossip"
	pb_kv "omnidict/proto/kv"
	"omnidict/ring"
	"omnidict/store"

	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OmnidictServer struct {
	pb_kv.UnimplementedOmnidictServiceServer
	Ring           *ring.HashRing
	ShardStores    map[string]*store.Store
	ShardRafts     map[string]*raft.Raft
	NodeAddress    string
	GossipManager  *gossip.GossipManager
	txnCoordinator *TransactionCoordinator
}

func NewOmnidictServerWithShards(
	shardStores map[string]*store.Store,
	shardRafts map[string]*raft.Raft,
	nodeAddress string,
	gossipManager *gossip.GossipManager,
) *OmnidictServer {
	hashRing := ring.NewHashRing(3)
	for i := 0; i < 3; i++ {
		shardID := fmt.Sprintf("shard%d", i)
		hashRing.AddNode(shardID)
	}

	return &OmnidictServer{
		Ring:           hashRing,
		ShardStores:    shardStores,
		ShardRafts:     shardRafts,
		NodeAddress:    nodeAddress,
		GossipManager:  gossipManager,
		txnCoordinator: NewTransactionCoordinator(shardRafts, shardStores),
	}
}

type Command struct {
	Op    string                `json:"op"`
	Key   string                `json:"key"`
	Value string                `json:"value,omitempty"`
	TTL   int64                 `json:"ttl,omitempty"`
	TxnID string                `json:"txn_id,omitempty"`
	Ops   []*pb_kv.TxnOperation `json:"ops,omitempty"`
}

func (s *OmnidictServer) getShardResources(key string) (string, *store.Store, *raft.Raft, error) {
	shardID := s.Ring.GetShardID(key)
	store, exists := s.ShardStores[shardID]
	if !exists {
		return "", nil, nil, status.Errorf(codes.NotFound, "shard store %s not found", shardID)
	}
	raftGroup, exists := s.ShardRafts[shardID]
	if !exists {
		return "", nil, nil, status.Errorf(codes.NotFound, "shard raft %s not found", shardID)
	}
	return shardID, store, raftGroup, nil
}

func (s *OmnidictServer) forwardToShardLeader(
	ctx context.Context,
	shardID string,
	req interface{},
) (interface{}, error) {
	shardRaft, exists := s.ShardRafts[shardID]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "shard %s not found", shardID)
	}

	leaderAddr := string(shardRaft.Leader())
	if leaderAddr == "" {
		return nil, status.Errorf(codes.Unavailable, "no leader available for shard %s", shardID)
	}

	// Convert Raft address to gRPC address
	parts := strings.Split(leaderAddr, ":")
	if len(parts) < 2 {
		return nil, status.Errorf(codes.Internal, "invalid leader address: %s", leaderAddr)
	}
	grpcPort := "8080" // Default gRPC port
	if len(parts) > 1 {
		raftPort := parts[1]
		if portNum, err := strconv.Atoi(raftPort); err == nil {
			grpcPort = strconv.Itoa(portNum - 1) // Infer gRPC port from Raft port
		}
	}
	grpcAddr := net.JoinHostPort(parts[0], grpcPort)

	// Avoid self-forwarding
	if grpcAddr == s.NodeAddress {
		return nil, status.Error(codes.Internal, "detected self-forwarding loop")
	}

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "failed to connect to leader %s: %v", grpcAddr, err)
	}
	defer conn.Close()

	client := pb_kv.NewOmnidictServiceClient(conn)

	switch r := req.(type) {
	case *pb_kv.PutRequest:
		return client.Put(ctx, r)
	case *pb_kv.GetRequest:
		return client.Get(ctx, r)
	case *pb_kv.DeleteRequest:
		return client.Delete(ctx, r)
	case *pb_kv.UpdateRequest:
		return client.Update(ctx, r)
	case *pb_kv.ExpireRequest:
		return client.Expire(ctx, r)
	default:
		return nil, status.Errorf(codes.Unimplemented, "unsupported request type for forwarding")
	}
}

func (s *OmnidictServer) Put(ctx context.Context, req *pb_kv.PutRequest) (*pb_kv.PutResponse, error) {
	shardID, _, raftGroup, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}

	if raftGroup.State() != raft.Leader {
		res, err := s.forwardToShardLeader(ctx, shardID, req)
		if err != nil {
			return nil, err
		}
		return res.(*pb_kv.PutResponse), nil
	}

	cmd := Command{Op: "put", Key: req.Key, Value: req.Value}
	data, _ := json.Marshal(cmd)
	f := raftGroup.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		return &pb_kv.PutResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.PutResponse{Success: true}, nil
}

func (s *OmnidictServer) Delete(ctx context.Context, req *pb_kv.DeleteRequest) (*pb_kv.DeleteResponse, error) {
	shardID, _, raftGroup, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}

	if raftGroup.State() != raft.Leader {
		res, err := s.forwardToShardLeader(ctx, shardID, req)
		if err != nil {
			return nil, err
		}
		return res.(*pb_kv.DeleteResponse), nil
	}

	cmd := Command{Op: "delete", Key: req.Key}
	data, _ := json.Marshal(cmd)
	f := raftGroup.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		return &pb_kv.DeleteResponse{Success: false}, nil
	}
	return &pb_kv.DeleteResponse{Success: true}, nil
}

func (s *OmnidictServer) Update(ctx context.Context, req *pb_kv.UpdateRequest) (*pb_kv.UpdateResponse, error) {
	shardID, _, raftGroup, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}

	if raftGroup.State() != raft.Leader {
		res, err := s.forwardToShardLeader(ctx, shardID, req)
		if err != nil {
			return nil, err
		}
		return res.(*pb_kv.UpdateResponse), nil
	}

	cmd := Command{Op: "put", Key: req.Key, Value: req.Value}
	data, _ := json.Marshal(cmd)
	f := raftGroup.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		return &pb_kv.UpdateResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.UpdateResponse{Success: true}, nil
}

func (s *OmnidictServer) Expire(ctx context.Context, req *pb_kv.ExpireRequest) (*pb_kv.ExpireResponse, error) {
	shardID, store, raftGroup, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}

	if raftGroup.State() != raft.Leader {
		res, err := s.forwardToShardLeader(ctx, shardID, req)
		if err != nil {
			return nil, err
		}
		return res.(*pb_kv.ExpireResponse), nil
	}

	value, exists := store.Get(req.Key)
	if !exists {
		return &pb_kv.ExpireResponse{Success: false, Message: "key not found"}, nil
	}

	cmd := Command{
		Op:    "put",
		Key:   req.Key,
		Value: value,
		TTL:   req.Ttl,
	}
	data, _ := json.Marshal(cmd)
	f := raftGroup.Apply(data, 5*time.Second)
	if err := f.Error(); err != nil {
		return &pb_kv.ExpireResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb_kv.ExpireResponse{Success: true}, nil
}

func (s *OmnidictServer) Get(ctx context.Context, req *pb_kv.GetRequest) (*pb_kv.GetResponse, error) {
	_, store, raftGroup, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}

	if err := raftGroup.Barrier(5 * time.Second).Error(); err != nil {
		return nil, status.Errorf(codes.Unavailable, "read barrier failed: %v", err)
	}

	value, exists := store.Get(req.Key)
	return &pb_kv.GetResponse{Found: exists, Value: value}, nil
}

func (s *OmnidictServer) Exists(ctx context.Context, req *pb_kv.ExistsRequest) (*pb_kv.ExistsResponse, error) {
	_, store, _, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}
	_, exists := store.Get(req.Key)
	return &pb_kv.ExistsResponse{Exists: exists}, nil
}

func (s *OmnidictServer) Keys(ctx context.Context, req *pb_kv.KeysRequest) (*pb_kv.KeysResponse, error) {
	allKeys := []string{}
	for shardID, store := range s.ShardStores {
		keys := store.GetAllKeys()
		if req.Pattern != "" {
			for _, key := range keys {
				if strings.HasPrefix(key, req.Pattern) {
					allKeys = append(allKeys, key)
				}
			}
		} else {
			allKeys = append(allKeys, keys...)
		}
		log.Printf("Shard %s contributed %d keys", shardID, len(keys))
	}
	return &pb_kv.KeysResponse{Keys: allKeys}, nil
}

func (s *OmnidictServer) TTL(ctx context.Context, req *pb_kv.TTLRequest) (*pb_kv.TTLResponse, error) {
	_, store, _, err := s.getShardResources(req.Key)
	if err != nil {
		return nil, err
	}
	ttl, exists := store.GetTTL(req.Key)
	if !exists {
		return &pb_kv.TTLResponse{Ttl: -2}, nil
	}
	return &pb_kv.TTLResponse{Ttl: int64(ttl.Seconds())}, nil
}

func (s *OmnidictServer) JoinCluster(
	ctx context.Context,
	req *pb_kv.JoinRequest,
) (*pb_kv.JoinResponse, error) {
	success := true
	var errorMsgs []string

	for shardID, shardRaft := range s.ShardRafts {
		// Skip if not leader for this shard
		if shardRaft.State() != raft.Leader {
			errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: not leader", shardID))
			success = false
			continue
		}

		configFuture := shardRaft.AddVoter(
			raft.ServerID(req.NodeId),
			raft.ServerAddress(req.RaftAddress),
			0, 10*time.Second, // Timeout added
		)
		if err := configFuture.Error(); err != nil {
			if err == raft.ErrNotLeader {
				errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: lost leadership during operation", shardID))
			} else {
				errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: %v", shardID, err))
			}
			success = false
		}
	}

	if !success {
		return &pb_kv.JoinResponse{
			Success: false,
			Error:   strings.Join(errorMsgs, "; "),
		}, nil
	}

	// Add new node to gossip
	host, _, err := net.SplitHostPort(req.RaftAddress)
	if err == nil {
		grpcAddr := net.JoinHostPort(host, "8080")
		s.GossipManager.AddPeer(grpcAddr)
	} else {
		log.Printf("Failed to parse raft address: %v", err)
	}

	// Update membership with new node
	s.GossipManager.AddMember(req.NodeId)

	return &pb_kv.JoinResponse{Success: true}, nil
}

func (s *OmnidictServer) RemoveNode(
	ctx context.Context,
	req *pb_kv.RemoveNodeRequest,
) (*pb_kv.RemoveNodeResponse, error) {
	success := true
	var errorMsgs []string

	for shardID, shardRaft := range s.ShardRafts {
		// Skip if not leader for this shard
		if shardRaft.State() != raft.Leader {
			errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: not leader", shardID))
			success = false
			continue
		}

		configFuture := shardRaft.RemoveServer(
			raft.ServerID(req.NodeId),
			0, 10*time.Second,
		)
		if err := configFuture.Error(); err != nil {
			// Handle specific error cases
			if err == raft.ErrNotLeader {
				errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: lost leadership during operation", shardID))
			} else if strings.Contains(err.Error(), "unknown node") {
				errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: node %s not found", shardID, req.NodeId))
			} else {
				errorMsgs = append(errorMsgs, fmt.Sprintf("shard %s: %v", shardID, err))
			}
			success = false
		}
	}

	if !success {
		return &pb_kv.RemoveNodeResponse{
			Success: false,
			Error:   strings.Join(errorMsgs, "; "),
		}, nil
	}
	return &pb_kv.RemoveNodeResponse{Success: true}, nil
}

func (s *OmnidictServer) Flush(ctx context.Context, req *pb_kv.FlushRequest) (*pb_kv.FlushResponse, error) {
	for shardID, store := range s.ShardStores {
		store.Flush()
		log.Printf("Flushed shard %s", shardID)
	}
	return &pb_kv.FlushResponse{Success: true}, nil
}

func (s *OmnidictServer) GetNodeInfo(ctx context.Context, req *pb_kv.NodeInfoRequest) (*pb_kv.NodeInfoResponse, error) {
	shardIDs := make([]string, 0, len(s.ShardRafts))
	for shardID := range s.ShardRafts {
		shardIDs = append(shardIDs, shardID)
	}
	return &pb_kv.NodeInfoResponse{
		NodeId:     s.NodeAddress,
		Address:    s.NodeAddress,
		Status:     "healthy",
		TotalNodes: int32(len(shardIDs)),
		Nodes:      shardIDs,
	}, nil
}

func (s *OmnidictServer) BeginTransaction(ctx context.Context, req *pb_kv.BeginTxnRequest) (*pb_kv.BeginTxnResponse, error) {
	return s.txnCoordinator.BeginTransaction(ctx, req)
}

func (s *OmnidictServer) Prepare(ctx context.Context, req *pb_kv.PrepareRequest) (*pb_kv.PrepareResponse, error) {
	return s.txnCoordinator.Prepare(ctx, req)
}

func (s *OmnidictServer) Commit(ctx context.Context, req *pb_kv.CommitRequest) (*pb_kv.CommitResponse, error) {
	return s.txnCoordinator.Commit(ctx, req)
}

func (s *OmnidictServer) Abort(ctx context.Context, req *pb_kv.AbortRequest) (*pb_kv.AbortResponse, error) {
	return s.txnCoordinator.Abort(ctx, req)
}
