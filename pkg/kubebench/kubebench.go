package kubebench

import (
	"encoding/json"
	"strings"

	kubebench "github.com/aquasecurity/kube-bench/check"
)

func convert(jsonString string) (*kubebench.OverallControls, error) {
	jsonDataReader := strings.NewReader(jsonString)
	decoder := json.NewDecoder(jsonDataReader)

	var controls kubebench.OverallControls
	if err := decoder.Decode(&controls); err != nil {
		return nil, err
	}

	return &controls, nil
}
