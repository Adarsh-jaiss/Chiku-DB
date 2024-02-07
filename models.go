package main

import (
	"encoding/json"
	"sync"
)

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Address Address
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type (
	Driver struct {
		Dir     string
		Mutex   sync.Mutex
		Mutexes map[string]*sync.Mutex
		Log     Logger
	}

	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}
)

type Options struct {
	Logger
}
