go build -o %~dp0..\bin\api.exe %~dp0..\apiserver\main.go
go build -o %~dp0..\bin\meta.exe %~dp0..\metaserver\main.go
go build -o %~dp0..\bin\object.exe %~dp0..\objectserver\main.go
go build -o %~dp0..\bin\admin.exe %~dp0..\adminserver\main.go