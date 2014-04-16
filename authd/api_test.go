/* authd/api_test.go
 */
package main

import (
	"net/http"
	"fmt"
	"testing"
	"io/ioutil"
)

var (
	defaultAddr = "127.0.0.1:8080"
	client = &http.Client{}
)

func check(bucket,key string) string {

	return fmt.Sprintf("http://%s/api/v1/check/%s/%s/",defaultAddr,bucket,key)
}

func add(bucket string) string {

	return fmt.Sprintf("http://%s/api/v1/add/%s/",defaultAddr,bucket)
}

func set(bucket string) string {
	
	return fmt.Sprintf("http://%s/api/v1/set/%s/",defaultAddr,bucket)
}

func del(bucket string) string {
	
	return fmt.Sprintf("http://%s/api/v1/del/%s/",defaultAddr,bucket)
}

func addKey(bucket,key string) string {

	return fmt.Sprintf("http://%s/api/v1/add/%s/%s/",defaultAddr,bucket,key)
}

func setKey(bucket,key string) string {

	return fmt.Sprintf("http://%s/api/v1/set/%s/%s/",defaultAddr,bucket,key)
}

func delKey(bucket,key string) string {

	return fmt.Sprintf("http://%s/api/v1/del/%s/%s/",defaultAddr,bucket,key)
}

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



func Test_AddBucket(t *testing.T) {

	status,msg,err := request(add("foo"))
	if err != nil {

		t.Fatalf(err.Error())
	}

	if status != 200 {
		
		t.Fatalf("incorrect status %d (200) - %s",status,msg)
	}

	/* test for error on reattempt */
	status,msg,err = request(add("foo"))
	if err != nil {

		t.Fatalf(err.Error())
	}

	if status != 500 {

		t.Fatalf("incorrect status %d (500) - %s",status,msg)
	}
}


func Test_SetBucket(t *testing.T) {

	status,msg,err := request(set("foo"))
	if err != nil {

		t.Fatalf(err.Error())
	}
	if status != 200 {

		t.Fatalf("incorrect status %d (200) - %s",status,msg)
	}
}

func Test_AddKey(t *testing.T) {

	status,msg,err := request(addKey("foo","bar"))
	if err != nil {

		t.Fatalf(err.Error())
	}
	if status != 200 {

		t.Fatalf("incorrect status %d (200) - %s",status,msg)
	}
}



func Test_DelBucket(t *testing.T) {

	status,msg,err := request(del("foo"))
	if err != nil {

		t.Fatalf(err.Error())
	}
	if status != 200 {

		t.Fatalf("incorrect status %d (200) - %s",status,msg)
	}
}

func Benchmark_SetKey(b *testing.B) {

	//request(set("soap"))
	url := setKey("soap","bar")

	for i := 0; i < b.N; i++ {

		status,msg,err := request(url)
		if err != nil {

			b.Fatalf(err.Error())
		}
		if status != 200 {

			b.Fatalf("incorrect status %d (200) - %s",status,msg)
		}
	}
}

func Benchmark_CheckKey(b *testing.B) {
	
	url := check("soap","bar")
	
	for i := 0; i < b.N; i++ {

		status,msg,err := request(url)
		if err != nil {

			b.Fatalf(err.Error())
		}
		if status != 200 {

			b.Fatalf("incorrect status %d (200) - %s",status,msg)
		}
	}
}

func Benchmark_Session(b *testing.B) {

	for i := 0; i < b.N; i++ {

		status,msg,err := request(setKey("tail","bar"))
		if err != nil {

			b.Fatalf(err.Error())
		}
		if status != 200 {

			b.Fatalf("incorrect status %d (200) - %s",status,msg)
		}
		
		status,msg,err = request(setKey("tail","barb"))
		if err != nil {

			b.Fatalf(err.Error())
		}
		if status != 200 {

			b.Fatalf("incorrect status %d (200) - %s",status,msg)
		}

		status,msg,err = request(delKey("tail","bar"))
		if err != nil {

			b.Fatalf(err.Error())
		}
		if status != 200 {

			b.Fatalf("incorrect status %d (200) - %s",status,msg)
		}
				
		status,msg,err = request(delKey("tail","barb"))
		if err != nil {

			b.Fatalf(err.Error())
		}
		if status != 200 {

			b.Fatalf("incorrect status %d (200) - %s",status,msg)
		}
	}
}
