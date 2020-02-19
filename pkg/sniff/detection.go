// Package sniff provides functionality to interact with networking interfaces.
//
// Functionality which ties directly to the 
package sniff

import "net"

// DetectInterfaces provides list of network interface handles.
// Returned names can be bound to for intercepting traffic.
func DetectInterfaces() ([]string, error) {
	output := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		output = append(output, i.Name)
	}
	return output, nil
}
