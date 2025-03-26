package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"	
	"strings"
	"syscall"
)

// InterSystemsSession represents a session with Cache or IRIS
type InterSystemsSession struct {
	instance string
	command  string
}

// NewSession creates a new InterSystems session
func NewSession() (*InterSystemsSession, error) {
	// Check for IRIS first, then Cache
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

// ExecuteCommand runs an ObjectScript command
func (s *InterSystemsSession) ExecuteCommand(command string) (int, error) {
	// Add namespace parameter for better session handling
	cmd := exec.Command(s.command, s.instance, "-U", "%SYS")
	cmd.Stdin = strings.NewReader(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// Try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), err
			}
		}
		return 1, err // Default to exit code 1 if we can't determine the actual code
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

	// Create session
	session, err := NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating session: %v\n", err)
		os.Exit(1)
	}

	// Execute integrity check query
	commands := []string{
		// Switch to %SYS namespace and terminate existing processes
		fmt.Sprintf(`set $namespace = "%%SYS"`),
		fmt.Sprintf(`do:(##class(%%SYS.Namespace).Exists("%s")'=1) ##class(%%SYSTEM.Process).Terminate(,1)`, *namespace),
		fmt.Sprintf(`set $namespace = "%s"`, *namespace),
		
		// Check for missing foreign keys
		fmt.Sprintf(`set class = ""`),
		fmt.Sprintf(`set sql = ##class(%%SQL.Statement).%%New()`),
		fmt.Sprintf(`do sql.%%Prepare("SELECT tClass.ID As class, tProp.Name As property, tProp.Type As type FROM %%Dictionary.CompiledClass AS tClass JOIN %%Dictionary.CompiledProperty AS tProp ON (tProp.parent = tClass.ID) LEFT JOIN %%Dictionary.CompiledForeignKey AS tFk ON (tFk.parent = tProp.parent AND tFk.Properties = tProp.Name) WHERE tClass.ID %%MATCHES '[a-z,A-Z]*' AND Super LIKE '%%Persistent%%' AND tProp.Type %%MATCHES '[a-z,A-Z]*' AND (SELECT COUNT(ID) FROM %%Dictionary.CompiledClass As tClass WHERE tClass.ID=tProp.Type AND Super LIKE '%%Persistent%%')<>0 AND tProp.Relationship = 0 AND tProp.Transient = 0 AND tFk.Name IS NULL")`),
		fmt.Sprintf(`set query = sql.%%Execute()`),
		fmt.Sprintf(`while query.%%Next() {`),
		fmt.Sprintf(`
			if (class '= query.%%Get("class")) {
				set class = query.%%Get("class")
				write !,""
				write !,"Classe: ["_$replace(class,".","/")_".cls"]"
			}
			write !,"ForeignKey fk{"_query.%%Get("property")_"}(_query.%%Get("type")_"();")"
		}`),
		
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
