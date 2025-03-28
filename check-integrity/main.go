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
	var (
		namespace = flag.String("namespace", "", "Target namespace")
	)
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
	
	commands := []string{
		fmt.Sprintf(`do:(##class(%%SYS.Namespace).Exists("%s")'=1) ##class(%%SYSTEM.Process).Terminate(,1)`, *namespace),		
    
		// Check for missing foreign keys
		fmt.Sprintf(`set $namespace = "%s"`, *namespace),
		fmt.Sprintf(`set class = ""`),
		fmt.Sprintf(`set sql = ##class(%%SQL.Statement).%%New()`),
		fmt.Sprintf(`do sql.%%Prepare("SELECT tClass.ID As class, tProp.Name As property, tProp.Type As type FROM %%Dictionary.CompiledClass AS tClass JOIN %%Dictionary.CompiledProperty AS tProp ON (tProp.parent = tClass.ID) LEFT JOIN %%Dictionary.CompiledForeignKey AS tFk ON (tFk.parent = tProp.parent AND tFk.Properties = tProp.Name) WHERE tClass.ID %%MATCHES '[a-z,A-Z]*' AND Super LIKE '%%Persistent%%' AND tProp.Type %%MATCHES '[a-z,A-Z]*' AND (SELECT COUNT(ID) FROM %%Dictionary.CompiledClass As tClass WHERE tClass.ID=tProp.Type AND Super LIKE '%%Persistent%%')<>0 AND tProp.Relationship = 0 AND tProp.Transient = 0 AND tFk.Name IS NULL")`),
		fmt.Sprintf(`set query = sql.%%Execute()`),
		fmt.Sprintf(`while query.%%Next() { if (class '= query.%%Get("class")) { set class = query.%%Get("class") write !,"" write !,"Classe: ["_$replace(class,".","/")_".cls]" } write !,"ForeignKey fk{nome_da_fk}("_query.%%Get("property")_") References "_query.%%Get("type")_"();" }`),
		fmt.Sprintf(`do query.%%Close()`),
		fmt.Sprintf(`if (class '= "") { do ##class(%%SYSTEM.Process).Terminate(,1)}`),		
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
