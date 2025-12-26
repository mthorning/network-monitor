package utils

import (
	"bytes"
	"testing"
	"time"
)

func getTestVals() (time.Time, []byte) {
	xmasLunch := time.Date(2025, time.December, 25, 13, 0, 0, 0, time.UTC)
	xmasBytes := []byte{24, 132, 118, 210, 110, 49, 32, 0}
	return xmasLunch, xmasBytes
}

func TestTimeToBinary(t *testing.T) {
	xmasLunch, xmasBytes := getTestVals()
	output := TimeToBinary(xmasLunch)
	if !bytes.Equal(output, xmasBytes) {
		t.Errorf("Sliceds are not equal: %v vs %v", output, xmasBytes)
	}
}

func TestBinaryToTime(t *testing.T) {
	xmasLunch, xmasBytes := getTestVals()
	output := BinaryToTime(xmasBytes)
	if !xmasLunch.Equal(output) {
		t.Errorf("Sliceds are not equal: %v vs %v", output, xmasBytes)
	}
}
