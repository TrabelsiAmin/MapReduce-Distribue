package main

import (
	"flag"
	"strings"
	"v_enonce/mapreduce"
)

// fonction main pour le programme master
func main() {
	jobName := flag.String("job", "testjob", "Job name")
	files := flag.String("files", "", "Comma-separated input files")
	nReduce := flag.Int("nreduce", 2, "Number of reduce tasks")
	flag.Parse()

	if *files == "" {
		mapreduce.CheckError(nil, "No input files provided\n")
	}
	fileList := strings.Split(*files, ",")

	// Start the master
	master := mapreduce.NewMaster(*jobName, fileList, *nReduce)
	master.Run()
}
