package dns

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockNetworkHandler struct {
	Domain                 string
	IncludeDomain          string
	MockDNSResponse        []string
	MockIncludeDNSResponse []string
}

func (s MockNetworkHandler) LookupTXT(cxt context.Context, host string) ([]string, error) {
	if host == s.Domain {
		return s.MockDNSResponse, nil
	}
	if host == s.IncludeDomain {
		return s.MockIncludeDNSResponse, nil
	}
	return nil, fmt.Errorf("Error: no such host")
}

func TestNew(t *testing.T) {
	dns := New()
	dnscompare := DNS{}
	dnscompare.NetworkHandler = DefaultNetworkInterface{}
	if !reflect.DeepEqual(dns, dnscompare) {
		t.Errorf("Expect: %v, Got: %v", dnscompare, dns)
	}

}

func TestLookupTXT(t *testing.T) {
	dns := DefaultNetworkInterface{}
	_, err := dns.LookupTXT(context.TODO(), "google.com")
	require.Nil(t, err)
}

func TestDNSLookupSPF(t *testing.T) {
	// Mocking DNS response for testing
	domain := "example.com."
	mockNetworkHandler := MockNetworkHandler{
		Domain:          domain,
		MockDNSResponse: []string{"v=spf1 include:_spf.example.com ~all"},
	}

	// Call the function
	dns := DNS{NetworkHandler: mockNetworkHandler}
	spfRecord, err := dns.DNSLookupSPF(domain)

	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if SPF record is parsed correctly
	expectedSPFRecord := &SPFRecord{
		Mechanisms: []string{"include:_spf.example.com", "~all"},
	}

	if spfRecord == nil || len(spfRecord.Mechanisms) != len(expectedSPFRecord.Mechanisms) {
		t.Fatalf("Unexpected SPF record. Got %v, expected %v", spfRecord, expectedSPFRecord)
	}

	for i, mech := range spfRecord.Mechanisms {
		if mech != expectedSPFRecord.Mechanisms[i] {
			t.Fatalf("Unexpected mechanism. Got %s, expected %s", mech, expectedSPFRecord.Mechanisms[i])
		}
	}
}

func TestFlattenSPF(t *testing.T) {
	// Call the function
	spfRecord := SPFRecord{
		Mechanisms: []string{"include:_spf1.example.com", "~all"},
	}

	domain1 := "example.com"
	domain2 := fmt.Sprintf("_spf1.%s", domain1)
	ipaddr := net.ParseIP("1.1.1.1")
	record1 := fmt.Sprintf("v=spf1 include:_spf1.%s ~all", domain1)
	record2 := "v=spf1 ip4:1.1.1.1 ~all"
	resolv := NewResolver()
	resolv.Txt[domain1] = []string{record1}
	resolv.Ip[domain1] = []net.IP{ipaddr}
	resolv.Txt[domain2] = []string{record2}

	dns := DNS{NetworkHandler: resolv}
	flattened, err := dns.FlattenSPF(spfRecord)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	fmt.Printf("%v\n", flattened)
	require.Equal(t, "ip4:1.1.1.1", flattened[0])
	spfRecord = SPFRecord{
		Mechanisms: []string{"include:Bogus", "~all"},
	}
	_, err = dns.FlattenSPF(spfRecord)
	require.EqualError(t, err, "lookup : domain not found (for testing)")

}

