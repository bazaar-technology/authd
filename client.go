/* authd/client.go */
package authd

import (
	"net/http"
	"strings"
	"io/ioutil"
	"fmt"
	"time"
	"errors"
)


var (
	defaultAddr = "http://127.0.0.1:8080"
	defaultTimeout = 5 * time.Second
	defaultAtLeast = 1 * time.Second /* requests always take at least n */

	client = &http.Client{}

	TimeOut = errors.New("Time Out")
)

func request(url string) (int,string,error) {

	resp,err := client.Get(url)
	if err != nil {
		return -1,"",err
	}

	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode,"",err
	}
	return resp.StatusCode,string(body),nil
}

func check(addr,bucket,key string) (bool,error) {

	url := fmt.Sprintf("%s/api/v1/check/%s/%s/",addr,bucket,key)
	status,msg,err := request(url)
	if err != nil {
		return false,err
	}
	if status != 200 {
		return false,nil
	}
	if strings.ToLower(msg) != "yes" {
		return false,nil
	}
	return true,nil
}
	
type response struct {

	checked bool 
	err error
}

func wrap(addr,bucket,key string) chan response {

	ch := make(chan response,1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		ok,err := check(addr,bucket,key)
		ch <- response{ok,err}
	}()
	return ch
}

type Client struct {

	Addr string /* service http address */
	Timeout time.Duration
	AtLeast time.Duration
}

func (c *Client) IsOnline() bool {

	url := fmt.Sprintf("%s/api/v1/status/",c.Addr)
	status,msg,err := request(url)
	if err != nil {
		return false
	}
	if status != 200 {
		return false
	}
	if strings.ToLower(msg) != "ok" {
		return false
	}
	return true
}

func (c *Client) Check(bucket,key string) (bool,error) {

	return check(c.Addr,bucket,key)
}

func (c *Client) CheckWithTimeout(bucket,key string) (bool,error) {

	select {
	case <- time.After(c.Timeout):
		return false,TimeOut
	case rep := <- wrap(c.Addr,bucket,key):
		return rep.checked,rep.err
	}
}

func (c *Client) AuthCheck(bucket,key string) (bool,error) {
		
	t0 := time.Now()
	ok,err := check(c.Addr,bucket,key)
	time.Sleep(c.AtLeast - time.Now().Sub(t0))

	return ok,err
}

func (c *Client) AuthCheckWithTimeout(bucket,key string) (bool,error) {

	t0 := time.Now()
	
	select {
	case <- time.After(c.Timeout):
		/* WARNING: assumption that timeout > atleast */
		return false,TimeOut
	case rep := <- wrap(c.Addr,bucket,key):
		
		time.Sleep(c.AtLeast - time.Now().Sub(t0))		
		return rep.checked,rep.err
	}
}

func NewClient(addr string) *Client {

	/* TODO: check for valid http address */
	if len(addr) == 0 {

		addr = defaultAddr
	}

	c := new(Client)
	c.Addr = addr
	c.Timeout = defaultTimeout
	c.AtLeast = defaultAtLeast
	return c
}
