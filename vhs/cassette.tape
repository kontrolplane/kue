Set FontSize 20
Set Width 2400
Set Height 1000
Set BorderRadius 10
Set Margin 30
Set MarginFill "#536c54"
Set Framerate 30

Output ./assets/cassette.gif

Hide
  Type "go build -o build/kontrolplane/kue main.go & clear"
  Enter
Show

Sleep 2s
Copy "./build/kontrolplane/kue"
Paste
Sleep 5s
Enter
Sleep 5s

Screenshot ./assets/pages/queue/overview.png

Sleep 2s
Enter
Sleep 2s
Down
Sleep 1s
Down

Screenshot ./assets/pages/queue/details.png

Type "q"

Sleep 1s
Down
Sleep 1s
Down
Sleep 1s
Down
Sleep 1s
Down
Sleep 1s
Down
Sleep 1s
Ctrl+D
Sleep 2s
Right

Screenshot ./assets/pages/queue/delete.png

Sleep 2s
Enter
Sleep 1s

Sleep 30s
