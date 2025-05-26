package tests

import (
	"os"
	"testing"
	"v_enonce/mapreduce"
)

func TestMapReduceSequential(t *testing.T) {
	input := "input_test.txt"
	_ = os.WriteFile(input, []byte("foo bar foo baz foo bar"), 0644)
	defer os.Remove(input)
	mapreduce.Sequential("testjob", []string{input}, 2, mapF, reduceF)
	filename := "mrtmp.testjob"
	expected := map[string]string{
		"foo": "3",
		"bar": "2",
		"baz": "1",
	}

	got := decodeMapFromFile(t, filename)
	defer os.Remove(filename)
	assertEqualMaps(t, got, expected)
	mapreduce.CleanIntermediary("testjob", 1, 2)
}
