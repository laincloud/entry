package main

import (
	"net"
	"os"

	"github.com/laincloud/entry/server"
)

func main() {
	swarmPort := os.Getenv("SWARM_PORT")
	server.StartServer("80", net.JoinHostPort("swarm.lain", swarmPort))
}
