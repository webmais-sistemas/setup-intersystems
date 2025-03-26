execute_process() {
if [ -x "$(command -v csession)" ]; then
  ISC_SESSION="csession CACHE"
elif [ -x "$(command -v irissession)" ]; then
  ISC_SESSION="irissession IRIS"
else
  echo "Error: irissession or csession not found in PATH" >&2
  exit 1
fi

$ISC_SESSION <<-EOF
 set \$namespace = "%SYS"
 do:(##class(%SYS.Namespace).Exists("$name")=1) ##class(%SYSTEM.Process).Terminate(,1)

 set path = ##class(%SYSTEM.Util).ManagerDirectory()_"$name"
 do ##class(%File).CreateDirectory(path)
 do ##class(SYS.Database).CreateDatabase(path,,,,"%DB_${resource^^}",2)

 set database("Directory") = path
 do ##class(Config.Databases).Create("$name",.database)

 set properties("Globals")="$name"
 do ##class(Config.Namespaces).Create("$name",.properties)

 do:(##class(Config.Namespaces).Exists("$name")'=1) ##class(%SYSTEM.Process).Terminate(,1)
 kill
 set \$namespace = "$name"
 set workcls = "$workdir/src"
 do ##class(%SYSTEM.OBJ).ImportDir(workcls,"*.inc","c","",1)
 do ##class(%SYSTEM.OBJ).ImportDir(workcls,"*.cls","c","",1)

 set ^UnitTestRoot = "$workdir/test"
 set lastError = \$get(%objlasterror)
 if (lastError'="") { do \$SYSTEM.OBJ.DisplayError(lastError) do ##class(%SYSTEM.Process).Terminate(,1) }

 do ##class(%SYSTEM.Process).Terminate(,0)
EOF
}