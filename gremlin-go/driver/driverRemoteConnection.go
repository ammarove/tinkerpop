/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package gremlingo

import (
	"crypto/tls"
	"golang.org/x/text/language"
	"time"
)

// DriverRemoteConnectionSettings are used to configure the DriverRemoteConnection.
type DriverRemoteConnectionSettings struct {
	TraversalSource   string
	TransporterType   TransporterType
	LogVerbosity      LogVerbosity
	Logger            Logger
	Language          language.Tag
	AuthInfo          *AuthInfo
	TlsConfig         *tls.Config
	KeepAliveInterval time.Duration
	WriteDeadline     time.Duration

	// TODO: Figure out exact extent of configurability for these and expose appropriate types/helpers
	Protocol   protocol
	Serializer serializer
}

// DriverRemoteConnection is a remote connection.
type DriverRemoteConnection struct {
	client *Client
}

// NewDriverRemoteConnection creates a new DriverRemoteConnection.
// If no custom connection settings are passed in, a connection will be created with "g" as the default TraversalSource,
// Gorilla as the default Transporter, Info as the default LogVerbosity, a default logger stuct, and English and as the
// default language
func NewDriverRemoteConnection(
	url string,
	configurations ...func(settings *DriverRemoteConnectionSettings)) (*DriverRemoteConnection, error) {
	settings := &DriverRemoteConnectionSettings{
		TraversalSource:   "g",
		TransporterType:   Gorilla,
		LogVerbosity:      Info,
		Logger:            &defaultLogger{},
		Language:          language.English,
		AuthInfo:          &AuthInfo{},
		TlsConfig:         &tls.Config{},
		KeepAliveInterval: keepAliveIntervalDefault,
		WriteDeadline:     writeDeadlineDefault,

		// TODO: Figure out exact extent of configurability for these and expose appropriate types/helpers
		Protocol:   nil,
		Serializer: nil,
	}
	for _, configuration := range configurations {
		configuration(settings)
	}

	logHandler := newLogHandler(settings.Logger, settings.LogVerbosity, settings.Language)
	connection, err := createConnection(url, settings.AuthInfo, settings.TlsConfig, logHandler, settings.KeepAliveInterval, settings.WriteDeadline)
	if err != nil {
		if err != nil {
			logHandler.logf(Error, logErrorGeneric, "NewDriverRemoteConnection", err.Error())
		}
		return nil, err
	}

	client := &Client{
		url:             url,
		traversalSource: settings.TraversalSource,
		transporterType: settings.TransporterType,
		logHandler:      logHandler,
		connection:      connection,
	}

	return &DriverRemoteConnection{client: client}, nil
}

// Close closes the DriverRemoteConnection.
func (driver *DriverRemoteConnection) Close() error {
	err := driver.client.Close()
	if err != nil {
		driver.client.logHandler.logf(Error, logErrorGeneric, "Driver.Close()", err.Error())
	}
	return err
}

// Submit sends a string traversal to the server.
func (driver *DriverRemoteConnection) Submit(traversalString string) (ResultSet, error) {
	result, err := driver.client.Submit(traversalString)
	if err != nil {
		driver.client.logHandler.logf(Error, logErrorGeneric, "Driver.Submit()", err.Error())
	}
	return result, err
}

// submitBytecode sends a bytecode traversal to the server.
func (driver *DriverRemoteConnection) submitBytecode(bytecode *bytecode) (ResultSet, error) {
	return driver.client.submitBytecode(bytecode)
}

// TODO: Bytecode, OptionsStrategy, RequestOptions
//func extractRequestOptions(bytecode Bytecode) RequestOptions {
//	var optionsStrategy OptionsStrategy = nil
//	for _, instruction := range bytecode.sourceInstructions {
//		if instruction[0] == "withStrategies" {
//			_, isOptionsStrategy := instruction[1].(OptionsStrategy)
//			if isOptionsStrategy {
//				optionsStrategy = instruction
//				break
//			}
//		}
//	}
//
//	var requestOptions RequestOptions = nil
//	if optionsStrategy != nil {
//		allowedKeys := []string{'evaluationTimeout', 'scriptEvaluationTimeout', 'batchSize', 'requestId', 'userAgent'}
//		requestOptions := make(map[string]string)
//		for _, allowedKey := range allowedKeys {
//			if isAllowedKeyInConfigs(allowedKey, optionsStrategy[1].configuration) {
//				requestOptions[allowedKey] = optionsStrategy[1].configuration[allowedKey]
//			}
//		}
//	}
//	return requestOptions
//}

//func isAllowedKeyInConfigs(allowedKey string, configs []string) bool {
//	for _, config := range configs {
//		if allowedKey == config {
//			return true
//		}
//	}
//	return false
//}
