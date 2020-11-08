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
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go-check-network-interface/convert"

	sknchk "github.com/pandaoc-io/go-shinken-check"
	log "github.com/sirupsen/logrus"
	g "github.com/soniah/gosnmp"
)

//IndexList is the variable to strore interface index information to generate the json index file per interface
var IndexList map[int]map[string]string

//BwInconsistency is used to keep the Bandwith inconsistency status and used it to reset the pckt calculation part
var BwInconsistency = false

//CreateIndexMap create the map of ifDescr and ifName per index found
func CreateIndexMap(snmpConnection *g.GoSNMP) error {
	IndexList = make(map[int]map[string]string)
	for _, oidTable := range []string{InterfaceOids["IfDescr"], InterfaceOids["IfName"]} {
		pdus, err := snmpConnection.BulkWalkAll(oidTable)
		if err != nil {
			return err
		}
		for _, variable := range pdus {
			switch variable.Type {
			case g.OctetString:
				bytes := variable.Value.([]byte)
				indexStr := strings.TrimPrefix(variable.Name, oidTable+".")
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return err
				}
				if IndexList[index] == nil {
					IndexList[index] = make(map[string]string)
				}
				if oidTable == InterfaceOids["IfDescr"] {
					IndexList[index]["IfDescr"] = string(bytes)
				} else {
					IndexList[index]["IfName"] = string(bytes)
				}
			}
		}
	}
	return nil
}

//FetchAllDatas grab all the interface details by SNMP
func FetchAllDatas(snmpConnection *g.GoSNMP, index string) (*InterfaceDetails, error) {

	networkinterface := &InterfaceDetails{}
	networkinterface.Index = new(int)
	*networkinterface.Index, _ = strconv.Atoi(index)

	elementList := []string{
		"IfName",
		"IfDescr",
		"IfSpeed",
		"IfAdminStatus",
		"IfOperStatus",
		"IfInOctets",
		"IfInUcastPkts",
		"IfInNUcastPkts",
		"IfInDiscards",
		"IfInErrors",
		"IfOutOctets",
		"IfOutUcastPkts",
		"IfOutNUcastPkts",
		"IfOutDiscards",
		"IfOutErrors",
		"IfHCInOctets",
		"IfHCInUcastPkts",
		"IfHCInMulticastPkts",
		"IfHCInBroadcastPkts",
		"IfHCOutOctets",
		"IfHCOutUcastPkts",
		"IfHCOutMulticastPkts",
		"IfHCOutBroadcastPkts",
		"IfHighSpeed",
		"IfAlias",
		"LocIfInCRC",
		"Dot3StatsDuplexStatus",
	}
	for _, elem := range elementList {
		err := networkinterface.GetData(snmpConnection, elem)
		if err != nil {
			return nil, err
		}
		if elem == "IfAlias" && networkinterface.IfAlias != nil {
			log.Debug("Replace the characters of the alias '|' by '!'")
			*networkinterface.IfAlias = strings.ReplaceAll(*networkinterface.IfAlias, "|", "!")
		}
	}
	networkinterface.Timestamp = time.Now().Unix()

	return networkinterface, nil
}

//Diff will return the difference between 2 values. Used to make the diff between 2 counters (32 or 64 bits)
func diff(newData interface{}, oldData interface{}, is64 bool) (uint, error) {
	if reflect.TypeOf(newData) != reflect.TypeOf(oldData) {
		return 0, fmt.Errorf("2 different value types provided : %v, %v", reflect.TypeOf(newData), reflect.TypeOf(oldData))
	}
	if reflect.ValueOf(oldData).IsNil() || reflect.ValueOf(oldData).IsNil() {
		log.Debug("New or Previous data are unavailable, skip calculation until next polling...")
		return 0, nil
	}

	newDataConverted, err := convert.ToUint(newData)
	if err != nil {
		return 0, err
	}
	oldDataConverted, err := convert.ToUint(oldData)
	if err != nil {
		return 0, err
	}

	log.Debugf("NewData : %v, type : %T", newDataConverted, newDataConverted)
	log.Debugf("OldData : %v, type : %T", oldDataConverted, oldDataConverted)
	diff := uint(0)
	if newDataConverted == oldDataConverted {
		diff = 0
	} else if BwInconsistency {
		//If Bandwidth inconsistency is found we also apply the calculation only on the new value
		log.Debug("Inconsistency found, diff based only on the new value")
		diff = 0
	} else if newDataConverted > oldDataConverted {
		diff = newDataConverted - oldDataConverted
	} else {
		if is64 {
			diff = math.MaxUint64 - oldDataConverted + newDataConverted
		} else {
			diff = math.MaxUint32 - oldDataConverted + newDataConverted
		}
	}

	return diff, nil
}

