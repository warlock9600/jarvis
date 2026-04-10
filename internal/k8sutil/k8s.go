package k8sutil

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type PodInfo struct {
	Name     string `json:"name"`
	Restarts int    `json:"restarts"`
	Status   string `json:"status"`
}

type podList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			Phase             string `json:"phase"`
			ContainerStatuses []struct {
				RestartCount int `json:"restartCount"`
			} `json:"containerStatuses"`
		} `json:"status"`
	} `json:"items"`
}

func runKubectl(kubeconfig string, args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	if kubeconfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kubectl %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

func Pods(namespace, kubeconfig string) ([]PodInfo, error) {
	out, err := runKubectl(kubeconfig, "get", "pods", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}
	var pl podList
	if err := json.Unmarshal(out, &pl); err != nil {
		return nil, err
	}
	res := make([]PodInfo, 0, len(pl.Items))
	for _, item := range pl.Items {
		restarts := 0
		for _, c := range item.Status.ContainerStatuses {
			restarts += c.RestartCount
		}
		res = append(res, PodInfo{Name: item.Metadata.Name, Restarts: restarts, Status: item.Status.Phase})
	}
	return res, nil
}

func ResolveNamespace(explicitNamespace, kubeconfig string) (string, error) {
	if strings.TrimSpace(explicitNamespace) != "" {
		return explicitNamespace, nil
	}

	out, err := runKubectl(kubeconfig, "config", "view", "--minify", "-o", "jsonpath={..namespace}")
	if err != nil {
		return "", fmt.Errorf("cannot resolve namespace from current context: %w", err)
	}
	ns := strings.TrimSpace(string(out))
	if ns == "" {
		ns = "default"
	}
	return ns, nil
}

func Images(namespace, kubeconfig string) ([]string, error) {
	out, err := runKubectl(kubeconfig, "get", "pods", "-n", namespace, "-o", "json")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Items []struct {
			Spec struct {
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, item := range raw.Items {
		for _, c := range item.Spec.Containers {
			set[c.Image] = struct{}{}
		}
	}
	images := make([]string, 0, len(set))
	for img := range set {
		images = append(images, img)
	}
	sort.Strings(images)
	return images, nil
}

func Clusters(kubeconfig string) ([]string, error) {
	out, err := runKubectl(kubeconfig, "config", "get-clusters")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	res := make([]string, 0, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i == 0 && strings.EqualFold(line, "NAME") {
			continue
		}
		res = append(res, line)
	}
	sort.Strings(res)
	return res, nil
}

func Contexts(kubeconfig string) ([]string, error) {
	out, err := runKubectl(kubeconfig, "config", "get-contexts", "-o", "name")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	res := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			res = append(res, line)
		}
	}
	sort.Strings(res)
	return res, nil
}

func CurrentContext(kubeconfig string) (string, error) {
	out, err := runKubectl(kubeconfig, "config", "current-context")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func UseContext(kubeconfig string, contextName string) error {
	_, err := runKubectl(kubeconfig, "config", "use-context", contextName)
	return err
}

func Namespaces(kubeconfig string) ([]string, error) {
	out, err := runKubectl(kubeconfig, "get", "namespaces", "-o", "json")
	if err != nil {
		return nil, err
	}
	var raw struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	res := make([]string, 0, len(raw.Items))
	for _, item := range raw.Items {
		if strings.TrimSpace(item.Metadata.Name) != "" {
			res = append(res, item.Metadata.Name)
		}
	}
	sort.Strings(res)
	return res, nil
}

func SetCurrentNamespace(kubeconfig string, namespace string) error {
	_, err := runKubectl(kubeconfig, "config", "set-context", "--current", "--namespace="+namespace)
	return err
}
