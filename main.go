package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
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
	flag.StringVar(&hostedZoneID, "hostedZoneID", "", "Hosted zone ID that contains the domain")
	flag.StringVar(&domain, "domain", "", "(sub)domain to manage")

	flag.Parse()

	if awsAccessKeyID == "" || awsSecretAccessKey == "" || domain == "" || hostedZoneID == "" {
		flag.Usage()
		os.Exit(2)
	}

	log.Println("Starting up")

	// Setup the route53 stuff and figure out what we have
	sess, err := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials:                   credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, ""),
	})
	if err != nil {
		log.Fatalln("Failed to create session:", err)
	}

	svc := route53.New(sess)
	listCNAMES(svc)

	// Get ipv4 address
	resp, err := http.Get("http://ipv4.icanhazip.com/")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("IPV4 address is: %s", body)
		ipv4 = string(body)
	}

	// Get ipv6 address
	resp, err = http.Get("http://ipv6.icanhazip.com/")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("IPV6 address is: %s", body)
		ipv6 = string(body)
	}

	if ipv4 == "" && ipv6 == "" {
		log.Println("Neither ipv4 nor ipv6 addresses found. Cowardly giving up.")
		os.Exit(2)
	}
}

func listCNAMES(svc *route53.Route53) {
	// Now lets list all of the records.
	// For the life of me, I can't figure out how to get these lists to actually constrain the results.
	// AFAICT, supplying only the HostedZoneId returns exactly the same results as any valid input in all params.
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID), // Required
		// MaxItems:              aws.String("100"),
		// StartRecordIdentifier: aws.String("Sample update."),
		// StartRecordName:       aws.String("com."),
		// StartRecordType:       aws.String("CNAME"),
	}
	respList, err := svc.ListResourceRecordSets(listParams)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	log.Println("All records:")
	log.Println(respList)
}