//Bandwidth will return the rate in bps and the usage in % of the link, the related perfdata and make the test with the thresholds to update the check
func Bandwidth(intNewData *InterfaceDetails, intOldData *InterfaceDetails, timeDiff time.Duration, chk *sknchk.Check, bw float64, bc float64) {
	var err error
	log.Debug("===== IfHCInOctets =====")
	if intNewData.IfHCInOctets != nil {
		intNewData.IfInRate, intNewData.IfInPrct, err = bwStats(intNewData.IfHCInOctets, intOldData.IfHCInOctets, *intNewData.IfHighSpeed*1000000, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInOctets counter available, skip...")
	}
	log.Debug("===== IfHCOutOctets =====")
	if intNewData.IfHCOutOctets != nil {
		intNewData.IfOutRate, intNewData.IfOutPrct, err = bwStats(intNewData.IfHCOutOctets, intOldData.IfHCOutOctets, *intNewData.IfHighSpeed*1000000, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutOctets counter available, skip...")
	}
	if intNewData.IfHCInOctets == nil && intNewData.IfHCOutOctets == nil {
		log.Debug("In/Out 64 bits counters not present, try 32 counters")
		log.Debug("===== IfInOctets =====")
		if intNewData.IfInOctets != nil {
			intNewData.IfInRate, intNewData.IfInPrct, err = bwStats(intNewData.IfInOctets, intOldData.IfInOctets, *intNewData.IfSpeed, timeDiff, false)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInOctets counter available, skip...")
		}
		log.Debug("===== IfOutOctets =====")
		if intNewData.IfOutOctets != nil {
			intNewData.IfOutRate, intNewData.IfOutPrct, err = bwStats(intNewData.IfOutOctets, intOldData.IfOutOctets, *intNewData.IfSpeed, timeDiff, false)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfOutOctets counter available, skip...")
		}
	}

	//Force 0% bandwidth usage for interfaces named vlanxxx
	if (intNewData.IfName != nil && strings.Contains(strings.ToLower(*intNewData.IfName), "vlan")) || (intNewData.IfDescr != nil && strings.Contains(strings.ToLower(*intNewData.IfDescr), "vlan")) {
		intNewData.IfInPrct = nil
		intNewData.IfOutPrct = nil
	}
	if intNewData.IfInRate != nil {
		//We suppress the UOM to be compatible with the Nagvis weathermap feature (value expressed in bps)
		chk.AddPerfData("in", strconv.FormatFloat(*intNewData.IfInRate, 'f', 2, 64), "", 0, 0, 0, *intNewData.SpeedInbit)
	}
	if intNewData.IfInPrct != nil {
		chk.AddPerfData("in_usage", strconv.FormatFloat(*intNewData.IfInPrct, 'f', 2, 64), "%", 0, 0, 0, 100)
	}

	if intNewData.IfOutRate != nil {
		//We suppress the UOM to be compatible with the Nagvis weathermap feature (value expressed in bps)
		chk.AddPerfData("out", strconv.FormatFloat(*intNewData.IfOutRate, 'f', 2, 64), "", 0, 0, 0, *intNewData.SpeedInbit)
	}
	if intNewData.IfOutPrct != nil {
		chk.AddPerfData("out_usage", strconv.FormatFloat(*intNewData.IfOutPrct, 'f', 2, 64), "%", 0, 0, 0, 100)
	}

	if intNewData.IfInPrct != nil && *intNewData.IfInPrct > bc {
		chk.AddShort(fmt.Sprintf(`Very high In Bandwidth : %v - %v (> %v%%)`,
			convert.HumanReadable(*intNewData.IfInRate, 1000, "bits/sec"),
			sknchk.FmtCritical(fmt.Sprintf("%.2f%%", *intNewData.IfInPrct)), bc),
			true)
		chk.AddCritical()
	} else if intNewData.IfInPrct != nil && *intNewData.IfInPrct > bw {
		chk.AddShort(fmt.Sprintf(`High In Bandwidth : %v - %v (> %v%%)`,
			convert.HumanReadable(*intNewData.IfInRate, 1000, "bits/sec"),
			sknchk.FmtWarning(fmt.Sprintf("%.2f%%", *intNewData.IfInPrct)), bw),
			true)
		chk.AddWarning()
	}

	if intNewData.IfOutPrct != nil && *intNewData.IfOutPrct > bc {
		chk.AddShort(fmt.Sprintf(`Very high Out Bandwidth : %v - %v (> %v%%)`,
			convert.HumanReadable(*intNewData.IfOutRate, 1000, "bits/sec"),
			sknchk.FmtCritical(fmt.Sprintf("%.2f%%", *intNewData.IfOutPrct)), bc),
			true)
		chk.AddCritical()
	} else if intNewData.IfOutPrct != nil && *intNewData.IfOutPrct > bw {
		chk.AddShort(fmt.Sprintf(`High Out Bandwidth : %v - %v (> %v%%)`,
			convert.HumanReadable(*intNewData.IfOutRate, 1000, "bits/sec"),
			sknchk.FmtWarning(fmt.Sprintf("%.2f%%", *intNewData.IfOutPrct)), bw),
			true)
		chk.AddWarning()
	}
}

//Packets returns the rate in pps of the total, unicast, multicast and broadcast packets
func Packets(intNewData *InterfaceDetails, intOldData *InterfaceDetails, timeDiff time.Duration) {
	log.Debug("===== Total pckts =====")
	var inUniPckt uint
	var inMultiPckt uint
	var inBroadPckt uint
	var err error
	log.Debug("==> In Unicast pckts")
	if intNewData.IfHCInUcastPkts != nil {
		inUniPckt, err = diff(intNewData.IfHCInUcastPkts, intOldData.IfHCInUcastPkts, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInUcastPkts counter available, skip...")
	}
	log.Debug("==> In Multicast pckts")
	if intNewData.IfHCInMulticastPkts != nil {
		inMultiPckt, err = diff(intNewData.IfHCInMulticastPkts, intOldData.IfHCInMulticastPkts, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInMulticastPkts counter available, skip...")
	}
	log.Debug("==> In Broadcast pckts")
	if intNewData.IfHCInBroadcastPkts != nil {
		inBroadPckt, err = diff(intNewData.IfHCInBroadcastPkts, intOldData.IfHCInBroadcastPkts, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInBroadcastPkts counter available, skip...")
	}

	// If the unicast packets 64 bits counter isn't available, we switch to the 32 bits counters
	if intNewData.IfHCInUcastPkts == nil {
		log.Debug("No IfHCInUcastPkts found, switch to 32 bits counter...")
		if intNewData.IfInUcastPkts != nil {
			inUniPckt, err = diff(intNewData.IfInUcastPkts, intOldData.IfInUcastPkts, false)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInUcastPkts counter available, skip...")
		}
		log.Debug("==> In Multicast pckts")
		if intNewData.IfInNUcastPkts != nil {
			inMultiPckt, err = diff(intNewData.IfInNUcastPkts, intOldData.IfInNUcastPkts, false)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInNUcastPkts counter available, skip...")
		}
	}
	log.Debug("==> In pckts Summary")
	log.Debugf("In Unicast : %v", inUniPckt)
	log.Debugf("In Multicast : %v", inMultiPckt)
	log.Debugf("In Broadcast : %v", inBroadPckt)

	//As the sum of all the elements can overflow a standard int64, we need to switch to bit.Int
	intNewData.IfInTotalPkts = new(uint)
	*intNewData.IfInTotalPkts = inUniPckt + inMultiPckt + inBroadPckt
	log.Debugf("Total in : %v, type %T", *intNewData.IfInTotalPkts, intNewData.IfInTotalPkts)

	intNewData.IfInTotalPktsRate = new(float64)
	if timeDiff.Seconds() > 0 {
		*intNewData.IfInTotalPktsRate = float64(*intNewData.IfInTotalPkts) / timeDiff.Seconds()
	} else {
		*intNewData.IfInTotalPktsRate = 0
	}
	log.Debugf("Total in rate : %.2f pps, type %T", *intNewData.IfInTotalPktsRate, intNewData.IfInTotalPktsRate)

	var outUniPckt uint
	var outMultiPckt uint
	var outBroadPckt uint
	log.Debug("==> Out Unicast pckts")
	if intNewData.IfHCOutUcastPkts != nil {
		outUniPckt, err = diff(intNewData.IfHCOutUcastPkts, intOldData.IfHCOutUcastPkts, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutUcastPkts counter available, skip...")
	}
	log.Debug("==> Out Multicast pckts")
	if intNewData.IfHCOutMulticastPkts != nil {
		outMultiPckt, err = diff(intNewData.IfHCOutMulticastPkts, intOldData.IfHCOutMulticastPkts, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutMulticastPkts counter available, skip...")
	}
	log.Debug("==> Out Broadcast pckts")
	if intNewData.IfHCOutBroadcastPkts != nil {
		outBroadPckt, err = diff(intNewData.IfHCOutBroadcastPkts, intOldData.IfHCOutBroadcastPkts, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutBroadcastPkts counter available, skip...")
	}

	// If the unicast packets 64 bits counter isn't available, we switch to the 32 bits counters
	if intNewData.IfHCOutUcastPkts == nil {
		log.Debug("No IfHCOutUcastPkts found, switch to 32 bits counter...")
		if intNewData.IfOutUcastPkts != nil {
			outUniPckt, err = diff(intNewData.IfOutUcastPkts, intOldData.IfOutUcastPkts, false)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInUcastPkts counter available, skip...")
		}
		log.Debug("==> In Multicast pckts")
		if intNewData.IfOutNUcastPkts != nil {
			outMultiPckt, err = diff(intNewData.IfOutNUcastPkts, intOldData.IfOutNUcastPkts, false)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInNUcastPkts counter available, skip...")
		}
	}

	log.Debug("==> Out pckts Summary")
	log.Debugf("Out Unicast : %v", outUniPckt)
	log.Debugf("Out Multicast : %v", outMultiPckt)
	log.Debugf("Out Broadcast : %v", outBroadPckt)

	intNewData.IfOutTotalPkts = new(uint)
	*intNewData.IfOutTotalPkts = outUniPckt + outMultiPckt + outBroadPckt
	log.Debugf("Total out : %v, type : %T", *intNewData.IfOutTotalPkts, intNewData.IfOutTotalPkts)

	intNewData.IfOutTotalPktsRate = new(float64)
	if timeDiff.Seconds() > 0 {
		*intNewData.IfOutTotalPktsRate = float64(*intNewData.IfOutTotalPkts) / timeDiff.Seconds()
	} else {
		*intNewData.IfOutTotalPktsRate = 0
	}
	log.Debugf("Total out rate : %.2f pps, type : %T", *intNewData.IfOutTotalPktsRate, intNewData.IfOutTotalPktsRate)

	log.Debug("===== In Unicast =====")
	if intNewData.IfHCInUcastPkts != nil {
		intNewData.InUniPcktRate, _, err = pckStats(intNewData.IfHCInUcastPkts, intOldData.IfHCInUcastPkts, intNewData.IfInTotalPkts, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInUcastPkts counter available, skip...")
	}

	log.Debug("===== In Multicast =====")
	if intNewData.IfHCInMulticastPkts != nil {
		intNewData.InMultiPcktRate, _, err = pckStats(intNewData.IfHCInMulticastPkts, intOldData.IfHCInMulticastPkts, intNewData.IfInTotalPkts, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInMulticastPkts counter available, skip...")
	}

	log.Debug("===== In Broadcast =====")
	if intNewData.IfHCInBroadcastPkts != nil {
		intNewData.InBroadPcktRate, _, err = pckStats(intNewData.IfHCInBroadcastPkts, intOldData.IfHCInBroadcastPkts, intNewData.IfInTotalPkts, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCInBroadcastPkts counter available, skip...")
	}

	log.Debug("===== Out Unicast =====")
	if intNewData.IfHCOutUcastPkts != nil {
		intNewData.OutUniPcktRate, _, err = pckStats(intNewData.IfHCOutUcastPkts, intOldData.IfHCOutUcastPkts, intNewData.IfOutTotalPkts, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutUcastPkts counter available, skip...")
	}

	log.Debug("===== Out Multicast =====")
	if intNewData.IfHCOutMulticastPkts != nil {
		intNewData.OutMultiPcktRate, _, err = pckStats(intNewData.IfHCOutMulticastPkts, intOldData.IfHCOutMulticastPkts, intNewData.IfOutTotalPkts, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutMulticastPkts counter available, skip...")
	}

	log.Debug("===== Out Broadcast =====")
	if intNewData.IfHCOutBroadcastPkts != nil {
		intNewData.OutBroadPcktRate, _, err = pckStats(intNewData.IfHCOutBroadcastPkts, intOldData.IfHCOutBroadcastPkts, intNewData.IfOutTotalPkts, timeDiff, true)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfHCOutBroadcastPkts counter available, skip...")
	}

	//32 bits counter
	if intNewData.IfHCInUcastPkts == nil {
		log.Debug("No IfHCInUcastPkts found, switch to 32 bits counter...")
		if intNewData.IfInUcastPkts != nil {
			log.Debug("===== In Unicast 32 bits =====")
			intNewData.InUniPcktRate, _, err = pckStats(intNewData.IfInUcastPkts, intOldData.IfInUcastPkts, intNewData.IfInTotalPkts, timeDiff, true)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInUcastPkts counter available, skip...")
		}

		if intNewData.IfInNUcastPkts != nil {
			log.Debug("===== In Multicast 32 bits =====")
			intNewData.InMultiPcktRate, _, err = pckStats(intNewData.IfInNUcastPkts, intOldData.IfInNUcastPkts, intNewData.IfInTotalPkts, timeDiff, true)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfInNUcastPkts counter available, skip...")
		}
	}

	if intNewData.IfHCOutUcastPkts == nil {
		log.Debug("No IfHCOutUcastPkts found, switch to 32 bits counter...")
		if intNewData.IfOutUcastPkts != nil {
			log.Debug("===== Out Unicast 32 bits =====")
			intNewData.OutUniPcktRate, _, err = pckStats(intNewData.IfOutUcastPkts, intOldData.IfOutUcastPkts, intNewData.IfOutTotalPkts, timeDiff, true)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfOutUcastPkts counter available, skip...")
		}

		if intNewData.IfOutNUcastPkts != nil {
			log.Debug("===== Out Multicast 32 bits =====")
			intNewData.OutMultiPcktRate, _, err = pckStats(intNewData.IfOutNUcastPkts, intOldData.IfOutNUcastPkts, intNewData.IfOutTotalPkts, timeDiff, true)
			if err != nil {
				sknchk.Unknown(fmt.Sprint(err), "")
			}
		} else {
			log.Debug("No IfOutNUcastPkts counter available, skip...")
		}
	}
}

//Errors returns the rate in pps and the % of packets in error, the related perfdata and make the test with the thresholds to update the check
func Errors(intNewData *InterfaceDetails, intOldData *InterfaceDetails, timeDiff time.Duration, chk *sknchk.Check, ew float64, ec float64, ev string) {
	log.Debug("===== IfInErrors =====")
	var err error
	if intNewData.IfInErrors != nil {
		intNewData.IfInErrorsRate, intNewData.IfInErrorsPrct, err = pckStats(intNewData.IfInErrors, intOldData.IfInErrors, intNewData.IfInTotalPkts, timeDiff, false)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfInErrors counter available, skip...")
	}

	if intNewData.LocIfInCRC != nil {
		intNewData.LocIfInCRCRate, intNewData.LocIfInCRCPrct, err = pckStats(intNewData.LocIfInCRC, intOldData.LocIfInCRC, intNewData.IfInTotalPkts, timeDiff, false)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No LocIfInCRC counter available, skip...")
	}

	log.Debug("===== IfOutErrors =====")
	if intNewData.IfOutErrors != nil {
		intNewData.IfOutErrorsRate, intNewData.IfOutErrorsPrct, err = pckStats(intNewData.IfOutErrors, intOldData.IfOutErrors, intNewData.IfOutTotalPkts, timeDiff, false)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfOutErrors counter available, skip...")
	}

	if intNewData.IfInErrorsRate != nil {
		chk.AddPerfData("in_errors", strconv.FormatFloat(*intNewData.IfInErrorsRate, 'f', 2, 64), "pps", 0, 0, 0, 0)
	}

	/* if intNewData.IfInErrorsPrct != nil {
		chk.AddPerfData("in_errors_prct", strconv.FormatFloat(*intNewData.IfInErrorsPrct, 'f', 2, 64), "%", 0, 0, 0, 0)
	} */

	if intNewData.IfOutErrorsRate != nil {
		chk.AddPerfData("out_errors", strconv.FormatFloat(*intNewData.IfOutErrorsRate, 'f', 2, 64), "pps", 0, 0, 0, 0)
	}

	/* if intNewData.IfOutErrorsPrct != nil {
		chk.AddPerfData("out_errors_prct", strconv.FormatFloat(*intNewData.IfOutErrorsPrct, 'f', 2, 64), "%", 0, 0, 0, 0)
	} */

	//Error thresholds
	if ev == "pps" {
		if intNewData.IfInErrorsRate != nil && *intNewData.IfInErrorsRate > ec {
			chk.AddShort(fmt.Sprintf(`Very high In Errors : %v - %.2f %% (> %v %v)`,
				sknchk.FmtCritical(fmt.Sprintf("%.2f pps", *intNewData.IfInErrorsRate)),
				*intNewData.IfInErrorsPrct, ec, ev),
				true)
			chk.AddCritical()
		} else if intNewData.IfInErrorsRate != nil && *intNewData.IfInErrorsRate > ew {
			chk.AddShort(fmt.Sprintf(`High In Errors : %v - %.2f %% (> %v %v)`,
				sknchk.FmtWarning(fmt.Sprintf("%.2f pps", *intNewData.IfInErrorsRate)),
				*intNewData.IfInErrorsPrct, ew, ev),
				true)
			chk.AddWarning()
		}
		if intNewData.IfOutErrorsRate != nil && *intNewData.IfOutErrorsRate > ec {
			chk.AddShort(fmt.Sprintf(`Very high Out Errors : %v - %.2f %% (> %v %v)`,
				sknchk.FmtCritical(fmt.Sprintf("%.2f pps", *intNewData.IfOutErrorsRate)),
				*intNewData.IfOutErrorsPrct, ec, ev),
				true)
			chk.AddCritical()
		} else if intNewData.IfOutErrorsRate != nil && *intNewData.IfOutErrorsRate > ew {
			chk.AddShort(fmt.Sprintf(`High Out Errors : %v - %.2f %% (> %v %v)`,
				sknchk.FmtWarning(fmt.Sprintf("%.2f pps", *intNewData.IfOutErrorsRate)),
				*intNewData.IfOutErrorsPrct, ew, ev),
				true)
			chk.AddWarning()
		}
	} else {
		if intNewData.IfInErrorsPrct != nil && *intNewData.IfInErrorsPrct > ec {
			chk.AddShort(fmt.Sprintf(`Very high In Errors : %.2f pps - %v (> %v %v)`,
				*intNewData.IfInErrorsRate,
				sknchk.FmtCritical(fmt.Sprintf("%.2f %%", *intNewData.IfInErrorsPrct)), ec, ev),
				true)
			chk.AddCritical()
		} else if intNewData.IfInErrorsPrct != nil && *intNewData.IfInErrorsPrct > ew {
			chk.AddShort(fmt.Sprintf(`High In Errors : %.2f pps - %v (> %v %v)`,
				*intNewData.IfInErrorsRate,
				sknchk.FmtWarning(fmt.Sprintf("%.2f %%", *intNewData.IfInErrorsPrct)), ew, ev),
				true)
			chk.AddWarning()
		}
		if intNewData.IfOutErrorsPrct != nil && *intNewData.IfOutErrorsPrct > ec {
			chk.AddShort(fmt.Sprintf(`Very high Out Errors : %.2f pps - %v (> %v %v)`,
				*intNewData.IfOutErrorsRate,
				sknchk.FmtCritical(fmt.Sprintf("%.2f %%", *intNewData.IfOutErrorsPrct)), ec, ev),
				true)
			chk.AddCritical()
		} else if intNewData.IfOutErrorsPrct != nil && *intNewData.IfOutErrorsPrct > ew {
			chk.AddShort(fmt.Sprintf(`High Out Errors : %.2f pps - %v (> %v %v)`,
				*intNewData.IfOutErrorsRate,
				sknchk.FmtWarning(fmt.Sprintf("%.2f %%", *intNewData.IfOutErrorsPrct)), ew, ev),
				true)
			chk.AddWarning()
		}
	}
}

//Discards returns the rate in pps and the % of packets in discard, the related perfdata and make the test with the thresholds to update the check
func Discards(intNewData *InterfaceDetails, intOldData *InterfaceDetails, timeDiff time.Duration, chk *sknchk.Check, dw float64, dc float64, dv string) {
	log.Debug("===== IfInDiscards =====")
	var err error
	if intNewData.IfInDiscards != nil {
		intNewData.IfInDiscardsRate, intNewData.IfInDiscardsPrct, err = pckStats(intNewData.IfInDiscards, intOldData.IfInDiscards, intNewData.IfInTotalPkts, timeDiff, false)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfInDiscards counter available, skip...")
	}
	log.Debug("===== IfOutDiscards =====")
	if intNewData.IfOutDiscards != nil {
		intNewData.IfOutDiscardsRate, intNewData.IfOutDiscardsPrct, err = pckStats(intNewData.IfOutDiscards, intOldData.IfOutDiscards, intNewData.IfOutTotalPkts, timeDiff, false)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	} else {
		log.Debug("No IfOutDiscards counter available, skip...")
	}

	if intNewData.IfInDiscardsRate != nil {
		chk.AddPerfData("in_discards", strconv.FormatFloat(*intNewData.IfInDiscardsRate, 'f', 2, 64), "pps", 0, 0, 0, 0)
	}

	/* if intNewData.IfInDiscardsPrct != nil {
		chk.AddPerfData("in_discards_prct", strconv.FormatFloat(*intNewData.IfInDiscardsPrct, 'f', 2, 64), "%", 0, 0, 0, 0)
	} */

	if intNewData.IfOutDiscardsRate != nil {
		chk.AddPerfData("out_discards", strconv.FormatFloat(*intNewData.IfOutDiscardsRate, 'f', 2, 64), "pps", 0, 0, 0, 0)
	}

	/* if intNewData.IfOutDiscardsPrct != nil {
		chk.AddPerfData("out_discards_prct", strconv.FormatFloat(*intNewData.IfOutDiscardsPrct, 'f', 2, 64), "%", 0, 0, 0, 0)
	} */

	if dv == "pps" {
		if intNewData.IfInDiscardsRate != nil && *intNewData.IfInDiscardsRate > dc {
			chk.AddShort(fmt.Sprintf(`Very high In Discards : %v - %.2f %% (> %v %v)`,
				sknchk.FmtCritical(fmt.Sprintf("%.2f pps", *intNewData.IfInDiscardsRate)),
				*intNewData.IfInDiscardsPrct, dc, dv),
				true)
			chk.AddCritical()
		} else if intNewData.IfInDiscardsRate != nil && *intNewData.IfInDiscardsRate > dw {
			chk.AddShort(fmt.Sprintf(`High In Discards : %v - %.2f %% (> %v %v)`,
				sknchk.FmtWarning(fmt.Sprintf("%.2f pps", *intNewData.IfInDiscardsRate)),
				*intNewData.IfInDiscardsPrct, dw, dv),
				true)
			chk.AddWarning()
		}
		if intNewData.IfOutDiscardsRate != nil && *intNewData.IfOutDiscardsRate > dc {
			chk.AddShort(fmt.Sprintf(`Very high Out Discards : %v - %.2f %% (> %v %v)`,
				sknchk.FmtCritical(fmt.Sprintf("%.2f pps", *intNewData.IfOutDiscardsRate)),
				*intNewData.IfOutDiscardsPrct, dc, dv),
				true)
			chk.AddCritical()
		} else if intNewData.IfOutDiscardsRate != nil && *intNewData.IfOutDiscardsRate > dw {
			chk.AddShort(fmt.Sprintf(`High Out Errors : %v - %.2f %% (> %v %v)`,
				sknchk.FmtWarning(fmt.Sprintf("%.2f pps", *intNewData.IfOutDiscardsRate)),
				*intNewData.IfOutDiscardsPrct, dw, dv),
				true)
			chk.AddWarning()
		}
	} else {
		if intNewData.IfInDiscardsPrct != nil && *intNewData.IfInDiscardsPrct > dc {
			chk.AddShort(fmt.Sprintf(`Very high In Discards : %.2f pps - %v (> %v %v)`,
				*intNewData.IfInDiscardsRate,
				sknchk.FmtCritical(fmt.Sprintf("%.2f %%", *intNewData.IfInDiscardsPrct)), dc, dv),
				true)
			chk.AddCritical()
		} else if intNewData.IfInDiscardsPrct != nil && *intNewData.IfInDiscardsPrct > dw {
			chk.AddShort(fmt.Sprintf(`High In Discards : %.2f pps - %v (> %v %v)`,
				*intNewData.IfInDiscardsRate,
				sknchk.FmtWarning(fmt.Sprintf("%.2f %%", *intNewData.IfInDiscardsPrct)), dw, dv),
				true)
			chk.AddWarning()
		}
		if intNewData.IfOutDiscardsPrct != nil && *intNewData.IfOutDiscardsPrct > dc {
			chk.AddShort(fmt.Sprintf(`Very high Out Discards : %.2f pps - %v (> %v %v)`,
				*intNewData.IfOutDiscardsRate,
				sknchk.FmtCritical(fmt.Sprintf("%.2f %%", *intNewData.IfOutDiscardsPrct)), dc, dv),
				true)
			chk.AddCritical()
		} else if intNewData.IfOutDiscardsPrct != nil && *intNewData.IfOutDiscardsPrct > dw {
			chk.AddShort(fmt.Sprintf(`High Out Discards : %.2f pps - %v (> %v %v)`,
				*intNewData.IfOutDiscardsRate,
				sknchk.FmtWarning(fmt.Sprintf("%.2f %%", *intNewData.IfOutDiscardsPrct)), dw, dv),
				true)
			chk.AddWarning()
		}
	}
}

//Speed returns the interface speed in bps and the related perfdata
func Speed(intNewData *InterfaceDetails, chk *sknchk.Check) {
	log.Debug("===== Speed =====")
	var speed uint
	if intNewData.IfHighSpeed != nil {
		log.Debug("ifHighSpeed found")
		speed = *intNewData.IfHighSpeed * 1000000
	} else if intNewData.IfSpeed != nil {
		log.Debug("No ifHighSpeed found, switch to ifSpeed")
		if *intNewData.IfSpeed == math.MaxUint32 {
			log.Debug("ifSpeed value is egal to MaxUint32 value")
			log.Debug("Arbitrarly set the speed to 10Gbps")
			speed = 10000000000
		} else {
			speed = *intNewData.IfSpeed
		}
	} else {
		log.Debug("No speed found")
	}
	intNewData.SpeedInbit = new(uint)
	*intNewData.SpeedInbit = speed
	chk.AddPerfData("speed", *intNewData.SpeedInbit, "", 0, 0, 0, 0)
}

//DuplexMode returns the Duplex Mode and the related perfdata
func DuplexMode(intNewData *InterfaceDetails, chk *sknchk.Check) {
	log.Debug("===== Duplex Mode =====")
	//1-unknown, 2-halfDuplex, 3-fullDuplex
	/* intNewData.Dot3StatsDuplexStatus = new(uint)
	*intNewData.Dot3StatsDuplexStatus = 3 */
	if intNewData.Dot3StatsDuplexStatus != nil {
		log.Debug("Duplex Mode found")
		chk.AddPerfData("duplexmode", *intNewData.Dot3StatsDuplexStatus, "", 0, 0, 0, 0)
		if *intNewData.Dot3StatsDuplexStatus == 2 {
			chk.AddShort(fmt.Sprintf(`Interface mode : %v `,
				sknchk.FmtCritical("Half-Duplex")),
				true)
			chk.AddCritical()
		}
	} else {
		log.Debug("No Duplex Mode found")
	}
}

//bwStats will return the rate and the percent usage of a specifique element
func bwStats(newData interface{}, oldData interface{}, speed interface{}, elapseTime time.Duration, is64 bool) (*float64, *float64, error) {
	if reflect.TypeOf(newData).Elem() != reflect.TypeOf(oldData).Elem() {
		return nil, nil, fmt.Errorf("2 different value types provided : %v, %v", reflect.TypeOf(newData), reflect.TypeOf(oldData))
	}
	if reflect.ValueOf(newData).IsNil() || reflect.ValueOf(oldData).IsNil() {
		log.Debug("New or Previous are unavailable, skip calculation until next polling...")
		return nil, nil, nil
	}

	log.Debugf("NewData : %v, type : %T", reflect.ValueOf(newData).Elem(), newData)
	log.Debugf("OldData : %v, type : %T", reflect.ValueOf(oldData).Elem(), oldData)

	newDataConverted, err := convert.ToUint(newData)
	if err != nil {
		return nil, nil, err
	}
	oldDataConverted, err := convert.ToUint(oldData)
	if err != nil {
		return nil, nil, err
	}

	var diff uint
	if newDataConverted == oldDataConverted {
		diff = 0
	} else if newDataConverted > oldDataConverted {
		diff = newDataConverted - oldDataConverted
	} else {
		if is64 {
			diff = math.MaxUint64 - oldDataConverted + newDataConverted
		} else {
			diff = math.MaxUint32 - oldDataConverted + newDataConverted
		}
	}

	log.Debugf("Time diff : %v sec (%v)\n", elapseTime.Seconds(), elapseTime.String())
	log.Debugf("Diff between new and old data : %v\n", diff)

	diffConverted := float64(diff)
	rate := 0.0
	if elapseTime.Seconds() > 0 {
		rate = diffConverted * 8 / elapseTime.Seconds()
	}

	log.Debugf("Rate : %v\n", convert.HumanReadable(rate, 1000, "bits/sec"))

	speedConverted, err := convert.ToFloat(speed)
	if err != nil {
		return nil, nil, err
	}
	prct := 0.0
	if speedConverted > 0 {
		prct = (rate / speedConverted) * 100
	}
	log.Debugf("Max value : %v\n", convert.HumanReadable(speedConverted, 1000, "bps"))
	log.Debugf("Percent : %.2f %%\n", prct)

	//Now we test is the values are relevant
	//Check if linkspeed is set
	if speed != 0 {
		//We check if the value if upper than 200%
		//We don't check directly 100% because for some operator links the overflow is sometime allowed
		if prct > 200 {
			log.Debug("The link percent usage is above 200%, something goes wrong.")
			log.Debug("The difference will be taken from 0 to newValue")
			//Inconsistency found, set the boolean to true to modify the packets calculation
			log.Debug("Set BwInconsistency to true.")
			BwInconsistency = true
			rate = 0
			prct = 0
			log.Debugf("New Rate : %v\n", convert.HumanReadable(rate, 1000, "bits/sec"))
			log.Debugf("New Percent : %.2f %%\n", prct)
		}
		//For some link like bond we don't have any speed, generaly set to 0.
		//In that case we check if the rate is upper than 1TB/sec.
		//It's arbitrary but it's enough to detected some inconsistency.
	} else if rate > 1000000000000 { //Test if rate is more than 1TB
		log.Debug("The rate value is inconsistant, more than 1 Tbits/sec")
		log.Debug("The difference will be taken from 0 to newValue")
		//Inconsistency found, set the boolean to true to modify the packets calculation
		log.Debug("Set BwInconsistency to true.")
		BwInconsistency = true
		rate = 0
		prct = 0
		log.Debugf("New Rate : %.2f\n", convert.HumanReadable(rate, 1000, "bits/sec"))
		log.Debugf("New Percent : %.2f %%\n", prct)
	}

	return &rate, &prct, nil
}

//pckStats will return the difference between 2 counters and the rate.
//Used by all the pckts elements (Total, Unicast, MultiCast and Broadcast)
func pckStats(newData *uint, oldData *uint, totalPckt *uint, elapseTime time.Duration, is64 bool) (*float64, *float64, error) {
	if reflect.TypeOf(newData) != reflect.TypeOf(oldData) {
		return nil, nil, fmt.Errorf("2 different value types provided : %v, %v", reflect.TypeOf(newData), reflect.TypeOf(oldData))
	}

	if reflect.ValueOf(oldData).IsNil() || reflect.ValueOf(oldData).IsNil() {
		log.Debug("New or Previous are unavailable, skip calculation until next polling...")
		return nil, nil, nil
	}

	var diff uint

	if *newData == *oldData {
		diff = 0
	} else if BwInconsistency {
		log.Debug("Inconsistency found, diff based only on the new value")
		//If Bandwidth inconsistency is found we also apply the calculation only on the new value
		diff = 0
	} else if *newData > *oldData {
		diff = *newData - *oldData
	} else {
		if is64 {
			diff = math.MaxUint64 - *oldData + *newData
		} else {
			diff = math.MaxUint32 - *oldData + *newData
		}
	}

	log.Debugf("Time diff : %v sec (%v)\n", elapseTime.Seconds(), elapseTime.String())
	log.Debugf("Total number of packets : %v, Type : %T\n", *totalPckt, totalPckt)
	log.Debugf("Diff between new and old data : %v, Type: %T\n", diff, diff)

	//Now we test is the values are relevant
	//We check if number of packets is lower than the total number of packets during the same time interval
	if diff > *totalPckt {
		log.Debug("The number of packets is higher than the total number of packets.")
		log.Debug("The difference will be taken from 0 to newValue")
		diff = 0
	}

	var rate float64
	var prct float64
	if elapseTime.Seconds() > 0 {
		rate = float64(diff) / elapseTime.Seconds()
	} else {
		rate = 0
	}
	if *totalPckt > 0 {
		prct = float64(diff) / float64(*totalPckt) * 100
	}

	log.Debugf("Rate : %.2f pps, type : %T\n", rate, rate)
	log.Debugf("Percent : %.2f %%, type : %T \n", prct, prct)

	return &rate, &prct, nil
}
