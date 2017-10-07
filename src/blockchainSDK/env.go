// +build !testpkcs11

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import "os"

var (
	// ConfigTestFile contains the path and filename of the config for integration tests
	ConfigTestFile = os.Getenv("GOPATH") + "/src/github.com/wadelee1986/guessNumByExample/src/blockchainSDK/config/config_test.yaml"
)
