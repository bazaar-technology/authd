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

	showapi := flag.Bool("api",false,"show the api")

	flag.Parse()

	ctx := NewContext()
	ctx.Namespace = *namespace
	ctx.AdminKey = *adminKey

	r := mux.NewRouter()

	api := NewApiV1Router(ctx,r,*addr)

	/* client api */
	api.ClientGetCall("/g/{bucket}",ApiV1GetBucketHandler)
	api.ClientGetCall("/g/{bucket}/{key}",ApiV1GetKeyHandler)

	/* admin api */
	//s.HandleFunc("/",ctx.admin(ApiV1PutRootHandler)).Methods("PUT") /* allows common tasks */
	//s.HandleFunc("/",ctx.admin(ApiV1DeleteRootHandler)).Methods("DELETE") /* allows common tasks */

	allowed := make(map[string]string,0)
	allowed["allow"] = "api-key"
	allowed["revoke"] = "api-key"
	allowed["enable"] = "yes"
	allowed["disable"] = "yes"

	api.AdminPutCall("/g/{bucket}",allowed,ApiV1PutBucketHandler)
	api.AdminDeleteCall("/g/{bucket}",allowed,ApiV1DeleteBucketHandler)

	allowed = make(map[string]string,0)
	api.AdminPutCall("/g/{bucket}/{key}",allowed,ApiV1PutKeyHandler)
	api.AdminDeleteCall("/g/{bucket}/{key}",allowed,ApiV1DeleteKeyHandler)

	api.AdminPutCall("/key",allowed,ApiV1PutApiKeyHandler)
	api.AdminDeleteCall("/key/{key}",allowed,ApiV1DeleteApiKeyHandler)

	/* print the api */
	if *showapi {
		for _,url := range api.api {
			
			fmt.Printf("%s\n",url)
		}
		
		for _,url := range api.curl {
			
			fmt.Printf("%s\n",url)
		}
		
		return
	}

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

type ApiV1Router struct {

	addr string
	sr * mux.Router
	ctx *Context

	api []string
	curl []string
}

func (a *ApiV1Router) ServiceGetCall(url string,fn func(http.ResponseWriter, *http.Request,*Context)) {

	r := func(w http.ResponseWriter,req *http.Request) {
		
		fn(w,req,a.ctx)
	}

	a.sr.HandleFunc(url,r).Methods("GET")
	a.api = append(a.api,fmt.Sprintf("GET %s",url))
	a.curl = append(a.curl,fmt.Sprintf("curl XGET http://%s/api/v1%s",a.addr,url))
}

func (a *ApiV1Router) ClientGetCall(url string,fn func(http.ResponseWriter,*http.Request,*Bucket)) {

	r := func(w http.ResponseWriter,req *http.Request) {
	
		vars := mux.Vars(req)
		bucket := vars["bucket"]
		
		b := a.ctx.GetBucket(Key(bucket))
		if b == nil {
			
			http.Error(w,"Unauthorized",401)
			return
		}
		
		api := ApiKey(req.Header.Get("X-ApiKey"))
		if valid,err := b.Allowed(api); !valid || err != nil {
			
			if err != nil {
				log.Printf("Invalid Api Key %s < %s (%v)\n",api.String(),req.RemoteAddr,err)
			} else {
				log.Printf("Invalid Api Key %s < %s\n",api.String(),req.RemoteAddr)
			}

			http.Error(w,"Unauthorized",401)
			return
		}			
		
		fn(w,req,b)
	}

	a.sr.HandleFunc(url,r).Methods("GET")
	a.api = append(a.api,fmt.Sprintf("GET /api/v1%s[/]",url))
	a.curl = append(a.curl,fmt.Sprintf("curl -XGET -H \"X-ApiKey:api-key\" http://%s/api/v1%s[/]",a.addr,url))
	
	a.sr.HandleFunc(url + "/",r).Methods("GET")
}

