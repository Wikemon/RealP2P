package main

import (
	"P2P/internal"
	"fmt"
	"os"
	"time"
)

func main() {
	IP, err := internal.GetLocalIP()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Println(IP)

	for {
		time.Sleep(2 * time.Second)
		fmt.Println("Waiting..")
		peers, err := internal.DiscoverPeers()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(peers)
	}
}
