package main

import (
	"encoding/json"
	"github.com/Rehtt/Kit/file"
	"github.com/Rehtt/Kit/host"
	"github.com/Rehtt/Kit/log/logs"
	"github.com/Rehtt/Kit/sudo"
	"github.com/Rehtt/service-go"
	"github.com/Rehtt/task/internal"
	"github.com/Rehtt/task/internal/client"
	"github.com/google/uuid"
	"os"
	"strings"
)

var taskservice service.Service

func init() {
	var sConfig = &service.Config{
		Name:        "task-service",
		DisplayName: "Task Service",
		Description: "Run Task Service",
	}
	switch client.SystemOS {
	case "windows":
		sConfig.Executable = "C:\\ProgramData\\Task\\task-service.exe"
	}

	var err error
	taskservice, err = service.New(client.NewProgram(), sConfig)
	if err != nil {
		logs.Fatal("service new error: %s", err)
	}

	client.SelfPath, err = os.Executable()
	if err != nil {
		logs.Fatal("service get path error: %s", err)
	}
	//param(sConfig)
	//_, err = taskservice.Status()
	//if err != nil && errors.Is(err, service.ErrNotInstalled) {
	//	c, err := sudo.SudoRunShell(client.SelfPath, "i")
	//	if err != nil {
	//		logs.Fatal("service install error: %s", err)
	//	}
	//	if err = c.Run(); err != nil {
	//		logs.Fatal("service install error: %s", err)
	//	}
	//	logs.Info("service background running")
	//}
	//if client.SelfPath != sConfig.Executable {
	//	os.Exit(0)
	//}

	j, _ := json.Marshal(map[string]any{
		"name": internal.StringMust(os.Hostname()),
		"os":   client.SystemOS,
	})
	u, err := host.GetBaseUUIDString()
	if err != nil {
		u = uuid.NameSpaceOID.String()
	}
	client.SystemUUID = uuid.NewSHA1(uuid.Must(uuid.Parse(u)), j).String()
}

func param(sConfig *service.Config) {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "i":
			install(sConfig)
		case "r":
			remove(sConfig)
		case "u":

		}
		os.Exit(0)
	}
}
func remove(sConfig *service.Config) {
	if err := taskservice.Uninstall(); err != nil && strings.Contains(strings.ToLower(err.Error()), "access is denied") {
		c, err := sudo.SudoRunShell(sConfig.Executable, "r")
		if err != nil {
			logs.Fatal("service remove error: %s", err)
		}
		if err = c.Run(); err != nil {
			logs.Fatal("service remove error: %s", err)
		}
	} else if err != nil {
		logs.Fatal("service remove error: %s", err)
	} else {
		taskservice.Stop()
	}
	logs.Info("service remove success")
}

func install(sConfig *service.Config) {
	if err := taskservice.Install(); err != nil && strings.Contains(strings.ToLower(err.Error()), "access is denied") {
		c, err := sudo.SudoRunShell(client.SelfPath, "i")
		if err != nil {
			logs.Fatal("service install error: %s", err)
		}
		if err = c.Run(); err != nil {
			logs.Fatal("service install error: %s", err)
		}
	} else if err != nil {
		logs.Fatal("service install error: %s", err)
	} else {
		if client.SelfPath != sConfig.Executable {
			data, _ := os.ReadFile(client.SelfPath)
			if err = file.CheckWriteFile(sConfig.Executable, data, 0644, true, 0755); err != nil {
				logs.Fatal("service copy self error: %s", err)
			}
		}
		taskservice.Start()
	}
	logs.Info("service install success")
}
