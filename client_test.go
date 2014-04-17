/* authd/client_test.go */
package authd

import (
	"net/http"
	"sync"
	"fmt"
	"time"
	"github.com/gorilla/mux"
	"testing"
	"io/ioutil"

	"log"
)

var (
	cert = "./authd/cert.pem"
	key = "./authd/key.pem"
	insecure = true /* if the cert is self-signed, the test cert is so tell the client to skip verify */
	addr = "127.0.0.1:8888"
	once = new(sync.Once)

	useTLS = true /* change to true to test TLS version */
	
)

func Test_Offline(t *testing.T) {

	/* do NOT start the dummy service yet */
	if useTLS {

		data, err := ioutil.ReadFile(cert)
		if err != nil {
			t.Fatalf(err.Error())
		}

		StartTLS(addr,data,insecure)
	} else {
		Start(addr)
	}


	if IsOnline() {

		t.Fatalf("expected service would be offline")
	}
}

func Test_Online(t *testing.T) {

	once.Do(dummy)

	if !IsOnline() {

		t.Fatalf("service offline")
	}
}

func Test_ClientCheckYes(t *testing.T) {
	
	once.Do(dummy)
		
	ok,err := Check("soap","bar")
	if err != nil {

		t.Fatalf(err.Error())
	}

	if ok != true {
		
		t.Fatalf("expecting YES, got NO")
	}
}

func Test_ClientAuthCheckYes(t *testing.T) {

	once.Do(dummy)

	t0 := time.Now()

	ok,err := AuthCheck("soap","bar")
	
	t1 := time.Now()

	if err != nil {

		t.Fatalf(err.Error())
	}
	if ok != true {
		t.Fatalf("expecting YES, got NO")
	}

	if t1.Sub(t0) < c.AtLeast {
		t.Fatalf("did not take at least the required time - took %v\n",t1.Sub(t0))
	}	
}

func Test_ClientCheckNo(t *testing.T) {

	once.Do(dummy)

	ok,err := Check("soap","tin")
	if err != nil {

		t.Fatalf(err.Error())
	}

	if ok != false {

		t.Fatalf("expecting NO, got YES")
	}
}

func Test_ClientAuthCheckNo(t *testing.T) {

	once.Do(dummy)

	t0 := time.Now()

	ok,err := AuthCheck("soap","tin")
	
	t1 := time.Now()
	
	if err != nil {

		t.Fatalf(err.Error())
	}

	if ok != false {

		t.Fatalf("expecting No, got YES")
	}

	if t1.Sub(t0) < c.AtLeast {
		t.Fatalf("did not take at least the required time - took %v\n",t1.Sub(t0))
	}
		
}

func Test_ClientCheckTimeout(t *testing.T) {

	once.Do(dummy)

	_,err := CheckWithTimeout("soap","bar")
	if err != nil {

		t.Fatalf(err.Error())
	}
}

func Test_ClientAuthCheckTimeout(t *testing.T) {

	once.Do(dummy)

	t0 := time.Now()
	_,err := AuthCheckWithTimeout("soap","bar")
	t1 := time.Now()
	if err != nil {

		t.Fatalf(err.Error())
	}
	if t1.Sub(t0) < c.AtLeast {
		t.Fatalf("did not take at least the required time - took %v\n",t1.Sub(t0))
	}
}

func Test_ClientCheckTimeoutError(t *testing.T) {

	once.Do(dummy)

	var err error

	if _,err = CheckWithTimeout("soap","bubble"); err == nil {

		t.Fatalf("unknown error, should have failed with timeout")
	}
	if err != TimeOut {

		t.Fatalf("Should have failed with timeout not - %v",err.Error())
	}
}

func Test_ClientAuthCheckTimeoutError(t *testing.T) {

	once.Do(dummy)
	
	var err error
	
	t0 := time.Now()
	if _,err = AuthCheckWithTimeout("soap","bubble"); err == nil {

		t.Fatalf("unknown error, should have failed with timeout")
	}
	t1 := time.Now()
	if err != TimeOut {

		t.Fatalf("Should have failed with timeout not - %v",err.Error())
	}
	if t1.Sub(t0) < c.AtLeast {
		t.Fatalf("did not take at least the required time - took %v\n",t1.Sub(t0))
	}
}

/* dummy server for testing client api */

type DummyServe struct {

	r *mux.Router	
}

func (d DummyServe) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	d.r.ServeHTTP(w,req)
}


func dummy() {

	go func() {
		proto := "http://"
		if useTLS {
			proto = "https://"
		}
		log.Printf("starting dummy service @%s%s",proto,addr)
		dumb := new(DummyServe)
		dumb.r = mux.NewRouter()
		dumb.r.StrictSlash(false)
		dumb.r.HandleFunc("/api/v1/check/{bucket}/{key}/",CheckHandler)
		dumb.r.HandleFunc("/api/v1/status/",StatusHandler)
	
		srv := &http.Server{
		Addr:           addr,
		Handler:        dumb,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		}
	
		if useTLS {
			log.Fatal(srv.ListenAndServeTLS(cert,key))
			
		} else {
			
			log.Fatal(srv.ListenAndServe())
		}		
	}()

	if useTLS {
		log.Printf("....setting up")
		time.Sleep(5 * time.Second) /* give the server a chance to start up */
	} else {
		time.Sleep(1 * time.Second)
	}
}

func CheckHandler(w http.ResponseWriter,req *http.Request) {

	vars := mux.Vars(req)
	bucket := vars["bucket"]
	key := vars["key"]

	if bucket == "soap" && key == "bubble" {
		
		/* the default timeout for the client is 5 seconds */
		time.Sleep(defaultTimeout + 100 * time.Millisecond) 
	}

	if key == "tin" {
		/* cause every key other than 'tin' to return YES */
		http.Error(w,"Not Found",404)
		return
	}
	
	fmt.Fprintf(w,"yes")
}

func StatusHandler(w http.ResponseWriter,req *http.Request) {

	fmt.Fprintf(w,"ok")
}
		

