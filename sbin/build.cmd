cd %~dp0..\src
go build -o %~dp0..\bin\api.exe apiserver\main.go
go build -o %~dp0..\bin\meta.exe metaserver\main.go
go build -o %~dp0..\bin\object.exe objectserver\main.go
go build -o %~dp0..\bin\admin.exe adminserver\main.go