package provider

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/server"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

func TestUserResource(t *testing.T) {

	userID := "87c570e2-c795-44b0-aefa-ebdcffd4d048"
	taskID := "fd4d3cbc-1ba0-4d21-be4b-b63ffe3af4f1"
	email := openapi_types.Email("prod@example.com")
	role := models.Role("customer_readonly")

	si := mocks.NewServerInterface(t)

	si.On("TeamsCreateUser", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(200, &models.TeamsCreateUserEvent{
			TaskId: aws.String(taskID),
		})
	})

	si.On("TasksReadTask", mock.AnythingOfType("*echo.context"), mock.AnythingOfType("string")).Return(func(c echo.Context, taskId string) error {
		return c.JSON(200, &models.TasksReadTask{Status: staxsdk.TaskSucceeded, Logs: []string{fmt.Sprintf("Successfully created user %s", userID)}})
	})

	si.On("TeamsReadUser", mock.AnythingOfType("*echo.context"), userID).Return(func(c echo.Context, userID string) error {
		return c.JSON(200, &models.TeamsReadUsers{
			Users: []models.User{
				{
					Id:        aws.String(userID),
					FirstName: aws.String("prod"),
					LastName:  aws.String("duction"),
					Email:     &email,
					Role:      &role,
				},
			},
		})
	})

	si.On("TeamsDeleteUser", mock.AnythingOfType("*echo.context"), userID).Return(func(c echo.Context, userID string) error {
		return c.JSON(200, &models.TeamsDeleteUserResponse{})
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
				Config: testAccCheckStaxUserConfig("production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_user.production", "id", userID),
				),
			},
		},
	})

}

func testAccCheckStaxUserConfig(label string) string {
	configTemplate := `
resource "stax_user" "${label}" {
	first_name = "prod"
	last_name  = "duction"
	email      = "prod@example.com"
	role       = "customer_readonly"
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label": label,
		},
	)
}

func TestExtractUserID(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected string
		wantErr  bool
	}{
		{
			name:     "Extract user ID",
			message:  "Successfully created user 28a6b88b-80d7-4ecd-8dad-2d956d5132e8",
			expected: "28a6b88b-80d7-4ecd-8dad-2d956d5132e8",
			wantErr:  false,
		},
		{
			name:     "No match",
			message:  "Random message",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := extractUserID(tc.message)
			if tc.wantErr && err == nil {
				t.Errorf("Expected error, got none")
			} else if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %s", err)
			}

			if actual != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
