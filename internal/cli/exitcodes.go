package cli

const (
	ExitSuccess              = 0
	ExitUnknownError         = 1
	ExitInvalidArguments     = 2
	ExitMissingDependency    = 3
	ExitConnectorError       = 4
	ExitDataValidationFailed = 5
	ExitPipelineFailed       = 6
	ExitLLMError             = 7
)
