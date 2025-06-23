// Remove when pushing to main
/*
	go run ring.go test_ring.go - test krlo
*/
package main

import (
	"fmt"
	// "log"
	"time"
)

func main() {
	fmt.Println("🚀 Starting Hash Ring Tests...")
	
	// Run all test scenarios
	testBasicOperations()
	testReplication()
	testHealthManagement()
	testConsistencyLevels()
	testLoadBalancing()
	testMigration()
	testBulkOperations()
	testEdgeCases()
	demonstrateUsage()

	fmt.Println("✅ All tests completed!")
}

func testBasicOperations() {
	fmt.Println("\n=== Testing Basic Operations ===")
	
	// Create a hash ring
	hr := NewHashRing(150)
	
	// Add nodes
	hr.AddNodeWithAddress("node1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("node2", "192.168.1.2:8080")
	hr.AddNodeWithAddress("node3", "192.168.1.3:8080")
	
	// Test basic set/get operations
	testKeys := []string{"user:1", "user:2", "user:3", "product:100", "order:500"}
	
	fmt.Println("Storing test data...")
	for i, key := range testKeys {
		value := fmt.Sprintf("value_%d", i+1)
		hr.Set(key, value)
	}
	
	fmt.Println("Retrieving test data...")
	for _, key := range testKeys {
		if value, found := hr.Get(key); found {
			fmt.Printf("✓ %s = %s\n", key, value)
		} else {
			fmt.Printf("✗ Key %s not found\n", key)
		}
	}
	
	// Test key existence
	fmt.Println("\nTesting key existence...")
	for _, key := range testKeys {
		if hr.Exists(key) {
			fmt.Printf("✓ Key %s exists\n", key)
		} else {
			fmt.Printf("✗ Key %s does not exist\n", key)
		}
	}
	
	hr.PrintRingStatus()
}

func testReplication() {
	fmt.Println("\n=== Testing Replication ===")
	
	// Create hash ring with replication
	hr := NewHashRingWithReplication(100, 3) // 3 replicas
	
	// Add nodes
	hr.AddNodeWithAddress("replica1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("replica2", "192.168.1.2:8080")
	hr.AddNodeWithAddress("replica3", "192.168.1.3:8080")
	hr.AddNodeWithAddress("replica4", "192.168.1.4:8080")
	
	// Test replication
	testKey := "replicated:key"
	testValue := "replicated_value"
	
	fmt.Printf("Storing %s with replication...\n", testKey)
	if err := hr.SetWithReplication(testKey, testValue); err != nil {
		fmt.Printf("✗ Error storing with replication: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully stored with replication\n")
	}
	
	// Check replication info
	repInfo := hr.GetReplicationInfo(testKey)
	fmt.Printf("Primary node: %s\n", repInfo.PrimaryNode)
	fmt.Printf("Replica nodes: %v\n", repInfo.Replicas)
	
	// Test replicated get
	if value, err := hr.GetWithReplication(testKey); err != nil {
		fmt.Printf("✗ Error getting with replication: %v\n", err)
	} else {
		fmt.Printf("✓ Retrieved value: %s\n", value)
	}
	
	hr.PrintRingStatus()
}

func testHealthManagement() {
	fmt.Println("\n=== Testing Health Management ===")
	
	hr := NewHashRingWithReplication(100, 2)
	
	// Add nodes
	hr.AddNodeWithAddress("health1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("health2", "192.168.1.2:8080")
	hr.AddNodeWithAddress("health3", "192.168.1.3:8080")
	
	// Check initial health
	fmt.Println("Initial health status:")
	for _, node := range []string{"health1", "health2", "health3"} {
		fmt.Printf("Node %s healthy: %v\n", node, hr.IsNodeHealthy(node))
	}
	
	// Mark a node as unhealthy
	fmt.Println("\nMarking health2 as unhealthy...")
	hr.MarkNodeUnhealthy("health2")
	
	// Check healthy nodes
	healthyNodes := hr.GetHealthyNodes()
	fmt.Printf("Healthy nodes: %v\n", healthyNodes)
	
	// Test operations with unhealthy node
	testKey := "health:test"
	if err := hr.SetWithReplication(testKey, "test_value"); err != nil {
		fmt.Printf("✗ Error with unhealthy node: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully handled unhealthy node\n")
	}
	
	// Restore health
	fmt.Println("Restoring health2...")
	hr.MarkNodeHealthy("health2")
	fmt.Printf("Health2 is now healthy: %v\n", hr.IsNodeHealthy("health2"))
}

func testConsistencyLevels() {
	fmt.Println("\n=== Testing Consistency Levels ===")
	
	hr := NewHashRingWithReplication(100, 3)
	
	// Add nodes
	hr.AddNodeWithAddress("consist1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("consist2", "192.168.1.2:8080")
	hr.AddNodeWithAddress("consist3", "192.168.1.3:8080")
	
	// Test different consistency levels
	consistencyLevels := []ConsistencyLevel{ONE, QUORUM, ALL}
	levelNames := []string{"ONE", "QUORUM", "ALL"}
	
	for i, level := range consistencyLevels {
		fmt.Printf("\nTesting consistency level: %s\n", levelNames[i])
		hr.SetConsistencyLevel(level)
		
		testKey := fmt.Sprintf("consistency:%s", levelNames[i])
		testValue := fmt.Sprintf("value_%s", levelNames[i])
		
		if err := hr.SetWithReplication(testKey, testValue); err != nil {
			fmt.Printf("✗ Error with %s consistency: %v\n", levelNames[i], err)
		} else {
			fmt.Printf("✓ Successfully stored with %s consistency\n", levelNames[i])
		}
		
		if value, err := hr.GetWithReplication(testKey); err != nil {
			fmt.Printf("✗ Error reading with %s consistency: %v\n", levelNames[i], err)
		} else {
			fmt.Printf("✓ Successfully read with %s consistency: %s\n", levelNames[i], value)
		}
	}
}

func testLoadBalancing() {
	fmt.Println("\n=== Testing Load Balancing ===")
	
	hr := NewHashRing(100)
	
	// Add nodes
	hr.AddNodeWithAddress("load1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("load2", "192.168.1.2:8080")
	hr.AddNodeWithAddress("load3", "192.168.1.3:8080")
	
	// Add some load
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("load:key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		hr.Set(key, value)
	}
	
	// Check load distribution
	fmt.Println("Load distribution:")
	clusterInfo := hr.GetClusterInfo()
	for nodeName, info := range clusterInfo {
		fmt.Printf("Node %s: %d keys\n", nodeName, info.KeyCount)
	}
	
	// Get least loaded node
	leastLoaded := hr.GetLeastLoadedNode()
	fmt.Printf("Least loaded node: %s\n", leastLoaded)
	
	// Get detailed node info
	if nodeInfo, err := hr.GetNodeInfo(leastLoaded); err != nil {
		fmt.Printf("✗ Error getting node info: %v\n", err)
	} else {
		fmt.Printf("Node info for %s: KeyCount=%d, Healthy=%v\n", 
			nodeInfo.Name, nodeInfo.KeyCount, nodeInfo.Healthy)
	}
}

func testMigration() {
	fmt.Println("\n=== Testing Migration ===")
	
	hr := NewHashRing(100)
	
	// Add initial nodes
	hr.AddNodeWithAddress("migrate1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("migrate2", "192.168.1.2:8080")
	
	// Add test data
	testKeys := []string{"migrate:1", "migrate:2", "migrate:3", "migrate:4", "migrate:5"}
	for i, key := range testKeys {
		hr.Set(key, fmt.Sprintf("value_%d", i))
	}
	
	fmt.Println("Before adding new node:")
	hr.PrintRingStatus()
	
	// Add a new node
	fmt.Println("Adding new node for migration...")
	hr.AddNode("migrate3")
	
	// Get keys that need migration
	keysToMigrate := hr.GetKeysToMigrate("migrate3")
	fmt.Printf("Keys to migrate to migrate3: %v\n", keysToMigrate)
	
	// Perform migration
	for _, key := range keysToMigrate {
		// Find current owner
		for _, nodeName := range hr.GetNodeNames() {
			if nodeName == "migrate3" {
				continue
			}
			nodeKeys := hr.GetNodeKeys(nodeName)
			for _, nodeKey := range nodeKeys {
				if nodeKey == key {
					if hr.MigrateKey(key, nodeName, "migrate3") {
						fmt.Printf("✓ Migrated %s from %s to migrate3\n", key, nodeName)
					}
					break
				}
			}
		}
	}
	
	fmt.Println("After migration:")
	hr.PrintRingStatus()
}

func testBulkOperations() {
	fmt.Println("\n=== Testing Bulk Operations ===")
	
	hr := NewHashRing(100)
	
	// Add nodes
	hr.AddNodeWithAddress("bulk1", "192.168.1.1:8080")
	hr.AddNodeWithAddress("bulk2", "192.168.1.2:8080")
	
	// Add bulk data
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("bulk:item_%d", i)
		value := fmt.Sprintf("bulk_value_%d", i)
		hr.Set(key, value)
	}
	
	// Test getting all keys
	allKeys := hr.GetAllKeys()
	fmt.Printf("Total keys in cluster: %d\n", len(allKeys))
	
	// Test prefix search
	prefixKeys := hr.GetKeysWithPrefix("bulk:item_1")
	fmt.Printf("Keys with prefix 'bulk:item_1': %v\n", prefixKeys)
	
	// Test bulk migration
	hr.AddNode("bulk3")
	keysToMove := []string{"bulk:item_1", "bulk:item_2", "bulk:item_3"}
	
	// Find source node for first key
	sourceNode := ""
	for _, nodeName := range hr.GetNodeNames() {
		if nodeName == "bulk3" {
			continue
		}
		nodeKeys := hr.GetNodeKeys(nodeName)
		for _, nodeKey := range nodeKeys {
			if nodeKey == keysToMove[0] {
				sourceNode = nodeName
				break
			}
		}
		if sourceNode != "" {
			break
		}
	}
	
	if sourceNode != "" {
		migrated := hr.BulkMigrateKeys(keysToMove, sourceNode, "bulk3")
		fmt.Printf("Bulk migrated %d keys from %s to bulk3\n", migrated, sourceNode)
	}
}

func testEdgeCases() {
	fmt.Println("\n=== Testing Edge Cases ===")
	
	// Test empty ring
	emptyRing := NewHashRing(100)
	if node := emptyRing.GetNode("test"); node != "" {
		fmt.Printf("✗ Expected empty string for empty ring, got: %s\n", node)
	} else {
		fmt.Printf("✓ Empty ring handled correctly\n")
	}
	
	// Test single node
	singleRing := NewHashRing(100)
	singleRing.AddNode("single")
	singleRing.Set("test", "value")
	if value, found := singleRing.Get("test"); found {
		fmt.Printf("✓ Single node operation successful: %s\n", value)
	} else {
		fmt.Printf("✗ Single node operation failed\n")
	}
	
	// Test node removal
	multiRing := NewHashRing(100)
	multiRing.AddNode("temp1")
	multiRing.AddNode("temp2")
	multiRing.Set("temp:key", "temp:value")
	
	fmt.Println("Before node removal:")
	multiRing.PrintRingStatus()
	
	multiRing.RemoveNode("temp1")
	fmt.Println("After removing temp1:")
	multiRing.PrintRingStatus()
	
	// Test deletion
	hr := NewHashRingWithReplication(100, 2)
	hr.AddNode("del1")
	hr.AddNode("del2")
	hr.SetWithReplication("delete:me", "delete_value")
	
	if err := hr.Delete("delete:me"); err != nil {
		fmt.Printf("✗ Delete operation failed: %v\n", err)
	} else {
		fmt.Printf("✓ Delete operation successful\n")
	}
	
	// Verify deletion
	if hr.Exists("delete:me") {
		fmt.Printf("✗ Key still exists after deletion\n")
	} else {
		fmt.Printf("✓ Key successfully deleted\n")
	}
	
	// Test serialization
	if state, err := hr.SerializeState(); err != nil {
		fmt.Printf("✗ Serialization failed: %v\n", err)
	} else {
		fmt.Printf("✓ Serialization successful, size: %d bytes\n", len(state))
	}
}

// Helper function to demonstrate usage patterns
func demonstrateUsage() {
	fmt.Println("\n=== Usage Demonstration ===")
	
	// Create production-like setup
	hr := NewHashRingWithReplication(150, 3)
	hr.SetConsistencyLevel(QUORUM)
	
	// Add cluster nodes
	nodes := []struct{ name, addr string }{
		{"web-server-1", "10.0.1.10:8080"},
		{"web-server-2", "10.0.1.11:8080"},
		{"web-server-3", "10.0.1.12:8080"},
		{"web-server-4", "10.0.1.13:8080"},
	}
	
	for _, node := range nodes {
		hr.AddNodeWithAddress(node.name, node.addr)
	}
	
	// Simulate real workload
	fmt.Println("Simulating real workload...")
	
	// User sessions
	for i := 1; i <= 5; i++ {
		sessionKey := fmt.Sprintf("session:%d", i)
		sessionData := fmt.Sprintf(`{"user_id":%d,"login_time":"%s"}`, i, time.Now().Format(time.RFC3339))
		hr.SetWithReplication(sessionKey, sessionData)
	}
	
	// Product cache
	for i := 1; i <= 3; i++ {
		productKey := fmt.Sprintf("product:%d", i)
		productData := fmt.Sprintf(`{"id":%d,"name":"Product %d","price":%.2f}`, i, i, float64(i)*10.99)
		hr.SetWithReplication(productKey, productData)
	}
	
	// Show final status
	fmt.Println("Final cluster status:")
	hr.PrintRingStatus()
	
	// Show cluster health
	fmt.Println("Cluster health summary:")
	clusterInfo := hr.GetClusterInfo()
	for nodeName, info := range clusterInfo {
		status := "🟢"
		if !info.Healthy {
			status = "🔴"
		}
		fmt.Printf("%s %s: %d keys, %s\n", status, nodeName, info.KeyCount, info.Address)
	}
}