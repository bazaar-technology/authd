/* authd/client.go */
package authd

import (
	"crypto/x509"
	"crypto/tls"
	"net/http"
	"strings"
	"io/ioutil"
	"fmt"
	"time"
	"errors"
)


var (
	defaultAddr = "127.0.0.1:8080"
	defaultTimeout = 5 * time.Second
	defaultAtLeast = 1 * time.Second /* requests always take at least n */
	
	TimeOut = errors.New("Time Out")

	c *client
)

func request(url string) (int,string,error) {

	resp,err := c.HttpClient.Get(url)
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

type client struct {

	Addr string /* service http address */
	Timeout time.Duration
	AtLeast time.Duration
	HttpClient *http.Client
}

func IsOnline() bool {

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

func Check(bucket,key string) (bool,error) {

	return check(c.Addr,bucket,key)
}

func CheckWithTimeout(bucket,key string) (bool,error) {

	select {
	case <- time.After(c.Timeout):
		return false,TimeOut
	case rep := <- wrap(c.Addr,bucket,key):
		return rep.checked,rep.err
	}
}

func AuthCheck(bucket,key string) (bool,error) {
		
	t0 := time.Now()
	ok,err := check(c.Addr,bucket,key)
	time.Sleep(c.AtLeast - time.Now().Sub(t0))

	return ok,err
}

func AuthCheckWithTimeout(bucket,key string) (bool,error) {

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

func Start(addr string) bool {

	if c != nil {
		return false
	}

	c = new(client)
	c.Addr = "http://" + addr
	c.Timeout = defaultTimeout
	c.AtLeast = defaultAtLeast
	c.HttpClient = &http.Client{}
	return true
}

func StartTLS(addr string,certData []byte,insecure bool) bool {

	if c != nil {
		return false
	}

	parts := strings.Split(addr,":")

	c = new(client)
	c.Addr = "https://" + addr
	c.Timeout = defaultTimeout
	c.AtLeast = defaultAtLeast
		
	config := &tls.Config {InsecureSkipVerify:insecure,ServerName:parts[0]}

	certs := x509.NewCertPool()

	certs.AppendCertsFromPEM(certData)
	config.RootCAs = certs
	
	tr := &http.Transport{
		TLSClientConfig: config,
	}

	c.HttpClient = &http.Client{Transport: tr}

	return true
}

