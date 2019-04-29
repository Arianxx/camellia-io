package camellia

// Server decorates some necessary components required to start to serve.
type Server struct {
	// El is the server's internal eventloop.
	El *EventLoop
	// Lis is a Listener slice listened on the machine.
	Lis []*Listener
}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{
		El:  NewEventLoop(),
		Lis: make([]*Listener, 0),
	}
}

// AddListener adds a listener to the server.
func (s *Server) AddListener(l *Listener) {
	s.Lis = append(s.Lis, l)
}

// AddEvent adds a Event to the server. The func will be triggered if the corresponding event is coming.
func (s *Server) AddEvent(e *Event) {
	s.El.AddEvent(e)
}

// AddPeriodTask adds a period task to the server that will be period triggered.
func (s *Server) AddPeriodTask(e *PeriodTask) {
	s.El.AddPeriodTask(e)
}

// StartServe blocks the thread to serve until the server has been broken.
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
