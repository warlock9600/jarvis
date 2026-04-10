package main

import (
	"fmt"
	"os"
	"sort"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/k8sutil"

	"github.com/spf13/cobra"
)

func newK8sCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kube",
		Aliases: []string{"k8s"},
		Short:   "Kubernetes inspection commands",
		Long:    "Kubernetes inspection commands using kubectl and kubeconfig from KUBECONFIG or --kubeconfig.",
		Example: "jarvis kube pods --namespace prod\njarvis kube ctx use prod-eu\njarvis kube ns list\njarvis kube ns use backend",
	}
	cmd.AddCommand(
		newK8sPodsCmd(state),
		newK8sImagesCmd(state),
		newK8sContextCmd(state),
		newK8sNamespaceCmd(state),
	)
	return cmd
}

func newK8sPodsCmd(state *app.State) *cobra.Command {
	var ns string
	var sortRestarts bool
	cmd := &cobra.Command{
		Use:   "pods [--namespace <ns>]",
		Short: "List pods in a namespace",
		Long: `List pods in a namespace with status and restart counters.

Return codes:
- 0: pods listed
- 1: kubectl failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			resolvedNS, err := k8sutil.ResolveNamespace(ns, state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot determine namespace (use --namespace to override)", err)
			}
			pods, err := k8sutil.Pods(resolvedNS, state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot get pods", err)
			}
			if sortRestarts {
				sort.Slice(pods, func(i, j int) bool { return pods[i].Restarts > pods[j].Restarts })
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"namespace": resolvedNS, "pods": pods})
			}
			rows := make([][]string, 0, len(pods))
			for _, p := range pods {
				rows = append(rows, []string{p.Name, fmt.Sprintf("%d", p.Restarts), p.Status})
			}
			state.Printer.PrintTable([]string{"Pod", "Restarts", "Status"}, rows)
			return nil
		},
		Example: "jarvis k8s pods --namespace kube-system\njarvis k8s pods --namespace prod --restarts",
	}
	cmd.Flags().StringVarP(&ns, "namespace", "n", "", "Kubernetes namespace (default: current context namespace)")
	cmd.Flags().BoolVarP(&sortRestarts, "restarts", "R", false, "Sort by restart count descending")
	return cmd
}

func newK8sImagesCmd(state *app.State) *cobra.Command {
	var ns string
	cmd := &cobra.Command{
		Use:   "images [--namespace <ns>]",
		Short: "List unique container images in a namespace",
		Long: `List unique container images from all pods in the target namespace.

Return codes:
- 0: images listed
- 1: kubectl failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			resolvedNS, err := k8sutil.ResolveNamespace(ns, state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot determine namespace (use --namespace to override)", err)
			}
			images, err := k8sutil.Images(resolvedNS, state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot get images", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"namespace": resolvedNS, "images": images})
			}
			for _, img := range images {
				fmt.Fprintln(os.Stdout, img)
			}
			return nil
		},
		Example: "jarvis k8s images --namespace prod\njarvis k8s images --namespace prod --json",
	}
	cmd.Flags().StringVarP(&ns, "namespace", "n", "", "Kubernetes namespace (default: current context namespace)")
	return cmd
}

func newK8sClustersCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "List clusters from kubeconfig",
		Long: `List Kubernetes clusters from kubeconfig.

Return codes:
- 0: clusters listed
- 1: kubectl failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			clusters, err := k8sutil.Clusters(state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot list clusters", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"clusters": clusters})
			}
			for _, c := range clusters {
				fmt.Fprintln(os.Stdout, c)
			}
			return nil
		},
		Example: "jarvis k8s clusters\njarvis k8s clusters --json",
	}
	return cmd
}

func newK8sNamespacesCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespaces",
		Short: "List namespaces in current cluster",
		Long: `List namespaces available in the currently selected cluster context.

Return codes:
- 0: namespaces listed
- 1: kubectl failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			nss, err := k8sutil.Namespaces(state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot list namespaces", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"namespaces": nss})
			}
			for _, ns := range nss {
				fmt.Fprintln(os.Stdout, ns)
			}
			return nil
		},
		Example: "jarvis k8s namespaces\njarvis k8s namespaces --json",
	}
	return cmd
}

func newK8sContextCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"ctx"},
		Short:   "Manage Kubernetes context",
		Long:    "Show or switch kubeconfig context.",
		Example: "jarvis k8s ctx current\njarvis k8s ctx list\njarvis k8s ctx use prod-eu",
	}
	cmd.AddCommand(newK8sContextCurrentCmd(state), newK8sContextListCmd(state), newK8sContextUseCmd(state))
	return cmd
}

