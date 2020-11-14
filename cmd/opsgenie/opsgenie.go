package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ricoberger/opsgenie/pkg/config"
	"github.com/ricoberger/opsgenie/pkg/opsgenie"
	"github.com/ricoberger/opsgenie/pkg/prompt"
	"github.com/ricoberger/opsgenie/pkg/version"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	configFile string
	limit      int
	logLevel   string
	logOutput  string
	query      string

	cfg config.Config
)

var rootCmd = &cobra.Command{
	Use:   "opsgenie",
	Short: "opsgenie - command line application to interact with Opsgenie.",
	Long:  "opsgenie - command line application to interact with Opsgenie.",
	Run: func(cmd *cobra.Command, args []string) {
		if logOutput == "json" {
			log.SetFormatter(&log.JSONFormatter{})
		} else {
			log.SetFormatter(&log.TextFormatter{})
		}

		lvl, err := log.ParseLevel(logLevel)
		if err != nil {
			log.WithError(err).Fatal("Could not set log level")
		}
		log.SetLevel(lvl)

		if lvl == log.DebugLevel {
			log.SetReportCaller(true)
		}

		log.Debugf(version.Info())
		log.Debugf(version.BuildContext())

		if configFile == "~/.opsgenie.yaml" {
			configFile = path.Join(os.Getenv("HOME"), ".opsgenie.yaml")
		}
		err = cfg.LoadConfig(configFile)
		if err != nil {
			log.WithError(err).Fatalf("Could not load configuration")
		}

		log.Debugf("Config loaded: %#v", cfg)

		for {
			alerts, err := opsgenie.GetAlerts(cfg, lvl, query, limit)
			if err != nil {
				log.WithError(err).Fatalf("Could not load alerts")
			}

			alert, err := prompt.SelectAlert(cfg, alerts)
			if err != nil {
				if err == promptui.ErrInterrupt {
					return
				}
				log.WithError(err).Fatalf("Could not select alert")
			}

			action, err := prompt.SelectAction(alert)
			if err != nil {
				if err == promptui.ErrInterrupt {
					return
				}
				log.WithError(err).Fatalf("Could not select action")
			}

			if action == "Cancel" {
				continue
			}

			var snoozeDuration time.Duration
			if action == "Snooze" {
				snoozeDuration, err = prompt.SetSnoozeDuration()
				if err != nil {
					if err == promptui.ErrInterrupt {
						return
					}
					log.WithError(err).Fatalf("Could set snooze duration")
				}
			}

			msg, err := opsgenie.AlertAction(cfg, lvl, alert, action, snoozeDuration)
			if err != nil {
				log.WithError(err).Fatalf("Could not apply action")
			}

			fmt.Println(msg)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information for opsgenie.",
	Long:  "Print version information for opsgenie.",
	Run: func(cmd *cobra.Command, args []string) {
		v, err := version.Print("opsgenie")
		if err != nil {
			log.WithError(err).Fatal("Failed to print version information")
		}

		fmt.Fprintln(os.Stdout, v)
		return
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "~/.opsgenie.yaml", "Path to the configuration file.")
	rootCmd.PersistentFlags().IntVar(&limit, "limit", 50, "Limit for the query results.")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "error", "Set the log level. Must be one of the following values: trace, debug, info, warn, error, fatal or panic.")
	rootCmd.PersistentFlags().StringVar(&logOutput, "log.output", "plain", "Set the output format of the log line. Must be plain or json.")
	rootCmd.PersistentFlags().StringVar(&query, "query", "status: open", "Query which should be used to get alerts.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Failed to initialize opsgenie")
	}
}
