package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"jarvis/internal/app"
	"jarvis/internal/common"
	"jarvis/internal/netutil"

	"github.com/spf13/cobra"
)

func newNetCmd(state *app.State) *cobra.Command {
	netCmd := &cobra.Command{
		Use:     "net",
		Short:   "Network diagnostics commands",
		Long:    "Network diagnostics commands for IP info, speed tests, DNS operations, connectivity checks and TLS certificate health.",
		Example: "jarvis net ip --public\njarvis net check --host example.com --port 443 --http https://example.com",
	}
	netCmd.AddCommand(newNetIPCmd(state), newNetSpeedtestCmd(state), newNetCheckCmd(state), newNetDNSCmd(state), newNetTLSCmd(state))
	return netCmd
}

func newNetIPCmd(state *app.State) *cobra.Command {
	var onlyPublic, onlyLocal, v4, v6 bool
	cmd := &cobra.Command{
		Use:   "ip",
		Short: "Show local and public IP addresses",
		Long: `Show local interface IPs and public IP using fallback providers.

Return codes:
- 0: all requested data fetched
- 2: partial data fetched (for example local ok, public failed)
- 1: command failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			timeout := time.Duration(state.Config.Net.TimeoutSeconds) * time.Second
			res := netutil.IPResult{}
			var localErr, publicErr error

			if !onlyPublic {
				res.Local, localErr = netutil.LocalIPs(v4, v6)
			}
			if !onlyLocal {
				res.Public, _ = netutil.PublicIP(state.Config.Net.PublicIPProviders, timeout, state.Config.Net.Retries, v4, v6)
				if len(res.Public) == 0 {
					publicErr = fmt.Errorf("public IP providers failed")
				}
			}

			if state.JSON {
				_ = state.Printer.PrintJSON(res)
			} else {
				if len(res.Local) > 0 {
					rows := make([][]string, 0, len(res.Local))
					for _, i := range res.Local {
						rows = append(rows, []string{i.Interface, i.Version, i.Address})
					}
					fmt.Fprintln(os.Stdout, "Local IPs:")
					state.Printer.PrintTable([]string{"Interface", "Version", "Address"}, rows)
				}
				if len(res.Public) > 0 {
					rows := make([][]string, 0, len(res.Public))
					for _, i := range res.Public {
						rows = append(rows, []string{i.Version, i.Address, i.Source})
					}
					fmt.Fprintln(os.Stdout, "Public IP:")
					state.Printer.PrintTable([]string{"Version", "Address", "Provider"}, rows)
				}
			}

			if (localErr == nil && publicErr != nil && !onlyLocal) || (localErr != nil && publicErr == nil && !onlyPublic) {
				return common.NewExitError(common.ExitPartial, "partial result: one source failed", publicErr)
			}
			if localErr != nil && publicErr != nil {
				return common.NewExitError(common.ExitError, "failed to fetch any IP data", fmt.Errorf("local: %v; public: %v", localErr, publicErr))
			}
			return nil
		},
		Example: "jarvis net ip\njarvis net ip --public --v4",
	}
	cmd.Flags().BoolVarP(&onlyPublic, "public", "p", false, "Show only public IP")
	cmd.Flags().BoolVarP(&onlyLocal, "local", "l", false, "Show only local interface IPs")
	cmd.Flags().BoolVarP(&v4, "v4", "4", false, "IPv4 only")
	cmd.Flags().BoolVarP(&v6, "v6", "6", false, "IPv6 only")
	return cmd
}

func newNetSpeedtestCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speedtest",
		Short: "Run internet speed test",
		Long: `Run speed test through external Speedtest CLI. This command depends on a compatible speedtest binary.

Return codes:
- 0: test completed
- 1: speedtest binary unavailable or execution failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := runSpeedtestWithSpinner(state)
			if err != nil {
				return common.NewExitError(common.ExitError, "speedtest failed", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(res)
			}
			rows := [][]string{{fmt.Sprintf("%.2f", res.PingMS), fmt.Sprintf("%.2f", res.JitterMS), fmt.Sprintf("%.2f", res.DownloadMB), fmt.Sprintf("%.2f", res.UploadMB)}}
			state.Printer.PrintTable([]string{"Ping ms", "Jitter ms", "Download Mbps", "Upload Mbps"}, rows)
			return nil
		},
		Example: "jarvis net speedtest\njarvis net speedtest --json",
	}
	return cmd
}

