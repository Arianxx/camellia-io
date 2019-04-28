package camellia

import (
	"fmt"
	"syscall"
)

type UnknownNetworkError struct {
	network string
	sa      *syscall.Sockaddr
}

func (u *UnknownNetworkError) Error() string {
	return fmt.Sprintf("unknow network: %s", u.network)
}
