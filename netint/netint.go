package netint

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
	"reflect"
	"strconv"

	"go-check-network-interface/convert"

	log "github.com/sirupsen/logrus"
	g "github.com/soniah/gosnmp"
)

// 1-up, 2-down, 3-testing, 4-unknown, 5-dormant, 6-notPresent, 7-lowerLayerDown
const (
	UP = iota + 1
	DOWN
	TESTING
	UNKNOWN
	DORMANT
	NOTPRESENT
	LOWERLAYERDOWN
)

//InterfaceDetails is the type hosting all the network interface information
type InterfaceDetails struct {
	UpTime                *uint
	Timestamp             int64
	Index                 *int
	IfName                *string
	IfDescr               *string
	IfAlias               *string
	IfSpeed               *uint
	IfAdminStatus         *uint
	IfOperStatus          *uint
	IfInOctets            *uint
	IfInRate              *float64
	IfInPrct              *float64
	IfInUcastPkts         *uint
	IfInNUcastPkts        *uint
	IfInDiscards          *uint
	IfInDiscardsRate      *float64
	IfInDiscardsPrct      *float64
	IfInErrors            *uint
	IfInErrorsRate        *float64
	IfInErrorsPrct        *float64
	IfOutOctets           *uint
	IfOutRate             *float64
	IfOutPrct             *float64
	IfOutUcastPkts        *uint
	IfOutNUcastPkts       *uint
	IfOutDiscards         *uint
	IfOutDiscardsRate     *float64
	IfOutDiscardsPrct     *float64
	IfOutErrors           *uint
	IfOutErrorsRate       *float64
	IfOutErrorsPrct       *float64
	IfHCInOctets          *uint
	IfHCInUcastPkts       *uint
	IfHCInMulticastPkts   *uint
	IfHCInBroadcastPkts   *uint
	IfInTotalPkts         *uint
	IfInTotalPktsRate     *float64
	IfHCOutOctets         *uint
	IfHCOutUcastPkts      *uint
	IfHCOutMulticastPkts  *uint
	IfHCOutBroadcastPkts  *uint
	InUniPcktRate         *float64
	InMultiPcktRate       *float64
	InBroadPcktRate       *float64
	OutUniPcktRate        *float64
	OutMultiPcktRate      *float64
	OutBroadPcktRate      *float64
	IfOutTotalPkts        *uint
	IfOutTotalPktsRate    *float64
	IfHighSpeed           *uint
	LocIfInCRC            *uint
	LocIfInCRCRate        *float64
	LocIfInCRCPrct        *float64
	Dot3StatsDuplexStatus *uint
	SpeedInbit            *uint
}

