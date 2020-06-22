package cmd

/*
go-shinken-check
Copyright © 2020 pandaoc-io <nicolas.bertaina@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "interface-snmp",
	Short: "Check network interface status and stats through SNMP",
	Long: `interface-snmp is a check build for Shinken and compatible with all the Nagios-like.
It check the interface oper/admin status and all the stats below :
- In/out Bandwidth in bps/%
- In/Out Errors in pps/%
- In/Out Discards in pps/%
- Half/Full duplex

It will also grap some additional informations like:
- Speed of the interface (Mb)
- Interface description also know as Alias
- Nb of pps per flow type in In/Out (Unicast/Multicast/Broadcast)
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func init() {
	rootCmd.Version = "0.1"
	rootCmd.PersistentFlags().StringP("hostname", "H", "127.0.0.1", "IP address or FQDN on which poll the information")
	rootCmd.PersistentFlags().StringP("interface", "i", "lo", "Interface name on which to grap the information (required)")
	rootCmd.PersistentFlags().IntP("timeout", "t", 5, "Timeout of the SNMP requests")
	rootCmd.PersistentFlags().IntP("retry", "r", 3, "Number of retry of the SNMP requests")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Activate the verbose mode to display more debuging information")

	rootCmd.PersistentFlags().Float64("bandwidth-warning", 80, "Warning threshold of the Bandwidth usage (in %%)")
	rootCmd.PersistentFlags().Float64("bandwidth-critical", 90, "Critical threshold of the Bandwidth usage (in %%)")

	rootCmd.PersistentFlags().String("error-warning", "50pps", "Warning threshold of the Errors (in %% or pps)")
	rootCmd.PersistentFlags().String("error-critical", "100pps", "Critical threshold of the Errors (in %% or pps)")

	rootCmd.PersistentFlags().String("discard-warning", "50pps", "Warning threshold of the Bandwidth usage (in %% or pps)")
	rootCmd.PersistentFlags().String("discard-critical", "100pps", "Critical threshold of the Bandwidth usage(in %% or pps)")

	rootCmd.PersistentFlags().Int("index-expiration", 60, "Expiration of the interfaces index file.")
}
