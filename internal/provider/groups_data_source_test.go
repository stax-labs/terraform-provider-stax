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

func TestGroupsDataSource(t *testing.T) {

	groupID := "28a6b88b-80d7-4ecd-8dad-2d956d5132e8"

	si := mocks.NewServerInterface(t)

	readGroupsParams := models.TeamsReadGroupsParams{
		IdFilter: aws.String(groupID),
	}

	si.On("TeamsReadGroups",
		mock.AnythingOfType("*echo.context"),
		readGroupsParams,
	).Return(func(c echo.Context, params models.TeamsReadGroupsParams) error {
		return c.JSON(200, &models.TeamsReadGroupsResponse{
			Groups: []models.Group{
				{
					Id:   aws.String(groupID),
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
				Config: fmt.Sprintf(`data "stax_groups" "dedicated_dev" {id = "%s"}`, groupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stax_groups.dedicated_dev", "id", groupID),
				),
			},
		},
	})
}
