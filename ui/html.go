package ui

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
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"go-check-network-interface/convert"
	"go-check-network-interface/netint"
)

//TableTmpl is the HTML code to generate the table into the long output
const TableTmpl = `
<table style="width: 100%; border-collapse: collapse; border-color: #000000; margin-left: auto; margin-right: auto; font-family:-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;" border="1">
            <tbody>
              <tr>
                <th colspan="6" style="color: #2160c4; background-color: #eef3fc; padding: 5; text-align: center;">Name : {{if .IfName}}{{.IfName}}{{else}}No name found{{end}} - Desc : {{if .IfDescr}}{{.IfDescr}}{{else}}No description found{{end}}</th>
              </tr>
              <tr>
                <th colspan="6" style="color: #1d72aa; background-color: #eef6fc; padding: 5; text-align: center;">Alias : {{if .IfAlias}}{{if eq (len .IfAlias) 0 }}Alias is empty{{else}}{{.IfAlias}}{{end}}{{else}}No alias found{{end}}</th>
              </tr>
              <tr style="background-color: rgba(0,0,0,.075);">
                <th colspan="2" style="padding: 5;">Oper Status</th>
                <th colspan="2" style="padding: 5;">Admin Status</th>
                <th style="padding: 5;">Speed</th>
                <th style="padding: 5;">Duplex Mode</th>
              </tr>
              <tr>
                {{if eq (StatusIntToStr .IfOperStatus) "UP" -}}
                  <td colspan="2" style="text-align: center; background-color: #d4edda; color: #155724; padding: 5;">{{StatusIntToStr .IfOperStatus}} &#10004;</td>
                {{else if eq (StatusIntToStr .IfOperStatus) "DOWN" -}}
                  <td colspan="2" style="text-align: center; background-color: #f8d7da; color: #721c24; padding: 5;">{{StatusIntToStr .IfOperStatus}} &#10006;</td>
                {{else -}}
                  <td colspan="2" style="text-align: center; background-color: #fff3cd; color: #856404; padding: 5;">{{StatusIntToStr .IfOperStatus}} &#8264;</td>
                {{end -}}
                {{if eq (StatusIntToStr .IfAdminStatus) "UP" -}}
                  <td colspan="2" style="text-align: center; background-color: #d4edda; color: #155724; padding: 5;">{{StatusIntToStr .IfAdminStatus}} &#10004;</td>
                {{else if eq (StatusIntToStr .IfAdminStatus) "DOWN" -}}
                  <td colspan="2" style="text-align: center; background-color: #f8d7da; color: #721c24; padding: 5;">{{StatusIntToStr .IfAdminStatus}} &#10006;</td>
                {{else -}}
                  <td colspan="2" style="text-align: center; background-color: #fff3cd; color: #856404; padding: 5;">{{StatusIntToStr .IfAdminStatus}} &#8264;</td>
                {{end -}}
                <td style="padding: 5;">{{HumanSpeed}}</td>
                {{if eq (DuplexIntToStr .Dot3StatsDuplexStatus) "Unknown" -}}
                  <td style="background-color: #fff3cd; color: #856404; padding: 5;">Unknown</td>
                {{else if eq (DuplexIntToStr .Dot3StatsDuplexStatus) "Half-Duplex" -}}
                  <td style="background-color: #f8d7da; color: #721c24; padding: 5;">Half-Duplex</td>
                {{else if eq (DuplexIntToStr .Dot3StatsDuplexStatus) "Full-Duplex" -}}
                  <td style="background-color: #d4edda; color: #155724; padding: 5;">Full-Duplex</td>
                {{else -}}
                  <td style="padding: 5;">{{ (DuplexIntToStr .Dot3StatsDuplexStatus) }}</td>
                {{end -}}
              </tr>
              <tr style="background-color: rgba(0,0,0,.075)">
                <th colspan="2" style="padding: 5;">In Bandwidth &#10563;</th>
                <th colspan="2" style="padding: 5;">Out Bandwidth &#10562;</th>
                <th style="padding: 5;">Usage Warning<br>threshold</th>
                <th style="padding: 5;">Usage Critical<br>threshold</th>
              </tr>
              <tr>
                {{if .IfInPrct -}}
                {{if eq (CompPnF .IfInPrct BwCritThreshold) 1 -}}
                  <td colspan="2" style="background-color: #f8d7da; color: #721c24; padding: 5;">{{if .IfInRate}}{{HumanBps .IfInRate}}{{end}} &#11020; {{Float2f .IfInPrct}} %</td>
                {{else if eq (CompPnF .IfInPrct BwWarnThreshold) 1 -}}
                  <td colspan="2" style="background-color: #fff3cd; color: #856404; padding: 5;">{{if .IfInRate}}{{HumanBps .IfInRate}}{{end}} &#11020;  {{Float2f .IfInPrct}} %</td>
                {{else -}}
                  <td colspan="2" style="background-color: #d4edda; color: #155724; padding: 5;">{{if .IfInRate}}{{HumanBps .IfInRate}}{{end}} &#11020; {{Float2f .IfInPrct}} %</td>
                {{end -}}
                {{else -}}
                <td colspan="2" style="padding: 5;">N/A</td>
                {{end -}}
                {{if .IfOutPrct -}}
                {{if eq (CompPnF .IfOutPrct BwCritThreshold) 1 -}}
                  <td colspan="2" style="background-color: #f8d7da; color: #721c24; padding: 5;">{{if .IfOutRate}}{{HumanBps .IfOutRate}}{{end}} &#11020; {{Float2f .IfOutPrct}} %</td>
                {{else if eq (CompPnF .IfOutPrct BwWarnThreshold) 1 -}}
                  <td colspan="2" style="background-color: #fff3cd; color: #856404; padding: 5;">{{if .IfOutRate}}{{HumanBps .IfOutRate}}{{end}} &#11020; {{Float2f .IfOutPrct}} %</td>
                {{else -}}
                  <td colspan="2" style="background-color: #d4edda; color: #155724; padding: 5;">{{if .IfOutRate}}{{HumanBps .IfOutRate}}{{end}} &#11020; {{Float2f .IfOutPrct}} %</td>
                {{end -}}
                {{else -}}
                <td colspan="2" style="padding: 5;">N/A</td>
                {{end -}}
                <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{BwWarnThreshold}} %</td>
                <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{BwCritThreshold}} %</td>
              </tr>
              <tr style="background-color: rgba(0,0,0,.075)">
                <th style="padding: 5;">In Packets</th>
                <th style="padding: 5;">In errors</th>
                <th style="padding: 5;">Out Packets</th>
                <th style="padding: 5;">Out errors</th>
                <th style="padding: 5;">Errors Warning<br>threshold</th>
                <th style="padding: 5;">Errors Critical<br>threshold</th>
              </tr>
              <tr>
                <td rowspan="3" style="padding: 5;">{{if .IfInTotalPktsRate}}Total: {{Float2f .IfInTotalPktsRate}} pps<br><br>
                {{if .InUniPcktRate}}Unicast: {{Float2f .InUniPcktRate}} pps<br>{{end -}}
                {{if .InMultiPcktRate}}Multicast: {{Float2f .InMultiPcktRate}} pps<br>{{end -}}
                {{if .InBroadPcktRate}}Broadcast: {{Float2f .InBroadPcktRate}} pps{{end -}}{{end -}}
                </td>
                {{if and (eq ErrUnitThreshold "pps") .IfInErrorsRate -}}
                  {{if eq (CompPnF .IfInErrorsRate ErrCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{Float2f .IfInErrorsRate}} pps &#11020; {{if .IfInErrorsPrct}}{{Float2f .IfInErrorsPrct}} %{{end}}</td>
                  {{else if eq (CompPnF .IfInErrorsRate ErrWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{Float2f .IfInErrorsRate}} pps &#11020; {{if .IfInErrorsPrct}}{{Float2f .IfInErrorsPrct}} %{{end}}</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{Float2f .IfInErrorsRate}} pps &#11020; {{if .IfInErrorsPrct}}{{Float2f .IfInErrorsPrct}} %{{end}}</td>
                  {{end -}}
                {{else if and (eq ErrUnitThreshold "%") .IfInErrorsPrct -}}
                  {{if eq (CompPnF .IfInErrorsPrct ErrCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{if .IfInErrorsRate}}{{Float2f .IfInErrorsRate}} pps{{end}} &#11020; {{Float2f .IfInErrorsPrct}} %</td>
                  {{else if eq (CompPnF .IfInErrorsPrct ErrWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{if .IfInErrorsRate}}{{Float2f .IfInErrorsRate}} pps{{end}} &#11020; {{Float2f .IfInErrorsPrct}} %</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{if .IfInErrorsRate}}{{Float2f .IfInErrorsRate}} pps{{end}} &#11020; {{Float2f .IfInErrorsPrct}} %</td>
                  {{end -}}
                {{else -}}
                  <td style="padding: 5;">N/A</td>
                {{end -}}
                <td rowspan="3" style="padding: 5;">{{if .IfOutTotalPktsRate}}Total : {{Float2f .IfOutTotalPktsRate}} pps<br><br>
                {{if .OutUniPcktRate}}Unicast: {{Float2f .OutUniPcktRate}} pps<br>{{end -}}
                {{if .OutMultiPcktRate}}Multicast: {{Float2f .OutMultiPcktRate}} pps<br>{{end -}}
                {{if .OutBroadPcktRate}}Broadcast: {{Float2f .OutBroadPcktRate}} pps<br>{{end -}}{{end -}}
                </td>
                {{if and (eq ErrUnitThreshold "pps") .IfOutErrorsRate -}}
                  {{if eq (CompPnF .IfOutErrorsRate ErrCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{Float2f .IfOutErrorsRate}} pps &#11020; {{if .IfOutErrorsPrct}}{{Float2f .IfOutErrorsPrct}} %{{end}}</td>
                  {{else if eq (CompPnF .IfOutErrorsRate ErrWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{Float2f .IfOutErrorsRate}} pps &#11020; {{if .IfOutErrorsPrct}}{{Float2f .IfOutErrorsPrct}} %{{end}}</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{Float2f .IfOutErrorsRate}} pps &#11020; {{if .IfOutErrorsPrct}}{{Float2f .IfOutErrorsPrct}} %{{end}}</td>
                  {{end -}}
                {{else if and (eq ErrUnitThreshold "%") .IfOutErrorsPrct -}}
                  {{if eq (CompPnF .IfOutErrorsPrct ErrCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{if .IfOutErrorsRate}}{{Float2f .IfOutErrorsRate}} pps{{end}} &#11020; {{Float2f .IfOutErrorsPrct}} %</td>
                  {{else if eq (CompPnF .IfOutErrorsPrct ErrWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{if .IfOutErrorsRate}}{{Float2f .IfOutErrorsRate}} pps{{end}} &#11020; {{Float2f .IfOutErrorsPrct}} %</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{if .IfOutErrorsRate}}{{Float2f .IfOutErrorsRate}} pps{{end}} &#11020; {{Float2f .IfOutErrorsPrct}} %</td>
                  {{end -}}
                {{else -}}
                  <td style="padding: 5;">N/A</td>
                {{end -}}
                <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{ErrWarnThreshold}} {{ErrUnitThreshold}}</td>
                <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{ErrCritThreshold}} {{ErrUnitThreshold}}</td>
              </tr>
              <tr style="background-color: rgba(0,0,0,.075)">
                <th style="padding: 5;">In discards</th>
                <th style="padding: 5;">Out discards</th>
                <th style="padding: 5;">Discards Warning<br>threshold</th>
                <th style="padding: 5;">Discards Critical<br>threshold</th>
              </tr>
              <tr>
                {{if and (eq DisUnitThreshold "pps") .IfInDiscardsRate -}}
                  {{if eq (CompPnF .IfInDiscardsRate DisCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{Float2f .IfInDiscardsRate}} pps &#11020; {{if .IfInDiscardsPrct}}{{Float2f .IfInDiscardsPrct}} %{{end}}</td>
                  {{else if eq (CompPnF .IfInDiscardsRate DisWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{Float2f .IfInDiscardsRate}} pps &#11020; {{if .IfInDiscardsPrct}}{{Float2f .IfInDiscardsPrct}} %{{end}}</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{Float2f .IfInDiscardsRate}} pps &#11020; {{if .IfInDiscardsPrct}}{{Float2f .IfInDiscardsPrct}} %{{end}}</td>
                  {{end -}}
                {{else if and (eq DisUnitThreshold "%") .IfInDiscardsPrct -}}
                  {{if eq (CompPnF .IfInDiscardsPrct DisCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{if .IfInDiscardsRate}}{{Float2f .IfInDiscardsRate}} pps{{end}} &#11020; {{Float2f .IfInDiscardsPrct}} %</td>
                  {{else if eq (CompPnF .IfInDiscardsPrct DisWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{if .IfInDiscardsRate}}{{Float2f .IfInDiscardsRate}} pps{{end}} &#11020; {{Float2f .IfInDiscardsPrct}} %</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{if .IfInDiscardsRate}}{{Float2f .IfInDiscardsRate}} pps{{end}} &#11020; {{Float2f .IfInDiscardsPrct}} %</td>
                  {{end -}}
                {{else -}}
                  <td style="padding: 5;">N/A</td>
                {{end -}}

                {{if and (eq DisUnitThreshold "pps") .IfOutDiscardsRate -}}
                  {{if eq (CompPnF .IfOutDiscardsRate DisCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{Float2f .IfOutDiscardsRate}} pps &#11020; {{if .IfOutDiscardsPrct}}{{Float2f .IfOutDiscardsPrct}} %{{end}}</td>
                  {{else if eq (CompPnF .IfOutDiscardsRate DisWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{Float2f .IfOutDiscardsRate}} pps &#11020; {{if .IfOutDiscardsPrct}}{{Float2f .IfOutDiscardsPrct}} %{{end}}</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{Float2f .IfOutDiscardsRate}} pps &#11020; {{if .IfOutDiscardsPrct}}{{Float2f .IfOutDiscardsPrct}} %{{end}}</td>
                  {{end -}}
                {{else if and (eq DisUnitThreshold "%") .IfOutDiscardsPrct -}}
                  {{if eq (CompPnF .IfOutDiscardsPrct DisCritThreshold) 1 -}}
                    <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{if .IfOutDiscardsRate}}{{Float2f .IfOutDiscardsRate}} pps{{end}} &#11020; {{Float2f .IfOutDiscardsPrct}} %</td>
                  {{else if eq (CompPnF .IfOutDiscardsPrct DisWarnThreshold) 1 -}}
                    <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{if .IfOutDiscardsRate}}{{Float2f .IfOutDiscardsRate}} pps{{end}} &#11020; {{Float2f .IfOutDiscardsPrct}} %</td>
                  {{else -}}
                    <td style="background-color: #d4edda; color: #155724; padding: 5;">{{if .IfOutDiscardsRate}}{{Float2f .IfOutDiscardsRate}} pps{{end}} &#11020; {{Float2f .IfOutDiscardsPrct}} %</td>
                  {{end -}}
                {{else -}}
                <td style="padding: 5;">N/A</td>
                {{end -}}
                <td style="background-color: #fff3cd; color: #856404; padding: 5;">{{DisWarnThreshold}} {{DisUnitThreshold}}</td>
                <td style="background-color: #f8d7da; color: #721c24; padding: 5;">{{DisCritThreshold}} {{DisUnitThreshold}}</td>
              </tr>
            </tbody>
            </table>
`

