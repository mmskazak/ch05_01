package main

import (
	"errors"
	"testing"
)

func TestPut(t *testing.T) {
	const key = "create-key"
	const value = "create-value"

	var val interface{}
	var contains bool

	defer delete(store.m, key)

	// Sanity check
	_, contains = store.m[key]
	if contains {
		t.Error("key/value already exists")
	}

	// err should be nil
	err := Put(key, value)
	if err != nil {
		t.Error(err)
	}

	val, contains = store.m[key]
	if !contains {
		t.Error("create failed")
	}

	if val != value {
		t.Error("value not match")
	}

}

func TestGet(t *testing.T) {
	const key = "ready-key"
	const value = "ready-value"

	var val interface{}
	var err error

	defer delete(store.m, key)

	// Ready a non-thing
	val, err = Get(key)
	if err == nil {
		t.Error("expected an error")
	}

	if !errors.Is(err, ErrorNoSuchKey) {
		t.Error("unexpected error:", err)
	}

	store.m[key] = value

	val, err = Get(key)
	if err != nil {
		t.Error("unexpected error", err)
	}

	if val != value {
		t.Error("val/value mismatch")
	}

}

func TestDelete(t *testing.T) {
	const key = "delete-key"
	const value = "delete-value"

	var contains bool

	defer delete(store.m, key)
	store.m[key] = value

	_, contains = store.m[key]

	if !contains {
		t.Error("key/value doesn't exist")
	}

	err := Delete(key)
	if err != nil {
		t.Error("Delete failed")
	}

	_, contains = store.m[key]
	if contains {
		t.Error("Delete failed")
	}
}
