package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
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
	awscfg.Retryer = func() aws.Retryer {
		return retry.NewStandard(func(o *retry.StandardOptions) {
			o.MaxAttempts = 5
			o.Backoff = retry.NewExponentialJitterBackoff(1 * time.Second)
		})
	}

	client := organizations.NewFromConfig(awscfg)
	return &OrganizationsClient{client: client}, nil
}
