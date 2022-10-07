package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var _varNames envVarNames
var _varValues envVarValuesSets

var _setsOfValuesModeEnabled bool

var VAR_NAMES_STORAGE_PATH = os.Getenv("VAR_NAMES_STORAGE_PATH")
var VAR_NAMES_STORAGE = os.Getenv("VAR_NAMES_STORAGE")
var VAR_VALUES_SETS_STORAGE = os.Getenv("VAR_VALUES_SETS_STORAGE")
var VAR_VALUES_SETS_CHOSEN_ID = os.Getenv("VAR_VALUES_SETS_CHOSEN_ID")

var announcementTemplate =
`
##### START SEV ANNOUNCEMENT #####

## Target environment variables
_{TARGET_VAR_NAMES_PLACEHOLDER}_

## Sets of values
_{SETS_OF_VALUES}_

## Mode
_{MODE}_

###### END SEV ANNOUNCEMENT ######
`

type envVarValuesSets map[string]interface{}
type envVarNames []string

func (evv *envVarValuesSets) fetch(){
	if VAR_VALUES_SETS_STORAGE != "" {

		var tmp interface{}

		err := json.Unmarshal([]byte(VAR_VALUES_SETS_STORAGE), &tmp)
		*evv = tmp.(map[string]interface{})

		if err != nil {
			log.Fatalf(`ERROR: Unable to unmarshal json structure VAR_VALUES_SETS_STORAGE: %v`, err)
		}

		if VAR_VALUES_SETS_CHOSEN_ID == "" {
			log.Fatalf("ERROR: Values Set ID not initialized: please set VAR_VALUES_SETS_CHOSEN_ID")
		}

		_setsOfValuesModeEnabled = true

	}
}

func (evv envVarValuesSets) getAnnouncement() string{

	if _setsOfValuesModeEnabled {

		var announcement string
		var activeValuesSetIdentified = false

		for id, _ := range evv {
			var pointer = "  "

			if id == VAR_VALUES_SETS_CHOSEN_ID {
				activeValuesSetIdentified = true
				pointer = "► "
			}

			announcement += fmt.Sprintf("%s%s\n", pointer, id)

		}

		if !activeValuesSetIdentified {
			log.Fatalf("ERROR: Chosen ID {%s} is missing in the list of available sets:\n %s", VAR_VALUES_SETS_CHOSEN_ID, announcement)
		} else {

			announcement += "\n\n"
			announcement += "## Fetched values\n"

			for key, value := range evv[VAR_VALUES_SETS_CHOSEN_ID].(map[string]interface{}) {


				var resValue string

				switch reflect.TypeOf(value).String(){
				case "float64":
					var f = value.(float64)
					ipart := int64(f)
					decpart := fmt.Sprintf("%.6g", f-float64(ipart))
					if len(decpart) == 1{
						resValue = fmt.Sprintf(`%.0f`, value)
					} else {
						resValue = fmt.Sprintf(`%f`, value)
					}

				case "bool":
					resValue = fmt.Sprintf(`%t`, value)
				default:
					resValue = value.(string)
				}


				//if reflect.TypeOf(value)
				announcement += fmt.Sprintf("  %s = %s\n", key, resValue)
			}
		}

		return announcement
	} else {
		return "N/A"
	}
}

func (evn envVarNames) getAnnouncement() string{
	var announcement string

	for _, ev := range evn {
		announcement += fmt.Sprintf("  %s = %s\n", ev, os.Getenv(ev))
	}

	return announcement

}

func (evn *envVarNames) fetch(){

	if VAR_NAMES_STORAGE == "" && VAR_NAMES_STORAGE_PATH == ""{
		log.Fatalf("ERROR: Not VAR_NAMES_STORAGE nor VAR_NAMES_STORAGE_PATH initialized")
	}

	if VAR_NAMES_STORAGE_PATH != "" {
		envVars, err := ioutil.ReadFile(VAR_NAMES_STORAGE_PATH)
		if err != nil {
			log.Fatalf("ERROR: Failed reading storage in %s:\n%v", VAR_NAMES_STORAGE_PATH)
		} else {
			*evn = strings.Split(string(envVars), "\n")
		}
	}

	if VAR_NAMES_STORAGE != ""{
		*evn = strings.Split(VAR_NAMES_STORAGE, ",")
	}

	// Remove whitespaces
	for i, ev := range *evn{
		(*evn)[i] = strings.ReplaceAll(strings.ReplaceAll(ev, " ", ""), "\t", "")
	}
}




