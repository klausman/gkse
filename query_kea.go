package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func getRawJSONFromFile(path string) ([]byte, error) {
	rawJSON, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could read JSON from file: %w", err)
	}
	return rawJSON, nil
}

func reader(r io.Reader, rc chan []byte) {
	var acc []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err != nil {
			break
		}
		acc = append(acc, buf[:n]...)
	}
	rc <- acc
}

func queryKeaOnce(query string) ([]byte, error) {
	c, err := net.Dial("unix", *sockPath)
	if err != nil {
		return nil, err
	}
	rc := make(chan []byte, 2)
	go reader(c, rc)
	_, err = c.Write([]byte(query))
	if err != nil {
		return nil, err
	}

	return <-rc, nil
}
