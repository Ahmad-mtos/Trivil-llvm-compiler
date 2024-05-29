@echo off

if "x%1" == "x" echo Usage: install "target directory for installation"
if "x%1" == "x" exit
echo installing to %1 ...

cd ..

del /s /q *.bak

xcopy /q /y /e /k /i стд %1\стд
xcopy /q /y /e /k /i runtime %1\runtime
xcopy /q /y /k /i doc\report\*.pdf %1\doc
xcopy /q /y /k /i doc\*.xml %1\doc
xcopy /q /y /k /i config\*.* %1\config

cd src
echo строим tric компилятор (Go)...
go build trivil.go
if  not exist trivil.exe (
 	echo ошибка при выполнении go build
	exit
)

cd ..
move src\trivil.exe tric.exe
del /q trik.exe

rem tric -exe=false трик
echo строим трик компилятор...
tric трик
rem cd _genllvm
rem call build.bat
rem cd ..
    
if not exist trik.exe (
        echo ошибка при компиляции трик компилятора
        exit
)

move tric.exe %1\tric.exe
move trik.exe %1\трик.exe

echo success
