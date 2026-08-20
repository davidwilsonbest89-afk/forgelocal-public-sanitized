package api

import (
	"net/http"

	"forgelocal/internal/workflow"
)

const WorkflowExecutionDisabledCode = "WORKFLOW_EXECUTION_DISABLED"

func WorkflowHandler(_ *workflow.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CR-02: workflows remain intentionally unavailable until a separately
		// authorized, loopback-only execution contract exists. Refuse before
		// body decoding so an unauthorised payload cannot trigger parsing work.
		writeError(w, http.StatusGone, WorkflowExecutionDisabledCode, "workflow execution is disabled by T27-R1 policy")
	}
}
