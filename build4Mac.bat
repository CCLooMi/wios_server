@echo off
echo=

echo copy config
if not exist %cd%\dist\conf\ md %cd%\dist\conf\
xcopy /Y /E %cd%\conf\config.yaml %cd%\dist\conf\

echo copy public resource
if not exist %cd%\dist\static\public\ md %cd%\dist\static\public\
xcopy /Y /E %cd%\static\public %cd%\dist\static\public\

echo build
set GOOS=darwin
set GOARCH=amd64
go build -o %cd%\dist\wios_server_mac .\main

echo done
