//go:build go1.19

package sing

import (
	"runtime/debug"

	"github.com/inazumav/sing-box/common/dialer/conntrack"
	"github.com/inazumav/sing-box/option"
)

func applyDebugOptions(options option.DebugOptions) {
	applyDebugListenOption(options)
	if options.GCPercent != nil {
		debug.SetGCPercent(*options.GCPercent)
	}
	if options.MaxStack != nil {
		debug.SetMaxStack(*options.MaxStack)
	}
	if options.MaxThreads != nil {
		debug.SetMaxThreads(*options.MaxThreads)
	}
	if options.PanicOnFault != nil {
		debug.SetPanicOnFault(*options.PanicOnFault)
	}
	if options.TraceBack != "" {
		debug.SetTraceback(options.TraceBack)
	}
	if options.MemoryLimit != 0 {
		debug.SetMemoryLimit(int64(options.MemoryLimit))
		conntrack.MemoryLimit = int64(options.MemoryLimit)
	}
	if options.OOMKiller != nil {
		conntrack.KillerEnabled = *options.OOMKiller
	}
}
