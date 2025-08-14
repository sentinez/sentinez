package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type Instance struct {
	Name    string
	Address string
	Port    int
}

type Client struct {
	consul *api.Client
}

func New(address string) (*Client, error) {
	conf := api.DefaultConfig()
	conf.Address = address

	c, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{consul: c}, nil
}

func (c *Client) Register(
	id, name, address string,
	port int,
	healthPath string,
) error {
	url := fmt.Sprintf("http://%s:%d%s", address, port, healthPath)
	check := &api.AgentServiceCheck{
		HTTP:                           url,
		Interval:                       "10s",
		Timeout:                        "2s",
		DeregisterCriticalServiceAfter: "1m",
	}

	reg := &api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Check:   check,
	}

	return c.consul.Agent().ServiceRegister(reg)
}

func (c *Client) RegisterWithTTL(
	id, name, address string,
	port int,
	ttl time.Duration,
) error {
	check := &api.AgentServiceCheck{
		TTL:                            ttl.String(),
		DeregisterCriticalServiceAfter: "1m",
	}

	reg := &api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Check:   check,
	}

	return c.consul.Agent().ServiceRegister(reg)
}

func (c *Client) SendHeartbeat(checkID string) error {
	return c.consul.Agent().PassTTL("service:"+checkID, "alive")
}

func (c *Client) Discover(name string) ([]Instance, error) {
	services, _, err := c.consul.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}

	var result []Instance
	for _, entry := range services {
		svc := entry.Service
		result = append(result, Instance{
			Name:    svc.Service,
			Address: svc.Address,
			Port:    svc.Port,
		})
	}
	return result, nil
}
