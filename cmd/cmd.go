package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CmdFlags struct {
	RoleName         string
	CredentialSource string
	CredentialPath   string
	ConnectionsPath  string
	ImportSchema     string
	DefaultRegion    string
	TargetRegions    []string
	AssumeRoleArn    string
	TemplatePath     string
	SkipOUs          []string
}

func ParseFlags() (*CmdFlags, error) {
	flags := CmdFlags{}

	flag.StringVar(&flags.RoleName, "role", "", "AWS Role to use in AWS config credentials")
	flag.StringVar(&flags.CredentialSource, "credential", "Environment", "AWS Credential source. Valid values are: Ec2InstanceMetadata, Environment, EcsContainer")
	flag.StringVar(&flags.CredentialPath, "path", "", "AWS Credentials file path")
	flag.StringVar(&flags.ConnectionsPath, "connections", "", "Steampipe AWS connections file path")
	flag.StringVar(&flags.ImportSchema, "schema", "enabled", "AWS Connection import schema. Valid values are: enabled, disabled")
	flag.StringVar(&flags.DefaultRegion, "region", "", "AWS Connection default region")
	flag.StringVar(&flags.AssumeRoleArn, "assume", "", "AWS Role to assume for getting Organization accounts")
	flag.StringVar(&flags.TemplatePath, "template", "", "Custom connections template path")
	targetRegions := flag.String("regions", "all", "AWS Connection target regions")
	skipOUs := flag.String("skipOUs", "", "AWS OU IDs to skip from account connections")
	flag.Parse()

	if flags.RoleName == "" {
		flag.Usage()
		return nil, fmt.Errorf("-role flag is required")
	}

	if flags.CredentialSource != "Ec2InstanceMetadata" && flags.CredentialSource != "Environment" && flags.CredentialSource != "EcsContainer" {
		flag.Usage()
		return nil, fmt.Errorf("-credential flag doesn't contain a valid value")
	}

	if flags.ImportSchema != "enabled" && flags.ImportSchema != "disabled" {
		flag.Usage()
		return nil, fmt.Errorf("-schema flag doesn't contain a valid value")
	}

	if flags.CredentialPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user's home directory: %v", err)
		}
		defaultCredentialPath := ".aws/"
		flags.CredentialPath = filepath.Join(homeDir, defaultCredentialPath)
	}

	if flags.ConnectionsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user's home directory: %v", err)
		}
		defaultConnectionsPath := ".steampipe/config/"
		flags.ConnectionsPath = filepath.Join(homeDir, defaultConnectionsPath)
	}

	if flags.DefaultRegion == "" {
		flags.DefaultRegion = os.Getenv("AWS_REGION")
		if flags.DefaultRegion == "" {
			flags.DefaultRegion = "us-east-1"
			log.Warn("default region not defined, using:", flags.DefaultRegion)
		} else {
			log.Info("default region not defined, using value from env AWS_REGION: ", flags.DefaultRegion)
		}
	}

	if *targetRegions == "all" {
		flags.TargetRegions = []string{"*"}
	} else {
		flags.TargetRegions = strings.Split(*targetRegions, ",")
	}
	log.Debug("regions: ", flags.TargetRegions)

	flags.SkipOUs = strings.Split(*skipOUs, ",")
	log.Debug("skipOUs: ", flags.SkipOUs)

	return &flags, nil
}
