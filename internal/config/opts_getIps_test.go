package config

import (
	"testing"
)

func TestGetOneIp(t *testing.T) {
	testStr := "8.8.8.8"
	result, err := GetIps(testStr)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Errorf(`Expected 1 element in result, received %d: %#v`, len(result), result)
	}

	if result[0] != testStr {
		t.Errorf(`Expected "8.8.8.8", received "%v"`, result[0])
	}
}

func TestGetManyIp(t *testing.T) {
	testStr := "8.8.8.8,9.12.9.4"
	result, err := GetIps(testStr)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Errorf(`Expected 2 element in result, received %d: %#v`, len(result), result)
	}

	if result[0] != "8.8.8.8" {
		t.Errorf(`Expected "8.8.8.8", received "%v"`, result[0])
	}

	if result[1] != "9.12.9.4" {
		t.Errorf(`Expected "9.12.9.4", received "%v"`, result[1])
	}
}

func TestTrimsIp(t *testing.T) {
	testStr := " 8.8.8.8 , "
	result, err := GetIps(testStr)

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 {
		t.Errorf(`Expected 1 element in result, received %d: %#v`, len(result), result)
	}

	if result[0] != "8.8.8.8" {
		t.Errorf(`Expected "8.8.8.8", received "%v"`, result[0])
	}
}
