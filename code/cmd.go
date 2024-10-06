package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CmdFlags struct {
	roleName         string
	credentialSource string
	credentialPath   string
	connectionsPath  string
	importSchema     string
	defaultRegion    string
	targetRegions    []string
	assumeRoleArn    string
	skipOUs          []string
}

func ParseFlags() (*CmdFlags, error) {
	flags := CmdFlags{}

	flag.StringVar(&flags.roleName, "role", "", "AWS Role to use in AWS config credentials")
	flag.StringVar(&flags.credentialSource, "credential", "Environment", "AWS Credential source. Valid values are: Ec2InstanceMetadata, Environment, EcsContainer")
	flag.StringVar(&flags.credentialPath, "path", "", "AWS Credentials file path")
	flag.StringVar(&flags.connectionsPath, "connections", "", "Steampipe AWS connections file path")
	flag.StringVar(&flags.importSchema, "schema", "enabled", "AWS Connection import schema. Valid values are: enabled, disabled")
	flag.StringVar(&flags.defaultRegion, "region", "", "AWS Connection default region")
	flag.StringVar(&flags.assumeRoleArn, "assume", "", "AWS Role to assume for getting Organization accounts")
	targetRegions := flag.String("regions", "all", "AWS Connection target regions")
	skipOUs := flag.String("skipOUs", "", "AWS OU IDs to skip from account connections")
	flag.Parse()

	if flags.roleName == "" {
		flag.Usage()
		return nil, fmt.Errorf("-role flag is required")
	}

	if flags.credentialSource != "Ec2InstanceMetadata" && flags.credentialSource != "Environment" && flags.credentialSource != "EcsContainer" {
		flag.Usage()
		return nil, fmt.Errorf("-credential flag doesn't contain a valid value")
	}

	if flags.importSchema != "enabled" && flags.importSchema != "disabled" {
		flag.Usage()
		return nil, fmt.Errorf("-schema flag doesn't contain a valid value")
	}

	if flags.credentialPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user's home directory: %v", err)
		}
		defaultCredentialPath := ".aws/"
		flags.credentialPath = filepath.Join(homeDir, defaultCredentialPath)
	}

	if flags.connectionsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user's home directory: %v", err)
		}
		defaultConnectionsPath := ".steampipe/config/"
		flags.connectionsPath = filepath.Join(homeDir, defaultConnectionsPath)
	}

	if flags.defaultRegion == "" {
		flags.defaultRegion = os.Getenv("AWS_REGION")
		if flags.defaultRegion == "" {
			flags.defaultRegion = "us-east-1"
			log.Warn("default region not defined, using:", flags.defaultRegion)
		} else {
			log.Info("default region not defined, using value from env AWS_REGION: ", flags.defaultRegion)
		}
	}

	if *targetRegions == "all" {
		flags.targetRegions = []string{"*"}
	} else {
		flags.targetRegions = strings.Split(*targetRegions, ",")
	}
	log.Debug("regions: ", flags.targetRegions)

	flags.skipOUs = strings.Split(*skipOUs, ",")
	log.Debug("skipOUs: ", flags.skipOUs)

	return &flags, nil
}
