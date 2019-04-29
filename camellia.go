package camellia

type Server struct {
	El  *EventLoop
	Lis []*Listener
}

func NewServer(network, addr string) (*Server, error) {
	var (
		el     = NewEventLoop()
		server = &Server{
			El:  el,
			Lis: make([]*Listener, 0),
		}
	)

	if err := server.AddListener(network, addr); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) AddListener(network, addr string) error {
	listener, err := NewListener(network, addr, s.El)
	if err != nil {
		return err
	}
	s.Lis = append(s.Lis, listener)
	return nil
}

func (s *Server) AddEvent(e *Event) {
	s.El.AddEvent(e)
}

func (s *Server) AddPeriodTask(e *PeriodTask) {
	s.El.AddPeriodTask(e)
}

func (s *Server) StartServe() error {
	var err error
	for _, l := range s.Lis {
		if err = l.BindAndListen(); err != nil {
			s.closeAllLis()
			return err
		}
		if err = l.RegisterAccept(); err != nil {
			s.closeAllLis()
			return err
		}
	}

	s.El.Run()
	return nil
}

func (s *Server) closeAllLis() {
	for _, l := range s.Lis {
		_ = l.Close()
	}
}