func TestSPFRecordIsValid(t *testing.T) {
	domain1 := "domain1"
	ipaddr := net.ParseIP("1.1.1.1")
	record1 := "v=spf1 ip4:1.1.1.1 ~all"
	dns := NewResolver()
	dns.Txt[domain1] = []string{record1}
	dns.Ip[domain1] = []net.IP{ipaddr}

	// Simple single level record
	require.True(t, SPFRecordIsValid(dns, "1.1.1.1", domain1))
	require.False(t, SPFRecordIsValid(dns, "192.168.1.1", domain1))
	require.True(t, SPFRecordIsValid(dns, "1.1.1.1", domain1))

	// Two level record with valid ip behind include
	domain2 := fmt.Sprintf("_spf1.%s", domain1)
	record1 = fmt.Sprintf("v=spf1 include:%s ~all", domain2)
	record2 := "v=spf1 ip4:1.1.1.1 -all"
	dns.Txt[domain1] = []string{record1}
	dns.Txt[domain2] = []string{record2}
	require.True(t, SPFRecordIsValid(dns, "1.1.1.1", domain1))
	require.False(t, SPFRecordIsValid(dns, "192.168.1.1", domain1))

	// Too many lookups
	domain3 := fmt.Sprintf("_spf2.%s", domain1)
	domain4 := fmt.Sprintf("_spf3.%s", domain1)
	domain5 := fmt.Sprintf("_spf4.%s", domain1)
	domain6 := fmt.Sprintf("_spf5.%s", domain1)
	domain7 := fmt.Sprintf("_spf6.%s", domain1)
	domain8 := fmt.Sprintf("_spf7.%s", domain1)
	domain9 := fmt.Sprintf("_spf8.%s", domain1)
	domain10 := fmt.Sprintf("_spf9.%s", domain1)
	domain11 := fmt.Sprintf("_spf10.%s", domain1)
	domain12 := fmt.Sprintf("_spf11.%s", domain1)
	record3 := "v=spf1 ip4:1.1.1.2 -all"
	dns.Txt[domain2] = []string{record3}
	dns.Txt[domain3] = []string{record3}
	dns.Txt[domain4] = []string{record3}
	dns.Txt[domain5] = []string{record3}
	dns.Txt[domain6] = []string{record3}
	dns.Txt[domain7] = []string{record3}
	dns.Txt[domain8] = []string{record3}
	dns.Txt[domain9] = []string{record3}
	dns.Txt[domain10] = []string{record3}
	dns.Txt[domain11] = []string{record3}
	dns.Txt[domain12] = []string{record2}
	record1 = fmt.Sprintf("v=spf1 include:%s include:%s include:%s include:%s include:%s include:%s include:%s include:%s include:%s include:%s include:%s ~all", domain2, domain3, domain4, domain5, domain6, domain7, domain8, domain9, domain10, domain11, domain12)
	dns.Txt[domain1] = []string{record1}
	require.False(t, SPFRecordIsValid(dns, "1.1.1.1", domain1))
}

func TestSPFRecordsAreValid(t *testing.T) {
	domain1 := "domain1"
	domain2 := fmt.Sprintf("_spf1.%s", domain1)
	domain3 := fmt.Sprintf("_spf2.%s", domain1)
	ipaddr := net.ParseIP("1.1.1.1")
	record1 := fmt.Sprintf("v=spf1 include:_spf1.%s include:_spf2.%s ~all", domain1, domain1)
	resolv := NewResolver()
	resolv.Txt[domain1] = []string{record1}
	resolv.Ip[domain1] = []net.IP{ipaddr}
	resolv.Txt[domain2] = []string{"v=spf1 ip4:1.1.1.0/24 ~all"}
	resolv.Txt[domain3] = []string{"v=spf1 ip6:AAAA:AAAA:AAAA::/36 ~all"}
	dns := DNS{}
	dns.NetworkHandler = resolv
	dns.TestIP = "1.1.1.1"

	splitRecs := make(map[string]string)
	splitRecs[domain1] = "v=spf1 include:_spf1.domain1 ~all"
	splitRecs[domain2] = "v=spf1 ip4:1.1.1.0/24 ~all"
	splitRecs[domain3] = "v=spf1 ip6:AAAA:AAAA:AAAA::/36 ~all"

	_, err := dns.SPFRecordsAreValid(splitRecs)
	require.Nil(t, err)
	splitRecs[domain1] = "v=spf1 include:bogus ~all"
	_, err = dns.SPFRecordsAreValid(splitRecs)
	require.NotNil(t, err)

}

