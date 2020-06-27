package ui

import (
	"fmt"

	"go-check-network-interface/convert"
	"go-check-network-interface/netint"

	sknchk "github.com/pandaoc-io/go-shinken-check"
)

//CliSummary is used to generate a CLI human friendly output, used in debug mode
func CliSummary(intNewData *netint.InterfaceDetails, chk *sknchk.Check) {
	if sknchk.Output.Mode() == "html" {
		return
	}
	chk.AddShort("====== Interface Details =======", false)
	if intNewData.IfName != nil {
		chk.AddShort(fmt.Sprintf("Name : %v", *intNewData.IfName), true)
	}
	if intNewData.IfDescr != nil {
		chk.AddShort(fmt.Sprintf("Descr : %v", *intNewData.IfDescr), true)
	}
	if intNewData.IfAlias != nil {
		chk.AddShort(fmt.Sprintf("Alias : %v", *intNewData.IfAlias), true)
	}
	chk.AddShort(fmt.Sprintf("Speed : %v", convert.HumanReadable(float64(*intNewData.SpeedInbit), "bps")), true)
	if intNewData.IfOperStatus != nil {
		if *intNewData.IfOperStatus == netint.UP {
			chk.AddShort("Oper Status : UP", true)
		} else if *intNewData.IfOperStatus == netint.DOWN {
			chk.AddShort("Oper Status : DOWN", true)
			chk.AddCritical()
		} else {
			chk.AddShort(fmt.Sprintf("Oper Status : %v", intNewData.IfOperStatus), true)
		}
	} else {
		chk.AddShort("Oper Status : Can't be determined", true)
	}
	if intNewData.IfAdminStatus != nil {
		if *intNewData.IfAdminStatus == netint.UP {
			chk.AddShort("Admin Status : UP", true)
		} else {
			chk.AddShort("Admin Status : DOWN", true)
		}
	} else {
		chk.AddShort("Admin Status : Can't be determined", true)
	}
	chk.AddShort(fmt.Sprintf("Total pkts In : %.2f pps", *intNewData.IfInTotalPktsRate), true)
	chk.AddShort(fmt.Sprintf("Total pkts Out : %.2f pps", *intNewData.IfOutTotalPktsRate), true)

	if intNewData.InUniPcktRate != nil || intNewData.InMultiPcktRate != nil {
		if intNewData.InBroadPcktRate != nil {
			chk.AddShort(fmt.Sprintf("In : Uni : %.2f pps, Multi : %.2f pps, Broad :%.2f pps", *intNewData.InUniPcktRate, *intNewData.InMultiPcktRate, *intNewData.InBroadPcktRate), true)
		} else {
			chk.AddShort(fmt.Sprintf("In : Uni : %.2f pps, Multi/Broad : %.2f pps", *intNewData.InUniPcktRate, *intNewData.InMultiPcktRate), true)
		}
	} else {
		chk.AddShort("In Uni/Multi/Broadcast packets : Can't be determined", true)
	}
	if intNewData.OutUniPcktRate != nil || intNewData.OutMultiPcktRate != nil {
		if intNewData.OutBroadPcktRate != nil {
			chk.AddShort(fmt.Sprintf("Out : Uni : %.2f pps, Multi : %.2f pps, Broad :%.2f pps", *intNewData.OutUniPcktRate, *intNewData.OutMultiPcktRate, *intNewData.OutBroadPcktRate), true)
		} else {
			chk.AddShort(fmt.Sprintf("Out : Uni : %.2f pps, Multi/Broad : %.2f pps", *intNewData.OutUniPcktRate, *intNewData.OutMultiPcktRate), true)
		}
	} else {
		chk.AddShort("Out Uni/Multi/Broadcast packets : Can't be determined", true)
	}
	var inRateStr string
	if intNewData.IfInRate != nil {
		inRateStr = fmt.Sprintf("In BW : %v", convert.HumanReadable(*intNewData.IfInRate, "bits/sec"))
	} else {
		inRateStr = "In BW : Rate can't be determined"
	}
	var inPrctStr string
	if intNewData.IfInPrct != nil {
		inPrctStr = fmt.Sprintf(" (%.2f%%)", *intNewData.IfInPrct)
	} else {
		inPrctStr = ", Percentage can't be determined"
	}

	chk.AddShort(inRateStr+inPrctStr, true)

	var outRateStr string
	if intNewData.IfOutRate != nil {
		outRateStr = fmt.Sprintf("In BW : %v", convert.HumanReadable(*intNewData.IfOutRate, "bits/sec"))
	} else {
		outRateStr = "Out BW : Rate can't be determined"
	}
	var outPrctStr string
	if intNewData.IfOutPrct != nil {
		outPrctStr = fmt.Sprintf(" (%.2f%%)", *intNewData.IfOutPrct)
	} else {
		outPrctStr = ", Percentage can't be determined"
	}

	chk.AddShort(outRateStr+outPrctStr, true)

	if intNewData.IfInErrorsRate != nil {
		if intNewData.LocIfInCRCRate != nil {
			chk.AddShort(fmt.Sprintf("In Errors : %.2f pps (%.2f%%) with %.2f pps CRC Errors", *intNewData.IfInErrorsRate, *intNewData.IfInErrorsPrct, *intNewData.LocIfInCRCRate), true)
		} else {
			chk.AddShort(fmt.Sprintf("In Errors : %.2f pps (%.2f%%), no additional CRC Error stat", *intNewData.IfInErrorsRate, *intNewData.IfInErrorsPrct), true)
		}
	} else {
		chk.AddShort("In Errors : Can't be determined", true)
	}
	if intNewData.IfOutErrorsPrct != nil {
		chk.AddShort(fmt.Sprintf("Out Errors : %.2f pps (%.2f%%)", *intNewData.IfOutErrorsPrct, *intNewData.IfOutErrorsPrct), true)
	} else {
		chk.AddShort("Out Errors : Can't be determined", true)
	}
	if intNewData.IfInDiscardsRate != nil {
		chk.AddShort(fmt.Sprintf("In Discards : %.2f pps (%.2f%%)", *intNewData.IfInDiscardsRate, *intNewData.IfInDiscardsPrct), true)
	} else {
		chk.AddShort("In Discards : Can't be determined", true)
	}
	if intNewData.IfOutDiscardsRate != nil {
		chk.AddShort(fmt.Sprintf("Out Discards : %.2f pps (%.2f%%)", *intNewData.IfOutDiscardsRate, *intNewData.IfOutDiscardsPrct), true)
	} else {
		chk.AddShort("Out Discards : Can't be determined", true)
	}
}
