package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	dns "github.com/searchspring.com/spf-flatten/dns"
	r53 "github.com/searchspring.com/spf-flatten/route53"
)

func main() {
	envs := map[string]string{
		"aws_Region":      os.Getenv("AWS_REGION"),
		"template_Domain": os.Getenv("TEMPLATE_DOMAIN"),
		"update_Domain":   os.Getenv("UPDATE_DOMAIN"),
		"test_IP":         os.Getenv("TEST_IP"),
		"zone_ID":         os.Getenv("ZONEID"),
	}

	for k, v := range envs {
		if v == "" {
			log.Fatalf("You must test '%s' ENV variable.", strings.ToUpper(k))
		}
	}

	// Retrieve SPF record for the domain
	d := dns.New()
	record, err := d.DNSLookupSPF(envs["template_Domain"])
	if err != nil {
		log.Fatalf("DNSLookupSPF: %s", err)
	}
	d.UpdateDomain = envs["update_Domain"]
	d.TestIP = envs["test_IP"]

	// Flatten SPF record
	flat, err := d.FlattenSPF(*record)
	if err != nil {
		log.Fatal(err)
	}

	// Split up records into top level record and include records
	txtRecs := d.SplitSPFRecords(flat)

	// Check records for validity
	_, err = d.SPFRecordsAreValid(txtRecs)
	if err != nil {
		log.Fatal(err)
	}

	// Update Route53
	r53updater, err := r53.New(r53.Route53Updater{
		Region:         envs["aws_Region"],
		TemplateDomain: envs["template_Domain"],
		UpdateDomain:   envs["update_Domain"],
		Zoneid:         envs["zone_ID"],
	})
	if err != nil {
		log.Fatal(err)
	}

	for domain, rec := range txtRecs {
		fmt.Printf("%v\tTXT\t%v\n\n", domain, rec)
		err = r53updater.UpdateTXTRecord(envs["zone_ID"], domain, rec)
		if err != nil {
			log.Fatalf("Update Record Fail: %v", err)
		}
	}
}