func main(){

	if len(os.Args) == 1 {
		log.Fatalf("ERROR: destination path missing")
	}

	var _path = os.Args[1]

	_varNames.fetch()
	vna := _varNames.getAnnouncement()

	_varValues.fetch()
	vva := _varValues.getAnnouncement()

	mode := "  Pulling values for variables from environment"
	if _setsOfValuesModeEnabled{
		mode = "  Pulling sets of values for variables from JSON structure saved at {env.VAR_VALUES_SETS_STORAGE}"
	}

	var res = strings.ReplaceAll(announcementTemplate, "_{TARGET_VAR_NAMES_PLACEHOLDER}_", vna)
	res = strings.ReplaceAll(res, "_{SETS_OF_VALUES}_", vva)
	res = strings.ReplaceAll(res, "_{MODE}_", mode)
	fmt.Println(res)

	stat, pathExists, _ := pathExists(_path)

	if pathExists{
		switch {
		case stat.IsDir():
			_varNames.processDir(_path)
		case ! stat.IsDir():
			_varNames.processFile(_path, stat.Mode())
		}
	}
}

func (evn envVarNames) processDir(_path string){
	err := filepath.Walk(_path,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}
			if ! info.IsDir() {

				evn.processFile(path, info.Mode())

			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

func (evn envVarNames) processFile(_filePath string, _fileMode os.FileMode){

	var contentBytes, err = ioutil.ReadFile(_filePath)
	var resContent = string(contentBytes)

	if err != nil {
		log.Fatalf("ERROR: Error reading file %s:\n%v", _filePath, err)
	} else {

		//var content = string(contentBytes)
		//var resContent = content

		for _, envVar := range evn{

			envVarValue := os.Getenv(envVar)

			if envVarValue != "" {

				evPlaceholder := fmt.Sprintf("_{%s}_", envVar)
				resContent = strings.ReplaceAll(resContent, evPlaceholder, envVarValue)

			} else {
				log.Printf("WARNING: Missing environment variable {%s}", envVar)
			}
		}

		if _setsOfValuesModeEnabled {

			for key, value := range _varValues[VAR_VALUES_SETS_CHOSEN_ID].(map[string]interface{}) {

				var resValue string

				switch reflect.TypeOf(value).String() {
				case "float64":
					var f = value.(float64)
					ipart := int64(f)
					decpart := fmt.Sprintf("%.6g", f-float64(ipart))
					if len(decpart) == 1 {
						resValue = fmt.Sprintf(`%.0f`, value)
					} else {
						resValue = fmt.Sprintf(`%f`, value)
					}

				case "bool":
					resValue = fmt.Sprintf(`%t`, value)
				default:
					resValue = value.(string)
				}

				if resValue != "" {

					evPlaceholder := fmt.Sprintf("_{%s}_", key)
					resContent = strings.ReplaceAll(resContent, evPlaceholder, fmt.Sprintf("%v", resValue))

				} else {
					log.Printf("WARNING: empty value of {%s}", key)
				}
			}

		}

		err = ioutil.WriteFile(_filePath, []byte(resContent), _fileMode)
		if err != nil {
			log.Fatalf("ERROR: Error writing resContent to file %s:\n%v", _filePath, err)
		}

	}

	log.Printf("INFO: sev succeeded: %s", _filePath)
}

func (evv envVarValuesSets) processFile(_filePath string, _fileMode os.FileMode){

	contentBytes, err := ioutil.ReadFile(_filePath)
	if err != nil {
		log.Fatalf("ERROR: Error reading file %s:\n%v", _filePath, err)
	} else {

		var content = string(contentBytes)
		var resContent = content


		for key, value := range evv[VAR_VALUES_SETS_CHOSEN_ID].(map[string]interface{}) {

			var resValue string

			switch reflect.TypeOf(value).String(){
			case "float64":
				var f = value.(float64)
				ipart := int64(f)
				decpart := fmt.Sprintf("%.6g", f-float64(ipart))
				if len(decpart) == 1{
					resValue = fmt.Sprintf(`%.0f`, value)
				} else {
					resValue = fmt.Sprintf(`%f`, value)
				}

			case "bool":
				resValue = fmt.Sprintf(`%t`, value)
			default:
				resValue = value.(string)
			}

			if resValue != "" {

				evPlaceholder := fmt.Sprintf("_{%s}_", key)
				resContent = strings.ReplaceAll(resContent, evPlaceholder, fmt.Sprintf("%v", resValue))

				err = ioutil.WriteFile(_filePath, []byte(resContent), _fileMode)
				if err != nil {
					log.Fatalf("ERROR: Error writing resContent to file %s:\n%v", _filePath, err)
				}

			} else {
				log.Printf("WARNING: empty value of {%s}", key)
			}
		}

	}

	log.Printf("INFO: sev succeeded: %s", _filePath)
}

func pathExists(path string) (os.FileInfo, bool, error) {
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, err
			// file does not exist
		} else {
			return stat, true, err
		}
	} else {
		return stat, true, nil
	}

}