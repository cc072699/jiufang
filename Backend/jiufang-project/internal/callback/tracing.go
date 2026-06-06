// Package callback provides Eino callbacks for tracing and logging.
// These callbacks are used to track LLM calls, tool executions, and agent operations.
package callback

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// TracingCallback implements callbacks.Handler for tracing LLM operations.
type TracingCallback struct {
	logger  *zap.Logger
	handler callbacks.Handler
}

// NewTracingCallback creates a new tracing callback.
func NewTracingCallback(logger *zap.Logger) *TracingCallback {
	if logger == nil {
		logger = zap.NewNop()
	}

	tc := &TracingCallback{
		logger: logger,
	}

	// Build handler using NewHandlerBuilder
	tc.handler = callbacks.NewHandlerBuilder().
		OnStartFn(tc.onStart).
		OnEndFn(tc.onEnd).
		OnErrorFn(tc.onError).
		OnStartWithStreamInputFn(tc.onStartWithStreamInput).
		OnEndWithStreamOutputFn(tc.onEndWithStreamOutput).
		Build()

	return tc
}

// GetHandler returns the underlying callbacks.Handler.
func (c *TracingCallback) GetHandler() callbacks.Handler {
	return c.handler
}

// onStart is called when an operation starts.
func (c *TracingCallback) onStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if info == nil {
		c.logger.Warn("RunInfo is nil in OnStart")
		return ctx
	}

	c.logger.Info("Operation started",
		zap.String("name", info.Name),
		zap.String("type", info.Type),
		zap.String("component", string(info.Component)),
		zap.Time("start_time", time.Now()),
	)

	// Store start time in context
	return context.WithValue(ctx, "start_time", time.Now())
}

// onEnd is called when an operation ends successfully.
func (c *TracingCallback) onEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if info == nil {
		c.logger.Warn("RunInfo is nil in OnEnd")
		return ctx
	}

	// Calculate duration
	startTime, ok := ctx.Value("start_time").(time.Time)
	duration := time.Duration(0)
	if ok {
		duration = time.Since(startTime)
	}

	c.logger.Info("Operation completed",
		zap.String("name", info.Name),
		zap.String("type", info.Type),
		zap.String("component", string(info.Component)),
		zap.Duration("duration", duration),
		zap.Any("output", output),
	)

	return ctx
}

// onError is called when an operation fails.
func (c *TracingCallback) onError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	if info == nil {
		c.logger.Warn("RunInfo is nil in OnError")
		return ctx
	}

	// Calculate duration
	startTime, ok := ctx.Value("start_time").(time.Time)
	duration := time.Duration(0)
	if ok {
		duration = time.Since(startTime)
	}

	c.logger.Error("Operation failed",
		zap.String("name", info.Name),
		zap.String("type", info.Type),
		zap.String("component", string(info.Component)),
		zap.Duration("duration", duration),
		zap.Error(err),
	)

	return ctx
}

// onStartWithStreamInput is called when an operation starts with stream input.
func (c *TracingCallback) onStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	if info == nil {
		c.logger.Warn("RunInfo is nil in OnStartWithStreamInput")
		return ctx
	}

	defer input.Close()

	c.logger.Info("Operation started with stream input",
		zap.String("name", info.Name),
		zap.String("type", info.Type),
		zap.String("component", string(info.Component)),
		zap.Time("start_time", time.Now()),
	)

	// Store start time in context
	return context.WithValue(ctx, "start_time", time.Now())
}

// onEndWithStreamOutput is called when an operation ends with stream output.
func (c *TracingCallback) onEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	if info == nil {
		c.logger.Warn("RunInfo is nil in OnEndWithStreamOutput")
		return ctx
	}

	defer output.Close()

	// Calculate duration
	startTime, ok := ctx.Value("start_time").(time.Time)
	duration := time.Duration(0)
	if ok {
		duration = time.Since(startTime)
	}

	c.logger.Info("Operation completed with stream output",
		zap.String("name", info.Name),
		zap.String("type", info.Type),
		zap.String("component", string(info.Component)),
		zap.Duration("duration", duration),
	)

	return ctx
}

// RegisterTracingCallback registers the tracing callback with Eino.
func RegisterTracingCallback(logger *zap.Logger) {
	callback := NewTracingCallback(logger)
	callbacks.AppendGlobalHandlers(callback.GetHandler())
}

// GetRunInfo extracts run info from context.
func GetRunInfo(ctx context.Context) *callbacks.RunInfo {
	// RunInfo might not be available in context, return default
	return &callbacks.RunInfo{
		Name:      "unknown",
		Type:      "unknown",
		Component: components.ComponentOfChatModel, // Use a valid component type
	}
}

// LogWithContext logs a message with context information.
func LogWithContext(ctx context.Context, logger *zap.Logger, level string, msg string, fields ...zap.Field) {
	info := GetRunInfo(ctx)

	// Add run info to fields
	allFields := append(fields,
		zap.String("name", info.Name),
		zap.String("type", info.Type),
		zap.String("component", string(info.Component)),
	)

	switch level {
	case "debug":
		logger.Debug(msg, allFields...)
	case "info":
		logger.Info(msg, allFields...)
	case "warn":
		logger.Warn(msg, allFields...)
	case "error":
		logger.Error(msg, allFields...)
	default:
		logger.Info(msg, allFields...)
	}
}

// FormatDuration formats a duration for logging.
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%d ms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
	return fmt.Sprintf("%.2f m", d.Minutes())
}
