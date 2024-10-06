package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"

	log "github.com/sirupsen/logrus"
)

func (sts *StsClient) assumeRole(role, sessionName string) (aws.Credentials, error) {

	appCreds := stscreds.NewAssumeRoleProvider(sts.client, role, func(opts *stscreds.AssumeRoleOptions) {
		opts.RoleSessionName = sessionName
	})
	value, err := appCreds.Retrieve(context.Background())
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("assume role failed: %w", err)
	}

	log.Debugf("successfully generated sts credentials for role: %s", role)

	return value, err
}

func GetAssumeRoleConfig(sts *StsClient, roleArn, region, sessionName string) (aws.Config, error) {

	ctx := context.Background()

	creds, err := sts.assumeRole(roleArn, sessionName)
	if err != nil {
		return aws.Config{}, fmt.Errorf("error assuming role: %s", err)
	}

	awscfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				creds.AccessKeyID,
				creds.SecretAccessKey,
				creds.SessionToken,
			),
		),
		config.WithRegion(region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("error loading aws config: %s", err)
	}

	log.Debugf("successfully generated assume role config for: %s, %s, %s", roleArn, region, sessionName)

	return awscfg, nil
}
