//go:build curl
// +build curl

package okteto

import (
	"fmt"
)

func commandAndArgs(oktetoURL, namespace string) (command string, args []string) {
	command = "sh"
	args = []string{"-c", fmt.Sprintf("curl %s/auth/kubetoken/%s -L -H 'authorization: Bearer %s'", oktetoURL, namespace, Context().Token)}
	return
}
