package github

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v84/github"
)

type BatchTransferResult struct {
	TotalRepositories  int
	RequestedTransfers int
}

func TransferRepo(c *github.Client, sourceOwner, sourceRepo, destOwner string, newName *string, teamId []int64) (*github.Repository, *github.Response, error) {
	if c == nil {
		return nil, nil, fmt.Errorf("github client is nil")
	}

	context := context.Background()

	repo, res, err := c.Repositories.Transfer(context, sourceOwner, sourceRepo, github.TransferRequest{
		NewOwner: destOwner,
		NewName:  newName,
		TeamID:   teamId,
	})

	return repo, res, err
}

func TransferMultipleRepo(c *github.Client, sourceOwner, destOwner string, teamId []int64) (*BatchTransferResult, error) {
	allRepo, err := ListRepos(c, sourceOwner)
	if err != nil {
		return nil, err
	}

	result := &BatchTransferResult{}

	for _, repo := range allRepo {
		repoName := repo.GetName()
		result.TotalRepositories++

		confirmed, err := confirmTransfer(repoName, destOwner)
		if err != nil {
			return nil, err
		}

		if !confirmed {
			continue
		}

		result.RequestedTransfers++
		_, _, _ = TransferRepo(c, sourceOwner, repoName, destOwner, nil, teamId)
	}

	return result, nil
}

func confirmTransfer(repoName, destOwner string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Transfer %s to %s? [y/n]: ", repoName, destOwner)

		response, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println("Please answer y or n.")
		}
	}
}
