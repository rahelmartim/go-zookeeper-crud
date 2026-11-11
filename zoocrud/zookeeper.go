//go:build !js

package zoocrud

import (
	"errors"
	"time"

	"github.com/go-zookeeper/zk"
)

// ZooKeeperStore is a Store that uses a real ZooKeeper server.
type ZooKeeperStore struct {
	conn *zk.Conn
}

var _ Store = (*ZooKeeperStore)(nil)

// DialZooKeeper connects to ZooKeeper servers, for example "localhost:2181",
// and waits until the session is ready.
func DialZooKeeper(servers []string, timeout time.Duration) (*ZooKeeperStore, error) {
	conn, events, err := zk.Connect(servers, timeout, zk.WithLogger(quietLogger{}))
	if err != nil {
		return nil, err
	}
	deadline := time.After(timeout)
	for {
		select {
		case event := <-events:
			if event.State == zk.StateHasSession {
				return &ZooKeeperStore{conn: conn}, nil
			}
		case <-deadline:
			conn.Close()
			return nil, errors.New("timed out while connecting to ZooKeeper")
		}
	}
}

// quietLogger hides the library logs, for example a failed attempt on one
// of several addresses. Errors still reach the caller.
type quietLogger struct{}

func (quietLogger) Printf(string, ...any) {}

// Close ends the ZooKeeper session.
func (s *ZooKeeperStore) Close() {
	s.conn.Close()
}

func (s *ZooKeeperStore) Create(path string, value []byte) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	_, err := s.conn.Create(path, value, 0, zk.WorldACL(zk.PermAll))
	return translate(err)
}

func (s *ZooKeeperStore) Get(path string) ([]byte, Stat, error) {
	if err := ValidatePath(path); err != nil {
		return nil, Stat{}, err
	}
	value, stat, err := s.conn.Get(path)
	return value, convertStat(stat), translate(err)
}

func (s *ZooKeeperStore) Exists(path string) (bool, error) {
	if err := ValidatePath(path); err != nil {
		return false, err
	}
	ok, _, err := s.conn.Exists(path)
	return ok, translate(err)
}

func (s *ZooKeeperStore) Children(path string) ([]string, error) {
	if err := ValidatePath(path); err != nil {
		return nil, err
	}
	children, _, err := s.conn.Children(path)
	return children, translate(err)
}

func (s *ZooKeeperStore) Set(path string, value []byte) (Stat, error) {
	if err := ValidatePath(path); err != nil {
		return Stat{}, err
	}
	// Version -1 updates the node whatever its current version is.
	stat, err := s.conn.Set(path, value, -1)
	return convertStat(stat), translate(err)
}

func (s *ZooKeeperStore) Delete(path string) error {
	if err := ValidatePath(path); err != nil {
		return err
	}
	if path == "/" {
		return ErrRootNode
	}
	return translate(s.conn.Delete(path, -1))
}

func convertStat(stat *zk.Stat) Stat {
	if stat == nil {
		return Stat{}
	}
	return Stat{
		Version:     stat.Version,
		NumChildren: stat.NumChildren,
		Created:     time.UnixMilli(stat.Ctime),
		Modified:    time.UnixMilli(stat.Mtime),
	}
}

// translate maps ZooKeeper errors to the errors of this package.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, zk.ErrNoNode):
		return ErrNoNode
	case errors.Is(err, zk.ErrNodeExists):
		return ErrNodeExists
	case errors.Is(err, zk.ErrNotEmpty):
		return ErrNotEmpty
	default:
		return err
	}
}
