package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/madalinpopa/gocost-web/internal/infrastructure/config"
	"github.com/spf13/cobra"
)

var (
	logger *slog.Logger
	conf   *config.Config
	dsn    string
)

var rootCmd = &cobra.Command{
	Use: "gocost",
	Long: `GoCost is a simple and efficient cost tracking application built with Go.
For more information, visit https://gocost.app`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		conf = config.New().WithDatabaseDsn(dsn)

		err := conf.LoadEnvironments()
		if err != nil {
			logger.Error("failed to load environments", slog.String("error", err.Error()))
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dsn, "dsn", "data.sqlite", "database connection string")

	rootCmd.AddCommand(migrateCmd)

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