func (a *ApiV1Router) AdminPutCall(url string,allowed map[string]string,
	fn func(http.ResponseWriter,*http.Request,*Context)) {

	r := func(w http.ResponseWriter,req *http.Request) {

		adminKey := req.Header.Get("X-AdminKey")
		if adminKey != a.ctx.AdminKey {
			
			log.Printf("Invalid Admin Key %s < %s\n",adminKey,req.RemoteAddr)
			http.Error(w,"Unauthorized",401)
			return
		}
		
		req.ParseForm()

		/* check through all the key=value pairs */
		for k,_ := range req.Form {

			log.Printf("\t%s ?= %v\n",k,allowed)

			if _,isallowed := allowed[k]; !isallowed {

				http.Error(w,"Unauthorized",401)
				return
			}
		}		

		fn(w,req,a.ctx)
	}

	query := "?"
	for k,v := range allowed {
		
		if query == "?" {
			query += k + "=" + v
			continue
		} 
		query += "&" + k + "=" + v
	}

	if query == "?" {
		query = ""
	}

	a.sr.HandleFunc(url,r).Methods("PUT")
	if url != "/" {
		a.sr.HandleFunc(url + "/",r).Methods("PUT")
		a.api = append(a.api,fmt.Sprintf("PUT /api/v1%s[/]%s",url,query))
		a.curl = append(a.curl,
			fmt.Sprintf("curl -XPUT -H \"X-AdminKey:admin-key\" http://%s/api/v1%s[/]%s",a.addr,url,query))
		
	} else {
		a.api = append(a.api,fmt.Sprintf("PUT /api/v1%s%s",url,query))
		a.curl = append(a.curl,
			fmt.Sprintf("curl -XPUT -H \"X-AdminKey:admin-key\" http://%s/api/v1%s%s",a.addr,url,query))
	}
}

func (a *ApiV1Router) AdminDeleteCall(url string,allowed map[string]string,
	fn func(http.ResponseWriter,*http.Request,*Context)) {

	r := func(w http.ResponseWriter,req *http.Request) {

		adminKey := req.Header.Get("X-AdminKey")
		if adminKey != a.ctx.AdminKey {

			log.Printf("Invalid Admin Key %s < %s\n",adminKey,req.RemoteAddr)
			http.Error(w,"Unauthorized",401)
			return
		}

		req.ParseForm()

		for k,_ := range req.Form {
			if _,isallowed := allowed[k]; !isallowed {

				http.Error(w,"Unauthorized",401)
				return
			}
		}

		fn(w,req,a.ctx)
	}
	
	query := "?"
	for k,v := range allowed {
		
		if query == "?" {
			query += k + "=" + v
			continue
		} 
		query += "&" + k + "=" + v
	}

	if query == "?" {
		query = ""
	}

	a.sr.HandleFunc(url,r).Methods("DELETE")
	if url != "/" {
		a.sr.HandleFunc(url + "/",r).Methods("DELETE")
		a.api = append(a.api,fmt.Sprintf("DELETE /api/v1%s[/]%s",url,query))
		a.curl = append(a.curl,
			fmt.Sprintf("curl -XDELETE -H \"X-AdminKey:admin-key\" http://%s/api/v1%s[/]%s",a.addr,url,query))

	} else {
		a.api = append(a.api,fmt.Sprintf("DELETE /api/v1%s%s",url,query))
		a.curl = append(a.curl,
			fmt.Sprintf("curl -XDELETE -H \"X-AdminKey:admin-key\" http://%s/api/v1%s%s",a.addr,url,query))
	}
}	


func NewApiV1Router(ctx *Context,r *mux.Router,addr string) *ApiV1Router {

	a := new(ApiV1Router)
	a.sr = r.PathPrefix("/api/v1").Subrouter()
	a.ctx = ctx
	a.addr = addr
	a.api = make([]string,0)
	a.curl = make([]string,0)
	return a
}	





