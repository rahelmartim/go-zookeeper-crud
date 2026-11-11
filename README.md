# ZooCRUD

[![CI](https://github.com/rahelmartim/go-zookeeper-crud/actions/workflows/ci.yml/badge.svg)](https://github.com/rahelmartim/go-zookeeper-crud/actions/workflows/ci.yml)
[![Live demo](https://img.shields.io/badge/live%20demo-GitHub%20Pages-34D399)](https://rahelmartim.github.io/go-zookeeper-crud/)
[![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)

An interactive CRUD shell for [Apache ZooKeeper](https://zookeeper.apache.org), written in Go. You create, read, update, and delete znodes with simple commands.

ZooKeeper is a coordination service that stores data in a tree. Each node of the tree is a *znode*. A znode has a path such as `/zoo/lion`, a value, a version, and children.

Live demo: https://rahelmartim.github.io/go-zookeeper-crud/

The live demo runs the same command code as the CLI, compiled to WebAssembly. It uses an in-memory tree that follows the ZooKeeper rules, because a real ZooKeeper server cannot run in the browser. You type commands in a terminal and watch the tree change.

## Commands

Command names are not case-sensitive. A value is the rest of the line, so it can contain spaces.

| Command | Description |
| --- | --- |
| `CREATE <path> [value]` | Creates a node. Its parent must exist. |
| `READ <path>` | Shows the value, the version, and the number of children of a node. |
| `EXISTS <path>` | Tells if a node exists. |
| `LIST <path>` | Lists the children of a node. |
| `UPDATE <path> <value>` | Replaces the value of a node. The version grows by 1. |
| `DELETE <path>` | Deletes a node that has no children. |
| `DELETEC <path>` | Deletes a node and all of its children. |
| `HELP` | Shows the commands. |
| `QUIT` | Ends the session. |

Example session:

```text
zoocrud> CREATE /zoo
Created /zoo
zoocrud> CREATE /zoo/lion king of the jungle
Created /zoo/lion
zoocrud> UPDATE /zoo/lion still the king
Updated /zoo/lion to version 1
zoocrud> DELETE /zoo
Error: /zoo has children. Use DELETEC to delete it with its children.
zoocrud> DELETEC /zoo
Deleted /zoo and 1 descendant
```

## Getting started

You need Go 1.24 or later. To use a real ZooKeeper, you also need [Docker](https://www.docker.com) or a ZooKeeper installation.

1. Start ZooKeeper.

   ```bash
   docker run --rm -p 2181:2181 zookeeper:3.9
   ```

2. In another terminal, start the shell.

   ```bash
   go run ./cmd/zoocrud
   ```

   To connect to other servers, use `-servers`, for example `-servers zk1:2181,zk2:2181`.

To try the shell without ZooKeeper, use the in-memory tree. The data is lost when you quit.

```bash
go run ./cmd/zoocrud -memory
```

## Use the package

The `zoocrud` package has a `Store` interface with two implementations:

- `ZooKeeperStore` uses a real ZooKeeper server through [go-zookeeper/zk](https://github.com/go-zookeeper/zk).
- `MemoryStore` keeps the tree in memory. It is for tests and for the browser demo.

```go
store, err := zoocrud.DialZooKeeper([]string{"localhost:2181"}, 10*time.Second)
if err != nil {
	log.Fatal(err)
}
defer store.Close()

result := zoocrud.Execute(store, "CREATE /zoo/lion roar")
fmt.Println(result.Lines)
```

## Commands for development

| Command | Description |
| --- | --- |
| `make test` | Runs `go vet` and the tests with the race detector. |
| `make test-zookeeper` | Runs the tests, including the tests against ZooKeeper on `localhost:2181`. |
| `make build` | Builds `zoocrud` into `bin/`. |
| `make web` | Builds the browser demo into `site/`. |
| `make serve` | Builds the browser demo and serves it on http://localhost:8080. It needs Python 3. To use another port, run `make serve PORT=8090`. |

## Tests

The same contract tests run on both stores. They check the parent rule, the versions, the children, and the deletes. The tests against ZooKeeper run only when `ZOOKEEPER_SERVERS` is set. CI starts a ZooKeeper container and runs them on every push.

## Project structure

```text
cmd/
├── zoocrud/   # Interactive shell
└── zoowasm/   # WebAssembly entry point for the browser demo
zoocrud/       # Store interface, ZooKeeper and memory stores, command engine, tests
web/           # Browser demo: HTML, CSS, and JavaScript
```

## Deployment

The `Deploy demo to GitHub Pages` workflow builds the demo and publishes it on every push to `master`. The demo is a static site, so it needs no server.

## License

This project uses the MIT license. Read the [LICENSE](LICENSE) file for the full text.
