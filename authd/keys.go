/* authd/authd/keys.go */
package main

import (
	"github.com/nu7hatch/gouuid"
	"crypto/rand"
	"io"
	"fmt"
	"encoding/hex"
)

func GenerateApiKey(namespace string) (ApiKey,error) {

	b := make([]byte,4)
	_,err := io.ReadFull(rand.Reader,b)
	if err != nil {
		return InvalidApiKey,err
	}

	id := hex.EncodeToString(b) + "." + namespace
	u,err := uuid.NewV5(uuid.NamespaceURL,[]byte(id))
	return ApiKey(u.String()),err
}

const (
	InvalidApiKey = ApiKey("")
)

/* ApiKey is used to control access of clients on buckets */

type ApiKey string


func (k ApiKey) String() string {
	return string(k)
}

func (k ApiKey) IsValid() bool {

	if len(string(k)) != 36 { /* TODO: formalise */
		return false
	}
	return true
}

func (k ApiKey) Obf() string {

	/* TODO */
	str := string(k)
	return fmt.Sprintf("%s..%s..%s",str[:4],str[20:24],str[30:36])
}

/* Key is just a user supplied string that is not len(0) */

type Key string

func (k Key) String() string {
	return string(k)
}

func (k Key) IsValid() bool {
	
	if len(string(k)) == 0 { 
		return false
	}
	return true
}

func (k Key) Obf() string {
	
	/* TODO */
	return string(k)
}
