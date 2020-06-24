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
	"math"
	"strconv"
	"strings"
	"time"

	"go-check-network-interface/file"
	"go-check-network-interface/netint"
	"go-check-network-interface/snmp"
	"go-check-network-interface/ui"

	"github.com/mitchellh/mapstructure"
	sknchk "github.com/pandaoc-io/go-shinken-check"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func networkInterfaceCheck(snmpVersion string, cmd *cobra.Command, args []string) {
	//Discard thresholds
	dcflag, _ := cmd.Flags().GetString("discard-critical")
	dwflag, _ := cmd.Flags().GetString("discard-warning")
	//Error thresholds
	ecflag, _ := cmd.Flags().GetString("error-critical")
	ewflag, _ := cmd.Flags().GetString("error-warning")
	//Bandwidth thresholds
	bcflag, _ := cmd.Flags().GetFloat64("bandwidth-critical")
	bwflag, _ := cmd.Flags().GetFloat64("bandwidth-warning")

	//Check if Error/Dicard thresholds have the same type
	if (strings.Contains(dcflag, "pps") && !strings.Contains(dwflag, "pps")) || (strings.Contains(dcflag, "%") && !strings.Contains(dwflag, "%")) {
		sknchk.Unknown("Discard thresholds haven't the same type. See usage for more details.", "")
	}
	if (strings.Contains(ecflag, "pps") && !strings.Contains(ewflag, "pps")) || (strings.Contains(ecflag, "%") && !strings.Contains(ewflag, "%")) {
		sknchk.Unknown("Error thresholds haven't the same type. See usage for more details.", "")
	}

	var ec float64
	var ew float64
	var ev string
	if strings.Contains(ecflag, "pps") {
		ec, _ = strconv.ParseFloat(strings.Split(ecflag, "pps")[0], 64)
		ew, _ = strconv.ParseFloat(strings.Split(ewflag, "pps")[0], 64)
		ev = "pps"
	} else {
		ec, _ = strconv.ParseFloat(strings.Split(ecflag, "%")[0], 64)
		ew, _ = strconv.ParseFloat(strings.Split(ewflag, "%")[0], 64)
		ev = "%"
	}

	var dc float64
	var dw float64
	var dv string
	if strings.Contains(dcflag, "pps") {
		dc, _ = strconv.ParseFloat(strings.Split(dcflag, "pps")[0], 64)
		dw, _ = strconv.ParseFloat(strings.Split(dwflag, "pps")[0], 64)
		dv = "pps"
	} else {
		dc, _ = strconv.ParseFloat(strings.Split(dcflag, "%")[0], 64)
		dw, _ = strconv.ParseFloat(strings.Split(dwflag, "%")[0], 64)
		dv = "%"
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		log.SetLevel(log.DebugLevel)
		sknchk.Output.SetDebug()
		log.Debugln("VERBOSE mode enable")
		log.Debugln("===== Command Flags =====")
		cmd.DebugFlags()
	} else {
		sknchk.Output.SetHTML()
	}

	//Create connection and prepare some variables
	snmpConnection, err := snmp.CreateConnection(snmpVersion, cmd)
	if err != nil {
		sknchk.Unknown(fmt.Sprintf("Error while Creating SNMP connection : %v", err), "")
	}
	file.DevicePath = file.GenDeviceDirName(snmpVersion, cmd)
	intFilename := strings.ReplaceAll(cmd.Flag("interface").Value.String(), "/", "_") + ".json"

	//Check and prepare the index file
	indexFileExp, _ := cmd.Flags().GetInt("index-expiration")
	asExp := false
	err = file.CheckFileExist(file.DevicePath, "index.json")
	if err == nil {
		asExp, err = file.AsExp(file.DevicePath, "index.json", time.Duration(indexFileExp)*time.Minute)
		if err != nil {
			sknchk.Unknown(fmt.Sprintf("Error while accessing Index file : %v", err), "")
		}
		if asExp {
			log.Debugln("Regeneration of the file...")
		}
	}
	if err != nil || asExp {
		err = netint.CreateIndexMap(snmpConnection)
		if err != nil {
			sknchk.Unknown(fmt.Sprintf("Error while Creating IndexMap : %v", err), "")
		}
		err = file.CreateJSONFile(file.DevicePath, "index.json", netint.IndexList)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	}

	//Check if Device directory is readable, avoid some snmp requests if the destination isn't writable
	err = file.IsPathWritable(file.DevicePath)
	if err != nil {
		sknchk.Unknown(fmt.Sprint(err), "")
	}

	var index string
	index, err = file.FindIntIndex(file.DevicePath, cmd.Flag("interface").Value.String())
	if err != nil {
		log.Debugln("No interface found, force the recreation of the index file...")
		err = netint.CreateIndexMap(snmpConnection)
		if err != nil {
			sknchk.Unknown(fmt.Sprintf("Error while Creating IndexMap : %v", err), "")
		}
		err = file.CreateJSONFile(file.DevicePath, "index.json", netint.IndexList)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
		index, err = file.FindIntIndex(file.DevicePath, cmd.Flag("interface").Value.String())
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
	}

	intNewData := new(netint.InterfaceDetails)
	chk := &sknchk.Check{}

	//Retrieve interface information
	intNewData, err = netint.FetchAllDatas(snmpConnection, index, snmpVersion, cmd)
	if err != nil {
		sknchk.Unknown(fmt.Sprint(err), "")
	}
	err = intNewData.GetUpTime(snmpConnection)
	if err != nil {
		sknchk.Unknown(fmt.Sprint(err), "")
	}

	//Check if interface is admin down, in that no need to process other information.
	if intNewData.IfAdminStatus != nil && *intNewData.IfAdminStatus == netint.DOWN {
		sknchk.Critical(`The interface is administratively <span style="color: hsl(348, 86%%, 61%%);">DOWN</span>`, "")
	}
	if intNewData.IfOperStatus != nil {
		for _, st := range []uint{2, 3, 4, 5, 6, 7} {
			if st == *intNewData.IfOperStatus {
				operStStrg := netint.OperToString(st)
				sknchk.Critical(fmt.Sprintf(`The interface status is <span style="color: hsl(348, 86%%, 61%%);">%v</span> (oper), <span style="color: hsl(348, 86%%, 61%%);">UP</span> (admin)`, operStStrg), "")
			}
		}
	}

	log.Debug("=====================")
	log.Debugf("New network interface values : %#v", *intNewData)

	err = file.CheckFileExist(file.DevicePath, intFilename)
	if err != nil {
		log.Debug("First polling, creation of the first json datas file")
		err = file.CreateJSONFile(file.DevicePath, intFilename, *intNewData)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
		sknchk.Ok("First polling, creation of the initial datas.", "")
	}
	log.Debug("Not First polling, calculation of the elements")
	log.Debug("Read of the old datas")
	data, _ := file.ReadJSONIntFile(file.DevicePath, intFilename)
	intOldData := &netint.InterfaceDetails{}
	err = mapstructure.Decode(data, &intOldData)
	if err != nil {
		sknchk.Unknown(fmt.Sprint(err), "")
	}

	var sysUpTime time.Duration
	if intNewData.UpTime != nil {
		sysUpTime = time.Duration(int64(*intNewData.UpTime/100)) * time.Second
		log.Debugf("Uptime : %v", sysUpTime)
	} else {
		log.Debugf("No Uptime found")
	}

	timeDiff := time.Unix(intNewData.Timestamp, 0).Sub(time.Unix(intOldData.Timestamp, 0))
	log.Debugf("Time diff between the 2 polling : %v", timeDiff.String())

	if intNewData.UpTime != nil && timeDiff > sysUpTime {
		log.Debugf("timediff is upper than the uptime : (%v > %v)", timeDiff.String(), sysUpTime.String())
		//If it's because of a overflow on the uptime counter (device up more than 497 days) we don't reset the old values.
		//we admit that if the old sysUptime is near from the value 2^32-1 (counter32) so the counter simply reset because of an overflow and not a reboot.
		//1 day = 86400 sec
		if intOldData.UpTime == nil {
			log.Debug("Old Uptime value isn't available, skip check of the counter overflow...")
		} else {
			if *intOldData.UpTime > (math.MaxUint32 - 86400) {
				log.Debug("sysUptime counter overflow")
				log.Debug("Don't force the old counter values to 0")
			} else {
				//As the system have rebooted the counter have normally reset. We force the old values to O.
				//We force the time diff to the uptime value
				timeDiff = sysUpTime
				*intOldData.IfInOctets = 0
				*intOldData.IfInUcastPkts = 0
				*intOldData.IfInNUcastPkts = 0
				*intOldData.IfInDiscards = 0
				*intOldData.IfInErrors = 0
				*intOldData.IfOutOctets = 0
				*intOldData.IfOutUcastPkts = 0
				*intOldData.IfOutNUcastPkts = 0
				*intOldData.IfOutDiscards = 0
				*intOldData.IfOutErrors = 0
				*intOldData.IfHCInOctets = 0
				*intOldData.IfHCInUcastPkts = 0
				*intOldData.IfHCInMulticastPkts = 0
				*intOldData.IfHCInBroadcastPkts = 0
				*intOldData.IfHCOutOctets = 0
				*intOldData.IfHCOutUcastPkts = 0
				*intOldData.IfHCOutMulticastPkts = 0
				*intOldData.IfHCOutBroadcastPkts = 0
			}
		}
	}

	//speed is also used for the creation of the bandwidtch perfdata. Need to be called before Bandwidth function
	netint.Speed(intNewData, chk)

	netint.Bandwidth(intNewData, intOldData, timeDiff, chk, bwflag, bcflag)

	netint.Packets(intNewData, intOldData, timeDiff)

	netint.Errors(intNewData, intOldData, timeDiff, chk, ew, ec, ev)

	netint.Discards(intNewData, intOldData, timeDiff, chk, dw, dc, dv)

	netint.DuplexMode(intNewData, chk)

	log.Debug("===== Write New Data to JSON file =====")
	err = file.CreateJSONFile(file.DevicePath, intFilename, *intNewData)
	if err != nil {
		sknchk.Unknown(fmt.Sprint(err), "")
	}

	switch chk.Rc() {
	case sknchk.RcOk:
		chk.PrependShort("No error found on the interface.", false)
	case sknchk.RcWarning:
		chk.PrependShort("Error(s) found on the interface:", false)
	case sknchk.RcCritical:
		chk.PrependShort("Critical Error(s) found on the interface:", false)
	}

	if verbose {
		ui.CliSummary(intNewData, chk)
	} else {
		var tableHTML string
		thresholds := &ui.Thresholds{
			Bw:     bwflag,
			Bc:     bcflag,
			Ewflag: ewflag,
			Ecflag: ecflag,
			Dwflag: dwflag,
			Dcflag: dcflag,
			Ec:     ec,
			Ew:     ew,
			Dc:     dc,
			Dw:     dw,
		}
		tableHTML, err = ui.GenerateHTMLTable(intNewData, thresholds)
		if err != nil {
			sknchk.Unknown(fmt.Sprint(err), "")
		}
		chk.AddLong(tableHTML, false)
	}
	sknchk.Exit(chk)
}
