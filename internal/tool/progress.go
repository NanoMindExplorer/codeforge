package tool

// ProgressFunc reports incremental tool output to the agent loop / UI.
type ProgressFunc func(chunk string)

// StreamingExecutor is an optional interface tools can implement so the
// agent loop can emit EventToolProgress while the tool runs.
type StreamingExecutor interface {
	ExecuteStream(input []byte, progress ProgressFunc) Result
}
