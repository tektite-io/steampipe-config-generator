package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/sirupsen/logrus"
)

//go:embed templates/*.tmpl
var templates embed.FS

type CredentialAccount struct {
	Name             string
	RoleARN          string
	CredentialSource string
	ImportSchema     string
	DefaultRegion    string
	TargetRegions    []string
}

type ConnectionsTemplateData struct {
	Accounts []CredentialAccount
	Tags     map[string][]string
}

func createAWSCredentialsFile(credentialPath string, organizationAccounts []CredentialAccount) error {
	tmplFile := "templates/aws_credentials.tmpl"

	t, err := template.ParseFS(templates, tmplFile)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	err = os.MkdirAll(credentialPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating aws credentials path: %v", err)
	}
	filePath := filepath.Join(credentialPath, "credentials")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating aws credentials file: %v", err)
	}
	defer file.Close()

	err = t.Execute(file, organizationAccounts)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	log.Debug("AWS credentials file created in:", filePath)
	return nil
}

func createAWSConnectionsFile(connectionsPath string, data ConnectionsTemplateData) error {
	tmplFile := "templates/aws_connections.tmpl"

	t, err := template.ParseFS(templates, tmplFile)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	err = os.MkdirAll(connectionsPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating aws credentials path: %v", err)
	}
	filePath := filepath.Join(connectionsPath, "aws.spc")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating aws connections file: %v", err)
	}
	defer file.Close()

	err = t.Execute(file, data)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	log.Debug("AWS Connections file created in:", filePath)
	return nil
}

func main() {
	flags, err := ParseFlags()

	if err != nil {
		log.Error("error parsing flags:", err)
		return
	}

	log.Debug("parsed flags:", flags)

	roleName := flags.roleName
	credentialSource := flags.credentialSource
	credentialPath := flags.credentialPath
	connectionsPath := flags.connectionsPath
	importSchema := flags.importSchema
	defaultRegion := flags.defaultRegion
	targetRegions := flags.targetRegions
	assumeRoleArn := flags.assumeRoleArn
	skipOUs := flags.skipOUs

	ctx := context.Background()
	awscfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("error loading aws config:", err)
		return
	}

	if assumeRoleArn != "" {
		log.Info("assuming role: ", assumeRoleArn)

		sts, err := NewStsClient(awscfg)
		if err != nil {
			log.Error("error getting sts client:", err)
			return
		}

		awscfg, err = GetAssumeRoleConfig(sts, assumeRoleArn, defaultRegion, "steampipeConfigGenerator")
		if err != nil {
			log.Error("error getting aws config:", err)
			return
		}
	}

	organizationsClient, err := NewOrganizationsClient(awscfg)
	if err != nil {
		log.Error("error loading aws config:", err)
		return
	}

	accounts, err := organizationsClient.ListOrganizationAccounts()
	if err != nil {
		log.Error("error retrieving organization accounts:", err)
		return
	}

	var organizationAccounts []CredentialAccount
	taggedAccounts := make(map[string][]string)

	for _, acc := range accounts {
		if slices.Contains(skipOUs, acc.AccountOU) {
			log.Infof("Skipping account %v included skipOUs argument", acc.AccountID)
			continue
		}

		name := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(acc.Name, " ", "_"), "-", "_"))

		for key, value := range acc.Tags {
			tagKey := key + "," + value
			taggedAccounts[tagKey] = append(taggedAccounts[tagKey], name)
		}

		organizationAccounts = append(organizationAccounts, CredentialAccount{
			Name:             name,
			RoleARN:          "arn:aws:iam::" + acc.AccountID + ":role/" + roleName,
			CredentialSource: credentialSource,
			ImportSchema:     importSchema,
			DefaultRegion:    defaultRegion,
			TargetRegions:    targetRegions,
		})
	}

	data := ConnectionsTemplateData{
		Accounts: organizationAccounts,
		Tags:     taggedAccounts,
	}

	err = createAWSCredentialsFile(credentialPath, organizationAccounts)
	if err != nil {
		log.Error("error creating aws credentials file:", err)
		return
	}

	err = createAWSConnectionsFile(connectionsPath, data)
	if err != nil {
		log.Error("error creating aws connections file:", err)
		return
	}

	log.Info("config files created successfully")
}
