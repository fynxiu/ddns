package main

import (
	"errors"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

type dnsClient struct {
	*dns.Client
	resolvers []*dnsResolver
}

type dnsResolver struct {
	questionDomain string
	host           string
}

func DefaultDnsClient() *dnsClient {
	return &dnsClient{
		Client: new(dns.Client),
		resolvers: []*dnsResolver{
			{
				"myip.opendns.com",
				"resolver1.opendns.com:53",
			},
			{
				"o-o.myaddr.l.google.com",
				"ns1.google.com:53",
			},
		},
	}
}

// GetIP Get public ip address
func GetIP() (string, error) {
	return DefaultDnsClient().getIP()
}

func (c *dnsClient) getIP() (ip string, err error) {
	for _, r := range c.resolvers {
		msg := newDnsMsg(r.questionDomain)
		resp, _, err := c.Exchange(msg, r.host)
		if err != nil {
			logrus.WithError(err).Infoln(111)
			continue
		}

		if resp.Rcode != dns.RcodeSuccess {
			continue
		}

		ip, err := extractIP(resp)
		if err != nil {
			logrus.WithError(err).Infoln(222)
			continue
		}
		return ip, nil
	}
	return "", errors.New("failed to query ip")
}

func newDnsMsg(questionDomain string) *dns.Msg {
	msg := new(dns.Msg)
	msg.RecursionDesired = false
	msg.SetQuestion(dns.Fqdn(questionDomain), dns.TypeANY)
	return msg
}

func extractIP(msg *dns.Msg) (ip string, err error) {
	for _, rr := range msg.Answer {
		if t, ok := rr.(*dns.TXT); ok {
			return t.Txt[0], nil
		}

		if a, ok := rr.(*dns.A); ok {
			return a.A.String(), nil
		}
	}

	return "", errors.New("failed to extract ip")
}
