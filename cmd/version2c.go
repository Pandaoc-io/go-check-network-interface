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
	"github.com/spf13/cobra"
)

// version2cCmd represents the version2c command
var version2cCmd = &cobra.Command{
	Use:   "version2c",
	Short: "SNMP request in version 2c",
	Long:  `Poll the interface information in SNMP version 2c`,
	Run: func(cmd *cobra.Command, args []string) {
		networkInterfaceCheck("2c", cmd)
	},
}

func init() {
	rootCmd.AddCommand(version2cCmd)

	version2cCmd.Flags().StringP("community", "c", "public", "SNMP community used for polling (required)")
	version2cCmd.MarkFlagRequired("community")
}
