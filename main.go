package main

import (
	"flag"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"log"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
)

type arrayFlag []string

func (af *arrayFlag) String() string {
	return strings.Join(*af, ",")
}

func (af *arrayFlag) Set(v string) error {
	*af = append(*af, v)
	return nil
}

func cleanSlice(f []string) (s []string) {
	for i := 0; i < len(f); i++ {
		ts := strings.TrimSpace(f[i])
		if ts != "" {
			s = append(s, ts)
		}
	}
	return
}

var (
	RECORDS            arrayFlag = cleanSlice(strings.Split(os.Getenv("RECORDS"), ","))
	CLOUDFLARE_EMAIL             = flag.String("cloudflare-email", os.Getenv("CLOUDFLARE_EMAIL"), "Cloudflare API email")
	CLOUDFLARE_API_KEY           = flag.String("cloudflare-api-key", os.Getenv("CLOUDFLARE_API_KEY"), "Cloudflare API key")
	PRINT_RECORDS                = flag.Bool("p", false, "Print DNS records")
	DRY                          = flag.Bool("dry", false, "Dry run")

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
	flag.Var(&RECORDS, "r", "DNS record (\"<provider>:<zone>:<type>:<name>\"), multiple values supported")
	flag.Parse()
	// fetch IP
	wg := sync.WaitGroup{}
	fetchedIPs := make(map[string]int)
	fetchLock := sync.Mutex{}
	provChan := make(chan Provider, len(PROVIDERS))
	for _, prov := range PROVIDERS {
		provChan <- prov
		go func() {
			defer wg.Done()
			prov := <-provChan
			log.Printf("Fetching from %v...", prov.Name())
			ip, err := prov.Fetch()
			if err != nil {
				log.Println(err)
				return
			}
			ip = strings.TrimSpace(ip)
			if ip == "" {
				log.Printf("Invalid IP from %v ignored", prov.Name())
				return
			}
			fetchLock.Lock()
			defer fetchLock.Unlock()
			fetchedIPs[ip]++
		}()
	}
	wg.Add(len(PROVIDERS))
	wg.Wait()
	// best match
	fetchedIP := ""
	max := 0
	tot := 0
	for ip, count := range fetchedIPs {
		tot += count
		if count > max {
			max = count
			fetchedIP = ip
		}
	}
	if fetchedIP == "" {
		log.Fatalln("Failed to detect external IP")
	}
	log.Printf("Picking %v with %v/%v matches", fetchedIP, max, tot)
	// process all records
	for _, record := range RECORDS {
		parts := strings.Split(record, ":")
		if len(parts) != 4 {
			log.Fatalf("Invalid record format: %v", record)
		}
		switch parts[0] {
		case "cloudflare":
			cf, err := cloudflare.New(*CLOUDFLARE_API_KEY, *CLOUDFLARE_EMAIL)
			if err != nil {
				log.Fatalf("Failed to init Cloudflare API client: %v", err)
			}
			zoneID, err := cf.ZoneIDByName(parts[1])
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
				if r.Type == parts[2] && r.Name == parts[3] {
					record = r
				}
			}
			if *PRINT_RECORDS {
				w.Flush()
				continue
			}
			log.Printf("Found DNS record %v-%v as %v -> \"%v\"", record.ZoneID, record.ID, record.Name, record.Content)
			if record.ID == "" {
				log.Fatal("No matching DNS record found.")
			}
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
	}
}
