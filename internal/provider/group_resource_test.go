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

func TestGroupResource(t *testing.T) {

	groupID := "87c570e2-c795-44b0-aefa-ebdcffd4d048"
	taskID := "fd4d3cbc-1ba0-4d21-be4b-b63ffe3af4f1"

	si := mocks.NewServerInterface(t)

	si.On("TeamsCreateGroup", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(200, &models.TeamsCreateGroupEvent{
			Detail: struct {
				Message         *string                "json:\"Message,omitempty\""
				Operation       models.Operation       "json:\"Operation\""
				OperationStatus models.OperationStatus "json:\"OperationStatus\""
				Severity        *string                "json:\"Severity,omitempty\""
				TaskId          *string                "json:\"TaskId,omitempty\""
			}{
				TaskId: aws.String(taskID),
			},
			GroupId: aws.String(groupID),
		})
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
				},
			},
		})
	})

	si.On("TeamsDeleteGroup", mock.AnythingOfType("*echo.context"), groupID).Return(func(c echo.Context, accountTypeId string) error {
		return c.JSON(200, &models.TeamsDeleteGroupEvent{})
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
				Config: testAccCheckStaxGroupConfig("production", "production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_group.production", "id", groupID),
				),
			},
		},
	})

}

func testAccCheckStaxGroupConfig(label, name string) string {
	configTemplate := `
resource "stax_group" "${label}" {
	name = "${name}"
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label": label,
			"name":  name,
		},
	)
}
