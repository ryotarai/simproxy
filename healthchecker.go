package simproxy

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Healthchecker struct {
	Logger    *log.Logger
	Backend   *Backend
	Balancer  Balancer
	Interval  time.Duration
	FallCount int
	RiseCount int

	active       bool
	errorCount   int
	successCount int
	client       http.Client
}

func (c *Healthchecker) Start() error {
	c.client = http.Client{
		Transport: http.DefaultTransport,
		Timeout:   c.Interval,
	}

	c.Logger.Printf("[healthchecker] [%s] start", c.Backend.HealthcheckURL)

	c.successCount = c.RiseCount - 1
	c.check() // sync

	go func() {
		for {
			after := time.After(c.Interval)
			c.check()
			<-after
		}
	}()

	return nil
}

func (c *Healthchecker) check() {
	res, err := c.request()
	if err != nil {
		c.onError(err.Error())
	} else {
		msg := fmt.Sprintf("status %d", res.StatusCode)
		if 200 <= res.StatusCode && res.StatusCode < 300 {
			c.onSuccess(msg)
		} else {
			c.onError(msg)
		}
	}
}

func (c *Healthchecker) request() (*http.Response, error) {
	req, err := http.NewRequest("GET", c.Backend.HealthcheckURL.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Healthchecker) onSuccess(msg string) {
	if c.active {
		c.errorCount = 0
		return
	}

	c.successCount++
	c.Logger.Printf("[healthchecker] [%s] success %d/%d: %s", c.Backend.HealthcheckURL, c.successCount, c.RiseCount, msg)
	if c.RiseCount <= c.successCount {
		c.Logger.Printf("[healthchecker] [%s] adding to balancer", c.Backend.HealthcheckURL)
		c.Balancer.AddBackend(c.Backend)
		c.active = true
	}
}

func (c *Healthchecker) onError(msg string) {
	if !c.active {
		c.Logger.Printf("[healthchecker] [%s] error: %s", c.Backend.HealthcheckURL, msg)
		c.successCount = 0
		return
	}

	c.errorCount++
	c.Logger.Printf("[healthchecker] [%s] error %d/%d: %s", c.Backend.HealthcheckURL, c.errorCount, c.FallCount, msg)
	if c.FallCount <= c.errorCount {
		c.Logger.Printf("[healthchecker] [%s] removing from balancer", c.Backend.HealthcheckURL)
		c.Balancer.RemoveBackend(c.Backend)
		c.active = false
	}
}
