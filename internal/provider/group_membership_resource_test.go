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
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

func TestGroupMembershipResource(t *testing.T) {

	groupID := "87c570e2-c795-44b0-aefa-ebdcffd4d048"
	taskID := "fd4d3cbc-1ba0-4d21-be4b-b63ffe3af4f1"
	userID := "ad74c49f-4854-4651-b606-c1ada37cd231"

	si := mocks.NewServerInterface(t)

	teamsUpdateEvent := &models.TeamsUpdateGroupMembersEvent{
		Detail: struct {
			Message         *string                "json:\"Message,omitempty\""
			Operation       models.Operation       "json:\"Operation\""
			OperationStatus models.OperationStatus "json:\"OperationStatus\""
			Severity        *string                "json:\"Severity,omitempty\""
			TaskId          *string                "json:\"TaskId,omitempty\""
		}{
			TaskId: aws.String(taskID),
		},
	}

	si.On("TeamsUpdateGroupMembers", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		tugm := new(models.TeamsUpdateGroupMembers)
		if err := c.Bind(tugm); err != nil {
			t.Fatalf("failed to bind request: %s", err)
		}

		if tugm.AddMembers == nil {
			t.Errorf("no members to add")
		}

		if tugm.RemoveMembers != nil {
			t.Errorf("members to remove not expected: %v", *tugm.RemoveMembers)
		}

		return c.JSON(200, teamsUpdateEvent)
	}).Once()

	si.On("TeamsUpdateGroupMembers", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		tugm := new(models.TeamsUpdateGroupMembers)
		if err := c.Bind(tugm); err != nil {
			t.Errorf("failed to bind request: %s", err)
		}

		if tugm.AddMembers != nil {
			t.Errorf("members to add not expected: %v", *tugm.AddMembers)
		}

		if tugm.RemoveMembers == nil {
			t.Errorf("no members to delete")
		}

		return c.JSON(200, teamsUpdateEvent)
	})

	si.On("TasksReadTask", mock.AnythingOfType("*echo.context"), mock.AnythingOfType("string")).Return(func(c echo.Context, taskId string) error {
		return c.JSON(200, &models.TasksReadTask{Status: staxsdk.TaskSucceeded})
	})

	si.On("TeamsReadGroup", mock.AnythingOfType("*echo.context"), groupID).Return(func(c echo.Context, groupID string) error {
		return c.JSON(200, &models.TeamsReadGroupsResponse{
			Groups: []models.Group{
				{
					Id:        aws.String(groupID),
					Name:      "production",
					GroupType: "LOCAL",
					Users: &[]string{
						userID,
					},
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
				Config: testAccCheckStaxGroupMembershipConfig("production", groupID, userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_group_membership.production", "id", groupID),
				),
			},
		},
	})

}

func testAccCheckStaxGroupMembershipConfig(label, groupID, userID string) string {
	configTemplate := `
resource "stax_group_membership" "${label}" {
	id = "${group_id}"
	user_ids = [
		"${user_id}"
	]
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label":    label,
			"group_id": groupID,
			"user_id":  userID,
		},
	)
}
