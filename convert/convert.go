package convert

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
)

//HumanReadable convert float64 value to something more readable in K, M, G, T with a 2 points precision
func HumanReadable(value float64, base float64, suffix string) string {
	if value < base {
		return fmt.Sprintf("%.0f %v", value, suffix)
	}
	div, exp := base, 0
	for n := value / base; n >= base; n /= base {
		div *= base
		exp++
	}
	return fmt.Sprintf("%.2f %c%v", float64(value)/float64(div), "kMGTPE"[exp], suffix)
}

//ToUint convert int and uint types to uint
func ToUint(value interface{}) (uint, error) {
	switch value := value.(type) { //shadow
	case int:
		return uint(value), nil
	case int8:
		return uint(value), nil
	case int16:
		return uint(value), nil
	case int32:
		return uint(value), nil
	case int64:
		return uint(value), nil
	case uint:
		return uint(value), nil
	case uint8:
		return uint(value), nil
	case uint16:
		return uint(value), nil
	case uint32:
		return uint(value), nil
	case uint64:
		return uint(value), nil
	case *uint:
		return uint(*value), nil
	case *uint32:
		return uint(*value), nil
	case *uint64:
		return uint(*value), nil
	default:
		return 0, fmt.Errorf("Unsupported Type %T for Float64 conversion", value)
	}
}

//ToFloat convert int, uint and float32 to float64
func ToFloat(value interface{}) (float64, error) {
	switch value := value.(type) {
	case int:
		return float64(value), nil
	case int8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	default:
		return 0, fmt.Errorf("Unsupported Type %T for Float64 conversion", value)
	}
}
