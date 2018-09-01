package stdlib

//
// Console tracer
//

import (
	"fmt"
	"time"
)

//
// TraceToConsole prints messages to console
//
func TraceToConsole() TracerFunc {
	return func(events ...Term) {
		for _, evt := range events {
			switch evt := evt.(type) {

			case *Call:
				fmt.Printf("%s %s %s %s %#v\n",
					evt.Time.Truncate(time.Microsecond),
					evt.Pid, "call:", evt.Tag, evt.Arg)

			case *CallResult:
				fmt.Println(evt.Time.Truncate(time.Microsecond),
					evt.Pid, "call result:", evt.Tag, evt.Arg,
					evt.Result, evt.Duration)
			}
		}
	}
}
