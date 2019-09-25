package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

var (
	ipv4 string
	ipv6 string
)

func main() {
	log.Println("Starting up")

	// Get ipv4 address
	resp, err := http.Get("http://ipv4.icanhazip.com/")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("IPV4 IP Address is: %s", body)
	ipv4 = string(body)

	// Get ipv6 address
	resp, err = http.Get("http://ipv6.icanhazip.com/")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("IPV6 IP Address is: %s", body)
	ipv6 = string(body)
}