//Thresholds is used to transfert thresholds value to build the table HTML template
type Thresholds struct {
	Bw, Bc                         float64
	Ewflag, Ecflag, Dwflag, Dcflag string
	Ew, Ec, Dw, Dc                 float64
}

//GenerateHTMLTable generate the HTML table of the long output details in string format
func GenerateHTMLTable(intNewData *netint.InterfaceDetails, threshold *Thresholds) (string, error) {
	t := template.Must(template.New("table").Funcs(template.FuncMap{
		"Float2f": func(f float64) string { return fmt.Sprintf("%.2f", f) },
		"StatusIntToStr": func(st *uint) string {
			if st != nil {
				return netint.OperToString(*st)
			}
			return "N/A"
		},
		"DuplexIntToStr": func(dp *uint) string {
			if dp != nil {
				return netint.DuplexToString(*dp)
			}
			return "N/A"
		},
		"HumanBps":         func(f float64) string { return convert.HumanReadable(f, "bits/sec") },
		"HumanSpeed":       func() string { return convert.HumanReadable(float64(*intNewData.SpeedInbit), "bps") },
		"BwCritThreshold":  func() float64 { return threshold.Bc },
		"BwWarnThreshold":  func() float64 { return threshold.Bw },
		"ErrCritThreshold": func() float64 { return threshold.Ec },
		"ErrWarnThreshold": func() float64 { return threshold.Ew },
		"ErrUnitThreshold": func() string {
			if strings.Contains(threshold.Ecflag, "pps") {
				return "pps"
			}
			return "%"
		},
		"DisCritThreshold": func() float64 { return threshold.Dc },
		"DisWarnThreshold": func() float64 { return threshold.Dw },
		"DisUnitThreshold": func() string {
			if strings.Contains(threshold.Dcflag, "pps") {
				return "pps"
			}
			return "%"
		},
		"CompPnF": func(f1 *float64, f2 float64) int {
			if *f1 > f2 {
				return 1
			} else if *f1 == f2 {
				return 0
			} else {
				return -1
			}
		},
	}).Parse(TableTmpl))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, intNewData)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}
