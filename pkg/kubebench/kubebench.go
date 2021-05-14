package kubebench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	kubebench "github.com/aquasecurity/kube-bench/check"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientSet(kubeconfigPath string) (*kubernetes.Clientset, error) {
	var kubeconfig *rest.Config

	if kubeconfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load kubeconfig from %s: %v", kubeconfigPath, err)
		}
		kubeconfig = config
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to load in-cluster config: %v", err)
		}
		kubeconfig = config
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		panic(err)
	}

	return clientset, nil

}
func RunJob(kubeconfig string, kubebenchYAML, kubebenchImg string, timeout time.Duration) (*kubebench.OverallControls, error) {

	clientset, err := getClientSet(kubeconfig)
	var jobName string
	jobName, err = deployJob(context.TODO(), clientset, kubebenchYAML, kubebenchImg)
	if err != nil {
		return nil, err
	}

	p, err := findPodForJob(context.TODO(), clientset, jobName, timeout)
	if err != nil {
		return nil, err
	}

	output := getPodLogs(context.TODO(), clientset, p)

	err = clientset.BatchV1().Jobs(apiv1.NamespaceDefault).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	controls, err := convert(output)
	if err != nil {
		return nil, err
	}

	return controls, nil

}

func deployJob(ctx context.Context, clientset *kubernetes.Clientset, kubebenchYAML, kubebenchImg string) (string, error) {
	//jobYAML, err := ioutil.ReadFile(kubebenchYAML)
	jobYAML, err := embedYAMLs(kubebenchYAML)
	if err != nil {
		return "", err
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(jobYAML), len(jobYAML))
	job := &batchv1.Job{}
	if err := decoder.Decode(job); err != nil {
		return "", err
	}
	jobName := job.GetName()
	job.Spec.Template.Spec.Containers[0].Image = kubebenchImg
	job.Spec.Template.Spec.Containers[0].Args = []string{"--json"}

	_, err = clientset.BatchV1().Jobs(apiv1.NamespaceDefault).Create(ctx, job, metav1.CreateOptions{})

	return jobName, err
}

func findPodForJob(ctx context.Context, clientset *kubernetes.Clientset, jobName string, duration time.Duration) (*apiv1.Pod, error) {
	failedPods := make(map[string]struct{})
	selector := fmt.Sprintf("job-name=%s", jobName)
	timeout := time.After(duration)
	for {
		time.Sleep(3 * time.Second)
	podfailed:
		select {
		case <-timeout:
			return nil, fmt.Errorf("podList - timed out: no Pod found for Job %s", jobName)
		default:
			pods, err := clientset.CoreV1().Pods(apiv1.NamespaceDefault).List(ctx, metav1.ListOptions{
				LabelSelector: selector,
			})
			if err != nil {
				return nil, err
			}
			fmt.Printf("Found (%d) pods\n", len(pods.Items))
			for _, cp := range pods.Items {
				if _, found := failedPods[cp.Name]; found {
					continue
				}

				if strings.HasPrefix(cp.Name, jobName) {
					fmt.Printf("pod (%s) - %#v\n", cp.Name, cp.Status.Phase)
					if cp.Status.Phase == apiv1.PodSucceeded {
						return &cp, nil
					}

					if cp.Status.Phase == apiv1.PodFailed {
						fmt.Printf("pod (%s) - %s - retrying...\n", cp.Name, cp.Status.Phase)
						fmt.Print(getPodLogs(ctx, clientset, &cp))
						failedPods[cp.Name] = struct{}{}
						break podfailed
					}
				}
			}
		}
	}
}

func getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, pod *apiv1.Pod) string {
	podLogOpts := corev1.PodLogOptions{}
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "getPodLogs - error in opening stream"
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "getPodLogs - error in copy information from podLogs to buf"
	}

	return buf.String()
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