func newK8sContextListCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available contexts",
		Long:  "List kubeconfig contexts.",
		RunE: func(_ *cobra.Command, _ []string) error {
			contexts, err := k8sutil.Contexts(state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot list contexts", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"contexts": contexts})
			}
			for _, c := range contexts {
				fmt.Fprintln(os.Stdout, c)
			}
			return nil
		},
		Example: "jarvis k8s ctx list\njarvis k8s ctx list --json",
	}
	return cmd
}

func newK8sContextCurrentCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current context",
		Long:  "Show the current kubeconfig context.",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, err := k8sutil.CurrentContext(state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot get current context", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"current_context": ctx})
			}
			fmt.Fprintln(os.Stdout, ctx)
			return nil
		},
		Example: "jarvis k8s context current\njarvis k8s context current --json",
	}
	return cmd
}

func newK8sContextUseCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <context>",
		Short: "Switch current context",
		Long: `Switch kubeconfig context to target value.

Return codes:
- 0: context switched
- 1: context switch failed`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := k8sutil.UseContext(state.Config.K8s.Kubeconfig, args[0]); err != nil {
				return common.NewExitError(common.ExitError, "cannot switch context", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"context": args[0], "switched": true})
			}
			fmt.Fprintf(os.Stdout, "Switched context to %s\n", args[0])
			return nil
		},
		Example: "jarvis k8s context use prod\njarvis k8s context use staging",
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items, err := k8sutil.Contexts(state.Config.K8s.Kubeconfig)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			res := make([]string, 0, len(items))
			for _, x := range items {
				if len(toComplete) == 0 || (len(toComplete) > 0 && len(x) >= len(toComplete) && x[:len(toComplete)] == toComplete) {
					res = append(res, x)
				}
			}
			return res, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}

func newK8sNamespaceCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns"},
		Short:   "Manage active namespace for current context",
		Long:    "Show or switch namespace bound to current kubeconfig context.",
		Example: "jarvis k8s ns current\njarvis k8s ns list\njarvis k8s ns use backend",
	}
	cmd.AddCommand(newK8sNamespaceCurrentCmd(state), newK8sNamespaceListCmd(state), newK8sNamespaceUseCmd(state))
	return cmd
}

func newK8sNamespaceListCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List namespaces in current cluster",
		Long:  "List namespaces available in the currently selected cluster context.",
		RunE: func(_ *cobra.Command, _ []string) error {
			nss, err := k8sutil.Namespaces(state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot list namespaces", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"namespaces": nss})
			}
			for _, ns := range nss {
				fmt.Fprintln(os.Stdout, ns)
			}
			return nil
		},
		Example: "jarvis k8s ns list\njarvis k8s ns list --json",
	}
	return cmd
}

func newK8sNamespaceCurrentCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current namespace",
		Long:  "Show namespace configured for current context (default if empty).",
		RunE: func(_ *cobra.Command, _ []string) error {
			ns, err := k8sutil.ResolveNamespace("", state.Config.K8s.Kubeconfig)
			if err != nil {
				return common.NewExitError(common.ExitError, "cannot get current namespace", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"current_namespace": ns})
			}
			fmt.Fprintln(os.Stdout, ns)
			return nil
		},
		Example: "jarvis k8s namespace current\njarvis k8s namespace current --json",
	}
	return cmd
}

func newK8sNamespaceUseCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <namespace>",
		Short: "Switch namespace for current context",
		Long: `Set namespace for current kubeconfig context using kubectl config set-context --current.

Return codes:
- 0: namespace switched
- 1: switch failed`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := k8sutil.SetCurrentNamespace(state.Config.K8s.Kubeconfig, args[0]); err != nil {
				return common.NewExitError(common.ExitError, "cannot switch namespace", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"namespace": args[0], "switched": true})
			}
			fmt.Fprintf(os.Stdout, "Switched namespace to %s\n", args[0])
			return nil
		},
		Example: "jarvis k8s namespace use default\njarvis k8s namespace use backend",
		ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items, err := k8sutil.Namespaces(state.Config.K8s.Kubeconfig)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			res := make([]string, 0, len(items))
			for _, x := range items {
				if len(toComplete) == 0 || (len(toComplete) > 0 && len(x) >= len(toComplete) && x[:len(toComplete)] == toComplete) {
					res = append(res, x)
				}
			}
			return res, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}
