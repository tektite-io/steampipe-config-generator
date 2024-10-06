package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	log "github.com/sirupsen/logrus"
)

type OrganizationAccount struct {
	Name      string
	AccountID string
	Tags      map[string]string
	AccountOU string
}

func (c *OrganizationsClient) listOrganizationAccountTags(accountId *string) (map[string]string, error) {
	tags := make(map[string]string)

	params := &organizations.ListTagsForResourceInput{
		ResourceId: accountId,
	}
	paginator := organizations.NewListTagsForResourcePaginator(c.client, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.Background())
		if err != nil {
			return tags, fmt.Errorf("failed to list tags: %v", err)
		}

		for _, tag := range resp.Tags {
			tags[*tag.Key] = *tag.Value
		}
	}

	return tags, nil
}

func (c *OrganizationsClient) getAccountOU(accountId *string) (string, error) {

	params := &organizations.ListParentsInput{
		ChildId: accountId,
	}

	resp, err := c.client.ListParents(context.Background(), params)
	if err != nil {
		return "", fmt.Errorf("failed to get accountOU: %v", err)
	}

	if len(resp.Parents) > 0 {
		return *resp.Parents[0].Id, nil
	}

	return "", fmt.Errorf("no parent OU found for account ID: %s", *accountId)
}

func (c *OrganizationsClient) ListOrganizationAccounts() ([]OrganizationAccount, error) {
	params := &organizations.ListAccountsInput{}
	paginator := organizations.NewListAccountsPaginator(c.client, params)

	var allAccounts []OrganizationAccount

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to list accounts: %v", err)
		}

		for _, acc := range resp.Accounts {
			if string(acc.Status) == "ACTIVE" {
				account := OrganizationAccount{
					Name:      *acc.Name,
					AccountID: *acc.Id,
				}
				allAccounts = append(allAccounts, account)
			}
		}
	}

	semAccountTags := make(chan struct{}, 50)
	semAccountOU := make(chan struct{}, 3)

	var wgAccountTags sync.WaitGroup
	for i, acc := range allAccounts {
		wgAccountTags.Add(1)

		go func(i int, accountId string) {
			semAccountTags <- struct{}{}
			defer func() { <-semAccountTags }()
			defer wgAccountTags.Done()

			log.Debug("getting tags for account: ", accountId)
			tags, err := c.listOrganizationAccountTags(&accountId)
			if err != nil {
				log.Errorf("error retrieving organization accounts: %v", err)
				return
			}

			allAccounts[i].Tags = tags
		}(i, acc.AccountID)
	}

	var wgAccountOUs sync.WaitGroup
	for i, acc := range allAccounts {
		wgAccountOUs.Add(1)

		go func(i int, accountId string) {
			semAccountOU <- struct{}{}
			defer func() { <-semAccountOU }()
			defer wgAccountOUs.Done()

			log.Debug("getting OU for account: ", accountId)
			accountOU, err := c.getAccountOU(&accountId)
			if err != nil {
				log.Errorf("error retrieving accountOU: %v", err)
				return
			}

			allAccounts[i].AccountOU = accountOU
		}(i, acc.AccountID)
	}

	wgAccountTags.Wait()
	wgAccountOUs.Wait()

	return allAccounts, nil
}
