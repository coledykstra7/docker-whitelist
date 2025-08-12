package main

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

// squidStatus checks TCP connectivity to squid service (container hostname) port 3128
func squidStatus() string {
	address := fmt.Sprintf("%s:%s", SquidHost, SquidPort)
	conn, err := net.DialTimeout("tcp", address, time.Duration(ConnectionTimeout)*time.Millisecond)
	if err != nil {
		return "DOWN"
	}
	if closeErr := conn.Close(); closeErr != nil {
		// Log close error but don't affect status
		fmt.Printf("Warning: failed to close connection: %v\n", closeErr)
	}
	return "UP"
}

// reloadSquid executes squid -k reconfigure inside the squid container via docker CLI
func reloadSquid() error {
	// Try graceful HUP first
	cmd := exec.Command("docker", "kill", "-s", "HUP", SquidHost)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	
	// Fallback to reconfigure
	cmd2 := exec.Command("docker", "exec", SquidHost, "squid", "-k", "reconfigure")
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return fmt.Errorf("reload failed: %v output1: %s output2: %s", err2, string(out), string(out2))
	}
	return nil
}
