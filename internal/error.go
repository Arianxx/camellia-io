package internal

import "fmt"

type UnknownMask struct {
	mask uint32
}

func (u *UnknownMask) Error() string {
	return fmt.Sprintf("unknow event mask: %d", u.mask)
}

type FdExecLimit struct {
	fd int
}

func (f *FdExecLimit) Error() string {
	return fmt.Sprintf("fd %d exceed the limit", f.fd)
}
