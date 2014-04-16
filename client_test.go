/* authd/client_test.go */
package authd

import (
	"net/http"
	"sync"
	"fmt"
	"time"
	"github.com/gorilla/mux"
	"testing"

	"log"
)

var (
	addr = "127.0.0.1:8888"
	once = new(sync.Once)
)

func Test_Offline(t *testing.T) {

	/* do NOT start the dummy service yet */

	c := NewClient("http://" + addr)
	
	if c.IsOnline() {

		t.Fatalf("expected service would be offline")
	}
}

func Test_Online(t *testing.T) {

	once.Do(dummy)

	c := NewClient("http://" + addr)

	if !c.IsOnline() {

		t.Fatalf("service offline")
	}
}

func Test_ClientCheckYes(t *testing.T) {
	
	once.Do(dummy)

	c := NewClient("http://" + addr) /* use the default address */
	
	ok,err := c.Check("soap","bar")
	if err != nil {

		t.Fatalf(err.Error())
	}

	if ok != true {
		
		t.Fatalf("expecting YES, got NO")
	}
}

func Test_ClientAuthCheckYes(t *testing.T) {

	once.Do(dummy)

	c := NewClient("http://" + addr)

	t0 := time.Now()

	ok,err := c.AuthCheck("soap","bar")
	
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

	c := NewClient("http://" + addr)

	ok,err := c.Check("soap","tin")
	if err != nil {

		t.Fatalf(err.Error())
	}

	if ok != false {

		t.Fatalf("expecting NO, got YES")
	}
}

func Test_ClientAuthCheckNo(t *testing.T) {

	once.Do(dummy)

	c := NewClient("http://" + addr)

	t0 := time.Now()

	ok,err := c.AuthCheck("soap","tin")
	
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

	c := NewClient("http://" + addr)

	_,err := c.CheckWithTimeout("soap","bar")
	if err != nil {

		t.Fatalf(err.Error())
	}
}

func Test_ClientAuthCheckTimeout(t *testing.T) {

	once.Do(dummy)

	c := NewClient("http://" + addr)
	
	t0 := time.Now()
	_,err := c.AuthCheckWithTimeout("soap","bar")
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

	c := NewClient("http://" + addr)
	var err error

	if _,err = c.CheckWithTimeout("soap","bubble"); err == nil {

		t.Fatalf("unknown error, should have failed with timeout")
	}
	if err != TimeOut {

		t.Fatalf("Should have failed with timeout not - %v",err.Error())
	}
}

func Test_ClientAuthCheckTimeoutError(t *testing.T) {

	once.Do(dummy)
	
	c := NewClient("http://" + addr)
	var err error
	
	t0 := time.Now()
	if _,err = c.AuthCheckWithTimeout("soap","bubble"); err == nil {

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

func dummy() {

	go func() {
		log.Printf("starting dummy service @%s",addr)
		r := mux.NewRouter()
		r.StrictSlash(false)
		r.HandleFunc("/api/v1/check/{bucket}/{key}/",CheckHandler)
		r.HandleFunc("/api/v1/status/",StatusHandler)
		http.Handle("/",r)
		log.Fatalf(http.ListenAndServe(addr,nil).Error())
	}()
	time.Sleep(1 * time.Second)
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
		

