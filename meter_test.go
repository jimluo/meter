package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTempWellsTestFile(f string) {
	meters = map[string]Meter{
		"00613311": {"00613311", "123.456.789.001", "xxx1", "x-xx-xxx", "机采", "30"},
		"00613312": {"00613312", "123.456.789.002", "xxx2", "x-xx-xxx", "注水", "30"},
		"00613313": {"00613313", "123.456.789.003", "xxx3", "x-xx-xxx", "集输", "30"},
	}

	writeFileMeters(f)
}

func TestReadWells(t *testing.T) {
	setupTempWellsTestFile("./tmptest1")

	readFileMeters("./tmptest1")
	assert.Equal(t, meters["00613311"].ID, "00613311", "meters[0] != w[0]")
	assert.Equal(t, meters["00613312"].ID, "00613312", "meters[1] != w[1]")

	os.Remove("./tmptest1")
}
