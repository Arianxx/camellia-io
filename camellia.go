package camellia

// Server decorates some necessary components required to start to serve.
// El is the server's internal eventloop.
// Lis is a Listener slice listened on the machine.
type Server struct {
	El  *EventLoop
	Lis []*Listener
}

// NewServer create a new server.
func NewServer() *Server {
	return &Server{
		El:  NewEventLoop(),
		Lis: make([]*Listener, 0),
	}
}

// AddListener add a listener to the server.
func (s *Server) AddListener(l *Listener) {
	s.Lis = append(s.Lis, l)
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
