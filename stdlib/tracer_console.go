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
				fmt.Printf("%s %s call -> %s(%#v)\n",
					evt.Time.Truncate(time.Microsecond),
					evt.Pid, evt.Tag, evt.Arg)

			case *CallResult:
				fmt.Printf("%s %s call <- %s(%#v)=%#v, %s\n",
					evt.Time.Truncate(time.Microsecond),
					evt.Pid, evt.Tag, evt.Arg, evt.Result, evt.Duration)
			}
		}
	}
}
