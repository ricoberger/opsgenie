package opsgenie

import (
	"fmt"
	"github.com/ricoberger/opsgenie/pkg/config"
	"time"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	log "github.com/sirupsen/logrus"
)

func GetAlerts(cfg config.Config, lvl log.Level, query string, limit int) ([]alert.GetAlertResult, error) {
	alertClient, err := alert.NewClient(&client.Config{
		ApiKey:         cfg.ApiKey,
		OpsGenieAPIURL: client.ApiUrl(cfg.ApiUrl),
		LogLevel:       lvl,
	})
	if err != nil {
		return nil, err
	}

	res, err := alertClient.List(nil, &alert.ListAlertRequest{
		Limit: limit,
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	var alerts []alert.GetAlertResult

	for _, a := range res.Alerts {
		alertRes, err := alertClient.Get(nil, &alert.GetAlertRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: a.Id,
		})
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, *alertRes)
	}

	return alerts, nil
}

func AlertAction(cfg config.Config, lvl log.Level, a alert.GetAlertResult, action string, snoozeDuration time.Duration) (string, error) {
	alertClient, err := alert.NewClient(&client.Config{
		ApiKey:         cfg.ApiKey,
		OpsGenieAPIURL: client.ApiUrl(cfg.ApiUrl),
		LogLevel:       lvl,
	})
	if err != nil {
		return "", err
	}

	if action == "Acknowledge" {
		_, err = alertClient.Acknowledge(nil, &alert.AcknowledgeAlertRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: a.Id,
			User:            cfg.User,
		})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Alert '%s' is acknowledge.", a.Message), nil
	} else if action == "Close" {
		_, err = alertClient.Close(nil, &alert.CloseAlertRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: a.Id,
			User:            cfg.User,
		})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Alert '%s' is closed.", a.Message), nil
	} else if action == "Snooze" {
		_, err = alertClient.Snooze(nil, &alert.SnoozeAlertRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: a.Id,
			EndTime:         time.Now().Add(snoozeDuration),
			User:            cfg.User,
		})
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Alert '%s' is snoozed until %s.", a.Message, time.Now().Add(snoozeDuration).Format("2006-02-01 15:04")), nil
	}

	return "", nil
}
