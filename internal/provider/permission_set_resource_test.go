package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/permissionssets/server"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

func TestPermissionSetResource(t *testing.T) {

	permissionSetID := "efa64ad8-2a44-41a9-8bbd-14343547af4a"

	si := mocks.NewServerInterface(t)

	si.On("CreatePermissionSet", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(201, &models.PermissionSetRecord{
			Id:   uuid.MustParse(permissionSetID),
			Name: "production",
		})
	})

	si.On("GetPermissionSet",
		mock.AnythingOfType("*echo.context"),
		uuid.MustParse(permissionSetID),
	).Return(func(c echo.Context, permissionSetId uuid.UUID) error {
		return c.JSON(200, &models.PermissionSetRecord{
			Id:   permissionSetId,
			Name: "production",
		})
	})

	si.On("DeletePermissionSet", mock.AnythingOfType("*echo.context"), uuid.MustParse(permissionSetID)).Return(func(c echo.Context, permissionSetId uuid.UUID) error {
		return c.JSON(200, &models.PermissionSetRecord{})
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
				Config: testAccCheckStaxPermissionSetConfig("production", "production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_permission_set.production", "id", permissionSetID),
				),
			},
		},
	})
}

func testAccCheckStaxPermissionSetConfig(label, name string) string {
	configTemplate := `
resource "stax_permission_set" "${label}" {
	name = "${name}"
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label": label,
			"name":  name,
		},
	)
}
