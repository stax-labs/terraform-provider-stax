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

func TestUsersDataSource(t *testing.T) {

	userID := "28a6b88b-80d7-4ecd-8dad-2d956d5132e8"

	si := mocks.NewServerInterface(t)

	readUsersParams := models.TeamsReadUsersParams{
		IdFilter: aws.String(userID),
	}

	si.On("TeamsReadUsers",
		mock.AnythingOfType("*echo.context"),
		readUsersParams,
	).Return(func(c echo.Context, params models.TeamsReadUsersParams) error {
		return c.JSON(200, &models.TeamsReadUsers{
			Users: []models.User{
				{
					Id:   aws.String(userID),
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
				Config: fmt.Sprintf(`data "stax_users" "dedicated_dev" {id = "%s"}`, userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stax_users.dedicated_dev", "id", userID),
				),
			},
		},
	})
}
