package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testAccClickhouseName string = fmt.Sprintf("%v-clickhouse", testPrefix)
	testAccClickhouseId   string = fmt.Sprintf("doublecloud_clickhouse_cluster.%v", testAccClickhouseName)
)

func TestAccClickhouseClusterResource(t *testing.T) {
	m := clickhouseClusterModel{
		ProjectId: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccClickhouseName),
		RegionId:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),

		Resources: &clickhouseClusterResources{
			Clickhouse: &clickhouseClusterResourcesClickhouse{
				ResourcePresetId: types.StringValue("s1-c2-m4"),
				DiskSize:         types.Int64Value(34359738368),
				ReplicaCount:     types.Int64Value(1),
			},
		},
	}

	m2 := clickhouseClusterModel(m)
	m2.Name = types.StringValue(fmt.Sprintf("%v-changed", testAccClickhouseName))
	m2.Resources = &clickhouseClusterResources{
		Clickhouse: &clickhouseClusterResourcesClickhouse{
			ResourcePresetId: types.StringValue("s1-c2-m4"),
			DiskSize:         types.Int64Value(51539607552),
			ReplicaCount:     types.Int64Value(1),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClickhouseClusterResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "name", m.Name.ValueString()),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id", "s1-c2-m4"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.disk_size", "34359738368"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.log_level", "LOG_LEVEL_INFORMATION"),
				),
			},
			// Update and Read testing
			{
				Config: testAccClickhouseClusterResourceConfigUpdated(&m2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "name", m2.Name.ValueString()),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id", "s1-c2-m4"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.disk_size", "51539607552"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.log_level", "LOG_LEVEL_TRACE"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.max_connections", "120"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccClickhouseClusterResourceConfig(m *clickhouseClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_clickhouse_cluster" "tf-acc-clickhouse" {
  project_id = %[1]q
  name = %[2]q
  region_id = %[3]q
  cloud_type = %[4]q
  network_id = %[5]q

  resources {
	clickhouse {
		resource_preset_id = %[6]q
		disk_size =  %[7]q
		replica_count = %[8]q
	}
  }

  config {
	log_level = "LOG_LEVEL_INFORMATION"
  }
}
`, m.ProjectId.ValueString(),
		m.Name.ValueString(),
		m.RegionId.ValueString(),
		m.CloudType.ValueString(),
		m.NetworkId.ValueString(),
		m.Resources.Clickhouse.ResourcePresetId.ValueString(),
		m.Resources.Clickhouse.DiskSize.String(),
		m.Resources.Clickhouse.ReplicaCount.String(),
	)
}

func testAccClickhouseClusterResourceConfigUpdated(m *clickhouseClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_clickhouse_cluster" "tf-acc-clickhouse" {
  project_id = %[1]q
  name = %[2]q
  region_id = %[3]q
  cloud_type = %[4]q
  network_id = %[5]q

  resources {
	clickhouse {
		resource_preset_id = %[6]q
		disk_size =  %[7]q
		replica_count = %[8]q
	}
  }

  config {
	log_level = "LOG_LEVEL_TRACE"
	max_connections = 120
  }
}
`, m.ProjectId.ValueString(),
		m.Name.ValueString(),
		m.RegionId.ValueString(),
		m.CloudType.ValueString(),
		m.NetworkId.ValueString(),
		m.Resources.Clickhouse.ResourcePresetId.ValueString(),
		m.Resources.Clickhouse.DiskSize.String(),
		m.Resources.Clickhouse.ReplicaCount.String(),
	)
}

func init() {
	resource.AddTestSweepers("clickhouse", &resource.Sweeper{
		Name:         "clickhouse",
		F:            sweepClickhouses,
		Dependencies: []string{},
	})
}

func sweepClickhouses(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &clickhouse.ListClustersRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.ClickHouse().Cluster()
	it := svc.ClusterIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepClickhouse(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepClickhouse(conf *Config, t *clickhouse.Cluster) error {
	op, err := conf.sdk.WrapOperation(conf.sdk.ClickHouse().Cluster().Delete(conf.ctx, &clickhouse.DeleteClusterRequest{ClusterId: t.Id}))
	if err != nil {
		return err
	}
	return op.Wait(conf.ctx)
}
