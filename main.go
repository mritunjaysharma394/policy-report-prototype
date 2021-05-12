// package
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mritunjaysharma394/policy-report-prototype/pkg/report"

	"github.com/mritunjaysharma394/policy-report-prototype/pkg/kubebench"

	"k8s.io/client-go/util/homedir"
)

var (
	name       string
	namespace  string
	category   string
	kubeconfig string
)

func parseArguments() {
	flag.StringVar(&name, "name", "kube-bench", "name of policy report")
	flag.StringVar(&namespace, "namespace", "default", "namespace of the cluster")
	flag.StringVar(&category, "category", "CIS Benchmarks", "category of the policy report")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()
}

func main() {
	parseArguments()

	var kubebenchImg = flag.String("kubebenchImg", "aquasec/kube-bench:latest", "kube-bench image used as part of this test")
	var timeout = flag.Duration("timeout", 10*time.Minute, "Test Timeout")

	var testdataDir string
	ctx, err := kubebench.SetupCluster("kube-bench", fmt.Sprintf("./testdata/%s/add-tls-kind.yaml", testdataDir), *timeout)
	if err != nil {
		fmt.Errorf("failed to setup KIND cluster error: %v", err)
	}
	defer func() {
		*ctx.Delete()
	}()

	if err := kubebench.LoadImageFromDocker(*kubebenchImg, *ctx); err != nil {
		fmt.Errorf("failed to load kube-bench image from Docker to KIND error: %v", err)
	}

	clientset, err := kubebench.GetClientSet(ctx.KubeConfigPath())
	if err != nil {
		fmt.Errorf("failed to connect to Kubernetes cluster error: %v", err)
	}

	resultData, err := kubebench.RunWithKind(ctx, clientset, c.TestName, c.KubebenchYAML, *kubebenchImg, *timeout)
	if err != nil {
		fmt.Errorf("unexpected error: %v", err)
	}

	// run kubebench
	cis, err := kubebench.Run([]string{"--json"})
	if err != nil {
		fmt.Printf("failed to run kube-bench: %v \n", err)
		os.Exit(-1)
	}

	// create policy report
	r, err := report.New(cis, name, category)
	if err != nil {
		fmt.Printf("failed to create policy reports: %v \n", err)
		os.Exit(-1)
	}

	// write policy report
	r, err = report.Write(r, namespace, kubeconfig)
	if err != nil {
		fmt.Printf("failed to create policy reports: %v \n", err)
		os.Exit(-1)
	}

	fmt.Printf("wrote policy report %s/%s \n", r.Namespace, r.Name)
}
