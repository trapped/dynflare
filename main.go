package main

import (
	"flag"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

var (
	CLOUDFLARE_EMAIL   = flag.String("cloudflare-email", os.Getenv("CLOUDFLARE_EMAIL"), "Cloudflare API email")
	CLOUDFLARE_API_KEY = flag.String("cloudflare-api-key", os.Getenv("CLOUDFLARE_API_KEY"), "Cloudflare API key")
	CLOUDFLARE_ZONE    = flag.String("cloudflare-zone", os.Getenv("CLOUDFLARE_ZONE"), "Cloudflare DNS zone (your domain)")
	CLOUDFLARE_RECORD  = flag.String("cloudflare-record", os.Getenv("CLOUDFLARE_RECORD"), "Cloudflare record (\"TYPE subdomain.domain\")")
	PRINT_RECORDS      = flag.Bool("print-records", false, "Print DNS records")
	DRY                = flag.Bool("dry", false, "Dry run")

	PROVIDERS = []Provider{
		&IPify{},
		&MyIPCom{},
	}
)

type Provider interface {
	Name() string
	Fetch() (string, error)
}

func main() {
	flag.Parse()
	target_record := strings.Split(*CLOUDFLARE_RECORD, " ")
	cf, err := cloudflare.New(*CLOUDFLARE_API_KEY, *CLOUDFLARE_EMAIL)
	if err != nil {
		log.Fatalf("Failed to init Cloudflare API client: %v", err)
	}
	zoneID, err := cf.ZoneIDByName(*CLOUDFLARE_ZONE)
	if err != nil {
		log.Fatalf("Failed to get Cloudflare zone ID: %v", err)
	}
	records, err := cf.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		log.Fatalf("Failed to list Cloudflare DNS records for zone %v: %v", zoneID, err)
	}
	var record cloudflare.DNSRecord
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	for _, r := range records {
		if *PRINT_RECORDS {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\n", r.ZoneID, r.ID, r.Name, r.Type, r.TTL, r.ModifiedOn, r.Content)
		}
		if r.Type == target_record[0] && r.Name == target_record[1] {
			record = r
		}
	}
	if *PRINT_RECORDS {
		w.Flush()
		os.Exit(0)
	}
	log.Printf("Found DNS record %v-%v as %v -> \"%v\"", record.ZoneID, record.ID, record.Name, record.Content)
	if record.ID == "" {
		log.Fatal("No matching DNS record found.")
	}
	// fetch IP
	fetchedIPs := make(map[string]int)
	for _, prov := range PROVIDERS {
		log.Printf("Fetching from %v...", prov.Name())
		ip, err := prov.Fetch()
		if err != nil {
			log.Println(err)
			continue
		}
		ip = strings.TrimSpace(ip)
		if ip == "" {
			log.Printf("Invalid IP from %v ignored", prov.Name())
			continue
		}
		fetchedIPs[ip]++
	}
	// best match
	fetchedIP := ""
	max := 0
	for ip, count := range fetchedIPs {
		if count > max {
			max = count
			fetchedIP = ip
		}
	}
	if fetchedIP == "" {
		log.Fatalln("Failed to detect external IP")
	}
	log.Printf("Picking %v with %v matches", fetchedIP, max)
	// update IP
	if record.Content == fetchedIP {
		log.Printf("Cloudflare DNS record content is already up to date (%v).", fetchedIP)
		os.Exit(0)
	}
	log.Printf("Cloudflare DNS record content is outdated; updating to %v...", fetchedIP)
	record.Content = fetchedIP
	if !*DRY {
		err = cf.UpdateDNSRecord(zoneID, record.ID, record)
	} else {
		err = nil
	}
	if err != nil {
		log.Fatalf("Failed to update DNS record: %v", err)
	}
	log.Println("Cloudflare DNS record updated.")
}
