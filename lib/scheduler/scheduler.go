package scheduler

import (
	"regexp"
	"strings"
	"time"
)

var jobs []*Job
var running = false

type Job struct {
	JobID     string               `gorm:"column:job_id" json:"job_id"`
	Every     *regexp.Regexp       `gorm:"column:every" json:"every"`
	LastRun   time.Time            `gorm:"column:last_run" json:"last_run"`
	Action    func(job *Job) error `gorm:"-" json:"-"`
	Pause     bool                 `gorm:"column:pause" json:"pause"`
	Running   bool
	OnError   func(job *Job, err error)
	OnSuccess func(job *Job)
	OnFinish  func(job *Job)
}

func (Job) TableName() string {
	return "scheduler_job"
}

func Register() {
	if running {
		return
	}
	running = true
	go func() {
		for {
			var now = time.Now().Format("Mon,2006-01-02,15:04:05")
			for idx, _ := range jobs {
				var job = jobs[idx]
				if !job.Running && !job.Pause && job.Every.MatchString(now) {
					go func() {
						job.Running = true
						err := job.Action(job)

						job.Running = false
						if err != nil && job.OnError != nil {
							job.OnError(job, err)
						}

						if err == nil && job.OnSuccess != nil {
							job.OnSuccess(job)
						}
						if job.OnFinish != nil {
							job.OnFinish(job)
						}

					}()
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func CreateJob(id string, every string, action func(job *Job) error) *Job {
	every = strings.Replace(every, "*", `[a-zA-Z0-9]+`, -1)
	var job = Job{
		JobID:  id,
		Every:  regexp.MustCompile("(?i)" + every),
		Action: action,
	}
	return &job
}

func (job *Job) Start() {
	jobs = append(jobs, job)
	if !running {
		Register()
	}
}
