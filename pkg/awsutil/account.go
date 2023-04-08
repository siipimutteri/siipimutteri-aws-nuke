package awsutil

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"github.com/rebuy-de/aws-nuke/v2/pkg/config"
)

type Account struct {
	Credentials

	id string
}

func NewAccount(creds Credentials, endpoints config.CustomEndpoints) (*Account, error) {
	creds.CustomEndpoints = endpoints
	account := Account{
		Credentials: creds,
	}

	customStackSupportSTSAndIAM := true
	if endpoints.GetRegion(DefaultRegionID) != nil {
		if endpoints.GetURL(DefaultRegionID, "sts") == "" {
			customStackSupportSTSAndIAM = false
		} else if endpoints.GetURL(DefaultRegionID, "iam") == "" {
			customStackSupportSTSAndIAM = false
		}
	}
	if !customStackSupportSTSAndIAM {
		account.id = "account-id-of-custom-region-" + DefaultRegionID
		return &account, nil
	}

	defaultSession, err := account.NewSession(DefaultRegionID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create default session in %s", DefaultRegionID)
	}

	identityOutput, err := sts.New(defaultSession).GetCallerIdentity(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed get caller identity")
	}

	account.id = *identityOutput.Account

	return &account, nil
}

func (a *Account) ID() string {
	return a.id
}

func (a *Account) ResourceTypeToServiceType(regionName, resourceType string) string {
	customRegion := a.CustomEndpoints.GetRegion(regionName)
	if customRegion == nil {
		return "-" // standard public AWS.
	}
	for _, e := range customRegion.Services {
		if strings.HasPrefix(strings.ToLower(resourceType), e.Service) {
			return e.Service
		}
	}
	return ""
}
