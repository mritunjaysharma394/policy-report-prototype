package kubebench

import (
	"embed"
)

//go:embed jobs/*.yaml
var yamlDir embed.FS

func embedYAMLs(kubebenchYAML string) ([]byte, error) {

	var (
		data []byte
		err  error
	)

	// jobs, _ := fs.ReadDir(yamlDir, "jobs")

	// for _, job := range jobs {
	// 	if kubebenchYAML == job.Name() {
	// 		data, err = yamlDir.ReadFile("jobs/" + job.Name())
	// 	} else {
	// 		continue
	// 	}
	// }

	data, err = yamlDir.ReadFile("jobs/" + kubebenchYAML)
	if err != nil {
		return nil, err
	}
	return data, nil
}
