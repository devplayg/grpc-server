@echo off

go build -o receiver.exe ..\receiver\main.go
go build -o classifier.exe ..\classifier\main.go
go build -o calculator.exe ..\calculator\main.go
go build -o generator.exe ..\generator\main.go
go build -o notifier.exe ..\notifier\main.go