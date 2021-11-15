package semrelay

// Notification is the JSON object sent by Semaphore describing the build
// results. See https://docs.semaphoreci.com/essentials/webhook-notifications/.
// Note that not all its fields are mapped.
type Notification struct {
	Version string `json:"version"`
	Project struct {
		Name string `json:"name"`
	} `json:"project"`
	Organization struct {
		Name string `json:"name"`
	}
	Repository struct {
		Url  string `json:"url"`
		Slug string `json:"slug"`
	} `json:"repository"`
	Revision struct {
		Tag    string `json:"tag"`
		Sender struct {
			Login string `json:"login"`
		} `json:"sender"`
		ReferenceType string `json:"reference_type"`
		Reference     string `json:"reference"`
		PullRequest   string `json:"pull_request"`
		CommitSHA     string `json:"commit_sha"`
		CommitMessage string `json:"commit_message"`
		Branch        struct {
			Name        string `json:"name"`
			CommitRange string `json:"commit_range"`
		} `json:"branch"`
	} `json:"revision"`
	Pipeline struct {
		Id           string `json:"id"`
		State        string `json:"state"`
		RunningAt    string `json:"running_at"`
		DoneAt       string `json:"done_at"`
		ResultReason string `json:"result_reason"`
		Result       string `json:"result"`
		YamlFileName string `json:"yaml_file_name"`
	} `json:"pipeline"`
	Workflow struct {
		Id                string `json:"id"`
		CreatedAt         string `json:"created_at"`
		InitialPipelineId string `json:"initial_pipeline_id"`
	} `json:"workflow"`
	Blocks []*struct {
		Name         string `json:"name"`
		Result       string `json:"result"`
		ResultReason string `json:"result_reason"`
		State        string `json:"state"`
		Jobs         []*struct {
			Id     string `json:"id"`
			Index  int    `json:"index"`
			Name   string `json:"name"`
			Result string `json:"result"`
			Status string `json:"status"`
		} `json:"jobs"`
	} `json:"blocks"`
}
