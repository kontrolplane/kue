package tui

// MaxMessageBytes defines maximum allowed size (256 KB) for SQS message body.
const MaxMessageBytes = 256 * 1024 // 262144 bytes

// BytesRemaining returns the number of bytes remaining before the message
// reaches MaxMessageBytes. It uses the UTF-8 byte length of the provided
// string. A negative return value means the limit has been exceeded.
func BytesRemaining(body string) int {
	return MaxMessageBytes - len([]byte(body))
}
