Set FontSize 20
Set Width 2000
Set Height 1000
Set BorderRadius 10
Set TypingSpeed 100ms

Output ./assets/cassette.gif

Hide
  Type@10ms "go build -o build/kontrolplane/kue . && clear && ./build/kontrolplane/kue"
  Enter
  Wait+Screen@60s /kontrolplane-users/
  Sleep 1s
Show

Sleep 6s

Screenshot ./assets/pages/queue/overview.png

Sleep 4s
Down
Sleep 2s
Enter
Sleep 5s
Down@800ms 2

Sleep 5s

Screenshot ./assets/pages/queue/details.png

Sleep 4s
Enter

Sleep 6s

Screenshot ./assets/pages/message/details.png

Sleep 4s
Type "q"

Sleep 4s
Ctrl+N
Sleep 4s
Type '{"userId": "20319", "action": "create", "timestamp": "2025-01-01T12:00:00Z"}'
Sleep 5s
Tab@1s 2

Sleep 2s

Screenshot ./assets/pages/message/creation.png

Sleep 2s
Enter
Sleep 5s

Ctrl+D
Sleep 4s
Tab

Sleep 2s

Screenshot ./assets/pages/message/delete.png

Sleep 2s
Enter
Sleep 4s

Type "q"

Sleep 4s
Ctrl+N
Sleep 4s
Type "kontrolplane-demonstration"
Enter
Sleep 3s
Down
Sleep 3s

Screenshot ./assets/pages/queue/creation-01.png

Sleep 4s
Enter
Sleep 2s
Type "60"
Sleep 3s

Screenshot ./assets/pages/queue/creation-02.png

Sleep 3s
Enter@1s 3
Sleep 3s

Screenshot ./assets/pages/queue/creation-03.png

Sleep 3s
Enter@1s 2
Sleep 2s
Right
Sleep 3s

Screenshot ./assets/pages/queue/creation-04.png

Sleep 3s
Enter@1s 3
Sleep 6s

Type "/"
Sleep 2s
Type "demonstration"
Sleep 2s
Enter
Sleep 3s
Ctrl+D
Sleep 3s
Type "kontrolplane-demonstration.fifo"
Sleep 3s

Screenshot ./assets/pages/queue/delete.png

Sleep 2s
Enter
Sleep 5s
Escape
Sleep 6s
