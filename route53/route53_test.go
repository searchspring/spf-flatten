package route53

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/require"
)

type MockRoute53Interface struct {
	Zoneid                         string
	ListResourceRecordSetsInput    *route53.ListResourceRecordSetsInput
	ListResourceRecordSetsOutput   *route53.ListResourceRecordSetsOutput
	ChangeResourceRecordSetsInput  *route53.ChangeResourceRecordSetsInput
	ChangeResourceRecordSetsOutput *route53.ChangeResourceRecordSetsOutput
}

func (s MockRoute53Interface) ListResourceRecordSetsWithContext(cxt context.Context, input *route53.ListResourceRecordSetsInput, option ...request.Option) (*route53.ListResourceRecordSetsOutput, error) {
	if s.Zoneid != *input.HostedZoneId {
		return nil, fmt.Errorf("Zone not found.")
	}
	return s.ListResourceRecordSetsOutput, nil
}

func (s MockRoute53Interface) ChangeResourceRecordSetsWithContext(cxt context.Context, input *route53.ChangeResourceRecordSetsInput, option ...request.Option) (*route53.ChangeResourceRecordSetsOutput, error) {
	if s.Zoneid != *input.HostedZoneId {
		return nil, fmt.Errorf("Zone not found.")
	}
	return s.ChangeResourceRecordSetsOutput, nil
}

func TestUpdateTXTRecord(t *testing.T) {
	zoneid := "ZONEID"
	route53updater := Route53Updater{
		Region:         "us-east-1",
		TemplateDomain: "_template.example.com",
		UpdateDomain:   "example.com",
		Zoneid:         zoneid,
		Route53: MockRoute53Interface{
			Zoneid:                      zoneid,
			ListResourceRecordSetsInput: &route53.ListResourceRecordSetsInput{HostedZoneId: aws.String(zoneid)},
			ListResourceRecordSetsOutput: &route53.ListResourceRecordSetsOutput{
				IsTruncated:        aws.Bool(false),
				MaxItems:           aws.String("100"),
				ResourceRecordSets: []*route53.ResourceRecordSet{{Name: aws.String("example.com"), Type: aws.String("TXT")}},
			},
			ChangeResourceRecordSetsInput: &route53.ChangeResourceRecordSetsInput{HostedZoneId: aws.String(zoneid)},
			ChangeResourceRecordSetsOutput: &route53.ChangeResourceRecordSetsOutput{
				ChangeInfo: &route53.ChangeInfo{
					Id:          aws.String("bogus"),
					Status:      aws.String("bogus"),
					SubmittedAt: aws.Time(time.Now()),
				},
			},
		},
	}
	err := route53updater.UpdateTXTRecord(zoneid, "example.com", "v=spf1 ip:192.168.1.1 ~all")
	require.Nil(t, err)
	route53updater.Zoneid = "bogus"
	err = route53updater.UpdateTXTRecord(zoneid, "example.com", "v=spf1 ip:192.168.1.1 ~all")
	require.Error(t, err, fmt.Errorf("Zone not found."))

}

func TestNew(t *testing.T) {
	_, err := New(Route53Updater{
		Region:         "us-east-1",
		TemplateDomain: "_template.example.com",
		UpdateDomain:   "example.com",
		Zoneid:         "ZONEID",
	})
	require.Nil(t, err)
}

func TestNewDefaultRoute53Interface(t *testing.T) {
	route53updater, _ := New(Route53Updater{
		Region:         "us-east-1",
		TemplateDomain: "_template.example.com",
		UpdateDomain:   "example.com",
		Zoneid:         "ZONEID",
	})
	_, err := route53updater.NewDefaultRoute53Interface()
	require.Nil(t, err)
	route53updater.Region = "bogus"
}
