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

var ifEntryBaseOid = ".1.3.6.1.2.1.2.2.1"
var ifXEntryBaseOid = ".1.3.6.1.2.1.31.1.1.1"

//InterfaceOids containe all the OIDs used to grab the interface information
var InterfaceOids = map[string]string{
	"HrSystemUptime":        ".1.3.6.1.2.1.25.1.1.0",
	"SysUpTime":             ".1.3.6.1.2.1.1.3.0",
	"LocIfInCRC":            ".1.3.6.1.4.1.9.2.2.1.1.12",
	"Dot3StatsDuplexStatus": ".1.3.6.1.2.1.10.7.2.1.19",
	"IfIndex":               ifEntryBaseOid + ".1",
	"IfDescr":               ifEntryBaseOid + ".2",
	"IfSpeed":               ifEntryBaseOid + ".5",
	"IfAdminStatus":         ifEntryBaseOid + ".7",
	"IfOperStatus":          ifEntryBaseOid + ".8",
	"IfInOctets":            ifEntryBaseOid + ".10",
	"IfInUcastPkts":         ifEntryBaseOid + ".11",
	"IfInNUcastPkts":        ifEntryBaseOid + ".12",
	"IfInDiscards":          ifEntryBaseOid + ".13",
	"IfInErrors":            ifEntryBaseOid + ".14",
	"IfOutOctets":           ifEntryBaseOid + ".16",
	"IfOutUcastPkts":        ifEntryBaseOid + ".17",
	"IfOutNUcastPkts":       ifEntryBaseOid + ".18",
	"IfOutDiscards":         ifEntryBaseOid + ".19",
	"IfOutErrors":           ifEntryBaseOid + ".20",
	"IfName":                ifXEntryBaseOid + ".1",
	"IfHCInOctets":          ifXEntryBaseOid + ".6",
	"IfHCInUcastPkts":       ifXEntryBaseOid + ".7",
	"IfHCInMulticastPkts":   ifXEntryBaseOid + ".8",
	"IfHCInBroadcastPkts":   ifXEntryBaseOid + ".9",
	"IfHCOutOctets":         ifXEntryBaseOid + ".10",
	"IfHCOutUcastPkts":      ifXEntryBaseOid + ".11",
	"IfHCOutMulticastPkts":  ifXEntryBaseOid + ".12",
	"IfHCOutBroadcastPkts":  ifXEntryBaseOid + ".13",
	"IfHighSpeed":           ifXEntryBaseOid + ".15",
	"IfAlias":               ifXEntryBaseOid + ".18",
}
