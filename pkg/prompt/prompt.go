package prompt

import (
	"errors"
	"strings"
	"time"

	"github.com/ricoberger/opsgenie/pkg/config"

	"github.com/manifoldco/promptui"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
)

func SelectAlert(cfg config.Config, alerts []alert.GetAlertResult) (alert.GetAlertResult, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   cfg.Templates.Active,
		Inactive: cfg.Templates.Inactive,
		Selected: cfg.Templates.Selected,
		Details:  cfg.Templates.Details,
	}

	searcher := func(input string, index int) bool {
		a := alerts[index]
		name := strings.Replace(strings.ToLower(a.Message), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Alerts",
		Items:     alerts,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
		Stdout:    &bellSkipper{},
	}

	i, _, err := prompt.Run()

	if err != nil {
		return alert.GetAlertResult{}, err
	}

	return alerts[i], nil
}

func SelectAction(a alert.GetAlertResult) (string, error) {
	prompt := promptui.Select{
		Label:  "Action for " + a.Message,
		Items:  []string{"Acknowledge", "Close", "Snooze", "Quit Opsgenie"},
		Stdout: &bellSkipper{},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func SetSnoozeDuration() (time.Duration, error) {
	validate := func(input string) error {
		_, err := time.ParseDuration(input)
		if err != nil {
			return errors.New("invalid duration")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Duration",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return time.ParseDuration(result)
}
