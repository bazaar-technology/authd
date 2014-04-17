/* authd/service.go
 */
package main

import (
	"net/http"
	"flag"
	"log"
	"fmt"
	"errors"
	"time"

	"github.com/gorilla/mux"
)

var (
	AlreadyPresent = errors.New("Already Present")
	NotFound = errors.New("Not Found")
	KeyInvalid = errors.New("Invalid Key")
)

const (
	DefaultAdminKey = "change-me"
)

func main() {

	addr := flag.String("addr","127.0.0.1:8080","http service address")
	namespace := flag.String("ns","namespace.authd.bazaar.technology","Namespace to use for generating ApiKeys")
	adminKey := flag.String("admin",DefaultAdminKey,"admin key to use")
	tls := flag.Bool("tls",false,"use TLS")
	cert := flag.String("cert","./cert.pem","certificate")
	pkey := flag.String("key","./key.pem","private key")

	flag.Parse()

	ctx := NewContext()
	ctx.Namespace = *namespace
	ctx.AdminKey = *adminKey

	r := mux.NewRouter()
	r.StrictSlash(false)

	r.HandleFunc("/",ctx.service(InformationHandler))
	r.HandleFunc("/status/",ctx.service(StatusHandler))

	s := r.PathPrefix("/api/v1").Subrouter()
	s.HandleFunc("/status/",ctx.service(StatusHandler))
	s.HandleFunc("/",ctx.service(ApiInformationHandler))
	
	/*  client api */
	s.HandleFunc("/check/{bucket}/{key}/",ctx.client(CheckKeyInBucketHandler))

	/* admin api */
	s.HandleFunc("/add/{bucket}/{key}/",ctx.admin(AddKeyToBucketHandler))
	s.HandleFunc("/set/{bucket}/{key}/",ctx.admin(SetKeyInBucketHandler))
	s.HandleFunc("/del/{bucket}/{key}/",ctx.admin(DelKeyFromBucketHandler))
	
	s.HandleFunc("/add/{bucket}/",ctx.admin(AddBucketHandler))
	s.HandleFunc("/set/{bucket}/",ctx.admin(SetBucketHandler))
	s.HandleFunc("/del/{bucket}/",ctx.admin(DelBucketHandler))
	s.HandleFunc("/enable/{bucket}/",ctx.admin(EnableBucketHandler))
	s.HandleFunc("/disable/{bucket}/",ctx.admin(DisableBucketHandler))

	s.HandleFunc("/create/",ctx.admin(CreateApiKeyHandler))

	s.HandleFunc("/allow/{key}/",ctx.admin(AllowApiKeyHandler))
	s.HandleFunc("/allow/{key}/{bucket}/",ctx.admin(AllowApiKeyHandler))

	s.HandleFunc("/revoke/{key}/",ctx.admin(RevokeApiKeyHandler))
	s.HandleFunc("/revoke/{key}/{bucket}/",ctx.admin(RevokeApiKeyHandler))


	srv := &http.Server{
		Addr:           *addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	
	if *tls {
		
		log.Fatal(srv.ListenAndServeTLS(*cert,*pkey))
	} else {
		log.Fatal(srv.ListenAndServe())
	}
}


type Context struct {

	AdminKey string
	Namespace string
	Buckets map[Key]*Bucket
}

/* AllowApiKey - allow an api key across all buckets, a global api key */
func (ctx *Context) AllowApiKey(key ApiKey) (bool,error) {
	
	if !key.IsValid() {
		return false,KeyInvalid
	}

	for _,b := range ctx.Buckets {

		b.AllowApiKey(key) /* we don't care about the return */
	}
	return true,nil
}

/* RevokeApiKey - revoke an api key across all buckets on a global scale */
func (ctx *Context) RevokeApiKey(key ApiKey) (bool,error) {

	if !key.IsValid() {
		return false,KeyInvalid
	}

	for _,b := range ctx.Buckets {

		b.RevokeApiKey(key)
	}
	return true,nil
}

/* This wraps any 'service' calls, such as status and viewing information about the service */
func (ctx *Context) service(fn func(http.ResponseWriter,*http.Request,*Context)) func(http.ResponseWriter,*http.Request) {

	r := func(w http.ResponseWriter,req *http.Request) {

		fn(w,req,ctx)
	}
	return r
}

/* This wraps all 'client' api calls, security can be added at this level */
func (ctx *Context) client(fn func(http.ResponseWriter,*http.Request,ApiKey,*Context)) func(http.ResponseWriter,*http.Request) {

	r := func(w http.ResponseWriter,req *http.Request) {
		
		api := ApiKey(req.Header.Get("X-ApiKey"))
		if !api.IsValid() {
			
			log.Printf("Invalid Api Key %s < %s\n",api.String(),req.RemoteAddr)
			http.Error(w,"Invalid Api Key",401)
			return
		}
		
		fn(w,req,api,ctx)
	}
	return r
}

/* This wraps all 'admin' api calls */
func (ctx *Context) admin(fn func(http.ResponseWriter,*http.Request,*Context)) func(http.ResponseWriter,*http.Request) {

	r := func(w http.ResponseWriter,req *http.Request) {

		adminKey := req.Header.Get("X-AdminKey")
		if adminKey != ctx.AdminKey {
			
			log.Printf("Invalid Admin Key %s < %s\n",adminKey,req.RemoteAddr)
			http.Error(w,"Invalid Admin Key",401)
			return
		}

		fn(w,req,ctx)
	}
	return r
}

/* GetBucket - find a global bucket by key */
func (ctx *Context) GetBucket(key Key) *Bucket {

	if !key.IsValid() {
		return nil
	}

	if b,exists := ctx.Buckets[key]; exists {
		
		return b
	}
	return nil
}

/* AddBucket - add a new bucket to the global space, fails on existing bucket by same key */
func (ctx *Context) AddBucket(name Key) (*Bucket,error) {

	if !name.IsValid() {
		return nil,KeyInvalid
	}

	if b,exists := ctx.Buckets[name]; exists {

		return b,AlreadyPresent
	}

	b := NewBucket(name)
	
	ctx.Buckets[name] = b
	return b,nil
}

/* SetBucket - add a new bucket to the global space, if not existing add new, else return previous */
func (ctx *Context) SetBucket(name Key) (*Bucket,error) {

	if !name.IsValid() {
		return nil,KeyInvalid
	}

	if b,exists := ctx.Buckets[name]; exists {

		return b,nil
	}


	b := new(Bucket)
	b.Name = name
	b.Records = make(map[Key]Record,0)

	ctx.Buckets[name] = b
	return b,nil
}

/* DelBucket - delete bucket in the global space */
func (ctx *Context) DelBucket(name Key) (error) {

	if !name.IsValid() {
		return KeyInvalid
	}

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

func StatusHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {


}

func InformationHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {


}

func ApiInformationHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {


}

func AllowApiKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	key := vars["key"]
	bucket := vars["bucket"]

	/* bucket only : */
	if b := ctx.GetBucket(Key(bucket)); b != nil {

		ok,err := b.AllowApiKey(ApiKey(key))
		if err != nil {
			http.Error(w,err.Error(),500)
			return
		}
		
		rep := "NO"
		if ok {
			rep = "OK"
		}

		fmt.Fprintf(w,rep)
		return
	}

	/* global : */
	ok,err := ctx.AllowApiKey(ApiKey(key))
	if err != nil {
		http.Error(w,err.Error(),500)
		return
	}

	rep := "NO"
	if ok {
		rep = "OK"
	}
	fmt.Fprintf(w,rep)
}

func RevokeApiKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	key := vars["key"]
	bucket := vars["bucket"]

	/* bucket only : */
	if b := ctx.GetBucket(Key(bucket)); b != nil {

		ok,err := b.RevokeApiKey(ApiKey(key))
		if err != nil {
			http.Error(w,err.Error(),500)
			return
		}

		rep := "NO"
		if ok {
			rep = "OK"
		}
		fmt.Fprintf(w,rep)
		return
	}

	/* global : */
	ok,err := ctx.RevokeApiKey(ApiKey(key))
	if err != nil {
		http.Error(w,err.Error(),500)
		return
	}

	rep := "NO"
	if ok {
		rep = "OK"
	}
	fmt.Fprintf(w,rep)
}

func CreateApiKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	/* generate new key */
	nkey,err := GenerateApiKey(ctx.Namespace)
	if err != nil {

		http.Error(w,err.Error(),500)
		return
	}
	fmt.Fprintf(w,nkey.String())
}
