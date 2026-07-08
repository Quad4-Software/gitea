// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package reticulum

// DefaultNetworkConfig is the initial Reticulum network configuration seeded on install.
const DefaultNetworkConfig = `# Reticulum network configuration.
# Edit interfaces here. See https://reticulum.network/manual/interfaces.html

[reticulum]
  share_instance = Yes

[logging]
  loglevel = 4

[interfaces]

  [[Default Interface]]
    type = AutoInterface
    enabled = Yes

  # Example TCP client to a public gateway:
  # [[Internet Gateway]]
  #   type = TCPClientInterface
  #   enabled = No
  #   target_host = amsterdam.connect.reticulum.network
  #   target_port = 7825

  # Example inbound TCP server:
  # [[TCP Server]]
  #   type = TCPServerInterface
  #   enabled = No
  #   listen_ip = 0.0.0.0
  #   listen_port = 4242
`
