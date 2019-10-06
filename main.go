package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

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

	sleepDuration int
	sleepSplay    int

	ipv4 string
	ipv6 string
)

func main() {
	// Parse command-line options
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: ./go-r53-dyndns\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&awsAccessKeyID, "awsAccessKeyID", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key ID (required)")
	flag.StringVar(&awsSecretAccessKey, "awsSecretAccessKey", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key (required)")
	flag.StringVar(&hostedZoneID, "hostedZoneID", os.Getenv("HOSTED_ZONE_ID"), "Hosted zone ID that contains the domain (required)")
	flag.StringVar(&domain, "domain", os.Getenv("HOSTED_DOMAIN"), "(sub)domain to manage (required)")

	flag.IntVar(&sleepDuration, "sleepDuration", 60, "sleep duration between checks (in seconds)")
	flag.IntVar(&sleepSplay, "sleepSplay", 5, "plus/minus seconds to splay sleep by")

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
	existingIPV4, existingIPV6, err := findExistingRecords(svc)
	if err != nil {
		log.Fatalln("Could not fetch existing records:", err)
	}

	log.Printf("Existing IPV4 record is: %s", existingIPV4)
	log.Printf("Existing IPV6 record is: %s", existingIPV6)

	// loop until killed
	var shouldExit bool
	for ok := true; ok; ok = !shouldExit {
		// Get ipv4 address
		ipv4, err := getCurrentAddress("ipv4")
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Current IPV4 address is: %s", ipv4)

		// Get ipv6 address
		ipv6, err := getCurrentAddress("ipv6")
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Current IPV6 address is: %s", ipv6)

		if ipv4 == "" && ipv6 == "" {
			log.Println("Neither IPV4 nor IPV6 addresses found. Cowardly giving up with no updates.")
			shouldExit = true
		}

		var hasChanges bool
		if ipv4 != existingIPV4 {
			log.Println("IPV4 addresses do not match. Updating...")
			hasChanges = true
		}

		if ipv6 != existingIPV6 {
			log.Println("IPV6 addresses do not match. Updating...")
			hasChanges = true
		}

		if hasChanges {
			err = setRecords(svc, ipv4, ipv6)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("OK")

			existingIPV4 = ipv4
			existingIPV6 = ipv6
		} else {
			log.Println("No changes.")
		}

		// Sleep until next check
		log.Println("Sleeping until next check...")
		time.Sleep((time.Duration(sleepDuration) * time.Second) + (time.Duration(rand.Intn(sleepSplay*2)-sleepSplay) * time.Second))
	}

	log.Println("Shutting down")
}

func getCurrentAddress(addressType string) (address string, err error) {
	resp, err := http.Get("http://" + addressType + ".icanhazip.com/")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(body)), nil
	}

	return "", nil
}

func findExistingRecords(svc *route53.Route53) (existingIPV4, existingIPV6 string, err error) {
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
	}
	respList, err := svc.ListResourceRecordSets(listParams)

	if err != nil {
		return "", "", err
	}

	for _, rs := range respList.ResourceRecordSets {
		if aws.StringValue(rs.Name) != domain && aws.StringValue(rs.Name) != domain+"." {
			continue
		}

		if aws.StringValue(rs.Type) == "A" {
			existingIPV4 = aws.StringValue(rs.ResourceRecords[0].Value)
		}

		if aws.StringValue(rs.Type) == "AAAA" {
			existingIPV6 = aws.StringValue(rs.ResourceRecords[0].Value)
		}
	}

	return existingIPV4, existingIPV6, nil
}

func setRecords(svc *route53.Route53, ipv4Address, ipv6Address string) (err error) {
	var changes []*route53.Change

	if ipv4Address != "" {
		changes = append(changes, &route53.Change{
			Action: aws.String("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(domain),
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: aws.String(ipv4Address),
					},
				},
				TTL:  aws.Int64(300),
				Type: aws.String("A"),
			},
		})
	}

	if ipv6Address != "" {
		changes = append(changes, &route53.Change{
			Action: aws.String("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(domain),
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: aws.String(ipv6Address),
					},
				},
				TTL:  aws.Int64(300),
				Type: aws.String("AAAA"),
			},
		})
	}

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
			Comment: aws.String("Mainted by go-r53-dyndns"),
		},
		HostedZoneId: aws.String(hostedZoneID),
	}

	_, err = svc.ChangeResourceRecordSets(input)
	if err != nil {
		return err
	}

	return nil
}
