@echo off
echo Building...
go env -w GOOS=linux
go build -o bin/
go env -w GOOS=windows
echo OK