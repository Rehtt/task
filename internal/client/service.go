package client

import (
	"context"
	"github.com/Rehtt/Kit/log/logs"
	"github.com/Rehtt/service-go"
	"github.com/Rehtt/task/internal"
	"github.com/Rehtt/task/rpc"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Program struct {
	sTime   time.Duration
	sTick   *time.Ticker
	service service.Service
}

var ServerAddr string

func (p *Program) Start(s service.Service) error {
	logs.Info("service start")
	p.service = s

	conn, err := grpc.Dial(ServerAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c := rpc.NewJobClient(conn)
	p.sTick = time.NewTicker(p.sTime)
	go func(c rpc.JobClient) {
		ctx := context.Background()
		info := &rpc.ServiceInfo{
			Version: Version,
			Uuid:    SystemUUID,
			Os:      SystemOS,
		}
		for {
			resp, err := c.GetJob(ctx, info)
			if err == nil {
				for _, job := range resp.GetJobList() {
					go func() {
						if msg, err := p.handle(job); err != nil {
							c.JobErr(ctx, &rpc.Err{
								Id:       job.Id,
								Error:    err.Error(),
								ErrorMsg: msg,
								Info:     info,
							})
						}
					}()
				}
			}
			<-p.sTick.C
		}
	}(c)
	return nil
}

func (p *Program) Stop(s service.Service) error {
	logs.Info("service stop")
	return nil
}

func (p *Program) handle(job *rpc.JobInfo) (string, error) {
	if job.RunTime != nil {
		t, err := time.Parse(time.RFC3339, *job.RunTime)
		if err != nil {
			return "", err
		}
		time.Sleep(t.Sub(time.Now()))
	}
	var rep = 1
	if job.RunRepetition != nil {
		rep = int(*job.RunRepetition)
	}
	if job.Url != nil {
		dir, path, err := p.downloadFile(*job.Url)
		if err != nil {
			return "", err
		}
		job.Command = strings.ReplaceAll(job.Command, internal.JobFileMark, path)
		if rep == 1 {
			defer os.RemoveAll(dir)
		}
	}
	var timeout = internal.CommandTimeout
	if job.RunTimeout != nil {
		timeout = time.Duration(*job.RunTimeout)
	}

	c := strings.Split(job.Command, " ")
	var (
		name = c[0]
		arg  = c[1:]
	)
	for i := 0; i < rep; i++ {
		var ctx, _ = context.WithTimeout(context.Background(), timeout)
		out, err := exec.CommandContext(ctx, name, arg...).CombinedOutput()
		if err != nil {
			return string(out), err
		}
		if job.RunInterval != nil {
			time.Sleep(time.Duration(*job.RunInterval))
		}
	}

	return "", nil
}

func (p *Program) downloadFile(url string) (dir, path string, err error) {
	dir, _ = filepath.Split(SelfPath)
	dir, err = os.MkdirTemp(dir, "cache")
	if err != nil {
		return "", "", err
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	path = filepath.Join(dir, resp.Header.Get("filename"))
	data, _ := io.ReadAll(resp.Body)
	err = os.WriteFile(path, data, 0644)
	return
}
