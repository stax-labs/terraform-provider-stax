package provider

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/server"
	"github.com/stretchr/testify/mock"
)

func TestWorkloadsDataSource(t *testing.T) {

	workloadID := "28a6b88b-80d7-4ecd-8dad-2d956d5132e8"

	si := mocks.NewServerInterface(t)

	readWorkloadsParams := models.WorkloadsReadWorkloadsParams{
		IdFilter: aws.String(workloadID),
	}

	si.On("WorkloadsReadWorkloads",
		mock.AnythingOfType("*echo.context"),
		readWorkloadsParams,
	).Return(func(c echo.Context, params models.WorkloadsReadWorkloadsParams) error {
		return c.JSON(200, &models.WorkloadsReadWorkloadsResponse{
			Workloads: []models.Workload{
				{
					Id:   aws.String(workloadID),
					Name: "production",
				},
			},
		})

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
		ProtoV6ProviderFactories:  testAccProtoV6ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: fmt.Sprintf(`data "stax_workloads" "production" {id = "%s"}`, workloadID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stax_workloads.production", "id", workloadID),
				),
			},
		},
	})
}
