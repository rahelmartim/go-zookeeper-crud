// Command zoocrud is an interactive CRUD shell for ZooKeeper.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rahelmartim/go-zookeeper-crud/zoocrud"
)

func main() {
	servers := flag.String("servers", "localhost:2181", "comma-separated list of ZooKeeper servers")
	memory := flag.Bool("memory", false, "use an in-memory tree instead of a ZooKeeper server")
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	var store zoocrud.Store
	if *memory {
		store = zoocrud.NewMemoryStore()
		fmt.Println("Using an in-memory tree. The data is lost when you quit.")
	} else {
		zkStore, err := zoocrud.DialZooKeeper(strings.Split(*servers, ","), *timeout)
		if err != nil {
			log.Fatalf("Could not connect to %s: %v", *servers, err)
		}
		defer zkStore.Close()
		store = zkStore
		fmt.Printf("Connected to %s.\n", *servers)
	}
	fmt.Println("Type HELP to see the commands, or QUIT to exit.")

	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("zoocrud> ")
		if !input.Scan() {
			fmt.Println()
			return
		}
		line := strings.TrimSpace(input.Text())
		if strings.EqualFold(line, "quit") || strings.EqualFold(line, "exit") {
			return
		}
		for _, output := range zoocrud.Execute(store, line).Lines {
			fmt.Println(output)
		}
	}
}
