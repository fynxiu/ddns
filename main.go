package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/sirupsen/logrus"
)

type RecordType string

const (
	typeA RecordType = "A"
)

var (
	cachedIP = ""
)

const (
	envAliAccessKeyID     = "ALI_ACCESS_KEY_ID"
	envAliAccessKeySecret = "ALI_ACCESS_KEY_SECRET"
	envAliDomain          = "ALI_DOMAIN"
	envAliRR              = "ALI_RR"
)

func main() {
	var rr = os.Getenv(envAliRR)
	var domainName = os.Getenv(envAliDomain)
	var accessKeyID = os.Getenv(envAliAccessKeyID)
	var accessKeySecret = os.Getenv(envAliAccessKeySecret)
	flag.StringVar(&rr, "rr", rr, "rr='git ssh'")
	flag.StringVar(&domainName, "domain", domainName, "domain name， e.g. example.com")
	flag.StringVar(&accessKeyID, "key", accessKeyID, "access key id")
	flag.StringVar(&accessKeySecret, "secret", accessKeySecret, "access key secret")
	var checkIntervalRaw = flag.Int("ci", 2, "check interval unit minute")

	flag.Parse()

	records := strings.Split(rr, " ")
	checkInterval := time.Duration(*checkIntervalRaw) * time.Minute

	client, err := NewDNDClient(accessKeyID, accessKeySecret, domainName)
	if err != nil {
		panic(err)
	}
	origin := 5 * time.Second
	interval := origin
	for {
		ip, err := GetIP()
		// 获取ip失败, 在暂停interval后继续尝试
		if err != nil {
			logrus.WithError(err).Warningln("failed to get ip")
			time.Sleep(interval)
			interval *= 2
			continue
		} else {
			interval = origin
		}
		if ip != cachedIP {
			logrus.Infof("update domain : %s\n", ip)
			err := client.refreshAllRecords(records, ip)
			if err != nil {
				logrus.WithError(err).Errorln("failed to update records")
			} else {
				cachedIP = ip
			}
		}
		time.Sleep(checkInterval)
	}
}

func NewDNDClient(id string, secret string, domainName string) (*DnsClient, error) {
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", id, secret)
	return &DnsClient{client, domainName}, err
}

type DnsClient struct {
	*alidns.Client
	domainName string
}

func (c *DnsClient) refreshAllRecords(records []string, ipAddr string) error {
	for _, v := range records {
		_, err := c.AddOrUpdateDomainRecord(typeA, v, ipAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *DnsClient) GetDomainRecords() (*alidns.DescribeDomainRecordsResponse, error) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = c.domainName
	return c.DescribeDomainRecords(request)
}

func (c *DnsClient) GetRecord(rr string) (*alidns.Record, error) {
	response, err := c.GetDomainRecords()
	if err != nil {
		return nil, err
	}
	for _, v := range response.DomainRecords.Record {
		if v.RR == rr {
			return &v, nil
		}
	}
	return nil, errors.New("not found")
}

func (c *DnsClient) AddOrUpdateDomainRecord(recordType RecordType, rr string, ipAddr string) (interface{}, error) {
	record, err := c.GetRecord(rr)
	if err != nil && err.Error() != "not found" {
		return nil, err
	}

	if err != nil {
		request := alidns.CreateAddDomainRecordRequest()
		request.Scheme = "https"
		request.DomainName = c.domainName
		request.Type = string(recordType)
		request.RR = rr
		request.Value = ipAddr
		return c.Client.AddDomainRecord(request)
	} else {
		if record.Value == ipAddr {
			return nil, nil
		}
		request := alidns.CreateUpdateDomainRecordRequest()
		request.Scheme = "https"
		request.Type = string(recordType)
		request.RR = rr
		request.RecordId = record.RecordId
		request.Value = ipAddr
		return c.Client.UpdateDomainRecord(request)
	}
}
