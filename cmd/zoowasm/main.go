//go:build js && wasm

// Command zoowasm runs the zoocrud commands in the browser, on an in-memory
// tree, for the web demo. It exposes a global "zoo" object to JavaScript:
//
//	zoo.run(line) -> { lines: [...], error: "" }
//	zoo.tree()    -> { name, path, value, version, children: [...] }
//	zoo.commands  -> [{ name, usage, help }]
package main

import (
	"syscall/js"

	"github.com/rahelmartim/go-zookeeper-crud/zoocrud"
)

// maxLineLength keeps a pasted text from filling the browser memory.
const maxLineLength = 4096

func main() {
	store := zoocrud.NewMemoryStore()
	api := js.Global().Get("Object").New()

	api.Set("run", js.FuncOf(func(_ js.Value, args []js.Value) any {
		line := args[0].String()
		if len(line) > maxLineLength {
			return map[string]any{"lines": []any{"Error: the line is too long"}, "error": "too long"}
		}
		result := zoocrud.Execute(store, line)
		lines := make([]any, len(result.Lines))
		for i, l := range result.Lines {
			lines[i] = l
		}
		errText := ""
		if result.Err != nil {
			errText = result.Err.Error()
		}
		return map[string]any{"lines": lines, "error": errText}
	}))

	api.Set("tree", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return node(store, "/", "/")
	}))

	commands := make([]any, len(zoocrud.Commands))
	for i, c := range zoocrud.Commands {
		commands[i] = map[string]any{"name": c.Name, "usage": c.Usage, "help": c.Help}
	}
	api.Set("commands", commands)

	js.Global().Set("zoo", api)
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("zoo-ready"))
	select {}
}

func node(store zoocrud.Store, path, name string) map[string]any {
	value, stat, _ := store.Get(path)
	names, _ := store.Children(path)
	children := make([]any, len(names))
	for i, child := range names {
		childPath := "/" + child
		if path != "/" {
			childPath = path + "/" + child
		}
		children[i] = node(store, childPath, child)
	}
	return map[string]any{
		"name":     name,
		"path":     path,
		"value":    string(value),
		"version":  int(stat.Version),
		"children": children,
	}
}
