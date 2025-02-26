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
	"regexp"

	"github.com/devfile/alizer/pkg/apis/model"
	"github.com/devfile/alizer/pkg/utils"
)

type NuxtDetector struct{}

func (n NuxtDetector) GetSupportedFrameworks() []string {
	return []string{"Nuxt"}
}

// DoFrameworkDetection uses a tag to check for the framework name
func (n NuxtDetector) DoFrameworkDetection(language *model.Language, config string) {
	if hasFramework(config, "nuxt") {
		language.Frameworks = append(language.Frameworks, "Nuxt", "Nuxt.js")
	}
}

// DoPortsDetection searches for the port in package.json, and nuxt.config.js
func (n NuxtDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	regexes := []string{`--port=(\d*)`}
	// check if port is set in start script in package.json
	port := getPortFromStartScript(component.Path, regexes)
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}

	// check if port is set in dev script in package.json
	port = getPortFromDevScript(component.Path, regexes)
	if utils.IsValidPort(port) {
		component.Ports = []int{port}
		return
	}

	//check inside the nuxt.config.js file
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "nuxt.config.js",
		},
	}, ctx)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`port:\s*(\d+)*`)
	component.Ports = utils.FindAllPortsSubmatch(re, string(bytes), 1)
}
