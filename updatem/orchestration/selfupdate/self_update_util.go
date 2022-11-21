// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Apache License 2.0 which is available at
// https://www.apache.org/licenses/LICENSE-2.0
//
// SPDX-License-Identifier: Apache-2.0

package selfupdate

import (
	"bytes"
	"io"

	"github.com/eclipse-kanto/container-management/containerm/log"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func parseMultiYAML(multiYamlData []byte) ([][]byte, []*unstructured.Unstructured, error) {
	mf := []*unstructured.Unstructured{}
	singleYamlDoc, yamlReadErr := readResources(multiYamlData)
	if yamlReadErr != nil {
		return nil, nil, yamlReadErr
	}
	if len(singleYamlDoc) == 0 {
		return nil, nil, log.NewError("no k8s content provided for processing")
	}

	for _, pod := range singleYamlDoc {
		bs, yamlErr := yaml.YAMLToJSON(pod)
		if yamlErr != nil {
			log.ErrorErr(yamlErr, "error reading yaml content")
			return nil, nil, yamlErr
		}
		podSpec := &unstructured.Unstructured{}
		podSpec.UnmarshalJSON(bs)
		mf = append(mf, podSpec)
	}
	log.Debug("downloaded update manifest to install from Things:\n%+v", mf)
	return singleYamlDoc, mf, nil
}

func readResources(multiYamlData []byte) ([][]byte, error) {
	var documentList [][]byte
	yamlDecoder := yamlv3.NewDecoder(bytes.NewReader(multiYamlData))
	for {
		var singleDoc interface{}
		err := yamlDecoder.Decode(&singleDoc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, log.NewErrorf("error splitting multi yaml content:\n%+v", err)
		}

		if singleDoc != nil {
			document, marshalErr := yamlv3.Marshal(singleDoc)
			if marshalErr != nil {
				return nil, log.NewErrorf("error marshalling single read yaml document:\n%+v", marshalErr)
			}
			documentList = append(documentList, document)
		}
	}
	return documentList, nil
}
