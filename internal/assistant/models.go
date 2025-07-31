package assistant

type Task string

var (
	TaskCodeGeneration Task = "code_generation"
	TaskCodeReview     Task = "code_review"
	TaskCodeCompletion Task = "code_completion"
)
