package zoocrud

import (
	"slices"
	"sync"
	"time"
)

type memoryNode struct {
	value    []byte
	stat     Stat
	children map[string]struct{}
}

// MemoryStore is a Store that keeps the tree in memory. It follows the
// ZooKeeper rules that this package uses: the parent of a new node must
// exist, every update increments the version, and a node with children
// cannot be deleted.
type MemoryStore struct {
	mu    sync.RWMutex
	nodes map[string]*memoryNode
	now   func() time.Time
}

var _ Store = (*MemoryStore)(nil)

// NewMemoryStore returns a store that contains only the root node "/".
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{nodes: map[string]*memoryNode{}, now: time.Now}
	now := s.now()
	s.nodes["/"] = &memoryNode{stat: Stat{Created: now, Modified: now}, children: map[string]struct{}{}}
	return s
}

func (s *MemoryStore) Create(path string, value []byte) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	if path == "/" {
		return ErrNodeExists
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[path]; ok {
		return ErrNodeExists
	}
	parent, ok := s.nodes[parentPath(path)]
	if !ok {
		return ErrNoParent
	}
	now := s.now()
	s.nodes[path] = &memoryNode{
		value:    slices.Clone(value),
		stat:     Stat{Created: now, Modified: now},
		children: map[string]struct{}{},
	}
	parent.children[baseName(path)] = struct{}{}
	parent.stat.NumChildren++
	return nil
}

func (s *MemoryStore) Get(path string) ([]byte, Stat, error) {
	if err := ValidatePath(path); err != nil {
		return nil, Stat{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, ok := s.nodes[path]
	if !ok {
		return nil, Stat{}, ErrNoNode
	}
	return slices.Clone(node.value), node.stat, nil
}

func (s *MemoryStore) Exists(path string) (bool, error) {
	if err := ValidatePath(path); err != nil {
		return false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.nodes[path]
	return ok, nil
}

func (s *MemoryStore) Children(path string) ([]string, error) {
	if err := ValidatePath(path); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, ok := s.nodes[path]
	if !ok {
		return nil, ErrNoNode
	}
	names := make([]string, 0, len(node.children))
	for name := range node.children {
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}

func (s *MemoryStore) Set(path string, value []byte) (Stat, error) {
	if err := ValidatePath(path); err != nil {
		return Stat{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[path]
	if !ok {
		return Stat{}, ErrNoNode
	}
	node.value = slices.Clone(value)
	node.stat.Version++
	node.stat.Modified = s.now()
	return node.stat, nil
}

func (s *MemoryStore) Delete(path string) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	if path == "/" {
		return ErrRootNode
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[path]
	if !ok {
		return ErrNoNode
	}
	if len(node.children) > 0 {
		return ErrNotEmpty
	}
	delete(s.nodes, path)
	parent := s.nodes[parentPath(path)]
	delete(parent.children, baseName(path))
	parent.stat.NumChildren--
	return nil
}
