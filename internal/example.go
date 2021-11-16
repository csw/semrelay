package internal

var ExampleSuccess = []byte(`{
	"blocks": [
	  {
		"jobs": [
		  {
			"id": "86f955ab-323a-4d48-9437-c688dcb710eb",
			"index": 0,
			"name": "Test",
			"result": "passed",
			"status": "finished"
		  }
		],
		"name": "Test",
		"result": "passed",
		"result_reason": "test",
		"state": "done"
	  },
	  {
		"jobs": [
		  {
			"id": "864934ad-3b54-4305-bb4e-678d26619375",
			"index": 0,
			"name": "Lint",
			"result": "passed",
			"status": "finished"
		  }
		],
		"name": "Lint",
		"result": "passed",
		"result_reason": "test",
		"state": "done"
	  }
	],
	"organization": {
	  "id": "f0318086-1f86-4271-9020-7eb5b338cf44",
	  "name": "genomenon"
	},
	"pipeline": {
	  "created_at": "2021-11-15T05:16:11Z",
	  "done_at": "2021-11-15T05:20:04Z",
	  "error_description": "",
	  "id": "d57b4188-9c6f-4dae-9043-ceca7e372970",
	  "name": "mm-data-go",
	  "pending_at": "2021-11-15T05:16:11Z",
	  "queuing_at": "2021-11-15T05:16:12Z",
	  "result": "passed",
	  "result_reason": "test",
	  "running_at": "2021-11-15T05:16:12Z",
	  "state": "done",
	  "stopping_at": "1970-01-01T00:00:00Z",
	  "working_directory": ".semaphore",
	  "yaml_file_name": "semaphore.yml"
	},
	"project": {
	  "id": "78a69d7b-1ad6-4d1d-8fb8-66f929509574",
	  "name": "myproject"
	},
	"repository": {
	  "slug": "example/myproject",
	  "url": "https://github.com/example/myproject"
	},
	"revision": {
	  "branch": {
		"commit_range": "d2758e435aa592c3b3a23b4173571188c07c74e3...e97080ec66d19282aafae5ddec5f5b51314c4ed8",
		"name": "notify_test"
	  },
	  "commit_message": "more nothing",
	  "commit_sha": "e97080ec66d19282aafae5ddec5f5b51314c4ed8",
	  "pull_request": null,
	  "reference": "refs/heads/notify_test",
	  "reference_type": "branch",
	  "sender": {
		"avatar_url": "https://avatars.githubusercontent.com/u/587788?v=4",
		"email": "cswheeler@gmail.com",
		"login": "csw"
	  },
	  "tag": null
	},
	"version": "1.0.0",
	"workflow": {
	  "created_at": "2021-11-15T05:16:11Z",
	  "id": "af05f7b8-fd86-4e9f-87b6-cc5820aa5fea",
	  "initial_pipeline_id": "d57b4188-9c6f-4dae-9043-ceca7e372970"
	}
  }`)

var ExampleFailure = []byte(`{
	"blocks": [
	  {
		"jobs": [
		  {
			"id": "2b1c3afa-efef-44d1-9583-280ccca88840",
			"index": 0,
			"name": "Run tests",
			"result": "failed",
			"status": "finished"
		  }
		],
		"name": "Test",
		"result": "failed",
		"result_reason": "test",
		"state": "done"
	  }
	],
	"organization": {
	  "id": "f0318086-1f86-4271-9020-7eb5b338cf44",
	  "name": "genomenon"
	},
	"pipeline": {
	  "created_at": "2021-11-15T05:20:05Z",
	  "done_at": "2021-11-15T05:22:18Z",
	  "error_description": "",
	  "id": "d57b4188-9c6f-4dae-9043-ceca7e372970",
	  "name": "Test",
	  "pending_at": "2021-11-15T05:20:06Z",
	  "queuing_at": "2021-11-15T05:20:06Z",
	  "result": "failed",
	  "result_reason": "test",
	  "running_at": "2021-11-15T05:20:07Z",
	  "state": "done",
	  "stopping_at": "1970-01-01T00:00:00Z",
	  "working_directory": ".semaphore",
	  "yaml_file_name": "semaphore.yml"
	},
	"project": {
	  "id": "78a69d7b-1ad6-4d1d-8fb8-66f929509574",
	  "name": "myproject"
	},
	"repository": {
	  "slug": "example/myproject",
	  "url": "https://github.com/example/myproject"
	},
	"revision": {
	  "branch": {
		"commit_range": "d2758e435aa592c3b3a23b4173571188c07c74e3...e97080ec66d19282aafae5ddec5f5b51314c4ed8",
		"name": "notify_test"
	  },
	  "commit_message": "more nothing",
	  "commit_sha": "e97080ec66d19282aafae5ddec5f5b51314c4ed8",
	  "pull_request": null,
	  "reference": "refs/heads/notify_test",
	  "reference_type": "branch",
	  "sender": {
		"avatar_url": "https://avatars.githubusercontent.com/u/587788?v=4",
		"email": "cswheeler@gmail.com",
		"login": "csw"
	  },
	  "tag": null
	},
	"version": "1.0.0",
	"workflow": {
	  "created_at": "2021-11-15T05:16:11Z",
	  "id": "af05f7b8-fd86-4e9f-87b6-cc5820aa5fea",
	  "initial_pipeline_id": "d57b4188-9c6f-4dae-9043-ceca7e372970"
	}
  }`)
