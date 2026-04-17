package common

import (
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser opens the given URL in the user's default browser. It blocks
// until the browser command completes (which is usually as soon as the browser
// page is successfully opened, but could technically be when the browser page
// is closed on certain older Linux configurations). Prefer this function over
// OpenBrowserAsync if you can tolerate it potentially blocking while the
// browser remains open, or if handling async errors returned on a channel
// (see [OpenBrowserAsync]) is just too complex.
// Declared as a var so tests can replace it with a no-op.
var OpenBrowser = func(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Escape '&' so cmd.exe doesn't treat it as a command separator
		cmd = exec.Command("cmd", "/c", "start", strings.ReplaceAll(url, "&", "^&"))
	case "darwin":
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Run()
}

// OpenBrowserAsync opens the given URL in the user's default browser in a
// goroutine and returns a channel that receives an error if the browser command
// fails. This is used by the login flow to avoid blocking if the browser
// command hangs on certain Linux configurations. Prefer this function whenever
// you need to open the browser, then wait for something else to happen on a
// channel (because in that case, it's easy to also wait for errors on the
// returned error chan via a select statement), or if opening the browser is
// "best effort" and you don't care to handle errors at all.
func OpenBrowserAsync(url string) <-chan error {
	errCh := make(chan error, 1)
	// Capture the current function value before spawning the goroutine,
	// so that tests can safely restore the original via t.Cleanup without
	// racing with the goroutine.
	fn := OpenBrowser
	go func() {
		if err := fn(url); err != nil {
			errCh <- err
		}
	}()
	return errCh
}
