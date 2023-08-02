package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// add two flag
// one is a file path, which is the path of values.yaml, and
// another is a chart version, which this tool will change the values.yaml to.
var (
	inputFilePath     string
	chartVersion      string
	operatorFilePath  string
	starrocksFilePath string
)

const (
	NEW_VERSION = "v1.8.0"
	_operator   = "operator"
	_starrocks  = "starrocks"
)

var _starrocksKeys = []string{"nameOverride", "initPassword", "timeZone", "datadog", "starrocksCluster",
	"starrocksFESpec", "starrocksCnSpec", "starrocksBeSpec", "secrets", "configMaps", "feProxy"}

var _operatorKeys = []string{"global", "timeZone", "nameOverride", "starrocksOperator"}

// this program requires two parameters:
// one is a file path, which is the path of values.yaml, and
// another is a chart version, which this tool will change the values.yaml to.
func main() {
	flag.StringVar(&inputFilePath, "input", "", "the path of values.yaml")
	flag.StringVar(&chartVersion, "version", "", "the chart version, which this tool will change the values.yaml to")
	flag.StringVar(&operatorFilePath, "operatorFilePath", "", "the path of values.yaml of operator or kube-starrocks")
	flag.StringVar(&starrocksFilePath, "starrocksFilePath", "", "the path of values.yaml of starrocks")
	flag.Parse()

	if inputFilePath == "" || chartVersion == "" || operatorFilePath == "" {
		log.Println("input, version and operatorFilePath are required")
		flag.Usage()
		return
	}
	if chartVersion >= NEW_VERSION {
		if starrocksFilePath == "" {
			log.Println("starrocksFilePath is required")
			flag.Usage()
			return
		}
	}

	input, err := os.Open(inputFilePath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = input.Close() }()

	operator, err := os.Open(operatorFilePath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = operator.Close() }()

	starrocks, err := os.Open(starrocksFilePath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = starrocks.Close() }()

	err = Do(input, chartVersion, operator, starrocks)
	if err != nil {
		panic(err)
	}
}

// Do is the main function of this tool. It will read the values.yaml from the reader, and write the new values.yaml to the writer.
// If write to new version, w1 is the writer to the values.yaml of operator, w2 is the writer to the values.yaml of starrocks.
// If write to old version, w1 is the writer to the values.yaml of starrocks and operator.
func Do(reader io.Reader, chartVersion string, w1 io.Writer, w2 io.Writer) error {
	// read the content from the file by viper
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(reader)
	if err != nil {
		return err
	}
	input, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	s := map[string]interface{}{}
	if err := yaml.Unmarshal(input, &s); err != nil {
		return err
	}

	// find the value of the field "operator" or "starrocks"
	operator := s[_operator]
	starrocks := s[_starrocks]
	if operator != nil || starrocks != nil {
		log.Printf("this values.yaml is from new chart version >= %v,\n", NEW_VERSION)
		// values.yaml is from new chart version
		if chartVersion >= NEW_VERSION {
			fmt.Printf("no need to change to upgrade to %v\n", chartVersion)
			return nil
		}
		// change the new version to old version
		mapper := operator.(map[string]interface{})
		// remove duplicate fields from operator
		delete(mapper, "timeZone")
		delete(mapper, "nameOverride")
		data1, err := yaml.Marshal(operator)
		if err != nil {
			return err
		}
		data2, err := yaml.Marshal(starrocks)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w1, "%v\n%v", string(data1), string(data2)); err != nil {
			return err
		}
		return nil
	} else {
		log.Printf("this values.yaml is from old chart version < %v,\n", NEW_VERSION)
		// values.yaml is from old chart version
		if chartVersion < NEW_VERSION {
			log.Printf("no need to change to down grade to %v\n", chartVersion)
			return nil
		}
		// change the old version to new version
		if err := Write(w1, s, _operatorKeys, _operator); err != nil {
			return err
		}
		if err := Write(w2, s, _starrocksKeys, _starrocks); err != nil {
			return err
		}
	}
	return nil
}

func Write(w io.Writer, s map[string]interface{}, keys []string, header string) error {
	newS := map[string]interface{}{}
	for _, key := range keys {
		value := s[key]
		if value == nil {
			continue
		}
		newS[key] = value
	}
	output := map[string]interface{}{header: newS}

	data, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
