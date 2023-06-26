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

func TestPermissionSetAssignmentResource(t *testing.T) {
	permissionSetAssignmentID := "5d7b1228-427b-4df2-b562-8dca6ae715bb"
	permissionSetID := "efa64ad8-2a44-41a9-8bbd-14343547af4a"
	groupID := "01110535-057d-4fa6-bd83-cc48c2c1aee9"
	accountTypeID := "6b8d429c-7051-4580-bdc0-d0f34a887944"

	si := mocks.NewServerInterface(t)

	si.On("CreatePermissionSetAssignments", mock.AnythingOfType("*echo.context"), uuid.MustParse(permissionSetID)).
		Return(func(c echo.Context, permissionSetId uuid.UUID) error {
			return c.JSON(200, &models.AssignmentRecords{
				{
					Id:            uuid.MustParse(permissionSetAssignmentID),
					GroupId:       uuid.MustParse(groupID),
					AccountTypeId: uuid.MustParse(accountTypeID),
					Status:        models.DEPLOYMENTINPROGRESS,
				},
			})
		})

	si.On("ListPermissionSetAssignments", mock.AnythingOfType("*echo.context"), uuid.MustParse(permissionSetID), models.ListPermissionSetAssignmentsParams{}).
		Return(func(c echo.Context, permissionSetId uuid.UUID, params models.ListPermissionSetAssignmentsParams) error {
			return c.JSON(200, &models.ListAssignmentRecords{
				Assignments: []models.AssignmentRecord{
					{
						Id:            uuid.MustParse(permissionSetAssignmentID),
						GroupId:       uuid.MustParse(groupID),
						AccountTypeId: uuid.MustParse(accountTypeID),
						Status:        models.DEPLOYMENTCOMPLETE,
					},
				},
			})
		}).Once()

	si.On("DeletePermissionSetAssignment", mock.AnythingOfType("*echo.context"), uuid.MustParse(permissionSetID), uuid.MustParse(permissionSetAssignmentID)).
		Return(func(c echo.Context, permissionSetId uuid.UUID, permissionSetAssignmentId uuid.UUID) error {
			return c.JSON(200, &models.AssignmentRecord{
				Id:            uuid.MustParse(permissionSetAssignmentID),
				GroupId:       uuid.MustParse(groupID),
				AccountTypeId: uuid.MustParse(accountTypeID),
				Status:        models.DELETEREQUESTED,
			})
		})

	si.On("ListPermissionSetAssignments", mock.AnythingOfType("*echo.context"), uuid.MustParse(permissionSetID), models.ListPermissionSetAssignmentsParams{}).
		Return(func(c echo.Context, permissionSetId uuid.UUID, params models.ListPermissionSetAssignmentsParams) error {
			return c.JSON(200, &models.ListAssignmentRecords{
				Assignments: []models.AssignmentRecord{
					{
						Id:            uuid.MustParse(permissionSetAssignmentID),
						GroupId:       uuid.MustParse(groupID),
						AccountTypeId: uuid.MustParse(accountTypeID),
						Status:        models.DELETECOMPLETE,
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
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccCheckStaxPermissionSetAssignmentConfig("production", permissionSetID, accountTypeID, groupID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_permission_set_assignment.production", "id", permissionSetAssignmentID),
				),
			},
		},
	})
}

func testAccCheckStaxPermissionSetAssignmentConfig(label, permission_set_id, account_type_id, group_id string) string {
	configTemplate := `
resource "stax_permission_set_assignment" "${label}" {
	permission_set_id = "${permission_set_id}"
	account_type_id = "${account_type_id}"
	group_id = "${group_id}"
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label":             label,
			"permission_set_id": permission_set_id,
			"account_type_id":   account_type_id,
			"group_id":          group_id,
		},
	)
}