//GetData is used to get a specific data of a network interface.
//The available elements can be found into the InterfaceOids Map.
func (i *InterfaceDetails) GetData(snmpConnection *g.GoSNMP, elem string) error {
	log.Debug("=====================")
	log.Debugf("GetData information : %v", elem)
	oid := []string{InterfaceOids[elem] + "." + strconv.Itoa(*i.Index)}
	result, err := snmpConnection.Get(oid)
	if err != nil {
		return err
	}

	switch result.Variables[0].Type {
	case g.OctetString:
		bytes := result.Variables[0].Value.([]byte)
		if reflect.ValueOf(i).Elem().FieldByName(elem).Kind() == reflect.Ptr {
			strbytes := string(bytes)
			reflect.ValueOf(i).Elem().FieldByName(elem).Set(reflect.ValueOf(&strbytes))
			/* log.Debug(reflect.ValueOf(i).Elem().FieldByName(elem).Addr)
			log.Debug(reflect.ValueOf(i).Elem().FieldByName(elem).Elem())
			log.Debug(reflect.ValueOf(i).Elem().FieldByName(elem).Type().Elem()) */
			log.Debugf("elem : %v string: '%v' Type: %v\n", elem, reflect.ValueOf(i).Elem().FieldByName(elem).Elem(), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
		} else {
			reflect.ValueOf(i).Elem().FieldByName(elem).SetString(string(bytes))
			log.Debugf("elem : %v string: '%v' Type: %v\n", elem, reflect.ValueOf(i).Elem().FieldByName(elem), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
		}
	case g.NoSuchObject:
		log.Debugf("NoSuchObject for elem '%v'", elem)
	case g.NoSuchInstance:
		log.Debugf("NoSuchInstance for elem '%v'", elem)
	default:
		log.Debugf("received value '%v' of type %T", result.Variables[0].Value, result.Variables[0].Value)
		if reflect.ValueOf(i).Elem().FieldByName(elem).Kind() == reflect.Ptr {
			switch result.Variables[0].Value.(type) {
			case int:
				log.Debug("int value received")
				value := uint(result.Variables[0].Value.(int))
				log.Debugf("Value after conversion is '%v' of type %T", value, value)
				reflect.ValueOf(i).Elem().FieldByName(elem).Set(reflect.ValueOf(&value))
				log.Debugf("Value of %v is '%v' of type %v", elem, reflect.ValueOf(i).Elem().FieldByName(elem).Elem(), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
			case uint:
				log.Debug("uint value received")
				value := result.Variables[0].Value.(uint)
				log.Debugf("Value after conversion is '%v' of type %T", value, value)
				reflect.ValueOf(i).Elem().FieldByName(elem).Set(reflect.ValueOf(&value))
				log.Debugf("Value of %v is '%v' of type %v", elem, reflect.ValueOf(i).Elem().FieldByName(elem).Elem(), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
			case uint64:
				log.Debug("uint64 value received")
				value := uint(result.Variables[0].Value.(uint64))
				log.Debugf("Value after conversion is '%v' of type %T", value, value)
				reflect.ValueOf(i).Elem().FieldByName(elem).Set(reflect.ValueOf(&value))
				log.Debugf("Value of %v is '%v' of type %v", elem, reflect.ValueOf(i).Elem().FieldByName(elem).Elem(), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
			default:
				log.Debugf("Value of type %T, not handle...", result.Variables[0].Value)
			}
			log.Debugf("elem : %v number: %v Type: %v\n", elem, reflect.ValueOf(i).Elem().FieldByName(elem).Elem(), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
		} else {
			value, _ := convert.ToUint(result.Variables[0].Value)
			reflect.ValueOf(i).Elem().FieldByName(elem).SetUint(uint64(value))
			log.Debugf("elem : %v number: %v Type: %v\n", elem, reflect.ValueOf(i).Elem().FieldByName(elem), reflect.ValueOf(i).Elem().FieldByName(elem).Type())
		}
	}
	return nil
}

/* hrSystemUptime : 1.3.6.1.2.1.25.1.1
sysUptime : 1.3.6.1.2.1.1.3.0
ifLastChange : 1.3.6.1.2.1.2.2.1.9 */

//GetUpTime retreive the system uptime in order to validate some data results
func (i *InterfaceDetails) GetUpTime(snmpConnection *g.GoSNMP) error {
	log.Debug("=====================")
	log.Debugf("Get Uptime")
	oid := []string{InterfaceOids["HrSystemUptime"], InterfaceOids["SysUpTime"]}
	result, err := snmpConnection.Get(oid)
	if err != nil {
		return fmt.Errorf("Get() UpTime err: %v", err)
	}

	for _, variable := range result.Variables {
		if variable.Name == InterfaceOids["HrSystemUptime"] {
			log.Debugf("Try HrSystemUptime OID")
		} else {
			log.Debugf("Try SysUpTime OID")
		}

		switch variable.Type {
		case g.TimeTicks:
			value := uint(variable.Value.(uint32))
			i.UpTime = &value
			log.Debugf("elem : %v number: %v Type: %T", variable.Name, *i.UpTime, i.UpTime)
			return nil
		case g.NoSuchInstance:
			log.Debugf("NoSuchInstance for elem '%v'", variable.Name)
		case g.NoSuchObject:
			log.Debugf("NoSuchObject for elem '%v'", variable.Name)
		}
	}
	return nil
}

//OperToString take the numerical status representation UP/DOWN/TESTING... (see const) and convert it to string.
func OperToString(status uint) string {
	switch status {
	case UP:
		return "UP"
	case DOWN:
		return "DOWN"
	case TESTING:
		return "TESTING"
	case UNKNOWN:
		return "UNKNOWN"
	case DORMANT:
		return "DORMANT"
	case NOTPRESENT:
		return "NONPRESENT"
	case LOWERLAYERDOWN:
		return "LAYERDOWN"
	default:
		return "Undefined"
	}
}

//DuplexToString take the numerical duplex state and convert it to string.
func DuplexToString(duplex uint) string {
	switch duplex {
	case 1:
		return "Unknown"
	case 2:
		return "Half-Duplex"
	case 3:
		return "Full-Duplex"
	default:
		return "Undefined"
	}
}
