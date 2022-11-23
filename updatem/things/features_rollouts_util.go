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

package things

import (
	"bytes"
	cryptoMd5 "crypto/md5"
	cryptoSha1 "crypto/sha1"
	cryptoSha256 "crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func convertToUpdateAction(args interface{}) (datatypes.UpdateAction, error) {
	bytes, err := json.Marshal(args)
	if err != nil {
		return datatypes.UpdateAction{}, err
	}
	var uas datatypes.UpdateAction
	err = json.Unmarshal(bytes, &uas)
	return uas, err
}

func getUpdateManifest(saa *datatypes.SoftwareArtifactAction) ([]*unstructured.Unstructured, bool, error) {
	downloadURL := saa.Download[datatypes.HTTP]

	if downloadURL == nil {
		downloadURL = saa.Download[datatypes.HTTPS]
	}

	mfBytes, err := downloadManifestDescription(downloadURL.URL)
	if err != nil {
		return nil, false, err
	}

	if err = validateSoftareArtifactHash(mfBytes, saa.Checksums); err != nil {
		// status should be FinishedRejected
		return nil, true, err
	}
	_, manifest, err := parseMultiYAML(mfBytes)
	return manifest, false, err
}

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
	log.Debug("downloaded update manager deployment manifest to install from Things:\n%+v", mf)
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

func downloadManifestDescription(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)

}

func validateSoftareArtifactHash(value []byte, hashes map[datatypes.Hash]string) error {
	if hashes[datatypes.MD5] != "" {
		return validateHashMd5(value, hashes[datatypes.MD5])
	} else if hashes[datatypes.SHA1] != "" {
		return validateHashSha1(value, hashes[datatypes.SHA1])
	} else if hashes[datatypes.SHA256] != "" {
		return validateHashSha256(value, hashes[datatypes.SHA256])
	} else {
		log.Warn("no hash information is provided to veryfiy the downloaded artifact")
		return nil
	}
}

func validateHashMd5(value []byte, md5Hash string) error {
	md5HashBytes, err := convertStringHashToBytes16(md5Hash)
	if err != nil {
		return err
	}

	md5 := cryptoMd5.Sum(value)
	if md5 != md5HashBytes {
		return log.NewError("md5 checksum does not match")
	}
	return nil
}

func validateHashSha1(value []byte, sha1Hash string) error {
	sha1HashBytes, err := convertStringHashToBytes20(sha1Hash)
	if err != nil {
		return err
	}
	sha1 := cryptoSha1.Sum(value)
	if sha1 != sha1HashBytes {
		return log.NewError("sha1 checksum does not match")
	}
	return nil
}

func validateHashSha256(value []byte, sha256Hash string) error {
	sha256HashBytes, err := convertStringHashToBytes32(sha256Hash)
	if err != nil {
		return err
	}
	sha256 := cryptoSha256.Sum256(value)
	if sha256 != sha256HashBytes {
		return log.NewError("sha256 checksum does not match")
	}
	return nil
}

func convertStringHashToBytes16(checkSum string) ([16]byte, error) {
	checkSumBytes := bytes.TrimSpace([]byte(checkSum))
	dst := [16]byte{}
	if _, err := hex.Decode(dst[:], checkSumBytes); err != nil {
		return dst, log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 16 bytes")
	}
	return dst, nil
}

func convertStringHashToBytes20(checkSum string) ([20]byte, error) {
	checkSumBytes := bytes.TrimSpace([]byte(checkSum))
	dst := [20]byte{}
	if _, err := hex.Decode(dst[:], checkSumBytes); err != nil {
		return dst, log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 20 bytes")
	}
	return dst, nil
}

func convertStringHashToBytes32(checkSum string) ([32]byte, error) {
	checkSumBytes := bytes.TrimSpace([]byte(checkSum))
	dst := [32]byte{}
	if _, err := hex.Decode(dst[:], checkSumBytes); err != nil {
		return dst, log.NewError("the provided input hash is either invalid, not a hex string or the length exceeds 32 bytes")
	}
	return dst, nil
}

func validateSoftwareUpdateActionManifests(updateAction datatypes.UpdateAction) error {
	if len(updateAction.SoftwareModules) != 1 {
		return log.NewError("the number of SoftwareModules must be exactly 1")
	}
	if len(updateAction.SoftwareModules[0].Artifacts) != 1 {
		return log.NewErrorf("the number of SoftwareArtifacts referenced for SoftwareModule [Name.version] = [%s.%s] must be exactly 1", updateAction.SoftwareModules[0].SoftwareModule.Name, updateAction.SoftwareModules[0].SoftwareModule.Version)
	}
	return nil
}
