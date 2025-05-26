package main

import (
	"flag"
	"v_enonce/mapreduce"
)

// fonction main pour le programme worker
func main() {
	masterAddr := flag.String("master", "localhost:1234", "Master RPC address")
	id := flag.String("id", "", "Worker ID")
	flag.Parse()

	if *id == "" {
		mapreduce.CheckError(nil, "Worker ID not provided\n")
	}

	worker := mapreduce.NewWorker(*id, *masterAddr)
	worker.Run()
}
