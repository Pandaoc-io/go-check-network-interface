package file

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"go-check-network-interface/netint"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

//CheckPath is the global path where all the files read/write by the check will be hosted
const CheckPath = "/var/tmp/go_check_snmp_interface_foreach"

//DevicePath is the path where is will be stored all the JSON files of the device
var DevicePath string

//CreatePath will try to create the path give in argument.
func CreatePath(fPath string) error {
	fileInfo, err := os.Stat(fPath)
	if os.IsNotExist(err) {
		log.Debugln("path doesn't exist, create it")
		err := os.MkdirAll(fPath, 0755)
		if err != nil {
			return fmt.Errorf("%v directory can't be created : %v", fPath, err)
		}
	}
	err = unix.Access(fPath, unix.W_OK)
	if err != nil {
		return fmt.Errorf("%v directory isn't writable, mode : %v", fPath, fileInfo.Mode())
	}
	return nil
}

//AsExp check if the file is older than the given expiration duration
func AsExp(devicePath string, filename string, exptime time.Duration) (bool, error) {
	log.Debugln("===== File expiration check =====")
	file, err := os.Stat(path.Join(devicePath, filename))
	log.Debugf("Tested file : %v", path.Join(devicePath, filename))
	if err != nil {
		return false, err
	}
	log.Debugf("Expiration delay : %v", exptime)
	modifiedtime := file.ModTime()
	now := time.Now()
	elapsed := now.Sub(modifiedtime)
	log.Debugf("Last modification time : %s ago", elapsed)
	if elapsed > exptime {
		log.Debugln("The file have expired")
		return true, nil
	}
	log.Debugln("The file haven't expired")
	return false, nil
}

//CreateJSONFile will create a json file bases on the interface{} datas.
func CreateJSONFile(devicePath string, filename string, datas interface{}) error {
	log.Debugln("===== JSON file creation =====")
	log.Debugf("File to create : %v", path.Join(devicePath, filename))
	if _, err := os.Stat(path.Join(devicePath, filename)); err != nil {
		err = CreatePath(devicePath)
		if err != nil {
			return err
		}
	}
	datasBytes, err := json.MarshalIndent(datas, "", "    ")
	if err != nil {
		return fmt.Errorf("Can't Marshal the datas : %v, Datas : %#v", err, datas)
	}
	jsonStr := string(datasBytes)
	log.Debug("The JSON data is : ", jsonStr)

	err = ioutil.WriteFile(path.Join(devicePath, filename), datasBytes, 0644)
	if err != nil {
		return fmt.Errorf("Can't create the file : %v", err)
	}
	return nil

}

//GenDeviceDirName will string the full path used to read/write the interface information
//for a given device
func GenDeviceDirName(version string, cmd *cobra.Command) string {
	dIP, _ := cmd.Flags().GetString("hostname")
	context, _ := cmd.Flags().GetString("context")
	deviceDir := ""
	if len(context) > 0 {
		deviceDir = dIP + "_SNMPv" + version + "_" + context
	} else {
		deviceDir = dIP + "_SNMPv" + version
	}
	return path.Join(CheckPath, deviceDir)
}

//CheckFileExist will check if the file 'filename' exist
func CheckFileExist(devicePath string, filename string) error {
	fullPath := path.Join(devicePath, filename)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

//IsPathWritable check if the path is writable
func IsPathWritable(path string) error {
	fileInfo, _ := os.Stat(path)
	err := unix.Access(path, unix.W_OK)
	if err != nil {
		return fmt.Errorf("%v isn't writable, mode : %v", path, fileInfo.Mode())
	}
	return nil
}

//FindIntIndex read the index.json file and return the interface if found
// and an error if not found or if issue reading the file
func FindIntIndex(devicePath string, name string) (string, error) {
	log.Debugln("===== Search of the interface Index =====")
	jsonFile, err := os.Open(path.Join(devicePath, "index.json"))
	if err != nil {
		return "", fmt.Errorf("Find Interface Index: %v", err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var indexMap map[string]interface{}
	json.Unmarshal([]byte(byteValue), &indexMap)

	for k, v := range indexMap {
		value := v.(map[string]interface{})
		if value["IfDescr"] == name || value["IfName"] == name {
			log.Debugf("Interface Index found : %v", k)
			return k, nil
		}
	}
	return "", fmt.Errorf("Index for interface %v not found", name)
}

//ReadJSONIntFile will read the interface values from the interface JSON file and generate the corresponding struct
func ReadJSONIntFile(devicePath string, filename string) (interface{}, error) {
	fullPath := path.Join(devicePath, filename)
	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("Find Interface Index: %v", err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var elements netint.InterfaceDetails
	json.Unmarshal([]byte(byteValue), &elements)
	return elements, nil
}
