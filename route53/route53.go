package route53

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53Updater struct {
	Region       string
	UpdateDomain string
	Zoneid       string
	DryRun       bool
	Route53      Route53Interface
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
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	// Create a Route 53 client
	return route53.New(sess), nil

}

func (s *Route53Updater) UpdateTXTRecord(recordName, newValue string) error {

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
	if targetRecord == nil {
		targetRecord = &route53.ResourceRecordSet{
			Name: aws.String(recordName),
			Type: aws.String(route53.RRTypeTxt),
		}
	}

	err = targetRecord.Validate()
	if err != nil {
		return err
	}

	// If the TXT record is found, update it; otherwise, create a new record
	targetRecord.ResourceRecords = []*route53.ResourceRecord{
		{
			Value: aws.String(newValue),
		},
	}

	changeBatch := &route53.ChangeBatch{
		Changes: []*route53.Change{
			{
				Action:            aws.String("UPSERT"),
				ResourceRecordSet: targetRecord,
			},
		},
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  changeBatch,
		HostedZoneId: aws.String(s.Zoneid),
	}
	if s.DryRun {
		fmt.Printf("DryRun TXT record not updated\n: %v\n", input)
		return nil
	}
	_, err = s.Route53.ChangeResourceRecordSetsWithContext(context.TODO(), input)
	if err != nil {
		return err
	}

	fmt.Println("TXT record updated successfully")

	return nil
}
