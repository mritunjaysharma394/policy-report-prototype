// package
package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {

	//calls function that runs KubeBench
	out := runKubeBench()
	fmt.Println(out)
}

func runKubeBench() string {

	//executes kube-bench
	cmd := exec.Command("./kube-bench", "--json")
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal(err)
	}

	return string(out)
}
