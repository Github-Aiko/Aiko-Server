package rate

import (
	"github.com/AikoPanel/Xray-core/common"
	"github.com/AikoPanel/Xray-core/common/buf"
	"github.com/juju/ratelimit"
)

type Writer struct {
	writer  buf.Writer
	limiter *ratelimit.Bucket
}

func NewRateLimitWriter(writer buf.Writer, limiter *ratelimit.Bucket) buf.Writer {
	return &Writer{
		writer:  writer,
		limiter: limiter,
	}
}

func (w *Writer) Close() error {
	return common.Close(w.writer)
}

func (w *Writer) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.limiter.Wait(int64(mb.Len()))
	return w.writer.WriteMultiBuffer(mb)
}
