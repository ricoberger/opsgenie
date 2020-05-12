package opsgenie

import (
	"fmt"
	"sync"
	"time"

	"github.com/ricoberger/opsgenie/pkg/config"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	log "github.com/sirupsen/logrus"
)

// GetAlerts returns all alerts including there details for the given query.
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
	var waitgroup sync.WaitGroup

	for _, a := range res.Alerts {
		waitgroup.Add(1)

		go func(a alert.Alert) {
			alertRes, err := alertClient.Get(nil, &alert.GetAlertRequest{
				IdentifierType:  alert.ALERTID,
				IdentifierValue: a.Id,
			})
			if err != nil {
				log.WithError(err).Errorf("Could not get alert details for %s (%s)", a.Message, a.TinyID)
			}
			alerts = append(alerts, *alertRes)
			waitgroup.Done()
		}(a)
	}

	waitgroup.Wait()

	return alerts, nil
}

// AlertAction applies the given action to the given alert. Valid actions are "Acknowledge", "Close" and "Snooze".
// If the action is "Snooze" the duration until the alert should be snoozed is required.
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
