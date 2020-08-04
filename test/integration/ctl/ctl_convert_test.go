/*
Copyright 2020 The Jetstack cert-manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ctl

import (
	"bytes"
	"io/ioutil"
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/jetstack/cert-manager/cmd/ctl/pkg/convert"
)

const (
	testdataResource1                        = "./testdata/convert/input/resource1.yaml"
	testdataResource2                        = "./testdata/convert/input/resource2.yaml"
	testdataResource3                        = "./testdata/convert/input/resource3.yaml"
	testdataResourceWithOrganizationV1alpha2 = "./testdata/convert/input/resource_with_organization_v1alpha2.yaml"
	testdataResourcesAsListV1alpha2          = "./testdata/convert/input/resources_as_list_v1alpha2.yaml"

	testdataNoOutputError                    = "./testdata/convert/output/no_output_error.yaml"
	testdataResource1V1alpha2                = "./testdata/convert/output/resource1_v1alpha2.yaml"
	testdataResource1V1alpha3                = "./testdata/convert/output/resource1_v1alpha3.yaml"
	testdataResource2V1alpha2                = "./testdata/convert/output/resource2_v1alpha2.yaml"
	testdataResource2V1alpha3                = "./testdata/convert/output/resource2_v1alpha3.yaml"
	testdataResource3V1alpha2                = "./testdata/convert/output/resource3_v1alpha2.yaml"
	testdataResourceWithOrganizationV1alpha3 = "./testdata/convert/output/resource_with_organization_v1alpha3.yaml"
	testdataResourceWithOrganizationV1beta1  = "./testdata/convert/output/resource_with_organization_v1beta1.yaml"
	testdataResourcesOutAsListV1alpha2       = "./testdata/convert/output/resources_as_list_v1alpha2.yaml"
	testdataResourcesOutAsListV1alpha3       = "./testdata/convert/output/resources_as_list_v1alpha3.yaml"
	testdataResourcesOutAsListV1beta1        = "./testdata/convert/output/resources_as_list_v1beta1.yaml"

	targetv1alpha2 = "cert-manager.io/v1alpha2"
	targetv1alpha3 = "cert-manager.io/v1alpha3"
	targetv1beta1  = "cert-manager.io/v1beta1"
)

func TestCtlConvert(t *testing.T) {
	tests := map[string]struct {
		input, expOutputFile string
		targetVersion        string
		expErr               bool
	}{
		"a single cert-manager resource should convert to v1alpha2 with no target": {
			input:         testdataResource1,
			expOutputFile: testdataResource1V1alpha2,
		},
		"a single cert-manager resource should convert to v1alpha2 with target v1alpha2": {
			input:         testdataResource1,
			targetVersion: targetv1alpha2,
			expOutputFile: testdataResource1V1alpha2,
		},
		"a single cert-manager resource should convert to v1alpha3 with target v1alpha3": {
			input:         testdataResource1,
			targetVersion: targetv1alpha3,
			expOutputFile: testdataResource1V1alpha3,
		},
		"a list of cert-manager resources should convert to v1alpha2 with no target": {
			input:         testdataResource2,
			expOutputFile: testdataResource2V1alpha2,
		},
		"a list of cert-manager resources should convert to v1alpha2 with target v1alpha2": {
			input:         testdataResource2,
			expOutputFile: testdataResource2V1alpha2,
		},
		"a list of cert-manager resources should convert to v1alpha3 with target v1alpha3": {
			input:         testdataResource2,
			targetVersion: targetv1alpha3,
			expOutputFile: testdataResource2V1alpha3,
		},
		"a list of a mix of cert-manager and non cert-manager resources should convert to v1alpha2 with no target": {
			input:         testdataResource2,
			targetVersion: targetv1alpha3,
			expOutputFile: testdataResource2V1alpha3,
		},
		"a list of a mix of cert-manager and non cert-manager resources should error with target v1alpha2": {
			input:         testdataResource3,
			targetVersion: targetv1alpha2,
			expOutputFile: testdataNoOutputError,
			expErr:        true,
		},
		"a list of a mix of cert-manager and non cert-manager resources should error with target v1alpha3": {
			input:         testdataResource3,
			targetVersion: targetv1alpha3,
			expOutputFile: testdataNoOutputError,
			expErr:        true,
		},
		"an object in v1alpha2 that uses a field that has been renamed in v1alpha3 should be converted properly": {
			input:         testdataResourceWithOrganizationV1alpha2,
			targetVersion: targetv1alpha3,
			expOutputFile: testdataResourceWithOrganizationV1alpha3,
		},
		"an object in v1alpha2 that uses a field that has been renamed in v1beta1 should be converted properly": {
			input:         testdataResourceWithOrganizationV1alpha2,
			targetVersion: targetv1beta1,
			expOutputFile: testdataResourceWithOrganizationV1beta1,
		},
		"a list in v1alpha2 should parsed": {
			input:         testdataResourcesAsListV1alpha2,
			targetVersion: targetv1alpha2,
			expOutputFile: testdataResourcesOutAsListV1alpha2,
		},
		"a list in v1alpha2 should be converted to v1alpha3": {
			input:         testdataResourcesAsListV1alpha2,
			targetVersion: targetv1alpha3,
			expOutputFile: testdataResourcesOutAsListV1alpha3,
		},
		"a list in v1alpha2 should be converted to v1beta1": {
			input:         testdataResourcesAsListV1alpha2,
			targetVersion: targetv1beta1,
			expOutputFile: testdataResourcesOutAsListV1beta1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expOutput, err := ioutil.ReadFile(test.expOutputFile)
			if err != nil {
				t.Fatalf("%s: %s", test.expOutputFile, err)
			}

			// Run ctl convert command with input options
			streams, _, outBuf, _ := genericclioptions.NewTestIOStreams()

			opts := convert.NewOptions(streams)
			opts.OutputVersion = test.targetVersion
			opts.Filenames = []string{test.input}

			if err := opts.Complete(); err != nil {
				t.Fatal(err)
			}

			err = opts.Run()
			if test.expErr != (err != nil) {
				t.Errorf("got unexpected error, exp=%t got=%v",
					test.expErr, err)
			}

			if !bytes.Equal(bytes.TrimSpace(expOutput), bytes.TrimSpace(outBuf.Bytes())) {
				t.Errorf("got unexpected output, exp=%s\n got=%s",
					bytes.TrimSpace(expOutput), bytes.TrimSpace(outBuf.Bytes()))
			}
		})
	}
}

const (
	resource1v1alpha2 = `
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  creationTimestamp: null
  name: ca-issuer
  namespace: sandbox
spec:
  commonName: my-csi-app
  isCA: true
  issuerRef:
    group: cert-manager.io
    kind: Issuer
    name: selfsigned-issuer
  secretName: ca-key-pair
status: {}`
	resource1v1alpha3 = `
apiVersion: cert-manager.io/v1alpha3
kind: Certificate
metadata:
  creationTimestamp: null
  name: ca-issuer
  namespace: sandbox
spec:
  commonName: my-csi-app
  isCA: true
  issuerRef:
    group: cert-manager.io
    kind: Issuer
    name: selfsigned-issuer
  secretName: ca-key-pair
status: {}`

	resource2v1alpha2 = `
Items:
- apiVersion: cert-manager.io/v1alpha2
  kind: Certificate
  metadata:
    creationTimestamp: null
    name: ca-issuer
    namespace: sandbox
  spec:
    commonName: my-csi-app
    isCA: true
    issuerRef:
      group: cert-manager.io
      kind: Issuer
      name: selfsigned-issuer
    secretName: ca-key-pair
  status: {}
- apiVersion: cert-manager.io/v1alpha2
  kind: Issuer
  metadata:
    creationTimestamp: null
    name: ca-issuer
    namespace: sandbox
  spec:
    ca:
      secretName: ca-key-pair
  status: {}
- apiVersion: cert-manager.io/v1alpha2
  kind: Certificate
  metadata:
    creationTimestamp: null
    name: ca-issuer-2
    namespace: sandbox
  spec:
    commonName: my-csi-app
    isCA: true
    issuerRef:
      group: cert-manager.io
      kind: Issuer
      name: ca-issuer
    secretName: ca-key-pair
  status: {}
apiVersion: v1
kind: List`
	resource2v1alpha3 = `
Items:
- apiVersion: cert-manager.io/v1alpha3
  kind: Certificate
  metadata:
    creationTimestamp: null
    name: ca-issuer
    namespace: sandbox
  spec:
    commonName: my-csi-app
    isCA: true
    issuerRef:
      group: cert-manager.io
      kind: Issuer
      name: selfsigned-issuer
    secretName: ca-key-pair
  status: {}
- apiVersion: cert-manager.io/v1alpha3
  kind: Issuer
  metadata:
    creationTimestamp: null
    name: ca-issuer
    namespace: sandbox
  spec:
    ca:
      secretName: ca-key-pair
  status: {}
- apiVersion: cert-manager.io/v1alpha3
  kind: Certificate
  metadata:
    creationTimestamp: null
    name: ca-issuer-2
    namespace: sandbox
  spec:
    commonName: my-csi-app
    isCA: true
    issuerRef:
      group: cert-manager.io
      kind: Issuer
      name: ca-issuer
    secretName: ca-key-pair
  status: {}
apiVersion: v1
kind: List`

	resource3v1alpha2 = `
Items:
- apiVersion: v1
  kind: Namespace
  metadata:
    creationTimestamp: null
    name: sandbox
  spec: {}
  status: {}
- apiVersion: cert-manager.io/v1alpha2
  kind: Issuer
  metadata:
    creationTimestamp: null
    name: selfsigned-issuer
    namespace: sandbox
  spec:
    selfSigned: {}
  status: {}
apiVersion: v1
kind: List`
)
