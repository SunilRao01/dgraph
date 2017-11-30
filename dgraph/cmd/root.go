/*
 * Copyright (C) 2017 Dgraph Labs, Inc. and Contributors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/dgraph/dgraph/cmd/bulk"
	"github.com/dgraph-io/dgraph/dgraph/cmd/live"
	"github.com/dgraph-io/dgraph/dgraph/cmd/server"
	"github.com/dgraph-io/dgraph/dgraph/cmd/zero"
	"github.com/dgraph-io/dgraph/x"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dgraph",
	Short: "Dgraph: Distributed Graph Database",
	Long: `
Dgraph is an open source, horizontally scalable and distributed graph database,
providing ACID transactions, consistent replication and linearizable reads.
It's built from ground up to perform for a rich set of queries. Being a native
graph database, it tightly controls how the data is arranged on disk to optimize
for query performance and throughput, reducing disk seeks and network calls in a
cluster.
` + x.BuildDetails(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	profileMode := viper.GetString("profile_mode")
	switch profileMode {
	case "cpu":
		defer profile.Start(profile.CPUProfile).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile).Stop()
	case "block":
		blockRate := viper.GetInt("block_rate")
		runtime.SetBlockProfileRate(blockRate)
		defer profile.Start(profile.BlockProfile).Stop()
	case "":
		// do nothing
	default:
		fmt.Printf("Invalid profile mode: %q\n", profileMode)
		os.Exit(1)
	}
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viper.SetEnvPrefix("DGRAPH")
	viper.AutomaticEnv()

	for _, cmd := range []*cobra.Command{
		bulk.BulkCmd, live.LiveCmd, server.ServerCmd, zero.ZeroCmd,
	} {
		RootCmd.AddCommand(cmd)
		viper.BindPFlags(cmd.Flags())
	}

	RootCmd.PersistentFlags().String("profile_mode", "",
		"Enable profiling mode, one of [cpu, mem, mutex, block]")
	RootCmd.PersistentFlags().Int("block_rate", 0,
		"Block profiling rate. Must be used along with block profile_mode")
	RootCmd.PersistentFlags().String("config", "",
		"Configuration file. Takes precedence over default values, but is "+
			"overridden to values set with environment variables and flags.")
	viper.BindPFlags(RootCmd.PersistentFlags())

	cobra.OnInitialize(func() {
		if cfg := viper.GetString("config"); cfg != "" {
			viper.SetConfigFile(cfg)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Println("Could not read config:", err)
				os.Exit(1)
			}
		}
	})
}
