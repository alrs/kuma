package cmd

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Universal In-Memory test", func() {
	RunSmokeTest(`
xdsServer:
  grpcPort: 0
  httpPort: 0
  diagnosticsPort: %d
apiServer:
  port: 0
environment: universal
store:
  type: memory
`)
})