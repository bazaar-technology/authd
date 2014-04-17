/* authd/authd/bucket_test.go */
package main

import (
	"testing"
)

func Test_AllowApiKey(t *testing.T) {

	b := NewBucket("foo")
	ok,err := b.AllowApiKey(ApiKey("foo.bar.que"))
	if !ok {
		t.Fatalf("expected to work")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func Test_AllowApiKeyAgain(t *testing.T) {

	b := NewBucket("foo")
	ok,err := b.AllowApiKey(ApiKey("foo.bar.que"))
	if !ok {
		t.Fatalf("expected to work")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}

	ok,err = b.AllowApiKey(ApiKey("foo.bar.que"))
	if ok {
		t.Fatalf("expected to fail")
	}
	if err == nil {
		t.Fatalf("expected error")
	}
}

func Test_GlobalAccess(t *testing.T) {

	b := NewBucket("foo")
	if !b.HasGlobalAccess() {
		
		t.Fatalf("expected to have global access")
	}

	ok,err := b.AllowApiKey(ApiKey("foo.bar.que"))
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
	
