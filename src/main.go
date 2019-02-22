package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"../src/json"
)

var version string
var jsonFilePath *string
var outputFilePath *string
var packageName *string
var structName *string

func parseAppParams() {
	// parse parameters
	jsonFilePath = flag.String("i", "", "path of json file")
	outputFilePath = flag.String("o", "", "path of output go struct file")
	packageName = flag.String("p", "", "name of package")
	structName = flag.String("s", "", "name of struct")
	printVersion := flag.Bool("v", false, "Print version and build date")
	printUsage := flag.Bool("h", false, "Print usage")
	flag.Parse()

	// dump app usage
	if printUsage != nil && *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	// dupm app version
	if printVersion != nil && *printVersion {
		fmt.Printf("App Version : %s\n", version)
		os.Exit(0)
	}

	// check parameters: jsonFilePath
	if jsonFilePath == nil || len(*jsonFilePath) == 0 {
		fmt.Print("[Error] Missing parameter : path of json file\n")
		flag.Usage()
		os.Exit(1)
	}

	// check parameters: outputFilePath
	if outputFilePath == nil || len(*outputFilePath) == 0 {
		fmt.Print("[Error] Missing parameter : path of output go struct file\n")
		flag.Usage()
		os.Exit(1)
	}
	if !strings.HasSuffix(*outputFilePath, ".go") {
		fmt.Print("[Error] invalid parameter : path of output go struct file should ended with '.go', for example: -o output/somestruct.go\n")
		os.Exit(1)
	}

	// check parameters: packageName
	if packageName == nil || len(*packageName) == 0 {
		fmt.Print("[Error] Missing parameter : name of package\n")
		flag.Usage()
		os.Exit(1)
	}

	// check parameters: structName
	if structName == nil || len(*structName) == 0 {
		fmt.Print("[Error] Missing parameter : name of struct\n")
		flag.Usage()
		os.Exit(1)
	}
}

func formatGoFile(filepath string) error {
	cmd := exec.Command("gofmt", "-w", filepath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error occured when format go file '%s', details: \n%s", filepath, out.String())
	}

	return nil
}

func main() {
	parseAppParams()

	s, err := json.Generate(*packageName, *structName, *jsonFilePath)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(*outputFilePath, []byte(s), 0644)
	if err != nil {
		panic(err)
	}

	err = formatGoFile(*outputFilePath)
	if err != nil {
		panic(err)
	}
}
