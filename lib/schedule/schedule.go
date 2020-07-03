package schedule

import (
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Runnable interface {
	Run() error
}

type scheduler struct {
	mu   sync.Mutex
	Jobs map[string]*job
}

type job struct {
	Previous      time.Time
	PreviousError error
	Next          time.Time
	Object        Runnable
	Duration      time.Duration
	Repeat        int
	IsConcurrent  bool
	Active        bool
	delete        bool
	name          string
}

func New(precision ...interface{}) *scheduler {
	sch := &scheduler{
		Jobs: map[string]*job{},
	}
	var duration time.Duration
	duration = 1 * time.Second
	if len(precision) > 1 {
		if v, ok := precision[0].(time.Duration); ok {
			duration = v
		}
	}

	go func() {
		for {
			now := time.Now().Unix()
			sch.mu.Lock()
			for _, j := range sch.Jobs {
				if j.Active && j.Next.Unix() <= now {
					if j.RunNow().delete {
						delete(sch.Jobs, j.name)
					}
				}
			}
			sch.mu.Unlock()
			time.Sleep(duration)
		}
	}()
	return sch
}
func (s *scheduler) GetJobs() map[string]*job {
	return s.Jobs
}
func (s *scheduler) appendJob(j *job) {
	direct := reflect.ValueOf(j.Object)
	i := 1
	name := direct.Type().String() + " " + strconv.Itoa(i)
	s.mu.Lock()
	for _, ok := s.Jobs[name]; ok; {
		i++
		name = direct.Type().String() + " " + strconv.Itoa(i)
	}
	s.Jobs[name] = j
	j.name = name
	s.mu.Unlock()
}

func (s *scheduler) Every(duration time.Duration, object Runnable) *job {
	j := &job{
		Duration: duration,
		Next:     time.Now().Add(duration),
		Object:   object,
		Repeat:   -1,
		Active:   true,
	}
	s.appendJob(j)
	return j
}

func (s *scheduler) RepeatN(times int, duration time.Duration, object Runnable) *job {
	j := &job{
		Duration: duration,
		Next:     time.Now().Add(duration),
		Object:   object,
		Repeat:   times,
		Active:   true,
	}
	s.appendJob(j)
	return j
}

func (s *scheduler) Once(duration time.Duration, object Runnable) *job {
	j := &job{
		Duration: duration,
		Next:     time.Now().Add(duration),
		Object:   object,
		Repeat:   1,
		Active:   true,
	}
	s.appendJob(j)
	return j
}

func (j *job) Concurrent() *job {
	j.IsConcurrent = true
	return j
}

func (j *job) Stop() *job {
	j.Active = false
	return j
}

func (j *job) Start() *job {
	j.Active = true
	return j
}

func (j *job) RunNow() *job {

	if j.Repeat == 0 || !j.Active {
		return j
	}

	j.Next = time.Now().Add(j.Duration)
	if j.Repeat > 0 {
		j.Repeat--
	}
	j.Previous = time.Now()
	j.PreviousError = (j.Object).Run()
	j.Next = time.Now().Add(j.Duration)

	if j.Repeat == 0 {
		j.Active = false
		j.delete = true
	}
	return j
}
