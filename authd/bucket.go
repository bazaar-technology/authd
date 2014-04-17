/* authd/authd/bucket.go */
package main

import (
	"time"
	"errors"
)

var (
	ApiKeyAlreadyPresent = errors.New("Api Key Already Present")
	ApiKeyNotFound = errors.New("Api Key Not Found")
)


type Record struct {

	Created time.Time  /* when the record was added */
} 

type Bucket struct {

	Name Key
	live bool   /* is the bucket enabled, viewable to clients */
	ApiKeyList []ApiKey    /* basic Access Control List, all keys on list are accepted */
	Records map[Key]Record
}

func (b *Bucket) HasGlobalAccess() bool {

	if len(b.ApiKeyList) == 0 {
		return true
	}
	return false
}

func (b *Bucket) AllowApiKey(key ApiKey) (bool,error) {

	if !key.IsValid() {
		return false,KeyInvalid
	}

	for _,k := range b.ApiKeyList {
		
		if key == k {
			return false,ApiKeyAlreadyPresent
		}
	}

	b.ApiKeyList = append(b.ApiKeyList,key)
	return true,nil
}

func (b *Bucket) RevokeApiKey(key ApiKey) (bool,error) {

	if !key.IsValid() {
		return false,KeyInvalid
	}

	keys := make([]ApiKey,0)
	found := false
	
	for _,k := range b.ApiKeyList {

		if key != k {
			
			keys = append(keys,k)
		} else {
			found = true
		}
	}

	if !found {
		return false,ApiKeyNotFound
	}

	b.ApiKeyList = keys
	return true,nil
}


func (b *Bucket) RevokeAllApiKeys() {

	b.ApiKeyList = make([]ApiKey,0)
}

func (b *Bucket) Allowed(api ApiKey) (bool,error) {

	if !api.IsValid() {
		return false,KeyInvalid
	}

	if len(b.ApiKeyList) == 0 {
		return true,nil
	}

	for _,k := range b.ApiKeyList {

		if k == api {
			return true,nil
		}
	}
	return false,nil
}
			

func (b *Bucket) Add(key Key) bool {

	if b.Check(key) {
		return false
	}

	b.Records[key] = Record{time.Now()}
	return true
}

func (b *Bucket) Set(key Key) bool {

	b.Records[key] = Record{time.Now()}
	return true
}

func (b *Bucket) Del(key Key) bool {

	if !b.Check(key) {
		return false
	}
	delete(b.Records,key)
	return true
}

func (b *Bucket) Check(key Key) bool {

	if _,exists := b.Records[key]; !exists {
		
		return false
	}
	return true
}

func (b *Bucket) IsLive() bool {
	return b.live
}

func (b *Bucket) Enable() {
	b.live = true
}

func (b *Bucket) Disable() {
	b.live = false
}

func NewBucket(name Key) *Bucket {

	b := new(Bucket)
	b.Name = name
	b.Records = make(map[Key]Record,0)
	b.ApiKeyList = make([]ApiKey,0)
	b.live = false
	return b
}
