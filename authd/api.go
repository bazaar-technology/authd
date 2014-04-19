/* authd/api.go 
 */
package main

import (
	"net/http"
	"fmt"
	"log"
	"github.com/gorilla/mux"
)

const (

	BucketEmptyResponse = "empty"
	BucketNotEmptyResponse = "not empty"
	KeyFoundResponse = "yes"
	KeyNotFoundResponse = "no"
	ActionDoneResponse = "ok"
)
	
/* GetBucket - ask whether a bucket exists and if so whether it is empty or contains records */
func ApiV1GetBucketHandler(w http.ResponseWriter,req *http.Request,bucket *Bucket) {

	/* very terse for added security */

	response := BucketEmptyResponse

	if !bucket.IsEmpty() {
		
		response = BucketNotEmptyResponse
	}

	fmt.Fprintf(w,response) 
}

/* GetKey - ask whether a bucket has a certain key (record) */
func ApiV1GetKeyHandler(w http.ResponseWriter,req *http.Request,bucket *Bucket) {

	vars := mux.Vars(req)
	key := vars["key"]

	if !bucket.Check(Key(key)) {

		http.Error(w,KeyNotFoundResponse,404)
		return
	}

	fmt.Fprintf(w,KeyFoundResponse)

}

/* PutBucket - add a bucket or change it's state */
func ApiV1PutBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]

	log.Printf("PUT bucket %s\n",bucket)

	b,err := ctx.SetBucket(Key(bucket))
	if err != nil {
		
		http.Error(w,err.Error(),500)
		return
	}

	/* go through the key=values */
	for k,vs := range req.Form {

		switch k {
		case "enable":
			if vs[0] == "yes" {
				b.Enable()
				log.Printf("enabled bucket %s",bucket)
			}
			break
		case "disable":
			if vs[0] == "yes" {
				b.Disable()
				log.Printf("disabled bucket %s",bucket)
			}
			break
		case "allow":
			b.AllowApiKey(ApiKey(vs[0]))
			log.Printf("allowed api key %s @ bucket %s",bucket)
			break
		case "revoke":
			b.RevokeApiKey(ApiKey(vs[0]))
			log.Printf("revoked api key %s @ bucket %s",bucket)
			break
		}
	}

	fmt.Fprintf(w,ActionDoneResponse)

}

/* DeleteBucket - remove a bucket and it's contained records */
func ApiV1DeleteBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]

	log.Printf("Del bucket %s\n",bucket)

	err := ctx.DelBucket(Key(bucket))
	if err != nil {
		
		http.Error(w,err.Error(),500)
		return
	}

	fmt.Fprintf(w,ActionDoneResponse)

}

/* PutKey - put a key (record) into a bucket */
func ApiV1PutKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	key := vars["key"]
	bucket := vars["bucket"]

	log.Printf("Set %s @ %s\n",key,bucket)

	b := ctx.GetBucket(Key(bucket))
	if b == nil {

		http.Error(w,"Unknown Bucket",404)
		return
	}

	b.Set(Key(key))

	fmt.Fprintf(w,ActionDoneResponse)
}

/* DeleteKey - remove a key (record) from containing bucket */
func ApiV1DeleteKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	key := vars["key"]
	bucket := vars["bucket"]

	log.Printf("Del %s @ %s\n",key,bucket)

	b := ctx.GetBucket(Key(bucket))
	if b == nil {

		http.Error(w,"Unknown Bucket",404)
		return
	}

	b.Del(Key(key))

	fmt.Fprintf(w,ActionDoneResponse)
}

/* PutApiKey - create a new Api Key for a client to use */
func ApiV1PutApiKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	/* generate new key */
	nkey,err := GenerateApiKey(ctx.Namespace)
	if err != nil {

		http.Error(w,err.Error(),500)
		return
	}
	fmt.Fprintf(w,nkey.String())
}

/* DeleteApiKey - delete all (revoke globally) an Api Key, preventing its use in future actions */
func ApiV1DeleteApiKeyHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	key := vars["key"]

	/* global : */
	_,err := ctx.RevokeApiKey(ApiKey(key))
	if err != nil {
		http.Error(w,err.Error(),500)
		return
	}

	fmt.Fprintf(w,ActionDoneResponse)
}

