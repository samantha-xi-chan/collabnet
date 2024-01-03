package api_workflow

const (
	ExitCodeWorkflowDefault             = 0
	ExitCodeWorkflowStoppedByDagEnd     = 50043001
	ExitCodeWorkflowStoppedByBizTimeout = 50043002
	ExitCodeWorkflowStoppedByBizCmd     = 50043003
	ExitCodeWorkflowStoppedByUnknown    = 50043098
)
