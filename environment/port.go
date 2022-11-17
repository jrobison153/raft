// Package environment contains functions that interact with the host operating environment
package environment

import (
	"fmt"
	"os"
	"strconv"
)

const (
	MIN_PORT = 1024
	MAX_PORT = 65535
)

// ResolvePort returns the PORT environment variable value if set otherwise defaultVal is returned
// An error is returned if the port is <= 1023 or > 65535 or if the PORT environment variable is set
// to a non-unsigned integer value
func ResolvePort(defaultVal uint32) (uint32, error) {

	thePort, resolvePortErr := lookupOrDefault(defaultVal)

	if resolvePortErr == nil {

		resolvePortErr = verifyPortRange(thePort)
	}

	return thePort, resolvePortErr
}

func lookupOrDefault(defaultVal uint32) (uint32, error) {

	lookedUpPort, isPortEnvSet := os.LookupEnv("PORT")

	var thePort uint32
	var err error

	if !isPortEnvSet || len(lookedUpPort) == 0 {
		thePort = defaultVal
	} else {

		var formatErr error
		thePort, formatErr = formatPort(lookedUpPort)

		err = formatErr
	}

	return thePort, err
}

func verifyPortRange(thePort uint32) error {

	var err error

	if thePort < MIN_PORT || thePort > MAX_PORT {
		err = fmt.
			Errorf("port value %d is outside valid range. It must be >= %d and <= %d",
				thePort,
				MIN_PORT, MAX_PORT)
	}

	return err
}

func formatPort(port string) (uint32, error) {

	formattedPort, uintParseErr := strconv.ParseUint(port, 10, 32)

	var thePort uint32

	if uintParseErr == nil {
		thePort = uint32(formattedPort)
	}

	return thePort, uintParseErr
}