func TestJoinStringsByBytes(t *testing.T) {
	// Example usage
	splitStrings := []string{"Hello", "World", "This", "Is", "Golang", "Programming"}
	maxBytes := 20

	// Call the function
	joinedStrings, remainingStrings := JoinStringsByBytes(splitStrings, maxBytes)

	// Define the expected result
	expectedJoinedStrings := []string{"Hello", "World", "This", "Is"}
	expectedRemainingStrings := []string{"Golang", "Programming"}

	// Check if the result matches the expected values
	if !reflect.DeepEqual(joinedStrings, expectedJoinedStrings) {
		t.Errorf("Unexpected joined strings. Got %v, expected %v", joinedStrings, expectedJoinedStrings)
	}

	if !reflect.DeepEqual(remainingStrings, expectedRemainingStrings) {
		t.Errorf("Unexpected remaining strings. Got %v, expected %v", remainingStrings, expectedRemainingStrings)
	}
}

func TestSplitSPFRecords(t *testing.T) {
	// Example usage
	dnsInstance := DNS{
		UpdateDomain: "example.com",
	}

	records := []string{
		"ip4:192.168.0.0/24",
		"ip4:192.168.1.0/24",
		"ip4:192.168.2.0/24",
		"ip4:192.168.3.0/24",
		"ip4:192.168.4.0/24",
		"ip4:192.168.5.0/24",
		"ip4:192.168.6.0/24",
		"ip4:192.168.7.0/24",
		"ip4:192.168.8.0/24",
		"ip4:192.168.9.0/24",
		"ip4:192.168.10.0/24",
		"ip4:192.168.11.0/24",
		"ip4:192.168.12.0/24",
		"ip4:192.168.13.0/24",
		"ip4:192.168.14.0/24",
		"ip4:192.168.15.0/24",
	}

	// Call the function
	result := dnsInstance.SplitSPFRecords(records)

	// Define the expected result
	expectedTxtRecs := map[string]string{
		"_spf1.example.com": "v=spf1 ip4:192.168.0.0/24 ip4:192.168.1.0/24 ip4:192.168.2.0/24 ip4:192.168.3.0/24 ip4:192.168.4.0/24 ip4:192.168.5.0/24 ip4:192.168.6.0/24 ip4:192.168.7.0/24 ip4:192.168.8.0/24 ip4:192.168.9.0/24 ip4:192.168.10.0/24 ip4:192.168.11.0/24 ip4:192.168.12.0/24 ~all",
		"_spf2.example.com": "v=spf1 ip4:192.168.13.0/24 ip4:192.168.14.0/24 ip4:192.168.15.0/24 ~all",
		"example.com":       "v=spf1 include:_spf1.example.com include:_spf2.example.com ~all",
	}

	// Check if the result matches the expected values
	if !reflect.DeepEqual(result, expectedTxtRecs) {
		t.Errorf("Unexpected result. Got %v, \n\nexpected %v", result, expectedTxtRecs)
	}
}

func TestExtractIPAddressFromSPF(t *testing.T) {
	tests := []struct {
		name       string
		spfRecord  string
		expectedIP string
	}{
		{
			name:       "IPv4 Mechanism Mask",
			spfRecord:  "v=spf1 ip4:192.0.2.1/32",
			expectedIP: "192.0.2.2",
		},
		{
			name:       "IPv4 Mechanism",
			spfRecord:  "v=spf1 ip4:192.0.2.1",
			expectedIP: "192.0.2.1",
		},
		{
			name:       "IPv6 Mechanism Mask",
			spfRecord:  "v=spf1 ip6:2001:0db8:85a3:0000:0000:8a2e:0370:7334/64",
			expectedIP: "2001:db8:85a3::8a2e:370:7335",
		},
		{
			name:       "IPv6 Mechanism",
			spfRecord:  "v=spf1 ip6:2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expectedIP: "2001:db8:85a3::8a2e:370:7334",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualIP := extractIPAddressFromRecord(test.spfRecord)
			if actualIP.String() != test.expectedIP {
				t.Errorf("Expected IP: %s, but got: %s", test.expectedIP, actualIP)
			}
		})
	}
}
