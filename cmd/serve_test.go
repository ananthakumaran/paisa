package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeFlags(t *testing.T) {
	flags := serveCmd.Flags()
	hostFlag := flags.Lookup("host")
	portFlag := flags.Lookup("port")
	require.NotNil(t, hostFlag)
	require.NotNil(t, portFlag)

	originalHost, originalPort := host, port
	hostChanged, portChanged := hostFlag.Changed, portFlag.Changed
	t.Cleanup(func() {
		host, port = originalHost, originalPort
		hostFlag.Changed, portFlag.Changed = hostChanged, portChanged
	})

	for _, tc := range []struct {
		name string
		args []string
		host string
		port int
	}{
		{name: "defaults", host: "", port: 7500},
		{name: "IPv4 loopback", args: []string{"--host", "127.0.0.1"}, host: "127.0.0.1", port: 7500},
		{name: "IPv6 loopback", args: []string{"--host", "::1"}, host: "::1", port: 7500},
		{name: "hostname", args: []string{"--host", "localhost"}, host: "localhost", port: 7500},
		{name: "custom port", args: []string{"--host", "192.168.1.10", "--port", "8080"}, host: "192.168.1.10", port: 8080},
		{name: "short port flag", args: []string{"-p", "8080"}, host: "", port: 8080},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, hostFlag.Value.Set(hostFlag.DefValue))
			require.NoError(t, portFlag.Value.Set(portFlag.DefValue))
			require.NoError(t, serveCmd.ParseFlags(tc.args))
			assert.Equal(t, tc.host, host)
			assert.Equal(t, tc.port, port)
		})
	}
}
