package sknchk

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
	"strings"
)

//PerfData is the nagios style perf element with the following output pattern : 'label'=value[UOM];[warn];[crit];[min];[max]
type PerfData struct {
	Name  string
	Value interface{}
	Unit  string
	Warn  interface{}
	Crit  interface{}
	Min   interface{}
	Max   interface{}
}

//AddPerfData add a new perfdata to the check
func (c *Check) AddPerfData(name string, value interface{}, unit string, warn interface{}, crit interface{}, min interface{}, max interface{}) {
	c.perfData = append(c.perfData, &PerfData{
		Name:  name,
		Value: value,
		Unit:  unit,
		Warn:  warn,
		Crit:  crit,
		Min:   min,
		Max:   max})
}

func generatePerfOutput(perf []*PerfData) string {
	var perfsSlice []string
	for _, p := range perf {
		perfStr := fmt.Sprintf("%v=%v%v;%v;%v;%v;%v", p.Name, p.Value, p.Unit, p.Warn, p.Crit, p.Min, p.Max)
		perfsSlice = append(perfsSlice, perfStr)
	}
	return strings.Join(perfsSlice, " ")
}
