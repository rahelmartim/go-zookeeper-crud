// Package zoocrud runs CRUD commands on a ZooKeeper-style tree of znodes.
//
// A znode is a node in the tree. It has a path such as /zoo/lion, a value,
// a version, and children. The package has two stores with the same rules:
// ZooKeeperStore talks to a real ZooKeeper server, and MemoryStore keeps the
// tree in memory, for tests and for the browser demo.
package zoocrud

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNoNode     = errors.New("node does not exist")
	ErrNodeExists = errors.New("node already exists")
	ErrNotEmpty   = errors.New("node has children")
	ErrBadPath    = errors.New("invalid path")
	ErrNoParent   = errors.New("parent node does not exist")
	ErrRootNode   = errors.New("the root node cannot be changed")
)

// Stat is the metadata of a znode.
type Stat struct {
	// Version starts at 0 and grows by 1 on every update.
	Version     int32
	NumChildren int32
	Created     time.Time
	Modified    time.Time
}

// Store is a tree of znodes. All paths are absolute, for example /zoo/lion.
type Store interface {
	Create(path string, value []byte) error
	Get(path string) ([]byte, Stat, error)
	Exists(path string) (bool, error)
	Children(path string) ([]string, error)
	// Set replaces the value of a node and returns its new version.
	Set(path string, value []byte) (Stat, error)
	// Delete removes a node. It fails with ErrNotEmpty when the node has children.
	Delete(path string) error
}

// ValidatePath reports an error unless path is an absolute znode path.
func ValidatePath(path string) error {
	if path == "/" {
		return nil
	}
	if !strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return fmt.Errorf("%w %q: it must start with / and must not end with /", ErrBadPath, path)
	}
	for _, segment := range strings.Split(path[1:], "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("%w %q: it has an empty, . or .. segment", ErrBadPath, path)
		}
	}
	return nil
}

// parentPath returns the parent of a valid path that is not the root.
func parentPath(path string) string {
	i := strings.LastIndex(path, "/")
	if i == 0 {
		return "/"
	}
	return path[:i]
}

// childPath joins a parent path and a child name.
func childPath(parent, name string) string {
	if parent == "/" {
		return "/" + name
	}
	return parent + "/" + name
}

// baseName returns the last segment of a valid path that is not the root.
func baseName(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}
