package main

import (
	"log"
	"os"
	"testing"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func TestCreateLogger(t *testing.T) {
	log.Println("TestCreateLogger start")
	const filename = "/tmp/create-logger.txt"
	defer os.Remove(filename)

	t1, err := NewTransactionLogger(filename)

	if t1 == nil {
		t.Error("Logger is nil?")
	}

	if err != nil {
		t.Errorf("got error: %v", err)
	}

	if !fileExists(filename) {
		t.Errorf("File %s doesnt exist", filename)
	}

	log.Println("TestCreateLogger end")
}

func TestWriteAppend(t *testing.T) {
	const filename = "tmp/write-append.txt"
	defer os.Remove(filename)

	t1, err := NewTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}

	chev, cherr := t1.ReadEvents()
	for e := range chev {
		t.Log(e)
	}

	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	t1.Run()
	defer t1.Close()

	t1.WritePut("my-key", "my-value")
	t1.WritePut("my-key2", "my-value2")
	t1.Wait()

	log.Println("t12")
	t12, err := NewTransactionLogger(filename)
	if err != nil {
		t.Error(err)
	}

	log.Println("t12.ReadEvents")
	chev, cherr = t12.ReadEvents()

	for e := range chev {
		log.Println("t.Log(e)")
		t.Log(e)
	}

	err = <-cherr
	if err != nil {
		t.Error(err)
	}

	t12.Run()
	defer t12.Close()

	t12.WritePut("my-key3", "my-value3")
	t12.WritePut("my-key4", "my-value4")
	t12.WritePut("my-key5", "my-value5")
	t12.Wait()

	if t12.lastSequence != 5 {
		t.Errorf("last sequence mismatch (expected 5: got %d)", t12.lastSequence)
	}
}

//func TestWritePut(t *testing.T) {
//	const filename = "/tmp/write-put.txt"
//	defer os.Remove(filename)
//
//	t1, _ := NewTransactionLogger(filename)
//	t1.Run()
//
//	defer t1.Close()
//
//	t1.WritePut("my-key", "my-value")
//	t1.WritePut("my-key", "my-value2")
//	t1.WritePut("my-key", "my-value3")
//	t1.WritePut("my-key", "my-value4")
//	t1.Wait()
//
//	t12, _ := NewTransactionLogger(filename)
//	evin, errin := t12.ReadEvents()
//	defer t12.Close()
//
//	for e := range evin {
//		t.Log(e)
//	}
//
//	err := <-errin
//	if err != nil {
//		t.Error(err)
//	}
//
//	if t1.lastSequence != t12.lastSequence {
//		t.Errorf("Last sequence mismatch (%d vs %d)", t1.lastSequence, t12.lastSequence)
//	}
//}
