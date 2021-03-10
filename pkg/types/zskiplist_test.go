package types

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestZSipList(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var ten [10][]int
	for range ten {
		fmt.Println(rand.Intn(65535))
	}

}