func runSpeedtestWithSpinner(state *app.State) (netutil.SpeedtestResult, error) {
	if state.JSON || !state.Printer.IsTTY {
		return netutil.RunSpeedtest(state.Config.Speedtest.Bin)
	}

	type result struct {
		res netutil.SpeedtestResult
		err error
	}

	done := make(chan result, 1)
	go func() {
		r, err := netutil.RunSpeedtest(state.Config.Speedtest.Bin)
		done <- result{res: r, err: err}
	}()

	frames := []string{"|", "/", "-", "\\"}
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case out := <-done:
			fmt.Fprint(os.Stderr, "\r\033[K")
			return out.res, out.err
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\rRunning speedtest... %s", frames[i%len(frames)])
			i++
		}
	}
}

func newNetDNSCmd(state *app.State) *cobra.Command {
	dnsCmd := &cobra.Command{
		Use:     "dns",
		Short:   "DNS lookup and cache operations",
		Long:    "Run DNS queries and flush local DNS cache with OS-aware behavior.",
		Example: "jarvis net dns lookup example.com --type A\njarvis net dns flush --dry-run",
	}
	dnsCmd.AddCommand(newNetDNSLookupCmd(state), newNetDNSFlushCmd(state))
	return dnsCmd
}

func newNetDNSLookupCmd(state *app.State) *cobra.Command {
	var server, recordType string
	cmd := &cobra.Command{
		Use:   "lookup <name>",
		Short: "Resolve DNS records",
		Long: `Resolve DNS records (A, AAAA, MX, TXT) optionally through a specific DNS server.

Return codes:
- 0: records returned
- 1: lookup failed`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			timeout := time.Duration(state.Config.Net.TimeoutSeconds) * time.Second
			records, err := netutil.Lookup(args[0], server, recordType, timeout)
			if err != nil {
				return common.NewExitError(common.ExitError, "DNS lookup failed", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(map[string]any{"name": args[0], "records": records})
			}
			rows := make([][]string, 0, len(records))
			for _, r := range records {
				rows = append(rows, []string{r.Type, r.Value})
			}
			state.Printer.PrintTable([]string{"Type", "Value"}, rows)
			return nil
		},
		Example: "jarvis net dns lookup raketa.online\njarvis net dns lookup google.com --server 1.1.1.1 --type AAAA",
	}
	cmd.Flags().StringVarP(&server, "server", "s", "", "DNS server IP (for example 1.1.1.1)")
	cmd.Flags().StringVarP(&recordType, "type", "y", "A", "Record type: A|AAAA|MX|TXT")
	return cmd
}

func newNetDNSFlushCmd(state *app.State) *cobra.Command {
	var dryRun, force bool
	cmd := &cobra.Command{
		Use:   "flush",
		Short: "Flush local DNS cache",
		Long: `Flush DNS cache with OS-specific commands.

Safety:
- Requires interactive confirmation unless --force is used.
- Use --dry-run to inspect commands first.

Return codes:
- 0: flush done
- 1: unsupported OS or execution failure`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dryRun && !force {
				if term := cmd.InOrStdin(); term != nil {
					fmt.Fprint(os.Stderr, "Proceed with DNS cache flush? [y/N]: ")
					var answer string
					_, _ = fmt.Fscanln(term, &answer)
					if strings.ToLower(strings.TrimSpace(answer)) != "y" {
						return common.NewExitError(common.ExitError, "aborted by user", nil)
					}
				}
			}
			res, err := netutil.FlushDNS(dryRun)
			if state.JSON {
				_ = state.Printer.PrintJSON(res)
			}
			if !state.JSON {
				rows := make([][]string, 0, len(res.Steps))
				for _, s := range res.Steps {
					rows = append(rows, []string{s.Command, s.Status, s.Output})
				}
				if len(rows) > 0 {
					state.Printer.PrintTable([]string{"Command", "Status", "Output"}, rows)
				}
				fmt.Fprintln(os.Stdout, res.Message)
			}
			if err != nil {
				return common.NewExitError(common.ExitError, "dns flush failed", err)
			}
			return nil
		},
		Example: "jarvis net dns flush --dry-run\njarvis net dns flush --force",
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show commands without executing")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Run without interactive confirmation")
	return cmd
}

