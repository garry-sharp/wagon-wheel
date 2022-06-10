/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/garry-sharp/assessment/src/core"
	"github.com/garry-sharp/assessment/src/db"
	"github.com/garry-sharp/assessment/src/logger"
	"github.com/garry-sharp/assessment/src/web"
	"github.com/spf13/cobra"
)

var cfgFile string
var dbName string

//
func GetDBName() string {
	return dbName
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "assessment",
	Short: "Wagon Wheel. An application for the Kraken Wagon",
	Long: `
	 _    _                           _    _ _               _ 
	| |  | |                         | |  | | |             | |
	| |  | | __ _  __ _  ___  _ __   | |  | | |__   ___  ___| |
	| |/\| |/ _` + "`" + ` |/ _` + "`" + ` |/ _ \| '_ \  | |/\| | '_ \ / _ \/ _ \ |
	\  /\  / (_| | (_| | (_) | | | | \  /\  / | | |  __/  __/ |
	 \/  \/ \__,_|\__, |\___/|_| |_|  \/  \/|_| |_|\___|\___|_|
                       __/ |                                       
                      |___/                                        
	
	A CLI to get you all your favourite live price feeds!!
	
	`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		duration, _ := cmd.Flags().GetInt32("duration")
		quote, _ := cmd.Flags().GetString("quote")
		assets, _ := cmd.Flags().GetStringSlice("assets")
		dbfn, _ := cmd.Flags().GetString("dbfilename")
		dbName = dbfn
		webBool, _ := cmd.Flags().GetBool("web")
		logInt, _ := cmd.Flags().GetUint8("logs")
		logfilename, _ := cmd.Flags().GetString("logfilename")
		kafkaurl, _ := cmd.Flags().GetString("kafkaurl")

		logger.Init(logger.Config{
			WriteType:   logInt,
			KafkaURL:    kafkaurl,
			LogFilePath: logfilename,
		})
		defer logger.Destroy()

		if webBool {
			c := make(chan db.Price)
			if quote != "" && len(assets) != 0 && duration != 0 {
				go func() {
					core.Run(duration, quote, assets, dbfn, c)
				}()
			}
			web.Serve(c, dbfn)
		} else {
			core.Run(duration, quote, assets, dbfn, nil)
		}

		if len(assets) == 0 {
			logger.Log("At least one asset needs to be provided", logger.Fatal, logger.Price)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.Flags().Int32P("duration", "d", 5000, "A value in milliseconds of how often prices should be updated")
	rootCmd.Flags().StringP("quote", "q", "USD", "Prices will be returned in this quote format")
	rootCmd.Flags().StringSliceP("assets", "a", []string{}, "A comma separated list of assets to get prices for")
	rootCmd.Flags().String("dbfilename", "database.db", "A DB file to update (sqlite), if not provided one will be created")
	rootCmd.Flags().Bool("web", false, "Use this value to create a webserver. If other values are present the server will start collecting the prices automatically.")
	rootCmd.Flags().Uint8("logs", 1, "A Bitmapped number representing what to log, 1 being stdout, 2 being file, 4 being kafka. E.g. 5 would be stdout and kafta")
	rootCmd.Flags().String("logfilename", "output.log", "If logs & 2 == 1, then logs will be written to this file")
	rootCmd.Flags().String("kafkaurl", "", "If logs & 4 == 1, then log messages will be pushed to this kafka URL")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
