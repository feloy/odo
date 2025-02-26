/*******************************************************************************
 * Copyright (c) 2021 Red Hat, Inc.
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

type DjangoDetector struct{}

func (d DjangoDetector) GetSupportedFrameworks() []string {
	return []string{"Django"}
}

// DoFrameworkDetection uses a tag to check for the framework name
// with django files and django config files
func (d DjangoDetector) DoFrameworkDetection(language *model.Language, files *[]string) {
	managePy := utils.GetFile(files, "manage.py")
	urlsPy := utils.GetFile(files, "urls.py")
	wsgiPy := utils.GetFile(files, "wsgi.py")
	asgiPy := utils.GetFile(files, "asgi.py")
	requirementsTxt := utils.GetFile(files, "requirements.txt")
	projectToml := utils.GetFile(files, "pyproject.toml")

	var djangoFiles []string
	var configDjangoFiles []string
	utils.AddToArrayIfValueExist(&djangoFiles, managePy)
	utils.AddToArrayIfValueExist(&djangoFiles, urlsPy)
	utils.AddToArrayIfValueExist(&djangoFiles, wsgiPy)
	utils.AddToArrayIfValueExist(&djangoFiles, asgiPy)
	utils.AddToArrayIfValueExist(&configDjangoFiles, requirementsTxt)
	utils.AddToArrayIfValueExist(&configDjangoFiles, projectToml)

	if hasFramework(&djangoFiles, "from django.") || hasFramework(&configDjangoFiles, "django") || hasFramework(&configDjangoFiles, "Django") {
		language.Frameworks = append(language.Frameworks, "Django")
	}
}

type ApplicationPropertiesFile struct {
	Dir  string
	File string
}

// DoPortsDetection searches for the port in /manage.py
func (d DjangoDetector) DoPortsDetection(component *model.Component, ctx *context.Context) {
	bytes, err := utils.ReadAnyApplicationFile(component.Path, []model.ApplicationFileInfo{
		{
			Dir:  "",
			File: "manage.py",
		},
	}, ctx)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`.default_port\s*=\s*"([^"]*)`)
	component.Ports = utils.FindAllPortsSubmatch(re, string(bytes), 1)
}
