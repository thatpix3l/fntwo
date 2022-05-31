package cmd

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thatpix3l/fntwo/app"
	"github.com/thatpix3l/fntwo/cfg"
)

var (
	// Home of config and data files.
	// Neat and tidy according to freedesktop.org's base directory specifications.
	// Along with whatever Windows does, I guess...

	appName      = "fntwo"                  // Name of program. Duh...
	envPrefix    = strings.ToUpper(appName) // Prefix for all environment variables used for configuration
	cfgNameNoExt = "config"                 // Name of the default config file used, without an extension

	cfgHomePath  = path.Join(xdg.ConfigHome, appName)   // Default path to app's config directory
	cfgPathNoExt = path.Join(cfgHomePath, cfgNameNoExt) // Default path to app's config file, without extension
	dataHomePath = path.Join(xdg.DataHome, appName)     // Default path to app's runtime-related data files
)

// Entrypoint for command line
func Start() {
	cmd := newRootCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Take a command, create env variables that are mapped to most flags, load config
func initializeConfig(cmdFlags *pflag.FlagSet) {

	// Viper config that will be merged from different file sources and env variables
	v := viper.New()

	// Setting properties of the config file, before reading and processing
	v.SetConfigName(cfgNameNoExt) // Default config name, without extension
	v.AddConfigPath(cfgHomePath)  // Path to search for config files
	v.SetEnvPrefix(envPrefix)     // Prefix for all environment variables
	v.AutomaticEnv()              // Auto-check if any config keys match env keys

	// Create equivalent env var keys for each flag, replace value in flag if not
	// explicitly changed by the user on the command line
	cmdFlags.VisitAll(func(f *pflag.Flag) {

		// Config is a special case. We only want it to be configurable from the command line
		if f.Name != "config" {

			// Create an env var key mapped to a flag, e.g. "FOO_BAR" is created from "foo-bar".
			// Take same env var key name, and normalize it to env var naming specification, e.g. "FOO_BAR",
			// so when assigning FOO_BAR=baz, it maps to foo-bar
			envKey := envPrefix + "_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, envKey)

			// If current flag value has not been changed and viper config does have a value,
			// assign to flag the config value
			if !f.Changed && v.IsSet(f.Name) {
				flagVal := v.Get(f.Name)
				cmdFlags.Set(f.Name, fmt.Sprintf("%v", flagVal))
			}

		} else if f.Changed {
			// If config has been set by command line, set that to be loaded when reading
			v.SetConfigFile(f.Value.String())
		}

	})

	// Load config sources
	v.ReadInConfig()

}

// Loads and parses config files from different sources,
// parses them, and finally merges them together
func newRootCommand() *cobra.Command {

	// Config with all keys that will eventually be used by the actual application
	var runtimeCfg cfg.Keys

	// Base command of actual program
	rootCmd := &cobra.Command{
		Use:   appName,
		Short: `"v" for "fntwo"`,
		Long:  `An easy to use tool for loading, configuring and displaying your VTuber models`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

			// Load and merge config from different sources, based on command flags
			initializeConfig(cmd.Flags())

		},
		Run: func(cmd *cobra.Command, args []string) {

			// Entrypoint for actual program
			app.Start(&runtimeCfg)

		},
	}

	// Here, we start defining a load of flags
	rootFlags := rootCmd.Flags()
	rootFlags.StringVarP(&runtimeCfg.ConfigPath, "config", "c", cfgPathNoExt+".{json,yaml,toml,ini}", "Path to a config file.")
	rootFlags.StringVar(&runtimeCfg.VmcListenIP, "vmc-ip", "0.0.0.0", "Address to listen and receive on for VMC motion data")
	rootFlags.IntVar(&runtimeCfg.VmcListenPort, "vmc-port", 39540, "Port to listen and receive on for VMC motion data")
	rootFlags.StringVar(&runtimeCfg.WebListenIP, "web-ip", "127.0.0.1", "Address to serve frontend page on")
	rootFlags.IntVar(&runtimeCfg.WebListenPort, "web-port", 3579, "Port to serve frontend page on")
	rootFlags.IntVar(&runtimeCfg.ModelUpdateFrequency, "update-frequency", 60, "Times per second the live VRM model data is sent to each client")

	return rootCmd

}
