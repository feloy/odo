/*******************************************************************************
 * Copyright (c) 2022 Red Hat, Inc.
 * Distributed under license by Red Hat, Inc. All rights reserved.
 * This program is made available under the terms of the
 * Eclipse Public License v2.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 * Red Hat, Inc.
 ******************************************************************************/

package enricher

import (
	"context"
	"encoding/json"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type AngularCliJson struct {
	Defaults struct {
		Serve HostPort `json:"serve"`
	} `json:"defaults"`
}

type AngularJson struct {
	Projects map[string]ProjectBody `json:"projects"`
}

type ProjectBody struct {
	Architect struct {
		Serve struct {
			Options HostPort `json:"options"`
		} `json:"serve"`
	} `json:"architect"`
}

type HostPort struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type AngularDetector struct{}

func (a AngularDetector) GetSupportedFrameworks() []string {
	return []string{"Angular"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (a AngularDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "angular") {
		language.Frameworks = append(language.Frameworks, "Angular")
	}
}

// DoPortsDetection searches for the port in angular.json, package.json, and angular-cli.json
func (a AngularDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	// check if port is set on angular.json file
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "angular.json",
		},
	}, ctx)
	if err != nil {
		return
	}
	var data AngularJson
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return
	}

	if projectBody, exists := data.Projects[component.Name]; exists {
		port := projectBody.Architect.Serve.Options.Port
		if utils.IsValidPort(port) {
			component.Ports = []int{port}
			return
		}
	}

	// check if port is set in start script in package.json
	port := getPortFromStartScript(component.Path, []string{`--port (\d*)`})
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}

	// check if port is set on angular-cli.json file
	bytes, err = utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "angular-cli.json",
		},
	}, ctx)
	if err != nil {
		return
	}
	var dataCli AngularCliJson
	err = json.Unmarshal(bytes, &dataCli)
	if err != nil {
		return
	}

	port = dataCli.Defaults.Serve.Port
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}
}
