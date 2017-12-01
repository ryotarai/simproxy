package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ryotarai/simproxy/balancer"
	"github.com/ryotarai/simproxy/types"
	"github.com/sirupsen/logrus"
)

type HealthChecker struct {
	State     HealthStateStore
	Logger    *logrus.Logger
	Backend   *types.Backend
	Balancer  balancer.Balancer
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
	req.Header.Set("user-agent", "Simproxy")
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
			c.infof("Healthcheck success: %s", msg)
		}
		return
	}

	c.successCount++
	c.infof("success %d/%d: %s", c.successCount, c.RiseCount, msg)
	if c.RiseCount <= c.successCount {
		err := c.State.Mark(c.Backend.URL.String(), HEALTH_STATE_HEALTHY)
		if err != nil {
			c.errorf("Healthcheck error: %s", err)
		}
		c.addToBalancer()
	}
}

func (c *HealthChecker) addToBalancer() {
	c.infof("Healthchecker is adding a backend to balancer")
	c.Balancer.AddBackend(c.Backend)
	c.active = true
}

func (c *HealthChecker) onError(msg string) {
	if !c.active {
		c.warnf("Healthcheck error: %s", msg)
		c.successCount = 0
		return
	}

	c.errorCount++
	c.warnf("Healthcheck error %d/%d: %s", c.errorCount, c.FallCount, msg)
	if c.FallCount <= c.errorCount {
		c.State.Mark(c.Backend.URL.String(), HEALTH_STATE_DEAD)
		c.removeFromBalancer()
	}
}

func (c *HealthChecker) removeFromBalancer() {
	c.warnf("Healthchecker is removing a backend from balancer")
	c.Balancer.RemoveBackend(c.Backend)
	c.active = false
}

func (c *HealthChecker) logentry() *logrus.Entry {
	return c.Logger.WithField("backendURL", c.Backend.URL.String()).WithField("healthcheckURL", c.Backend.HealthcheckURL.String())
}

func (c *HealthChecker) infof(format string, args ...interface{}) {
	if c.Logger != nil {
		c.logentry().Infof(format, args...)
	}
}

func (c *HealthChecker) warnf(format string, args ...interface{}) {
	if c.Logger != nil {
		c.logentry().Warnf(format, args...)
	}
}

func (c *HealthChecker) errorf(format string, args ...interface{}) {
	if c.Logger != nil {
		c.logentry().Errorf(format, args...)
	}
}
