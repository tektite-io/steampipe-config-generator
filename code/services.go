package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type StsClient struct {
	client *sts.Client
}

func NewStsClient(awscfg aws.Config) (*StsClient, error) {
	client := sts.NewFromConfig(awscfg)
	return &StsClient{client: client}, nil
}

type OrganizationsClient struct {
	client *organizations.Client
}

func NewOrganizationsClient(awscfg aws.Config) (*OrganizationsClient, error) {
	client := organizations.NewFromConfig(awscfg)
	return &OrganizationsClient{client: client}, nil
}
