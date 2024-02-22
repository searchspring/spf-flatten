# SPF-Flatten
A tool for flattening SPF record hosted in AWS Route53

## Configure
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

## Use
You need to setup an template SPF record will all the `include` mechanisms you need to flatten. Point this at that template record and it will flatten all the includes to ip4 and ip6 mechanisms. It will also generate a number of seperate records so that no record is over the limit for [RFC720](https://tools.ietf.org/html/rfc7208). It then checks the validity of all created records. Finally it updates the domain's SPF records in route53.


# License and Author

- Author:: Greg Hellings 

Copyright 2024, Searchspring.com
Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

```
http://www.apache.org/licenses/LICENSE-2.0
```
