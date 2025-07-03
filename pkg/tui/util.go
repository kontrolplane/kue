package tui

import "strconv"

func parseInt32(s string) int32 {
    if s == "" {
        return 0
    }
    i, _ := strconv.Atoi(s)
    return int32(i)
}
