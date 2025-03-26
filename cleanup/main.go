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
	var namespace = flag.String("namespace", "", "Target namespace")
	flag.Parse()

	if *namespace == "" {
		fmt.Fprintf(os.Stderr, "Error: namespace is required")
		os.Exit(1)
	}
	
	session, err := NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	commands := []string{
		fmt.Sprintf(`do:(##class(%%SYS.Namespace).Exists("%s")'=1) ##class(%%SYSTEM.Process).Terminate(,1)`, *namespace),		
    
		fmt.Sprintf(`set $namespace = "%s"`, *namespace),
		fmt.Sprintf(`do ##class(%%SYSTEM.OBJ).DeletePackage("test")`),
		fmt.Sprintf(`set st = ##class(%%SQL.Statement).%%New()`),
		fmt.Sprintf(`do st.%%PrepareClassQuery("%%SYS.GlobalQuery","NameSpaceList")`),
		fmt.Sprintf(`set rs = st.%%Execute("%s")`, *namespace),
		fmt.Sprintf(`while rs.%%Next() { if (rs.%%Get("Name")="cspRule") { continue } write !,"Delete "_rs.%%Get("Name") do ##class(%%Studio.Global).Kill("^"_rs.%%Get("Name"),1) }`),		
		fmt.Sprintf(`set $namespace = "%%SYS"`),
		fmt.Sprintf(`set path = ##class(%%SYSTEM.Util).ManagerDirectory()_"%s"`, *namespace),
		fmt.Sprintf(`do ##class(SYS.Database).ReturnUnusedSpace(path)`),
		fmt.Sprintf(`do ##class(SYS.Database).DismountDatabase(path)`),
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