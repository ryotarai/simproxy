package simproxy

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Healthchecker struct {
	Backend   *Backend
	Balancer  Balancer
	Interval  time.Duration
	FallCount int
	RiseCount int

	active       bool
	errorCount   int
	successCount int
}

func (c *Healthchecker) Start() error {
	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout:   c.Interval,
	}

	req, err := http.NewRequest("GET", c.Backend.HealthcheckURL.String(), nil)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("[healthchecker] starting %s", c.Backend.HealthcheckURL)
		for {
			after := time.After(c.Interval)
			res, err := client.Do(req)
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
			<-after
		}
	}()

	return nil
}

func (c *Healthchecker) onSuccess(msg string) {
	if c.active {
		c.errorCount = 0
		return
	}

	c.successCount++
	log.Printf("[healthchecker] [%s] success %d/%d: %s", c.Backend.HealthcheckURL, c.successCount, c.RiseCount, msg)
	if c.RiseCount <= c.successCount {
		log.Printf("[healthchecker] [%s] adding to balancer", c.Backend.HealthcheckURL)
		c.Balancer.AddBackend(c.Backend)
		c.active = true
	}
}

func (c *Healthchecker) onError(msg string) {
	if !c.active {
		log.Printf("[healthchecker] [%s] error: %s", c.Backend.HealthcheckURL, msg)
		c.successCount = 0
		return
	}

	c.errorCount++
	log.Printf("[healthchecker] [%s] error %d/%d: %s", c.Backend.HealthcheckURL, c.errorCount, c.FallCount, msg)
	if c.FallCount <= c.errorCount {
		log.Printf("[healthchecker] [%s] removing from balancer", c.Backend.HealthcheckURL)
		c.Balancer.RemoveBackend(c.Backend)
		c.active = false
	}
}
