# SPF-Flatten
A tool for flattening SPF record hosted in AWS Route53

## Use
Configure the following environment variables
* AWS_REGION

The aws region IE us-east-1
* TEMPLATE_DOMAIN

A resovlable existing SPF to flatten
* UPDATE_DOMAIN

The actual domain you want to create SPF records for
* TEST_IP

An IP addres that should be valid via your SPF records
* ZONEID

The AWS Route53 ZONEID
