package http

import (
	"context"
	"encoding/json"
	"mylife-home-common/log"
	"net/http"
	"net/url"
	"time"
)

var logger = log.CreateLogger("mylife:home:core:plugins:absoluta:engine:itv2:http")

var httpClient = &http.Client{Timeout: 10 * time.Second}

type status struct {
	Present          bool `json:"present"`
	Busy             bool `json:"busy"`
	AppStatus        uint `json:"appStatus"`
	BossStatus       uint `json:"bossStatus"`
	ClientsWithNotif uint `json:"clientsWithNotif"`
}

func CheckAvailability(ctx context.Context, uid string) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://mobile.absoluta.info/admin-api/device/status?id="+url.QueryEscape(uid), nil)
	if err != nil {
		logger.WithError(err).Error("error building request")
		return false
	}

	res, err := httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Warn("error contacting cloud server")
		return true
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logger.Errorf("cloud server responded with an error: (%d) %s", res.StatusCode, res.Status)
		return true
	}

	var s status
	err = json.NewDecoder(res.Body).Decode(&s)
	if err != nil {
		logger.WithError(err).Error("could not read cloud server response")
		return true
	}

	logger.Debugf("got status: %+v", s)

	if s.Busy {
		logger.Debug("cloud server reported busy panel")
		return false
	}

	return true
}
