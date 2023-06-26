package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/server"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

func TestWorkloadResource(t *testing.T) {

	workloadID := "c55ef08c-82b0-4eac-9cac-42c3d32c8734"

	catalogID := "b705523f-e280-4c48-9483-31a2e446f743"
	accountID := "21c4183b-2886-4d8c-9663-95b0352d9f5e"
	region := "ap-southeast-2"
	// taskID := "fd4d3cbc-1ba0-4d21-be4b-b63ffe3af4f1"

	si := mocks.NewServerInterface(t)

	si.On("WorkloadsCreateWorkload", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(200, &models.CreateWorkloadEvent{
			WorkloadId: aws.String(workloadID),
		})
	})

	si.On("WorkloadsReadWorkload", mock.AnythingOfType("*echo.context"), workloadID, models.WorkloadsReadWorkloadParams{}).
		Return(func(c echo.Context, workloadID string, params models.WorkloadsReadWorkloadParams) error {
			return c.JSON(200, &models.WorkloadsReadWorkloadsResponse{
				Workloads: []models.Workload{
					{
						Id:          aws.String(workloadID),
						Name:        "production",
						AccountId:   accountID,
						CatalogueId: catalogID,
						Region:      region,
						Tags: &models.Tags{
							"test": "test",
						},
					},
				},
			})
		})

	si.On("WorkloadsDeleteWorkload", mock.AnythingOfType("*echo.context"), workloadID).
		Return(func(c echo.Context, workloadID string) error {
			return c.JSON(200, &models.DeleteWorkloadEvent{})
		})

	e := echo.New()

	server.RegisterHandlers(e, si)

	ts := httptest.NewServer(e.Server.Handler)
	defer ts.Close()

	t.Setenv("INTEGRATION_TEST_ENDPOINT_URL", ts.URL)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccCheckStaxWorkloadConfig("production", "production", catalogID, accountID, region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_workload.production", "id", workloadID),
				),
			},
		},
	})
}

func testAccCheckStaxWorkloadConfig(label, name, catalogID, accountID, region string) string {
	configTemplate := `
resource "stax_workload" "${label}" {
	name       = "${name}"
	catalog_id = "${catalog_id}"
	account_id = "${account_id}"
	region     = "${region}"
	tags       = {
		"test" : "test"
	}
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label":      label,
			"name":       name,
			"catalog_id": catalogID,
			"account_id": accountID,
			"region":     region,
		},
	)
}
