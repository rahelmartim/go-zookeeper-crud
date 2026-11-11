package zoocrud

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Command describes one CLI command.
type Command struct {
	Name  string
	Usage string
	Help  string
}

// Commands lists the commands that Execute accepts.
var Commands = []Command{
	{"CREATE", "CREATE <path> [value]", "Create a node. Its parent must exist."},
	{"READ", "READ <path>", "Show the value and the version of a node."},
	{"EXISTS", "EXISTS <path>", "Tell if a node exists."},
	{"LIST", "LIST <path>", "List the children of a node."},
	{"UPDATE", "UPDATE <path> <value>", "Replace the value of a node."},
	{"DELETE", "DELETE <path>", "Delete a node that has no children."},
	{"DELETEC", "DELETEC <path>", "Delete a node and all of its children."},
	{"HELP", "HELP", "Show this list."},
}

// Result is the output of one command.
type Result struct {
	Lines []string
	// Err is set when the command failed. Lines then contain the message.
	Err error
}

// ErrUsage is returned for an unknown command or wrong arguments.
var ErrUsage = errors.New("incorrect usage")

func ok(lines ...string) Result { return Result{Lines: lines} }

func fail(err error, lines ...string) Result {
	if len(lines) == 0 {
		lines = []string{"Error: " + err.Error()}
	}
	return Result{Lines: lines, Err: err}
}

func usage(name string) Result {
	for _, c := range Commands {
		if c.Name == name {
			return fail(ErrUsage, "Usage: "+c.Usage)
		}
	}
	return fail(ErrUsage, fmt.Sprintf("Unknown command %q. Type HELP to see the commands.", name))
}

// Execute parses one command line and runs it on the store. Command names
// are not case-sensitive. A value is the rest of the line, so it can
// contain spaces.
func Execute(store Store, line string) Result {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Result{}
	}
	name := strings.ToUpper(fields[0])
	args := fields[1:]
	value := ""
	if len(args) >= 2 {
		// Keep the original spacing of the value.
		rest := strings.TrimSpace(line)
		rest = strings.TrimSpace(rest[len(fields[0]):])
		value = strings.TrimSpace(rest[len(args[0]):])
	}

	switch name {
	case "HELP":
		lines := make([]string, 0, len(Commands))
		for _, c := range Commands {
			lines = append(lines, fmt.Sprintf("%-24s %s", c.Usage, c.Help))
		}
		return ok(lines...)

	case "CREATE":
		if len(args) < 1 {
			return usage(name)
		}
		if err := store.Create(args[0], []byte(value)); err != nil {
			return fail(err)
		}
		return ok("Created " + args[0])

	case "READ":
		if len(args) != 1 {
			return usage(name)
		}
		data, stat, err := store.Get(args[0])
		if err != nil {
			return fail(err)
		}
		children := "children"
		if stat.NumChildren == 1 {
			children = "child"
		}
		return ok(fmt.Sprintf("%s (version %d, %d %s)", quote(data), stat.Version, stat.NumChildren, children))

	case "EXISTS":
		if len(args) != 1 {
			return usage(name)
		}
		exists, err := store.Exists(args[0])
		if err != nil {
			return fail(err)
		}
		if exists {
			return ok(args[0] + " exists")
		}
		return ok(args[0] + " does not exist")

	case "LIST":
		if len(args) != 1 {
			return usage(name)
		}
		children, err := store.Children(args[0])
		if err != nil {
			return fail(err)
		}
		if len(children) == 0 {
			return ok(args[0] + " has no children")
		}
		slices.Sort(children)
		return ok(children...)

	case "UPDATE":
		if len(args) < 2 {
			return usage(name)
		}
		stat, err := store.Set(args[0], []byte(value))
		if err != nil {
			return fail(err)
		}
		return ok(fmt.Sprintf("Updated %s to version %d", args[0], stat.Version))

	case "DELETE":
		if len(args) != 1 {
			return usage(name)
		}
		if err := store.Delete(args[0]); err != nil {
			if errors.Is(err, ErrNotEmpty) {
				return fail(err, "Error: "+args[0]+" has children. Use DELETEC to delete it with its children.")
			}
			return fail(err)
		}
		return ok("Deleted " + args[0])

	case "DELETEC":
		if len(args) != 1 {
			return usage(name)
		}
		count, err := DeleteRecursive(store, args[0])
		if err != nil {
			return fail(err)
		}
		return ok(fmt.Sprintf("Deleted %s and %s", args[0], plural(count-1, "descendant")))

	default:
		return usage(name)
	}
}

// DeleteRecursive deletes a node and all of its descendants, deepest first.
// It returns the number of deleted nodes.
func DeleteRecursive(store Store, path string) (int, error) {
	if err := ValidatePath(path); err != nil {
		return 0, err
	}
	if path == "/" {
		return 0, ErrRootNode
	}
	children, err := store.Children(path)
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, name := range children {
		n, err := DeleteRecursive(store, childPath(path, name))
		deleted += n
		if err != nil {
			return deleted, err
		}
	}
	if err := store.Delete(path); err != nil {
		return deleted, err
	}
	return deleted + 1, nil
}

func quote(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}
	return fmt.Sprintf("%q", data)
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
