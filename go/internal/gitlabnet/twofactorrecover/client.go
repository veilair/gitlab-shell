package twofactorrecover

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/discover"
)

type Client struct {
	config *config.Config
	client *gitlabnet.GitlabClient
}

type Response struct {
	Success       bool     `json:"success"`
	RecoveryCodes []string `json:"recovery_codes"`
	Message       string   `json:"message"`
}

type RequestBody struct {
	KeyId  string `json:"key_id,omitempty"`
	UserId int64  `json:"user_id,omitempty"`
}

func NewClient(config *config.Config) (*Client, error) {
	client, err := gitlabnet.GetClient(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating http client: %v", err)
	}

	return &Client{config: config, client: client}, nil
}

func (c *Client) GetRecoveryCodes(args *commandargs.Shell) ([]string, error) {
	requestBody, err := c.getRequestBody(args)

	if err != nil {
		return nil, err
	}

	response, err := c.client.Post("/two_factor_recovery_codes", requestBody)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return parse(response)
}

func parse(hr *http.Response) ([]string, error) {
	response := &Response{}
	if err := gitlabnet.ParseJSON(hr, response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New(response.Message)
	}

	return response.RecoveryCodes, nil
}

func (c *Client) getRequestBody(args *commandargs.Shell) (*RequestBody, error) {
	client, err := discover.NewClient(c.config)

	if err != nil {
		return nil, err
	}

	var requestBody *RequestBody
	if args.GitlabKeyId != "" {
		requestBody = &RequestBody{KeyId: args.GitlabKeyId}
	} else {
		userInfo, err := client.GetByCommandArgs(args)

		if err != nil {
			return nil, err
		}

		requestBody = &RequestBody{UserId: userInfo.UserId}
	}

	return requestBody, nil
}
