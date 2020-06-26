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
	"os"
	"strings"
)

//Standard Shinken/Nagios-like Return Code
const (
	RcOk Status = iota
	RcWarning
	RcCritical
	RcUnknwon
)

//Prefix used for the final output
const (
	PrefixCliOk        string = "[OK]"
	PrefixCliWarning   string = "[WARNING]"
	PrefixCliCritical  string = "[CRITICAL]"
	PrefixCliUnknown   string = "[UNKNOWN]"
	PrefixHTMLOk       string = `<span style="align-items: center; background-color: #28a745; border-radius: 4px; color: white; display: inline-flex; font-size: 12px; height: 2rem; justify-content: center; line-height: 1.5; padding-left: .75rem; padding-right: .75rem; white-space: nowrap; margin-top: 0.25rem; margin-left: .25rem;">OK</span>`
	PrefixHTMLWarning  string = `<span style="align-items: center; background-color: #ffc107; border-radius: 4px; color: #212529; display: inline-flex; font-size: 112px; height: 2rem; justify-content: center; line-height: 1.5; padding-left: .75rem; padding-right: .75rem; white-space: nowrap; margin-top: 0.25rem; margin-left: .25rem;">Warning</span>`
	PrefixHTMLCritical string = `<span style="align-items: center; background-color: #dc3545; border-radius: 4px; color: white; display: inline-flex; font-size: 12px; height: 2rem; justify-content: center; line-height: 1.5; padding-left: .75rem; padding-right: .75rem; white-space: nowrap; margin-top: 0.25rem; margin-left: .25rem;">Critical</span>`
	PrefixHTMLUnknown  string = `<span style="align-items: center; background-color: #6c757d; border-radius: 4px; color: white; display: inline-flex; font-size: 12px; height: 2rem; justify-content: center; line-height: 1.5; padding-left: .75rem; padding-right: .75rem; white-space: nowrap; margin-top: 0.25rem; margin-left: .25rem;">Unknown</span>`
)

//Status type used to define the status of the check
type Status int

//OutputMode to display the check result, it can be 'cli' or 'html'
type OutputMode struct {
	mode          string
	newLine       string
	bullet        string
	classOk       string
	classWarning  string
	classCritical string
}

//Output is a global variable that define the type of output, 'cli' by default
var Output *OutputMode = &OutputMode{
	mode:          "cli",
	newLine:       "",
	bullet:        " - ",
	classOk:       "",
	classWarning:  "",
	classCritical: "",
}

//Check struct
type Check struct {
	short    []string
	long     []string
	perfData []*PerfData
	rc       []Status
}

//Mode return the current output mode type
func (o *OutputMode) Mode() string {
	return o.mode
}

//SetHTML will set the output format to html format
func (o *OutputMode) SetHTML() {
	Output.mode = "html"
	Output.newLine = "<br />"
	Output.bullet = "&#8226;&#8194;"
	Output.classOk = "color: #28a745!important;"
	Output.classWarning = "color: #947600!important;"
	Output.classCritical = "color: #dc3545!important;"
}

//SetDebug will set the output format to debug format
func (o *OutputMode) SetDebug() {
	Output.mode = "debug"
	Output.newLine = "\n"
	Output.bullet = "- "
	Output.classOk = ""
	Output.classWarning = ""
	Output.classCritical = ""
}

func fmtOutput(str string, class string) string {
	if Output.mode == "html" {
		return fmt.Sprintf(`<span style="%v">%v</span>`, class, str)
	}
	return str
}

//FmtOk will format the OK output string depending on the output choosen mode
func FmtOk(str string) string {
	return fmtOutput(str, Output.classOk)
}

//FmtWarning will format the Warning output string depending on the output choosen mode
func FmtWarning(str string) string {
	return fmtOutput(str, Output.classWarning)
}

//FmtCritical will format the Critical output string depending on the output choosen mode
func FmtCritical(str string) string {
	return fmtOutput(str, Output.classCritical)
}

