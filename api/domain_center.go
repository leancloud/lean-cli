package api

import (
	"time"
)

type Type string

const (
	Platform      Type = "platform"
	File          Type = "file"
	Engine        Type = "engine"
	EngineCdn     Type = "engine-cdn"
	EnginePreview Type = "engine-preview"
	Billboard     Type = "billboard"
)

type SSLType string

const (
	None      SSLType = "none"
	Automatic SSLType = "automatic"
	Uploaded  SSLType = "uploaded"
)

type State string

const (
	VerifyingIcp   State = "verifyingIcp"
	VerifyingCname State = "verifyingCname"
	IssuingCert    State = "issuingCert"
	Normal         State = "normal"
	Suspended      State = "suspended"
	Failed         State = "failed"
)

type DomainBinding struct {
	Type                  Type       `json:"type"`
	AppId                 string     `json:"appId"`
	GroupName             string     `json:"groupName"`
	Domain                string     `json:"domain"`
	CnameTarget           *string    `json:"cnameTarget"`
	IcpLicense            string     `json:"icpLicense"`
	State                 State      `json:"state"`
	FailedReason          *string    `json:"failedReason"`
	SslType               SSLType    `json:"sslType"`
	SslExpiredAt          *time.Time `json:"sslExpiredAt"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	MultiAppsOnThisDomain bool       `json:"multiAppsOnThisDomain"`
	SharedDomain          bool       `json:"sharedDomain"`
	DedicatedIPs          []string   `json:"dedicatedIPs"`
	ForceHttps            int        `json:"forceHttps"`
}

func GetDomainBindings(appID string, domainType Type, groupName string) ([]DomainBinding, error) {
	client := NewClientByApp(appID)
	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	url := "/1.1/domain-center/domain-bindings?type=" + string(domainType)
	resp, err := client.get(url, opts)
	if err != nil {
		return nil, err
	}

	var domainBindings []DomainBinding
	err = resp.JSON(&domainBindings)

	if err != nil {
		return nil, err
	}

	i := 0
	for _, domain := range domainBindings {
		if domain.GroupName == groupName {
			domainBindings[i] = domain
			i += 1
		}
	}
	domainBindings = domainBindings[:i]

	return domainBindings, err
}
