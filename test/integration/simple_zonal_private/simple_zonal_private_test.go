// Copyright 2022-2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package simple_zonal_private

import (
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/gcloud"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/golden"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/tft"
	"github.com/stretchr/testify/assert"
	"github.com/terraform-google-modules/terraform-google-kubernetes-engine/test/integration/testutils"
)

func TestSimpleZonalPrivate(t *testing.T) {
	bpt := tft.NewTFBlueprintTest(t,
		tft.WithRetryableTerraformErrors(testutils.RetryableTransientErrors, 3, 2*time.Minute),
	)

	bpt.DefineVerify(func(assert *assert.Assertions) {
		//Skipping Default Verify as the Verify Stage fails due to change in Client Cert Token
		// bpt.DefaultVerify(assert)
		testutils.TGKEVerify(t, bpt, assert) // Verify Resources

		projectId := bpt.GetStringOutput("project_id")
		location := bpt.GetStringOutput("location")
		clusterName := bpt.GetStringOutput("cluster_name")
		serviceAccount := bpt.GetStringOutput("service_account")

		op := gcloud.Runf(t, "container clusters describe %s --zone %s --project %s", clusterName, location, projectId)
		g := golden.NewOrUpdate(t, op.String(),
			golden.WithSanitizer(golden.StringSanitizer(serviceAccount, "SERVICE_ACCOUNT")),
			golden.WithSanitizer(golden.StringSanitizer(projectId, "PROJECT_ID")),
			golden.WithSanitizer(golden.StringSanitizer(clusterName, "CLUSTER_NAME")),
		)
		validateJSONPaths := []string{
			"status",
			"location",
			"locations",
			"privateClusterConfig.enablePrivateEndpoint",
			"privateClusterConfig.enablePrivateNodes",
			"addonsConfig.horizontalPodAutoscaling",
			"addonsConfig.httpLoadBalancing",
			"addonsConfig.kubernetesDashboard.disabled",
			"addonsConfig.networkPolicyConfig.disabled",
			"nodePools.name",
			"nodePools.initialNodeCount",
			"nodePools.autoscaling.enabled",
			"nodePools.autoscaling.minNodeCount",
			"nodePools.autoscaling.maxNodeCount",
			"nodePools.config.machineType",
			"nodePools.config.diskSizeGb",
			"nodePools.config.labels",
			"nodePools.config.tags",
			"nodePools.management.autoRepair",
			"masterAuthorizedNetworksConfig",
		}
		for _, pth := range validateJSONPaths {
			g.JSONEq(assert, op, pth)
		}

	})

	bpt.Test()
}
