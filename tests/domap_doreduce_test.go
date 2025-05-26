package tests

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"v_enonce/mapreduce"
)

var jobName = "jobwcount"

func checkErrFatal(t *testing.T, err error, msg string, args ...interface{}) {
	t.Helper()
	if err != nil {
		t.Fatalf(msg, args...)
	}
}

func assertEqualMaps(t *testing.T, m1, m2 map[string]string) {
	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("assertEqualMaps failed, got %v, want %v", m1, m2)
	}
}

func encodeMapInFile(t *testing.T, kvs map[string]string, filename string) {
	file, err := os.Create(filename)
	checkErrFatal(t, err, "cannot create file %s: %v", filename, err)

	enc := json.NewEncoder(file)
	for k, v := range kvs {
		err := enc.Encode(&mapreduce.KeyValue{Key: k, Value: v})
		checkErrFatal(t, err, "cannot encode kv: %v", err)
	}
	file.Close()
}

func decodeMapFromFile(t *testing.T, filename string) map[string]string {
	inFile, err := os.Open(filename)
	defer inFile.Close()
	checkErrFatal(t, err, "cannot open file %s: %v", filename, err)

	decoder := json.NewDecoder(inFile)
	kvs := make(map[string]string)
	var kv mapreduce.KeyValue
	for decoder.Decode(&kv) == nil {
		kvs[kv.Key] = kv.Value
	}
	return kvs
}

func TestDoMap(t *testing.T) {
	input := "orange banana banana apple orange banana"
	expectedKeys := map[string]string{
		"banana": "3",
		"orange": "2",
		"apple":  "1",
	}

	inputFile := "test_input.txt"
	file, err := os.Create(inputFile)
	checkErrFatal(t, err, "cannot create input file: %v", err)
	_, err = file.WriteString(input)
	file.Close()
	defer os.Remove(inputFile)

	mapTaskNumber := 555
	nReduce := 10
	mapreduce.DoMap(jobName, mapTaskNumber, inputFile, nReduce, mapF)

	gotKeys := map[string]string{}
	for r := 0; r < nReduce; r++ {
		fileName := mapreduce.ReduceName(jobName, mapTaskNumber, r)
		tmp := decodeMapFromFile(t, fileName)
		defer os.Remove(fileName)
		for k, v := range tmp {
			gotKeys[k] = v
		}
	}
	assertEqualMaps(t, gotKeys, expectedKeys)
}

func TestDoReduce(t *testing.T) {
	jobName := "job1"
	reduceTaskNumber := 0
	nMap := 2

	inputs := [][]mapreduce.KeyValue{
		{{"apple", "1"}, {"banana", "2"}},
		{{"apple", "1"}, {"orange", "2"}},
	}
	expectedKeys := map[string]string{
		"banana": "2",
		"orange": "2",
		"apple":  "2",
	}

	for i := 0; i < nMap; i++ {
		fileName := mapreduce.ReduceName(jobName, i, reduceTaskNumber)
		file, err := os.Create(fileName)
		defer os.Remove(fileName)
		checkErrFatal(t, err, "cannot create file %s: %v", fileName, err)

		enc := json.NewEncoder(file)
		for _, kv := range inputs[i] {
			err := enc.Encode(&kv)
			checkErrFatal(t, err, "cannot encode kv: %v", err)
		}
		file.Close()
	}

	mapreduce.DoReduce(jobName, reduceTaskNumber, nMap, reduceF)

	fileName := mapreduce.MergeName(jobName, reduceTaskNumber)
	defer os.Remove(fileName)
	gotKeys := decodeMapFromFile(t, fileName)

	assertEqualMaps(t, gotKeys, expectedKeys)
}