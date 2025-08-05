package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type InterSystemsSession struct {
	instance string
	command  string
}

func NewSession() (*InterSystemsSession, error) {	
	if path, err := exec.LookPath("irissession"); err == nil {
			return &InterSystemsSession{
				instance: "IRIS",
				command:  path,
			}, nil
		}

	if path, err := exec.LookPath("csession"); err == nil {
			return &InterSystemsSession{
				instance: "CACHE",
				command:  path,
			}, nil
		}

	return nil, fmt.Errorf("neither irissession nor csession found in PATH")
}

func (s *InterSystemsSession) ExecuteCommand(command string) (int, error) {
	cmd := exec.Command(s.command, s.instance, "-U", "%SYS")
	cmd.Stdin = strings.NewReader(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), err
			}
		}
		return 1, err
	}

	return 0, nil
}

func main() {
	var namespace = flag.String("namespace", "", "Namespace to destroy")
	flag.Parse()

	if *namespace == "" {
		fmt.Fprintf(os.Stderr, "Error: namespace is required\n")
		os.Exit(1)
	}
	
	session, err := NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
		os.Exit(1)
	}

	// Execute namespace destruction commands
	commands := []string{
		fmt.Sprintf(`do:(##class(%%SYS.Namespace).Exists("%s")'=1) ##class(%%SYSTEM.Process).Terminate(,0)`, *namespace),
		fmt.Sprintf(`set path = ##class(Config.Databases).GetDirectory("%s")`, *namespace),
		fmt.Sprintf(`do ##class(Security.Applications).Delete("/csp/%s")`, *namespace),	
		fmt.Sprintf(`do ##class(Config.Namespaces).Delete("%s")`, *namespace),
		fmt.Sprintf(`do ##class(Config.Databases).Delete("%s")`, *namespace),
		fmt.Sprintf(`do ##class(SYS.Database).DeleteDatabase(path)`),
		fmt.Sprintf(`do ##class(%%Library.File).RemoveDirectoryTree(path)`),
		fmt.Sprintf(`do ##class(%%SYSTEM.Process).Terminate(,0)`),
	}

	command := strings.Join(commands, "\n")
	exitCode, err := session.ExecuteCommand(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing commands: %v\n", err)
		os.Exit(exitCode)
	}
	os.Exit(exitCode)
}
