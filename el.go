package camellia

import (
	"github.com/arianxx/camellia-io/internal"
	"time"
)

type EventLoop struct {
	*internal.Selector
	events         []*Event
	interval       time.Duration
	done           bool
	triggerDataPtr *interface{}
	periodTasks    []*PeriodTask
}

func NewEventLoop() *EventLoop {
	return &EventLoop{
		Selector:    internal.New(1024),
		events:      []*Event{},
		periodTasks: []*PeriodTask{},
		interval:    100 * time.Millisecond,
	}
}

func (el *EventLoop) AddEvent(e *Event) {
	el.events = append(el.events, e)
}

func (el *EventLoop) AddPeriodTask(t *PeriodTask) {
	el.periodTasks = append(el.periodTasks, t)
}

func (el *EventLoop) SetTriggerDataPtr(data interface{}) {
	el.triggerDataPtr = &data
}

func (el *EventLoop) Run() {
	for _, e := range el.events {
		if e.Serving != nil {
			e.Serving(el, nil)
		}
	}

	for _, t := range el.periodTasks {
		t.setNextTriggerTime()
	}

	for !el.done {
		el.Tick()
	}
}

func (el *EventLoop) Tick() {
	var (
		ed        EventData
		sleepTime = el.interval
	)

	nearestTask := el.findNearestTask()
	if nearestTask != nil {
		sleepTime = nearestTask.nextTriggerTime.Sub(time.Now())
	}

	keys, _, _ := el.Selector.Poll(int(sleepTime / time.Millisecond))
	for _, k := range keys {
		ed = k.Data.(EventData)
		action := ed.e(el, ed.data)
		el.processAction(action, k.Fd)
	}

	if nearestTask != nil {
		nearestTask.Event(el, nil)
		nearestTask.setNextTriggerTime()
	}
}

func (el *EventLoop) Register(fd int, mask uint32, e EventProc, d interface{}) error {
	return el.Selector.Register(fd, mask, EventData{e, d})
}

func (el *EventLoop) Done() {
	el.done = true
}

func (el *EventLoop) findNearestTask() *PeriodTask {
	var ans *PeriodTask
	for _, t := range el.periodTasks {
		if ans == nil || t.nextTriggerTime.Before(ans.nextTriggerTime) {
			ans = t
		}
	}
	return ans
}

func (el *EventLoop) processAction(action Action, fd int) {
	switch action {
	case SHUTDOWN_RD:
		_, _ = el.Unregister(fd, internal.EV_READABLE)
	case SHUTDOWN_WR:
		_, _ = el.Unregister(fd, internal.EV_WRITABLE)
	case SHUTDOWN_RDWR:
		_, _ = el.Unregister(fd, internal.EV_READABLE)
		_, _ = el.Unregister(fd, internal.EV_WRITABLE)
	case TRIGGER_OPEN_EVENT:
		for _, t := range el.events {
			if t.Open != nil {
				t.Open(el, el.triggerDataPtr)
			}
		}
	case TRIGGER_DATA_EVENT:
		for _, t := range el.events {
			if t.Data != nil {
				t.Data(el, el.triggerDataPtr)
			}
		}
	case TRIGGER_CLOSE_EVENT:
		for _, t := range el.events {
			if t.Closed != nil {
				t.Closed(el, el.triggerDataPtr)
			}
		}
	case CONTINUE:
	}

	el.triggerDataPtr = nil
}

type Action int

const (
	CONTINUE Action = iota
	SHUTDOWN_RD
	SHUTDOWN_WR
	SHUTDOWN_RDWR
	TRIGGER_OPEN_EVENT
	TRIGGER_DATA_EVENT
	TRIGGER_CLOSE_EVENT
)

type Event struct {
	Serving, Open, Closed, Data TriggerProc
}

type EventProc func(el *EventLoop, data interface{}) Action

type EventData struct {
	e    EventProc
	data interface{}
}

type TriggerProc func(el *EventLoop, dataPtr *interface{})

type PeriodTask struct {
	Interval time.Duration
	Event    TriggerProc

	nextTriggerTime time.Time
}

func (t *PeriodTask) setNextTriggerTime() {
	t.nextTriggerTime = time.Now().Add(t.Interval)
}
