// run this code to check it eveything is working
// go run test_proto.go
// it should return in the terminal - created request : key:"testKey"  value:"testValue"

package main

import (
	"fmt"
	pb "omnidict/proto"
)

func main() {
	req := &pb.PutRequest{
		Key:   "testKey",
		Value: "testValue",
	}
	fmt.Printf("created request : %v\n", req)
}
