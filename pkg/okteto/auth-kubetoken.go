//go:build kubetoken
// +build kubetoken

package okteto

// What we aimed for
func commandAndArgs(oktetoURL, namespace string) (command string, args []string) {
	command = "okteto"
	args = []string{"kubetoken", "--context", oktetoURL, "--namespace", namespace}
	return
}
