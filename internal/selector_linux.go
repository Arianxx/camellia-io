package internal

import (
	"syscall"
)

type SelectorKey struct {
	fd    int
	event uint32
	data  map[uint32]interface{}
}

type Selector struct {
	epfd int
	Keys []*SelectorKey
}

func New(size int) *Selector {
	p := &Selector{}
	epfd, err := syscall.EpollCreate(size)
	if err != nil {
		panic(err)
	}
	p.epfd = epfd
	p.Keys = make([]*SelectorKey, size)
	return p
}

func (p *Selector) Close() {
	if err := syscall.Close(p.epfd); err != nil {
		panic(err)
	}
	p.Keys = nil
}

func (p *Selector) Register(fd int, mask uint32, data interface{}) error {
	if fd > len(p.Keys) {
		return &FdExecLimit{fd}
	}

	if p.Keys[fd] == nil {
		p.Keys[fd] = &SelectorKey{fd: fd, event: EV_NONE, data: make(map[uint32]interface{})}
	}
	key := p.Keys[fd]

	ee := &syscall.EpollEvent{Fd: int32(fd)}
	var op int
	if key.event == EV_NONE {
		op = syscall.EPOLL_CTL_ADD
	} else {
		op = syscall.EPOLL_CTL_MOD
	}

	key.event |= mask
	if err := p.setRawEvent(ee, mask, key, data); err != nil {
		return err
	}

	if err := syscall.EpollCtl(p.epfd, op, fd, ee); err != nil {
		return err
	}

	return nil
}

func (p *Selector) Unregister(fd int, mask uint32) error {
	if fd > len(p.Keys) {
		return &FdExecLimit{fd}
	}

	key := p.Keys[fd]
	if key == nil || key.event&mask == 0 {
		return nil
	}
	key.event &= ^mask

	var op int
	if key.event != 0 {
		op = syscall.EPOLL_CTL_MOD
	} else {
		op = syscall.EPOLL_CTL_DEL
		p.Keys[fd] = nil
	}

	ee := &syscall.EpollEvent{Fd: int32(fd)}
	if err := p.setRawEvent(ee, mask, key, nil); err != nil {
		return err
	}

	if err := syscall.EpollCtl(p.epfd, op, fd, ee); err != nil {
		return err
	}

	return nil
}

func (p *Selector) GetSelectorKey(fd int) *SelectorKey {
	return p.Keys[fd]
}

func (p *Selector) Poll(t int) ([]*SelectorKey, []uint32, error) {
	events := make([]syscall.EpollEvent, len(p.Keys))
	n, err := syscall.EpollWait(p.epfd, events, t)
	if err != nil {
		return nil, nil, err
	}

	keys, masks := make([]*SelectorKey, n), make([]uint32, n)
	var e *syscall.EpollEvent
	for i := 0; i < n; i++ {
		e = &events[i]
		keys[i] = p.Keys[e.Fd]
		if e.Events&syscall.EPOLLIN != 0 {
			masks[i] |= EV_READABLE
		}
		if e.Events&syscall.EPOLLOUT != 0 || e.Events&syscall.EPOLLERR != 0 || e.Events&syscall.EPOLLHUP != 0 {
			masks[i] |= EV_WRITABLE
		}
	}

	return keys, masks, nil
}

func (p *Selector) setRawEvent(e *syscall.EpollEvent, mask uint32, key *SelectorKey, data interface{}) error {
	if mask&EV_READABLE != 0 {
		e.Events |= syscall.EPOLLIN
		key.data[EV_READABLE] = data
	} else if mask&EV_WRITABLE != 0 {
		e.Events |= syscall.EPOLLOUT
		key.data[EV_WRITABLE] = data
	} else {
		return &UnknownMask{mask}
	}

	return nil
}
