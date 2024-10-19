package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type StsClient struct {
	client *sts.Client
}

type OrganizationsClient struct {
	client *organizations.Client
}

func newStsClient(awscfg aws.Config) (*StsClient, error) {
	client := sts.NewFromConfig(awscfg)
	return &StsClient{client: client}, nil
}

func newOrganizationsClient(awscfg aws.Config) (*OrganizationsClient, error) {
	client := organizations.NewFromConfig(awscfg)
	return &OrganizationsClient{client: client}, nil
}
