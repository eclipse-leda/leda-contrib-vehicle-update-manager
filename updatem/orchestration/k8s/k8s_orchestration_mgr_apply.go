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

package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"

	"io/ioutil"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/delete"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

type kubectlApply struct {
	factory      cmdutil.Factory
	applyOptions *apply.ApplyOptions
	restConfig   *rest.Config
	ioStreams    genericclioptions.IOStreams
}

func newKubectlApply(kubeconfig *string) (*kubectlApply, error) {
	factory := cmdutil.NewFactory(&genericclioptions.ConfigFlags{KubeConfig: kubeconfig})
	ioStreams := genericclioptions.IOStreams{
		Out:    ioutil.Discard,
		ErrOut: ioutil.Discard,
	}
	applyOptions, err := createApplyOptions(factory, ioStreams)
	if err != nil {
		return nil, err
	}
	restConfig, err := factory.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	return &kubectlApply{
		factory:      factory,
		restConfig:   restConfig,
		ioStreams:    ioStreams,
		applyOptions: applyOptions,
	}, nil
}

func createApplyOptions(factory cmdutil.Factory, ioStreams genericclioptions.IOStreams) (*apply.ApplyOptions, error) {
	dryRunStrategy := cmdutil.DryRunNone
	dynamicClient, err := factory.DynamicClient()
	if err != nil {
		return nil, err
	}
	dryRunVerifier := resource.NewDryRunVerifier(dynamicClient, factory.OpenAPIGetter())

	printFlags := genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme)
	toPrinter := func(operation string) (printers.ResourcePrinter, error) {
		printFlags.NamePrintFlags.Operation = operation
		cmdutil.PrintFlagsWithDryRunStrategy(printFlags, dryRunStrategy)
		return printFlags.ToPrinter()
	}

	recorder, err := genericclioptions.NewRecordFlags().ToRecorder()
	if err != nil {
		return nil, err
	}

	deleteOptions, err := delete.NewDeleteFlags("").ToOptions(dynamicClient, ioStreams)
	if err != nil {
		return nil, err
	}
	deleteOptions.Filenames = []string{"-"}

	openAPISchema, _ := factory.OpenAPISchema()
	validator, err := factory.Validator(true)
	if err != nil {
		return nil, err
	}
	builder := factory.NewBuilder()
	mapper, err := factory.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	namespace, enforceNamespace, err := factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}

	applyOptions := &apply.ApplyOptions{
		PrintFlags: printFlags,

		DeleteOptions:   deleteOptions,
		ToPrinter:       toPrinter,
		ServerSideApply: false,
		ForceConflicts:  false,
		FieldManager:    "kubectl-client-side-apply",
		DryRunStrategy:  dryRunStrategy,
		DryRunVerifier:  dryRunVerifier,
		Prune:           true,
		All:             true,
		Overwrite:       true,
		OpenAPIPatch:    true,

		Recorder:         recorder,
		Namespace:        namespace,
		EnforceNamespace: enforceNamespace,
		Validator:        validator,
		Builder:          builder,
		Mapper:           mapper,
		DynamicClient:    dynamicClient,
		OpenAPISchema:    openAPISchema,

		IOStreams: ioStreams,

		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
	}

	applyOptions.PostProcessorFn = applyOptions.PrintAndPrunePostProcessor()

	return applyOptions, nil
}

func (k *kubectlApply) apply(mf []*unstructured.Unstructured) error {
	objects, err := k.createResourceInfos(mf, k.restConfig, k.applyOptions.Mapper)
	if err != nil {
		return err
	}
	k.applyOptions.SetObjects(objects)

	if err := k.applyOptions.Run(); err != nil {
		return err
	}
	return nil
}

func (k *kubectlApply) createResourceInfos(mf []*unstructured.Unstructured, restConfig *rest.Config, mapper meta.RESTMapper) ([]*resource.Info, error) {
	infos := []*resource.Info{}
	for _, u := range mf {
		info, err := k.createResourceInfo(u, restConfig, mapper)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (k *kubectlApply) createResourceInfo(mf *unstructured.Unstructured, restConfig *rest.Config, mapper meta.RESTMapper) (*resource.Info, error) {
	gvk := mf.GroupVersionKind()
	gv := gvk.GroupVersion()

	restClient, err := k.newRestClient(restConfig, gv)
	if err != nil {
		return nil, err
	}

	restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	namespace := mf.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	info := &resource.Info{
		Source:          "",
		Namespace:       namespace,
		Name:            mf.GetName(),
		Client:          restClient,
		Mapping:         restMapping,
		ResourceVersion: restMapping.Resource.Version,

		Object: mf,
	}

	return info, nil
}

func (k *kubectlApply) newRestClient(restConfig *rest.Config, gv schema.GroupVersion) (rest.Interface, error) {
	restConfig.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv
	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}
	return rest.RESTClientFor(restConfig)
}
