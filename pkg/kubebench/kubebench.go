package kubebench

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	kubebench "github.com/aquasecurity/kube-bench/check"
)

func Run(args []string) (*kubebench.OverallControls, error) {
	out, err := execute(args)
	if err != nil {
		fmt.Print(out)
		return nil, err
	}

	controls, err := convert(out)
	if err != nil {
		return nil, err
	}

	return controls, nil
}

func execute(args []string) (string, error) {
	cmd := exec.Command("./kube-bench", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), err
}

func convert(jsonString string) (*kubebench.OverallControls, error) {
	jsonDataReader := strings.NewReader(jsonString)
	decoder := json.NewDecoder(jsonDataReader)

	var controls kubebench.OverallControls
	if err := decoder.Decode(&controls); err != nil {
		return nil, err
	}

	return &controls, nil
}
