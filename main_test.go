// package
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	data, err := ioutil.ReadFile("./check.json")
	if err != nil {
		fmt.Print(err)
	}

	var got, want interface{}
	// json from stream
	jsonDataReader1 := strings.NewReader(runKubeBench())
	//json from file
	jsonDataReader2 := strings.NewReader(string(data))

	d := json.NewDecoder(jsonDataReader1)
	if err := d.Decode(&got); err != nil {
		fmt.Print(err)
	}
	d = json.NewDecoder(jsonDataReader2)
	if err := d.Decode(&want); err != nil {
		fmt.Print(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
