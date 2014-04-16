/* authd/api.go 
 */
package main

import (
	"net/http"
	"fmt"
	"log"
	"github.com/gorilla/mux"
)

func CheckKeyInBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]
	key := vars["key"]

//	log.Printf("Check %s@%s\n",key,bucket)

	b := ctx.GetBucket(Key(bucket))
	if b == nil {

		http.Error(w,"NO",404)
		return
	}

	if !b.Check(Key(key)) {

		http.Error(w,"NO",404)
		return
	}

	fmt.Fprintf(w,"YES")
}


/* admin api */
func AddKeyToBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]
	key := vars["key"]

	log.Printf("Add %s@%s\n",key,bucket)

	b := ctx.GetBucket(Key(bucket))
	if b == nil {

		http.Error(w,"Bucket Not Found",404)
		return
	}

	if !b.Add(Key(key)) {

		http.Error(w,"Key Already Exists",404)
		return
	}

	fmt.Fprintf(w,"ok")
}

func SetKeyInBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]
	key := vars["key"]

	log.Printf("Set %s@%s\n",key,bucket)
	
	var err error

	b := ctx.GetBucket(Key(bucket))
	if b == nil {

		b,err = ctx.AddBucket(Key(bucket))
		if err != nil {

			http.Error(w,err.Error(),500)
			return
		}
	}

	if !b.Set(Key(key)) {

		http.Error(w,"Problem setting Key",500)
		return
	}

	fmt.Fprintf(w,"ok")
}

func DelKeyFromBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]
	key := vars["key"]

	log.Printf("Del %s@%s\n",key,bucket)

	b := ctx.GetBucket(Key(bucket))
	if b == nil {

		http.Error(w,"Bucket Not Found",404)
		return
	}

	if !b.Del(Key(key)) {

		http.Error(w,"Key Not Found",404)
		return
	}

	fmt.Fprintf(w,"ok")
}

func AddBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]

	log.Printf("Add %s\n",bucket)

	_,err := ctx.AddBucket(Key(bucket))
	if err != nil {
		
		http.Error(w,err.Error(),500)
		return
	}

	fmt.Fprintf(w,"ok")

}

func SetBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]

	log.Printf("Set %s\n",bucket)

	_,err := ctx.SetBucket(Key(bucket))
	if err != nil {
		
		http.Error(w,err.Error(),500)
		return
	}

	fmt.Fprintf(w,"ok")
}

func DelBucketHandler(w http.ResponseWriter,req *http.Request,ctx *Context) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]

	log.Printf("Del %s\n",bucket)

	err := ctx.DelBucket(Key(bucket))
	if err != nil {
		
		http.Error(w,err.Error(),500)
		return
	}

	fmt.Fprintf(w,"ok")
}
