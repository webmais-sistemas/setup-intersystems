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
		outputPath = flag.String("output-path", "", "Output path for test results")
		generateReport = flag.Bool("generate-report", true, "Generate test report")
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

	// Run tests	
	commands := []string{
		fmt.Sprintf(`do:(##class(%%SYS.Namespace).Exists("%s")'=1) ##class(%%SYSTEM.Process).Terminate(,1)`, *namespace),		
    
		fmt.Sprintf(`set $namespace = "%s"`, *namespace),
		fmt.Sprintf(`do ##class(%%UnitTest.Manager).RunTest("","/nodelete")`),
		fmt.Sprintf(`do ##class(%%SYSTEM.Process).Terminate(,0)`),
	}	
	
	command := strings.Join(commands, "\n")	
	exitCode, err := session.ExecuteCommand(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing commands: %v\n", err)
		os.Exit(exitCode)
	}
	os.Exit(exitCode)
	
	if *generateReport && *outputPath != "" {		
		if err := os.MkdirAll(filepath.Dir(*outputPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		// Generate JUnit XML report
		commands := []string{			
			fmt.Sprintf(`set $namespace = "%s"`, *namespace),
			fmt.Sprintf(`do ##class(%%File).CreateDirectoryChain("%s")`, *outputPath),
			fmt.Sprintf(`set File=##class(%%File).%%New("%s/test-report.xml")`, *outputPath),
			fmt.Sprintf(`set i=$order(^UnitTest.Result(""),-1)`),
			fmt.Sprintf(`if i="" do ##class(%%SYSTEM.Process).Terminate(,0)`),
			fmt.Sprintf(`kill ^||TMP`),
			fmt.Sprintf(`set suite="" for {
				set suite=$order(^UnitTest.Result(i,suite))
				quit:suite=""
				set ^||TMP("S",suite,"time")=$listget(^UnitTest.Result(i,suite),2)
				set ^||TMP("result")=$listget(^UnitTest.Result(i,suite),1)
				set case="" for {
					set case=$order(^UnitTest.Result(i,suite,case))
					quit:case=""
					if $increment(^||TMP("S",suite,"tests"))
					set ^||TMP("S",suite,"C",case,"time")=$listget(^UnitTest.Result(i,suite),2)
					set method="" for {
						set method=$order(^UnitTest.Result(i,suite,case,method))
						quit:method=""
						if $increment(^||TMP("S",suite,"C",case,"tests"))
						set ^||TMP("S",suite,"C",case,"M",method,"time")=$listget(^UnitTest.Result(i,suite,case,method),2)
						set assert="" for {
							set assert=$order(^UnitTest.Result(i,suite,case,method,assert))
							quit:assert=""
							if $increment(^||TMP("S",suite,"assertions"))
							if $increment(^||TMP("S",suite,"C",case,"assertions"))
							if $increment(^||TMP("S",suite,"C",case,"M",method,"assertions"))
							if $listget(^UnitTest.Result(i,suite,case,method,assert))=0 {
								if $increment(^||TMP("S",suite,"failures"))
								if $increment(^||TMP("S",suite,"C",case,"failures"))
								if $increment(^||TMP("S",suite,"C",case,"M",method,"failures"))
								set ^||TMP("S",suite,"C",case,"M",method,"failure")=$get(^||TMP("S",suite,"C",case,"M",method,"failure"))
									_$listget(^UnitTest.Result(i,suite,case,method,assert),2)
									_": "_$listget(^UnitTest.Result(i,suite,case,method,assert),3)
									_$char(13,10)
							}
						}
					}
				}
			}

			do File.Open("WSN")
			do File.WriteLine("<?xml version=""1.0"" encoding=""UTF-8"" ?>")
			do File.WriteLine("<testsuites>")
			set suite="" for {
				set suite=$order(^||TMP("S",suite))
				quit:suite=""
				do File.Write("<testsuite")
				do File.Write(" name="""_$zstrip($zconvert($zconvert(suite,"O","XML"),"O","UTF8"),"<=>W")_"""")
				do File.Write(" assertions="""_$get(^||TMP("S",suite,"assertions"))_"""")
				do File.Write(" time="""_$get(^||TMP("S",suite,"time"))_"""")
				do File.Write(" tests="""_$get(^||TMP("S",suite,"tests"))_"""")
				do File.WriteLine(">")
				set case="" for {
					set case=$order(^||TMP("S",suite,"C",case))
					quit:case=""
					do File.Write("<testsuite")
					do File.Write(" name="""_$zstrip($zconvert($zconvert(case,"O","XML"),"O","UTF8"),"<=>W")_"""")
					do File.Write(" assertions="""_$get(^||TMP("S",suite,"C",case,"assertions"))_"""")
					do File.Write(" time="""_$get(^||TMP("S",suite,"C",case,"time"))_"""")
					do File.Write(" tests="""_$get(^||TMP("S",suite,"C",case,"tests"))_"""")
					do File.WriteLine(">")
					set method="" for {
						set method=$order(^||TMP("S",suite,"C",case,"M",method))
						quit:method=""
						do File.Write("<testcase")
						do File.Write(" classname="""_$zstrip($zconvert($zconvert(case,"O","XML"),"O","UTF8"),"<=>W")_"""")
						do File.Write(" name="""_$zstrip($zconvert($zconvert(method,"O","XML"),"O","UTF8"),"<=>W")_"""")
						do File.Write(" assertions="""_$get(^||TMP("S",suite,"C",case,"M",method,"assertions"))_"""")
						do File.Write(" time="""_$get(^||TMP("S",suite,"C",case,"M",method,"time"))_"""")
						do File.WriteLine(">")
						if $data(^||TMP("S",suite,"C",case,"M",method,"failure")) {
							do File.Write("<failure type=""cache-error"" message=""Cache Error"">")
							do File.Write($zconvert($zconvert($zstrip($zstrip(^||TMP("S",suite,"C",case,"M",method,"failure"),"<=>W"),"*E'A'P'N'W"),"O","XML"),"O","UTF8"))
							do File.WriteLine("</failure>")
						}
						do File.WriteLine("</testcase>")
					}
					do File.WriteLine("</testsuite>")
				}
				do File.WriteLine("</testsuite>")
			}
			do File.WriteLine("</testsuites>")
			do File.Close()

			set result = $get(^||TMP("result"),1)
			kill ^||TMP
			if (result=0) { do ##class(%%SYSTEM.Process).Terminate(,1) }
			do ##class(%%SYSTEM.Process).Terminate(,0)`),
		}
		
		command := strings.Join(commands, "\n")	
		exitCode, err := session.ExecuteCommand(command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating test report: %v\n", err)
			os.Exit(exitCode)
		}
		os.Exit(exitCode)
	}
}
