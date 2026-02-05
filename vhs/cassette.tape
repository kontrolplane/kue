Set FontSize 20
Set Width 2000
Set Height 1000
Set BorderRadius 10

Output ./assets/cassette.gif

Hide
  Type "go build -o build/kontrolplane/kue main.go & clear"
  Enter
  Copy "./build/kontrolplane/kue"
  Paste
  Enter
Show

Sleep 10s

Screenshot ./assets/pages/queue/overview.png

Sleep 2s
Enter
Sleep 3s
Down
Sleep 5s
Down

Sleep 8s

Screenshot ./assets/pages/queue/details.png

Sleep 8s
Enter

Sleep 8s

Screenshot ./assets/pages/message/details.png

Sleep 2s

Type "q"

Sleep 3s
Ctrl+N
Sleep 4s
Type '{"userId": "20319", "action": "create", "timestamp": "2025-01-01T12:00:00Z"}'
Sleep 8s
Tab
Sleep 2s
Tab

Sleep 2s

Screenshot ./assets/pages/message/creation.png

Enter
Sleep 2s

Ctrl+D
Sleep 4s
Tab

Sleep 2s

Screenshot ./assets/pages/message/delete.png

Sleep 2s
Enter

Type "q"

Sleep 4s
Ctrl+N
Sleep 4s
Type "kontrolplane-demonstration"
Enter
Sleep 3s

Screenshot ./assets/pages/queue/creation-01.png

Sleep 2s
Enter
Sleep 5s
Enter
Sleep 3s

Screenshot ./assets/pages/queue/creation-02.png

Sleep 2s
Enter
Sleep 2s

Screenshot ./assets/pages/queue/creation-03.png

Sleep 2s
Enter
Sleep 2s
Enter
Sleep 4s
Enter
Sleep 2s

Screenshot ./assets/pages/queue/creation-04.png

Sleep 3s
Enter
Sleep 2s
Enter
Sleep 2s
Enter

Sleep 1s
Down
Sleep 1s
Ctrl+D
Sleep 2s
Right
Sleep 2s

Screenshot ./assets/pages/queue/delete.png

Sleep 2s
Enter
Sleep 1s

Sleep 15s
