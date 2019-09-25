package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	awsAccessKeyID     string
	awsSecretAccessKey string

	domain       string
	hostedZoneID string

	ipv4 string
	ipv6 string
)

func main() {
	// Parse command-line options
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: ./go-r53-dyndns\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&awsAccessKeyID, "awsAccessKeyID", "", "AWS access key ID")
	flag.StringVar(&awsSecretAccessKey, "awsSecretAccessKey", "", "AWS secret access key")
	flag.StringVar(&domain, "domain", "", "(sub)domain to manage")
	flag.StringVar(&hostedZoneID, "hostedZoneID", "", "Hosted zone ID that contains the domain")

	flag.Parse()

	if awsAccessKeyID == "" || awsSecretAccessKey == "" || domain == "" || hostedZoneID == "" {
		flag.Usage()
		os.Exit(2)
	}

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
