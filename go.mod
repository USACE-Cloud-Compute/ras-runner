module ras-runner

go 1.18

//replace github.com/usace/hdf5utils => /workspaces/hdf5utils

require (
	github.com/usace/cc-go-sdk v0.0.0-20230613194210-1928678a0098
	github.com/usace/go-hdf5 v0.0.0-20230626152743-72d0ae21fd0c
	github.com/usace/hdf5utils v0.0.0-20230731192430-e48a18481e69
)

require (
	github.com/aws/aws-sdk-go v1.44.189 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/usace/filestore v0.0.0-20230309205740-49d6e1f06e4a // indirect
)
