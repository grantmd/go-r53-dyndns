package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

var (
	ipv4 = ""
	ipv6 = ""
)

func main() {
	log.Println("Starting up")

	resp, err := http.Get("http://ifconfig.co/")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("IP Address is: %s", body)
	ipv4 = string(body)
	ipv6 = string(body)
}