//AddShort add a new string to the short output
func (c *Check) AddShort(short string, bullet bool) {
	if bullet {
		c.short = append(c.short, Output.bullet+short)
	} else {
		c.short = append(c.short, short)
	}
}

//PrependShort add a new string at the begining of the short output
func (c *Check) PrependShort(short string, bullet bool) {
	if bullet {
		c.short = append([]string{Output.bullet + short}, c.short...)
	} else {
		c.short = append([]string{short}, c.short...)
	}
}

//AddLong add a new string to the long output
func (c *Check) AddLong(long string, bullet bool) {
	if bullet {
		c.long = append(c.long, Output.bullet+long)
	} else {
		c.long = append(c.long, long)
	}
}

//AddOk will add an Ok status to the Check structure
//to prepare the final output/RC
func (c *Check) AddOk() {
	c.rc = append(c.rc, RcOk)
}

//AddWarning will add an Warning status to the Check structure
//to prepare the final output/RC
func (c *Check) AddWarning() {
	c.rc = append(c.rc, RcWarning)
}

//AddCritical will add an Critical status with some short and long information to the Check structure
//to prepare the final output/RC
func (c *Check) AddCritical() {
	c.rc = append(c.rc, RcCritical)
}

//AddUnknown will add an Unknown status with some short and long information to the Check structure
//to prepare the final output/RC
func (c *Check) AddUnknown() {
	c.rc = append(c.rc, RcUnknwon)
}

//Ok will exit the program with the OK status
func Ok(short string, long string) {
	var sLong []string
	if len(long) > 0 {
		sLong = append(sLong, long)
	}
	check := &Check{[]string{short}, sLong, nil, []Status{RcOk}}
	Exit(check)
}

//Warning will exit the program with the Warning status
func Warning(short string, long string) {
	var sLong []string
	if len(long) > 0 {
		sLong = append(sLong, long)
	}
	check := &Check{[]string{short}, sLong, nil, []Status{RcWarning}}
	Exit(check)
}

//Critical will exit the program with the Critical status
func Critical(short string, long string) {
	var sLong []string
	if len(long) > 0 {
		sLong = append(sLong, long)
	}
	check := &Check{[]string{short}, sLong, nil, []Status{RcCritical}}
	Exit(check)
}

//Unknown will exit the program with the Unknown status
func Unknown(short string, long string) {
	var sLong []string
	if len(long) > 0 {
		sLong = append(sLong, long)
	}
	check := &Check{[]string{short}, sLong, nil, []Status{RcUnknwon}}
	Exit(check)
}

//Rc return the Return Code of the check
func (c *Check) Rc() Status {
	var maxRc Status
	for _, value := range c.rc {
		if value > maxRc {
			maxRc = value
		}
	}
	return maxRc
}

//Exit quit the program displaying the short and long output with the
func Exit(c *Check) {
	rc := c.Rc()
	var prefix string
	if Output.mode == "html" {
		switch rc {
		case RcOk:
			prefix = PrefixHTMLOk
		case RcWarning:
			prefix = PrefixHTMLWarning
		case RcCritical:
			prefix = PrefixHTMLCritical
		case RcUnknwon:
			prefix = PrefixHTMLUnknown
		}
	} else {
		switch rc {
		case RcOk:
			prefix = PrefixCliOk
		case RcWarning:
			prefix = PrefixCliWarning
		case RcCritical:
			prefix = PrefixCliCritical
		case RcUnknwon:
			prefix = PrefixCliUnknown
		}
	}

	perfStr := ""
	if len(c.perfData) > 0 {
		perfStr = "|" + generatePerfOutput(c.perfData)
	}

	if len(c.long) > 0 {
		c.AddShort(Output.newLine+"For more details see long output.", false)
		fmt.Fprintf(os.Stdout, "%v %v\n%v%v", prefix, strings.Join(c.short, Output.newLine), strings.Join(c.long, Output.newLine), perfStr)
	} else {
		fmt.Fprintf(os.Stdout, "%v %v%v", prefix, strings.Join(c.short, Output.newLine), perfStr)
	}
	os.Exit(int(rc))
}
