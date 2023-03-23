@echo off
IF [%1]==[] echo "build.cmd [api/meta/object/admin/all]"
IF [%1]==[api] go build -o %~dp0..\bin\api.exe %~dp0..\src\apiserver\main.go
IF [%1]==[meta] go build -o %~dp0..\bin\meta.exe %~dp0..\src\metaserver\main.go
IF [%1]==[object] go build -o %~dp0..\bin\object.exe %~dp0..\src\objectserver\main.go
IF [%1]==[admin] cd %~dp0..\src\adminserver\ui && yarn build && go build -o %~dp0..\bin\admin.exe %~dp0..\src\adminserver\main.go
IF [%1]==[all] (
    go build -o %~dp0..\bin\api.exe %~dp0..\src\apiserver\main.go
    go build -o %~dp0..\bin\meta.exe %~dp0..\src\metaserver\main.go
    go build -o %~dp0..\bin\object.exe %~dp0..\src\objectserver\main.go
    cd %~dp0..\src\adminserver\ui && yarn build && go build -o %~dp0..\bin\admin.exe %~dp0..\src\adminserver\main.go
)
IF [%1]==[core] (
    go build -o %~dp0..\bin\api.exe %~dp0..\src\apiserver\main.go
    go build -o %~dp0..\bin\meta.exe %~dp0..\src\metaserver\main.go
    go build -o %~dp0..\bin\object.exe %~dp0..\src\objectserver\main.go
)