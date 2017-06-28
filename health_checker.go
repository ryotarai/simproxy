package simproxy

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type HealthChecker struct {
	State     *HealthStateStore
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

func (c *HealthChecker) Start() error {
	c.client = http.Client{
		Transport: http.DefaultTransport,
		Timeout:   c.Interval,
	}

	if c.State.State(c.Backend.URL.String()) == HEALTH_STATE_HEALTHY {
		c.addToBalancer()
	}

	go func() {
		for {
			after := time.After(c.Interval)
			c.check()
			<-after
		}
	}()

	return nil
}

func (c *HealthChecker) check() {
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

func (c *HealthChecker) request() (*http.Response, error) {
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

func (c *HealthChecker) onSuccess(msg string) {
	if c.active {
		if c.errorCount > 0 {
			c.errorCount = 0
			c.logf("success: %s", msg)
		}
		return
	}

	c.successCount++
	c.logf("success %d/%d: %s", c.successCount, c.RiseCount, msg)
	if c.RiseCount <= c.successCount {
		err := c.State.Mark(c.Backend.URL.String(), HEALTH_STATE_HEALTHY)
		if err != nil {
			c.logf("error: %s", err)
		}
		c.addToBalancer()
	}
}

func (c *HealthChecker) addToBalancer() {
	c.logf("adding to balancer")
	c.Balancer.AddBackend(c.Backend)
	c.active = true
}

func (c *HealthChecker) onError(msg string) {
	if !c.active {
		c.logf("error: %s", msg)
		c.successCount = 0
		return
	}

	c.errorCount++
	c.logf("error %d/%d: %s", c.errorCount, c.FallCount, msg)
	if c.FallCount <= c.errorCount {
		c.State.Mark(c.Backend.URL.String(), HEALTH_STATE_DEAD)
		c.removeFromBalancer()
	}
}

func (c *HealthChecker) removeFromBalancer() {
	c.logf("removing from balancer")
	c.Balancer.RemoveBackend(c.Backend)
	c.active = false
}

func (c *HealthChecker) logf(format string, args ...interface{}) {
	c.Logger.Printf(fmt.Sprintf("[healthchecker] [%s] ", c.Backend.HealthcheckURL)+format, args...)
}
