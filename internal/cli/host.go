package cli

import (
	"github.com/GMfatcat/goslide/internal/server"
	"github.com/spf13/cobra"
)

var (
	hostPort   int
	hostNoOpen bool
)

var hostCmd = &cobra.Command{
	Use:   "host <directory>",
	Short: "Serve a directory of presentations",
	Args:  cobra.ExactArgs(1),
	RunE:  runHost,
}

func init() {
	hostCmd.Flags().IntVarP(&hostPort, "port", "p", 8080, "port number")
	hostCmd.Flags().BoolVar(&hostNoOpen, "no-open", false, "don't auto-open browser")
	rootCmd.AddCommand(hostCmd)
}

func runHost(cmd *cobra.Command, args []string) error {
	return server.HostRun(server.HostOptions{
		Dir:    args[0],
		Port:   hostPort,
		NoOpen: hostNoOpen,
	})
}
