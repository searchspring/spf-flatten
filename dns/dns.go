package dns

import (
	"context"
	"fmt"
	"net"
	"strings"

	"blitiri.com.ar/go/spf"
)

type NetworkInterface interface {
	LookupTXT(context.Context, string) ([]string, error)
}

type DefaultNetworkInterface struct{}

type DNS struct {
	UpdateDomain   string
	TestIP         string
	NetworkHandler NetworkInterface
	Records        []string
	SPFRecord      *SPFRecord
}

// SPFRecord represents a parsed SPF record
type SPFRecord struct {
	Mechanisms []string
}

func New() DNS {
	dns := DNS{}
	dns.NetworkHandler = DefaultNetworkInterface{}
	return dns
}

func (s DefaultNetworkInterface) LookupTXT(cxt context.Context, str string) ([]string, error) {

	return net.LookupTXT(str)
}

// Split up flattened record into multiple legal sized spf records
func (s DNS) SplitSPFRecords(records []string) (txtRecs map[string]string) {
	var spfSubdomain string
	var newSpfRec = "v=spf1"
	var rec []string
	recnum := 1
	txtRecs = make(map[string]string)

	// build top level spf record and sub spf records
	for len(records) > 0 {
		rec, records = JoinStringsByBytes(records, 255)
		spfSubdomain = fmt.Sprintf("_spf%d.%s", recnum, s.UpdateDomain)
		txtRecs[spfSubdomain] = fmt.Sprintf("v=spf1 %v ~all", strings.Join(rec, " "))
		newSpfRec = fmt.Sprintf("%s include:%s", newSpfRec, spfSubdomain)
		recnum = recnum + 1
	}
	txtRecs[s.UpdateDomain] = fmt.Sprintf("%s ~all", newSpfRec)
	return
}

// Test individual SPF record for compliance https://tools.ietf.org/html/rfc7208
func SPFRecordIsValid(dns *TestResolver, ip string, domain string) bool {
	ipaddr := net.ParseIP(ip)
	result, err := spf.CheckHostWithSender(ipaddr, "helo", fmt.Sprintf("sender@%s", strings.Trim(domain, ".")), spf.WithResolver(dns))
	if result == spf.Pass {
		return true
	}
	fmt.Printf("IP: %v\nDomain: %v\nResult: %v\nError: %v\nDNS: %+v\n\n", ip, domain, result, err, dns)
	return false
}

func extractIPAddressFromRecord(spfRecord string) net.IP {
	for _, mech := range strings.Split(spfRecord, " ") {
		if !(strings.HasPrefix(mech, "ip4:") || strings.HasPrefix(mech, "ip6:")) {
			continue
		}
		ipString := strings.TrimPrefix(strings.TrimPrefix(mech, "ip4:"), "ip6:")
		var ip net.IP
		if !strings.Contains(ipString, "/") {
			ip = net.ParseIP(ipString)
		} else {
			var err error
			ip, _, err = net.ParseCIDR(ipString)
			if err != nil {
				fmt.Println("Error parsing CIDR:", err)
				break
			}
			ip[len(ip)-1]++
		}
		return ip
	}
	return nil
}

// Test validity for a collection of SPF records without doing a real DNS lookup and using an IP pulled from the record
func (s DNS) SPFRecordsAreValid(txtRecs map[string]string) (bool, error) {

	// Create new DNS resolver for fake DNS lookups
	dnsRes := NewResolver()
	for domain, txtrec := range txtRecs {
		dnsRes.Txt[strings.Trim(domain, ".")] = []string{txtrec}
	}
	ip := net.ParseIP(s.TestIP)
	dnsRes.Ip[strings.Trim(s.UpdateDomain, ".")] = []net.IP{ip}

	// Check all spf records for valid syntax
	for domain, rec := range txtRecs {
		ip := extractIPAddressFromRecord(rec)
		ipaddr := s.TestIP
		if ip != nil {
			ipaddr = ip.String()
		}
		// Actually validate the record
		if !SPFRecordIsValid(dnsRes, ipaddr, domain) {
			fmt.Printf("%v:InValid\n", domain)
			fmt.Printf("IPAddr:%v\n", ipaddr)
			return false, fmt.Errorf("invalid record for domain: %v", domain)
		}
		fmt.Printf("%v:Valid\n", domain)
	}
	return true, nil
}

// DNSLookupSPF performs a DNS lookup to retrieve the SPF record for a given domain
func (s DNS) DNSLookupSPF(domain string) (*SPFRecord, error) {

	var spfRecord SPFRecord
	txt, err := s.NetworkHandler.LookupTXT(context.TODO(), domain)
	if err != nil {
		return &spfRecord, err
	}
	for _, ans := range txt {
		if strings.Contains(ans, "v=spf1") {
			for i, mech := range strings.Split(ans, " ") {
				if i == 0 {
					continue
				}
				spfRecord.Mechanisms = append(spfRecord.Mechanisms, mech)
			}
		}
	}
	s.SPFRecord = &spfRecord
	return s.SPFRecord, nil
}

// FlattenSPF flattens the SPF record by resolving included mechanisms
func (s DNS) FlattenSPF(record SPFRecord) ([]string, error) {
	flattened := make([]string, 0)

	for _, mech := range record.Mechanisms {
		if strings.HasPrefix(mech, "include:") {
			// Resolve included mechanism
			includeDomain := strings.TrimPrefix(mech, "include:")
			includeRecord, err := s.DNSLookupSPF(includeDomain)
			if err != nil {
				return nil, err
			}

			// Recursively flatten included record
			includeFlattened, err := s.FlattenSPF(*includeRecord)
			if err != nil {
				return nil, err
			}

			flattened = append(flattened, includeFlattened...)
		} else {
			// Add other mechanisms as is
			if strings.Contains(mech, "all") {
				continue
			} else {
				flattened = append(flattened, mech)
			}
		}
	}
	s.Records = flattened
	return s.Records, nil
}

func JoinStringsByBytes(splitstrings []string, maxBytes int) ([]string, []string) {
	var result []string
	var remaining []string

	currentBytes := 0

	for _, str := range splitstrings {
		// Calculate the length of the string in bytes
		strBytes := []byte(str)
		strLen := len(strBytes)

		// Check if adding the current string exceeds the maximum bytes
		if currentBytes+strLen > maxBytes {
			remaining = append(remaining, str)
		} else {
			result = append(result, str)
			currentBytes += strLen
		}
	}

	return result, remaining
}
