package snmp

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
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	g "github.com/soniah/gosnmp"
	"github.com/spf13/cobra"
)

// CreateConnection create the SNMP connection depending on the version provided
func CreateConnection(version string, cmd *cobra.Command) (*g.GoSNMP, error) {
	timeout, _ := cmd.Flags().GetInt("timeout")
	retry, _ := cmd.Flags().GetInt("retry")
	params := &g.GoSNMP{
		Target:  cmd.Flag("hostname").Value.String(),
		Port:    161,
		Timeout: time.Duration(timeout) * time.Second,
		Retries: retry,
	}
	log.Debugf("Polling in version %v\n", version)
	switch version {
	case "2c":
		params.Version = g.Version2c
		params.Community = cmd.Flag("community").Value.String()
	case "3":
		params.Version = g.Version3
		params.SecurityModel = g.UserSecurityModel
		var authProto = g.SHA
		var privProto = g.AES
		var secLevel = g.AuthPriv

		switch cmd.Flag("auth-protocol").Value.String() {
		case "MD5":
			authProto = g.MD5
		case "SHA":
			authProto = g.SHA
		default:
			return nil, fmt.Errorf("%v is not a valid authentication protocol, check usage", cmd.Flag("auth-protocol").Value.String())
		}

		switch cmd.Flag("priv-protocol").Value.String() {
		case "DES":
			privProto = g.DES
		case "AES":
			privProto = g.AES
		default:
			return nil, fmt.Errorf("%v is not a valid privacy protocol, check usage", cmd.Flag("priv-protocol").Value.String())
		}

		switch cmd.Flag("sec-level").Value.String() {
		case "noAuthNoPriv":
			secLevel = g.NoAuthNoPriv
			authProto = g.NoAuth
			privProto = g.NoPriv
		case "authNoPriv":
			secLevel = g.AuthNoPriv
			privProto = g.NoPriv
		case "authPriv":
			secLevel = g.AuthPriv
		default:
			return nil, fmt.Errorf("%v is not a valid security level, check usage", cmd.Flag("sec-level").Value.String())
		}

		params.MsgFlags = secLevel
		params.SecurityParameters = &g.UsmSecurityParameters{
			UserName:                 cmd.Flag("username").Value.String(),
			AuthenticationProtocol:   authProto,
			AuthenticationPassphrase: cmd.Flag("auth-passphrase").Value.String(),
			PrivacyProtocol:          privProto,
			PrivacyPassphrase:        cmd.Flag("priv-passphrase").Value.String(),
		}
	}

	err := params.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}

	if version == "3" {
		authRes, err := params.Get([]string{"1.3.6.1.2.1.1.1.0"})
		if err != nil {
			return nil, err
		}
		if authRes.Error != g.NoError {
			return nil, errors.New(authRes.Error.String())
		}

		if err := getErrorFromVariables(authRes.Variables); err != nil {
			return nil, err
		}
	}

	return params, nil
}

func oidToError(name string) (err error) {
	switch name {
	case ".1.3.6.1.6.3.15.1.1.3.0":
		// usmStatsUnknownUserNames
		err = errors.New("unknown user name")
	case ".1.3.6.1.6.3.15.1.1.5.0":
		// usmStatsWrongDigests
		err = errors.New("wrong digests, possibly wrong password")
	case "1.3.6.1.6.3.15.1.1.1.0":
		err = errors.New("wrong or unavailable securityLevel")
	}
	return
}

func getErrorFromVariables(packets []g.SnmpPDU) error {
	for _, variable := range packets {
		if err := oidToError(variable.Name); err != nil {
			return err
		}
	}
	return nil
}
