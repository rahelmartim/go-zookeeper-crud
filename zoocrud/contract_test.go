package zoocrud

import (
	"errors"
	"slices"
	"testing"
)

// testStore checks the rules that every Store must follow. root is an
// existing node under which the test creates its nodes.
func testStore(t *testing.T, store Store, root string) {
	t.Helper()
	zoo := childPath(root, "zoo")
	lion := childPath(zoo, "lion")

	if err := store.Create(lion, []byte("roar")); !errors.Is(err, ErrNoParent) && !errors.Is(err, ErrNoNode) {
		t.Errorf("Create without parent: %v, want ErrNoParent or ErrNoNode", err)
	}
	mustNot(t, store.Create(zoo, nil))
	mustNot(t, store.Create(lion, []byte("roar")))
	if err := store.Create(lion, nil); !errors.Is(err, ErrNodeExists) {
		t.Errorf("Create twice: %v, want ErrNodeExists", err)
	}

	value, stat, err := store.Get(lion)
	mustNot(t, err)
	if string(value) != "roar" || stat.Version != 0 {
		t.Errorf("Get = %q version %d, want \"roar\" version 0", value, stat.Version)
	}

	for want := int32(1); want <= 2; want++ {
		stat, err = store.Set(lion, []byte("ROAR"))
		mustNot(t, err)
		if stat.Version != want {
			t.Errorf("Set version = %d, want %d", stat.Version, want)
		}
	}

	mustNot(t, store.Create(childPath(zoo, "zebra"), nil))
	children, err := store.Children(zoo)
	mustNot(t, err)
	slices.Sort(children)
	if !slices.Equal(children, []string{"lion", "zebra"}) {
		t.Errorf("Children = %v, want [lion zebra]", children)
	}

	if err := store.Delete(zoo); !errors.Is(err, ErrNotEmpty) {
		t.Errorf("Delete with children: %v, want ErrNotEmpty", err)
	}
	count, err := DeleteRecursive(store, zoo)
	mustNot(t, err)
	if count != 3 {
		t.Errorf("DeleteRecursive deleted %d nodes, want 3", count)
	}
	if exists, err := store.Exists(zoo); err != nil || exists {
		t.Errorf("Exists after delete = %v, %v; want false, nil", exists, err)
	}
	if _, _, err := store.Get(lion); !errors.Is(err, ErrNoNode) {
		t.Errorf("Get after delete: %v, want ErrNoNode", err)
	}
}

func mustNot(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemoryStore(t *testing.T) {
	testStore(t, NewMemoryStore(), "/")
}

func TestMemoryStoreRoot(t *testing.T) {
	store := NewMemoryStore()
	if err := store.Delete("/"); !errors.Is(err, ErrRootNode) {
		t.Errorf("Delete(/) = %v, want ErrRootNode", err)
	}
	if err := store.Create("/", nil); !errors.Is(err, ErrNodeExists) {
		t.Errorf("Create(/) = %v, want ErrNodeExists", err)
	}
}

func TestValidatePath(t *testing.T) {
	for _, path := range []string{"/", "/a", "/a/b-c/d_1"} {
		if err := ValidatePath(path); err != nil {
			t.Errorf("ValidatePath(%q) = %v, want nil", path, err)
		}
	}
	for _, path := range []string{"", "a", "/a/", "//a", "/a//b", "/a/./b", "/.."} {
		if err := ValidatePath(path); !errors.Is(err, ErrBadPath) {
			t.Errorf("ValidatePath(%q) = %v, want ErrBadPath", path, err)
		}
	}
}
