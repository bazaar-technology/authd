/* authd/service.go
 */
package main

import (
	"net/http"
	"flag"
	"log"
	"time"
	"errors"

	"github.com/gorilla/mux"
)

var (
	AlreadyPresent = errors.New("Already Present")
	NotFound = errors.New("Not Found")
)

func main() {

	addr := flag.String("addr","127.0.0.1:8080","http service address")
	flag.Parse()

	ctx := NewContext()

	r := mux.NewRouter()
	r.StrictSlash(false)
	s := r.PathPrefix("/api/v1").Subrouter()
	
	/*  client api */
	s.HandleFunc("/check/{bucket}/{key}/",ctx.client(CheckKeyInBucketHandler))

	/* admin api */
	s.HandleFunc("/add/{bucket}/{key}/",ctx.admin(AddKeyToBucketHandler))
	s.HandleFunc("/set/{bucket}/{key}/",ctx.admin(SetKeyInBucketHandler))
	s.HandleFunc("/del/{bucket}/{key}/",ctx.admin(DelKeyFromBucketHandler))
	
	s.HandleFunc("/add/{bucket}/",ctx.admin(AddBucketHandler))
	s.HandleFunc("/set/{bucket}/",ctx.admin(SetBucketHandler))
	s.HandleFunc("/del/{bucket}/",ctx.admin(DelBucketHandler))

	http.Handle("/",r)

	log.Fatal(http.ListenAndServe(*addr,nil))
}

type Key string

type Record struct {

	Created time.Time  /* when the record was added */
} 

type Bucket struct {

	Name Key
	Records map[Key]Record
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

type Context struct {

	Buckets map[Key]*Bucket
}

func (ctx *Context) client(fn func(http.ResponseWriter,*http.Request,*Context)) func(http.ResponseWriter,*http.Request) {

	r := func(w http.ResponseWriter,req *http.Request) {

		fn(w,req,ctx)
	}
	return r
}

func (ctx *Context) admin(fn func(http.ResponseWriter,*http.Request,*Context)) func(http.ResponseWriter,*http.Request) {

	r := func(w http.ResponseWriter,req *http.Request) {

		/* TODO, add check X-Headers here for api key etc */
		fn(w,req,ctx)
	}
	return r
}

func (ctx *Context) GetBucket(key Key) *Bucket {

	if b,exists := ctx.Buckets[key]; exists {
		
		return b
	}
	return nil
}

func (ctx *Context) AddBucket(name Key) (*Bucket,error) {

	if b,exists := ctx.Buckets[name]; exists {

		return b,AlreadyPresent
	}
	
	b := new(Bucket)
	b.Name = name
	b.Records = make(map[Key]Record,0)
	
	ctx.Buckets[name] = b
	return b,nil
}

func (ctx *Context) SetBucket(name Key) (*Bucket,error) {

	if b,exists := ctx.Buckets[name]; exists {

		return b,nil
	}


	b := new(Bucket)
	b.Name = name
	b.Records = make(map[Key]Record,0)

	ctx.Buckets[name] = b
	return b,nil
}

func (ctx *Context) DelBucket(name Key) (error) {

	if _,exists := ctx.Buckets[name]; !exists {

		return NotFound
	}

	delete(ctx.Buckets,name)
	return nil
}

func NewContext() *Context {

	c := new(Context)
	c.Buckets = make(map[Key]*Bucket,0)
	return c
}
