package environment

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func tearDown() {
	os.Unsetenv("PORT")
}

func TestWhenPortEnvVarNotSetThenDefaultValueReturned(t *testing.T) {

	var defaultPort uint32 = 1234

	resolvedPort, _ := ResolvePort(defaultPort)

	if resolvedPort != defaultPort {
		t.Errorf("Expected resolved port %d to be the default port %d", resolvedPort, defaultPort)
	}
}

func TestWhenPortEnvVarIsSetThenItIsReturned(t *testing.T) {

	var expectedPort uint32 = 9090

	os.Setenv("PORT", fmt.Sprintf("%d", expectedPort))
	defer tearDown()

	resolvedPort, _ := ResolvePort(1234)

	if resolvedPort != expectedPort {
		t.Errorf("Expected resolved port %d to be the default port %d", resolvedPort, expectedPort)
	}
}

func TestWhenPortEnvVarIsSetButIsNotValidUintThenAnErrorIsReturned(t *testing.T) {

	os.Setenv("PORT", "not valid port")
	defer tearDown()

	_, err := ResolvePort(1234)

	if err == nil {
		t.Error("PORT env value not a valid unsigned integer, an error should have been returned")
	}
}

func TestWhenPortValueIsLessThanTheMinimumValidPortThenErrorIsReturned(t *testing.T) {

	_, err := ResolvePort(MIN_PORT - 1)

	if matched, _ := regexp.MatchString(`^port value \d+ is outside valid range.*`, err.Error()); !matched {
		t.Error("Port value is out of valid range, an error should have been returned")
	}
}

func TestWhenPortValueIsGreaterThanTheMaximumValidPortThenErrorIsReturned(t *testing.T) {

	_, err := ResolvePort(MAX_PORT + 1)

	if matched, _ := regexp.MatchString(`^port value \d+ is outside valid range.*`, err.Error()); !matched {
		t.Error("Port value is out of valid range, an error should have been returned")
	}
}
