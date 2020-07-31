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
	Previous      time.Time     `json:"previous"`
	PreviousError error         `json:"previous_error"`
	Next          time.Time     `json:"next"`
	Object        Runnable      `json:"job"`
	Duration      time.Duration `json:"duration"`
	Repeat        int           `json:"repeat"`
	IsConcurrent  bool          `json:"is_concurrent"`
	Active        bool
	delete        bool
	Name          string `json:"Name"`
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
						delete(sch.Jobs, j.Name)
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
	for {
		if _, ok := s.Jobs[name]; ok {
			i++
			name = direct.Type().String() + " " + strconv.Itoa(i)
		} else {
			break
		}

	}
	s.Jobs[name] = j
	j.Name = name
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
	go func() {
		j.PreviousError = (j.Object).Run()
		j.Next = time.Now().Add(j.Duration)

		if j.Repeat == 0 {
			j.Active = false
			j.delete = true
		}
	}()

	return j
}