func newNetCheckCmd(state *app.State) *cobra.Command {
	var host string
	var port int
	var httpURL string
	cmd := &cobra.Command{
		Use:   "check --host <host>",
		Short: "Check host reachability (ping/tcp/http)",
		Long: `Run ping, TCP, and optional HTTP checks for a target host.

Return codes:
- 0: all requested checks passed
- 2: partial checks passed
- 1: all checks failed`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if host == "" {
				return common.NewExitError(common.ExitError, "--host is required", nil)
			}
			timeout := time.Duration(state.Config.Net.TimeoutSeconds) * time.Second
			res := netutil.CheckResult{Host: host, TCPPort: port, HTTPURL: httpURL, Timestamp: time.Now().Format(time.RFC3339)}
			res.PingOK = netutil.Ping(host, timeout)
			if port > 0 {
				res.TCPOK = netutil.TCPCheck(host, port, timeout)
			}
			if httpURL != "" {
				res.HTTPCode, res.HTTPOK = netutil.HTTPCheck(httpURL, timeout)
			}
			if !res.PingOK || (port > 0 && !res.TCPOK) || (httpURL != "" && !res.HTTPOK) {
				res.Advice = "Check firewall/routing/DNS or target service health"
			}

			if state.JSON {
				_ = state.Printer.PrintJSON(res)
			} else {
				state.Printer.PrintTable([]string{"Check", "Result", "Details"}, [][]string{{"ping", boolWord(res.PingOK), host}, {"tcp", boolWord(res.TCPOK), fmt.Sprintf("%s:%d", host, port)}, {"http", boolWord(res.HTTPOK), fmt.Sprintf("%s (code %d)", httpURL, res.HTTPCode)}})
				if res.Advice != "" {
					fmt.Fprintln(os.Stdout, "Advice:", res.Advice)
				}
			}

			requested := 1
			passed := 0
			if res.PingOK {
				passed++
			}
			if port > 0 {
				requested++
				if res.TCPOK {
					passed++
				}
			}
			if httpURL != "" {
				requested++
				if res.HTTPOK {
					passed++
				}
			}
			if passed == requested {
				return nil
			}
			if passed > 0 {
				return common.NewExitError(common.ExitPartial, "partial connectivity success", nil)
			}
			return common.NewExitError(common.ExitError, "all checks failed", nil)
		},
		Example: "jarvis net check --host 1.1.1.1 --port 443\njarvis net check --host example.com --http https://example.com",
	}
	cmd.Flags().StringVarP(&host, "host", "H", "", "Target host or IP")
	cmd.Flags().IntVarP(&port, "port", "p", 443, "TCP port to check")
	cmd.Flags().StringVarP(&httpURL, "http", "u", "", "HTTP URL to check")
	return cmd
}

func newNetTLSCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tls",
		Short:   "TLS certificate checks",
		Long:    "TLS certificate checks for target endpoints.",
		Example: "jarvis net tls expiry example.com\njarvis net tls expiry api.example.com:8443 --json",
	}
	cmd.AddCommand(newNetTLSExpiryCmd(state))
	return cmd
}

func newNetTLSExpiryCmd(state *app.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expiry <host>[:port]",
		Short: "Show days until TLS certificate expiry",
		Long: `Connect to target host over TLS and show certificate expiry details.

Return codes:
- 0: certificate fetched
- 1: connect or certificate retrieval failed`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			timeout := time.Duration(state.Config.Net.TimeoutSeconds) * time.Second
			res, err := netutil.TLSExpiryCheck(args[0], timeout)
			if err != nil {
				return common.NewExitError(common.ExitError, "TLS check failed", err)
			}
			if state.JSON {
				return state.Printer.PrintJSON(res)
			}
			state.Printer.PrintTable([]string{"Host", "CN", "Not After", "Days Left"}, [][]string{{fmt.Sprintf("%s:%s", res.Host, res.Port), res.ServerName, res.NotAfter.Format(time.RFC3339), fmt.Sprintf("%d", res.DaysLeft)}})
			return nil
		},
		Example: "jarvis net tls expiry github.com\njarvis net tls expiry internal.service.local:9443",
	}
	return cmd
}

func boolWord(v bool) string {
	if v {
		return "ok"
	}
	return "fail"
}
