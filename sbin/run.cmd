@echo off
%~dp0..\bin\%1.exe app %2

:end
echo "run.cmd [api/meta/object/admin] [xxx.yaml]"