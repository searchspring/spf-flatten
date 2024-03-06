package route53

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53Updater struct {
	Region         string
	TemplateDomain string
	UpdateDomain   string
	Zoneid         string
	Route53        Route53Interface
}

type Route53Interface interface {
	ListResourceRecordSetsWithContext(context.Context, *route53.ListResourceRecordSetsInput, ...request.Option) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSetsWithContext(context.Context, *route53.ChangeResourceRecordSetsInput, ...request.Option) (*route53.ChangeResourceRecordSetsOutput, error)
}

type DefaultRoute53Interface struct {
	Route53 route53.Route53
}

func New(s Route53Updater) (Route53Updater, error) {
	r53interface, err := s.NewDefaultRoute53Interface()
	if err != nil {
		return s, err
	}
	s.Route53 = r53interface
	return s, nil
}

func (s Route53Updater) NewDefaultRoute53Interface() (Route53Interface, error) {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region),
	})
	if err != nil {
		return nil, err
	}
	// Create a Route 53 client
	return route53.New(sess), nil

}

func (s Route53Updater) UpdateTXTRecord(recordName, newValue string) error {

	// Retrieve the existing record
	existingRecord, err := s.Route53.ListResourceRecordSetsWithContext(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(s.Zoneid),
	})
	if err != nil {
		return err
	}

	// Find the TXT record
	var targetRecord *route53.ResourceRecordSet
	for _, record := range existingRecord.ResourceRecordSets {
		if aws.StringValue(record.Name) == recordName && aws.StringValue(record.Type) == "TXT" {
			targetRecord = record
			break
		}
	}

	// If the TXT record is found, update it; otherwise, create a new record
	if targetRecord != nil {
		targetRecord.ResourceRecords = []*route53.ResourceRecord{
			{
				Value: aws.String(newValue),
			},
		}

		_, err = s.Route53.ChangeResourceRecordSetsWithContext(context.TODO(), &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action:            aws.String("UPSERT"),
						ResourceRecordSet: targetRecord,
					},
				},
			},
			HostedZoneId: aws.String(s.Zoneid),
		})
		if err != nil {
			return err
		}

		fmt.Println("TXT record updated successfully")
	} else {
		log.Fatal("TXT record not found")
	}

	return nil
}
