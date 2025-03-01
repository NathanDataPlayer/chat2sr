package logs

import (
    "log"
    "os"
)

func Init() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetOutput(os.Stdout)
}