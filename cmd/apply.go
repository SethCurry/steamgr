/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/SethCurry/steamgr/internal/steamcmd"
	"github.com/SethCurry/steamgr/internal/steamgr"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		configsDir := "."

		if len(args) == 1 {
			configsDir = args[0]
		}

		systemdDir, err := cmd.Flags().GetString("systemd")
		if err != nil {
			logger.Fatal("failed to get systemd flag", zap.Error(err))
		}

		username, err := cmd.Flags().GetString("username")
		if err != nil {
			logger.Fatal("failed to get steam username flag", zap.Error(err))
		}

		factory := steamcmd.NewSessionFactory(username)

		if _, err = os.Stat(systemdDir); os.IsNotExist(err) {
			err = os.MkdirAll(systemdDir, 0755)
			if err != nil {
				logger.Fatal("failed to make systemd dir", zap.String("path", systemdDir), zap.Error(err))
			}
		}

		configsList, err := os.ReadDir(configsDir)
		if err != nil {
			logger.Fatal("failed to list configs", zap.String("dir", configsDir), zap.Error(err))
		}

		for _, v := range configsList {
			if v.IsDir() {
				logger.Debug("skipping directory", zap.String("name", v.Name()))

				continue
			}

			if filepath.Ext(v.Name()) != ".json" {
				logger.Debug("skipping non-JSON file", zap.String("name", v.Name()))

				continue
			}

			var config steamgr.ApplicationConfig

			fd, err := os.Open(filepath.Join(configsDir, v.Name()))
			if err != nil {
				logger.Fatal("failed to open config file", zap.String("name", v.Name()), zap.Error(err))
			}
			defer fd.Close()

			configContents, err := io.ReadAll(fd)
			if err != nil {
				logger.Fatal("failed to read config", zap.Error(err))
			}

			err = json.Unmarshal(configContents, &config)
			if err != nil {
				logger.Fatal("failed to unmarshal config", zap.String("name", v.Name()), zap.Error(err))
			}

			err = steamgr.ApplyApplicationConfig(context.Background(), &config, factory)
			if err != nil {
				logger.Fatal("failed to apply config", zap.String("name", v.Name()), zap.Error(err))
			}

			systemdUnitText, err := steamgr.BuildSystemdUnitFile(&config)
			if err != nil {
				logger.Fatal("failed to generate systemd unit file", zap.String("name", v.Name()), zap.Error(err))
			}

			err = os.WriteFile(filepath.Join(systemdDir, config.Name+".service"), []byte(systemdUnitText), 0644)
			if err != nil {
				logger.Fatal("failed to write systemd unit file", zap.String("name", v.Name()), zap.Error(err))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	applyCmd.Flags().StringP("systemd", "s", "./systemd", "The directory to put generated systemd configs in")
	applyCmd.Flags().StringP("username", "u", "", "The username to use for login")
}
