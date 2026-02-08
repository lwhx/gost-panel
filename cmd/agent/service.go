package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	svc "github.com/kardianos/service"
)

var agentSvcConfig = &svc.Config{
	Name:        "gost-node",
	DisplayName: "GOST Panel Agent",
	Description: "GOST Panel Node Agent",
	Option:      makeAgentServiceOptions(),
}

type agentProgram struct {
	agent *Agent
}

func (p *agentProgram) Start(s svc.Service) error {
	go p.run()
	return nil
}

func (p *agentProgram) Stop(s svc.Service) error {
	if p.agent != nil {
		p.agent.stopping.Store(true)
		p.agent.stopGost()
	}
	return nil
}

func (p *agentProgram) run() {
	// Parse flags (os.Args has been adjusted)
	flag.Parse()

	if *panelURL == "" || *token == "" {
		log.Fatalf("Missing required flags: -panel and -token")
	}

	log.Printf("Starting gost-agent %s (%s/%s)", AgentVersion, runtime.GOOS, runtime.GOARCH)

	resolvedGostPath, err := findGost(*gostPath)
	if err != nil {
		log.Fatalf("GOST not found: %v", err)
	}
	log.Printf("Using GOST: %s", resolvedGostPath)

	p.agent = NewAgent(*panelURL, *token, *configPath, resolvedGostPath, *gostAPI, *gostUser, *gostPass, *autoUpdate)
	if err := p.agent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}

func handleAgentServiceCommand() {
	if len(os.Args) < 3 {
		printAgentServiceUsage()
		os.Exit(1)
	}

	action := os.Args[2]

	// Capture extra flags for install
	var svrArgs []string
	if action == "install" && len(os.Args) > 3 {
		svrArgs = os.Args[3:]
	}

	execPath, _ := os.Executable()
	agentSvcConfig.WorkingDirectory = filepath.Dir(execPath)
	agentSvcConfig.Arguments = append([]string{"service", "run"}, svrArgs...)

	p := &agentProgram{}
	svr, err := svc.New(p, agentSvcConfig)
	if err != nil {
		fmt.Printf("Error creating service: %v\n", err)
		os.Exit(1)
	}

	switch action {
	case "install":
		_ = svr.Stop()
		_ = svr.Uninstall()
		if err := svr.Install(); err != nil {
			fmt.Printf("Install failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service installed successfully")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  gost-agent service start    - Start the service")
		fmt.Println("  gost-agent service stop     - Stop the service")
		fmt.Println("  gost-agent service restart  - Restart the service")
		fmt.Println("  gost-agent service uninstall - Remove the service")

	case "uninstall":
		_ = svr.Stop()
		if err := svr.Uninstall(); err != nil {
			fmt.Printf("Uninstall failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service uninstalled successfully")

	case "start", "stop", "restart":
		if err := svc.Control(svr, action); err != nil {
			fmt.Printf("%s failed: %v\n", action, err)
			os.Exit(1)
		}
		fmt.Printf("Service %s successfully\n", action)

	case "status":
		status, err := svr.Status()
		if err != nil {
			fmt.Printf("Status check failed: %v\n", err)
			os.Exit(1)
		}
		switch status {
		case svc.StatusRunning:
			fmt.Println("Service is running")
		case svc.StatusStopped:
			fmt.Println("Service is stopped")
		default:
			fmt.Println("Service status unknown")
		}

	case "run":
		// Called by the service manager
		os.Args = append(os.Args[:1], os.Args[3:]...)
		_ = svr.Run()

	default:
		printAgentServiceUsage()
		os.Exit(1)
	}
}

func printAgentServiceUsage() {
	fmt.Println("GOST Panel Agent - Service Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gost-agent service <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install    Install as system service")
	fmt.Println("  uninstall  Remove system service")
	fmt.Println("  start      Start the service")
	fmt.Println("  stop       Stop the service")
	fmt.Println("  restart    Restart the service")
	fmt.Println("  status     Check service status")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gost-agent service install -panel http://panel:8080 -token YOUR_TOKEN")
	fmt.Println("  gost-agent service start")
	fmt.Println("  gost-agent service stop")
}
