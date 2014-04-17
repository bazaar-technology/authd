/* authd/authd/bucket_test.go */
package main

import (
	"testing"
)

const (
	DefaultNamespace = "namespace.authd.bazaar.technology"
)

func Test_AllowApiKey(t *testing.T) {

	key,_ := GenerateApiKey(DefaultNamespace)

	b := NewBucket("foo")
	ok,err := b.AllowApiKey(key)
	if !ok {
		t.Fatalf("expected to work")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_AllowApiKeyAgain(t *testing.T) {

	key,_ := GenerateApiKey(DefaultNamespace)

	b := NewBucket("foo")
	ok,err := b.AllowApiKey(key)
	if !ok {
		t.Fatalf("expected to work")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}

	ok,err = b.AllowApiKey(key)
	if ok {
		t.Fatalf("expected to fail")
	}
	if err == nil {
		t.Fatalf("expected error")
	}
}

func Test_GlobalAccess(t *testing.T) {

	key,_ := GenerateApiKey(DefaultNamespace)

	b := NewBucket("foo")
	if !b.HasGlobalAccess() {
		
		t.Fatalf("expected to have global access")
	}

	ok,err := b.AllowApiKey(key)
	if !ok {
		t.Fatalf("expected to work")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}

	if b.HasGlobalAccess() {

		t.Fatalf("expected not to have global access")
	}
}
	
