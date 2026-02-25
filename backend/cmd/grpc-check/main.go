package main

import (
	"fmt"
	"log"
	"os"

	"backend/internal/grpcclient"
)

func main() {
	addr := "localhost:9091"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	v, err := grpcclient.Check(addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(v)
}
