//go:build !js

package zoocrud

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestZooKeeperStore runs the Store contract against a real server. Set
// ZOOKEEPER_SERVERS, for example to localhost:2181, to run it.
func TestZooKeeperStore(t *testing.T) {
	servers := os.Getenv("ZOOKEEPER_SERVERS")
	if servers == "" {
		t.Skip("set ZOOKEEPER_SERVERS to run this test against ZooKeeper")
	}
	store, err := DialZooKeeper(strings.Split(servers, ","), 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	root := fmt.Sprintf("/zoocrud-test-%d", time.Now().UnixNano())
	mustNot(t, store.Create(root, nil))
	defer DeleteRecursive(store, root)
	testStore(t, store, root)
}
