package gossip

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	pb_gossip "omnidict/proto/gossip"

	"google.golang.org/grpc"
)

type GossipManager struct {
	nodeID    string
	members   map[string]struct{}
	peers     []string
	lock      sync.RWMutex
	gossipInt time.Duration
}

func NewGossipManager(nodeID string, initialPeers []string) *GossipManager {
	return &GossipManager{
		nodeID:    nodeID,
		members:   make(map[string]struct{}),
		peers:     initialPeers,
		gossipInt: 1 * time.Second,
	}
}

func (g *GossipManager) Start() {
	go g.gossipLoop()
}

func (g *GossipManager) gossipLoop() {
	for {
		time.Sleep(g.gossipInt)
		g.gossipRound()
	}
}

func (g *GossipManager) gossipRound() {
	g.lock.RLock()
	if len(g.peers) == 0 {
		g.lock.RUnlock()
		return
	}
	peer := g.peers[rand.Intn(len(g.peers))]

	members := make([]string, 0, len(g.members))
	for m := range g.members {
		members = append(members, m)
	}
	g.lock.RUnlock()

	conn, err := grpc.Dial(peer, grpc.WithInsecure())
	if err != nil {
		log.Printf("Gossip failed to connect to %s: %v", peer, err)
		return
	}
	defer conn.Close()

	client := pb_gossip.NewGossipServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Gossip(ctx, &pb_gossip.GossipMessage{
		NodeId:    g.nodeID,
		Members:   members,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		log.Printf("Gossip failed to %s: %v", peer, err)
	}
}

func (g *GossipManager) UpdateMembership(members []string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.members = make(map[string]struct{})
	for _, m := range members {
		g.members[m] = struct{}{}
	}
}

func (g *GossipManager) AddPeer(addr string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	// Avoid duplicate peers
	for _, p := range g.peers {
		if p == addr {
			return
		}
	}
	g.peers = append(g.peers, addr)
}

func (g *GossipManager) AddMember(nodeID string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.members[nodeID] = struct{}{}
}

type GossipServer struct {
	pb_gossip.UnimplementedGossipServiceServer
	mgr *GossipManager
}

func NewGossipServer(mgr *GossipManager) *GossipServer {
	return &GossipServer{mgr: mgr}
}

func (s *GossipServer) Gossip(ctx context.Context, msg *pb_gossip.GossipMessage) (*pb_gossip.GossipAck, error) {
	s.mgr.UpdateMembership(msg.Members)
	return &pb_gossip.GossipAck{Success: true}, nil
}
