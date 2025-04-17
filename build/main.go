package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
		namespace   = flag.String("namespace", "", "Target namespace")
		projectPath = flag.String("project", "", "Project path")
		snpm        = flag.Bool("snpm", false, "Enable SNPM process (default: false)")
	)
	flag.Parse()

	if *namespace == "" {
		fmt.Fprintln(os.Stderr, "Error: namespace is required")
		os.Exit(1)
	}

	// Get absolute path for workdir
	absWorkdir, err := filepath.Abs(filepath.Join(filepath.Dir(os.Getenv("GITHUB_WORKSPACE")), *projectPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving workdir path: %v\n", err)
		os.Exit(1)
	}

	session, err := NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Format the resource name
	resource := "%DB_" + strings.ToUpper(*namespace)

	// Build the ObjectScript commands
	commands := []string{
		fmt.Sprintf(`do:(##class(%%SYS.Namespace).Exists("%s")=1) ##class(%%SYSTEM.Process).Terminate(,1)`, *namespace),

		// Create database directory and database
		fmt.Sprintf(`set path = ##class(%%SYSTEM.Util).ManagerDirectory()_"%s"`, *namespace),
		fmt.Sprintf(`do ##class(%%File).CreateDirectory(path)`),
		fmt.Sprintf(`do ##class(SYS.Database).CreateDatabase(path,,,,"%s",2)`, resource),

		// Configure database and namespace
		fmt.Sprintf(`set database("Directory") = path`),
		fmt.Sprintf(`do ##class(Config.Databases).Create("%s",.database)`, *namespace),
		fmt.Sprintf(`set properties("Globals")="%s"`, *namespace),
		fmt.Sprintf(`do ##class(Config.Namespaces).Create("%s",.properties)`, *namespace),

		// Verify namespace creation
		fmt.Sprintf(`do:(##class(Config.Namespaces).Exists("%s")'=1) ##class(%%SYSTEM.Process).Terminate(,1)`, *namespace),
	}

	if *snpm {
		commands = append(commands, []string{
			// Switch to target namespace and import files
			fmt.Sprintf(`set file = ##class(%%Library.FileBinaryStream).%%New()`),
			fmt.Sprintf(`do file.Write("")`),
			fmt.Sprintf(`set filename = file.Filename`),
			fmt.Sprintf(`kill file`),
			fmt.Sprintf(`do $SYSTEM.OBJ.Export("Monitor.Sample.cls", .filename)`),
			fmt.Sprintf(`set $namespace = "%s"`, *namespace),
			fmt.Sprintf(`do $SYSTEM.OBJ.Load(filename, "ck")`),
			fmt.Sprintf(`do ##class(%%File).Delete(filename)`),
		}...)
	}

	commands = append(commands, []string{
		// Switch to target namespace and import files
		fmt.Sprintf(`kill`),
		fmt.Sprintf(`set $namespace = "%s"`, *namespace),
		fmt.Sprintf(`set workcls = "%s/src"`, strings.ReplaceAll(absWorkdir, "\\", "/")),
		fmt.Sprintf(`do ##class(%%SYSTEM.OBJ).ImportDir(workcls,"*.inc","c","",1)`),
		fmt.Sprintf(`do ##class(%%SYSTEM.OBJ).ImportDir(workcls,"*.cls","c","",1)`),

		// Set test root and check for errors
		fmt.Sprintf(`set ^UnitTestRoot = "%s/test"`, strings.ReplaceAll(absWorkdir, "\\", "/")),
		fmt.Sprintf(`set lastError = $get(%%objlasterror)`),
		fmt.Sprintf(`if (lastError'="") { do $SYSTEM.OBJ.DisplayError(lastError) do ##class(%%SYSTEM.Process).Terminate(,1) }`),
		fmt.Sprintf(`do ##class(%%SYSTEM.Process).Terminate(,0)`),
	}...)

	command := strings.Join(commands, "\n")
	exitCode, err := session.ExecuteCommand(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing commands: %v\n", err)
		os.Exit(exitCode)
	}
	os.Exit(exitCode)
}
