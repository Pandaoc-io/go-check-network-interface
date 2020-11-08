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

// version3Cmd represents the version3 command
var version3Cmd = &cobra.Command{
	Use:   "version3",
	Short: "SNMP request in version 3",
	Long:  `Poll the interface information in SNMP version 3`,
	Run: func(cmd *cobra.Command, args []string) {
		networkInterfaceCheck("3", cmd)
	},
}

func init() {
	rootCmd.AddCommand(version3Cmd)

	version3Cmd.Flags().StringP("username", "u", "admin", "Username used for SNMP v3 authentication.")
	version3Cmd.Flags().StringP("auth-protocol", "a", "SHA", "Authentication protocol (MD5|SHA|SHA-224|SHA-256|SHA-384|SHA-512).")
	version3Cmd.Flags().StringP("auth-passphrase", "A", "passphrase", "Authentication passphrase.")
	version3Cmd.Flags().StringP("sec-level", "l", "authPriv", "Security level (noAuthNoPriv|authNoPriv|authPriv).")
	version3Cmd.Flags().StringP("context", "n", "", "Context name.")
	version3Cmd.Flags().StringP("priv-protocol", "x", "AES", "Privacy protocol (DES|AES).")
	version3Cmd.Flags().StringP("priv-passphrase", "X", "passphrase", "Privacy passphrase.")
}
