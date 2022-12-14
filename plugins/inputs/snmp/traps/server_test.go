// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

// var freePort = getFreePort()

// func TestStartFailure(t *testing.T) {
// 	/*
// 		Start two servers with the same config to trigger an "address already in use" error.
// 	*/

// 	config := Config{Port: freePort, CommunityStrings: []string{"public"}}
// 	Configure(t, config)

// 	mockSender := mocksender.NewMockSender("snmp-traps-listener")
// 	mockSender.SetupAcceptAll()

// 	sucessServer, err := NewTrapServer(config, &DummyFormatter{}, mockSender)
// 	require.NoError(t, err)
// 	require.NotNil(t, sucessServer)
// 	defer sucessServer.Stop()

// 	failedServer, err := NewTrapServer(config, &DummyFormatter{}, mockSender)
// 	require.Nil(t, failedServer)
// 	require.Error(t, err)
// }
