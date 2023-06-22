package provider

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/server"
	"github.com/stretchr/testify/mock"
)

func TestPermissionSetsDataSource(t *testing.T) {

	permissionSetID := "28a6b88b-80d7-4ecd-8dad-2d956d5132e8"

	si := mocks.NewServerInterface(t)

	psetUUID := uuid.MustParse(permissionSetID)

	si.On("GetPermissionSet",
		mock.AnythingOfType("*echo.context"),
		psetUUID,
	).Return(func(c echo.Context, permissionSetId uuid.UUID) error {
		return c.JSON(200, &models.PermissionSetRecord{
			Id: permissionSetId,
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
				Config: fmt.Sprintf(`data "stax_permission_sets" "dedicated_dev" {id = "%s"}`, permissionSetID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stax_permission_sets.dedicated_dev", "id", permissionSetID),
				),
			},
		},
	})

}
