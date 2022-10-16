go build -o %~dp0..\bin\api.exe %~dp0..\src\apiserver\main.go
go build -o %~dp0..\bin\meta.exe %~dp0..\src\metaserver\main.go
go build -o %~dp0..\bin\object.exe %~dp0..\src\objectserver\main.go
cd %~dp0..\src\adminserver\ui && yarn build && go build -o %~dp0..\bin\admin.exe %~dp0..\src\adminserver\main.go