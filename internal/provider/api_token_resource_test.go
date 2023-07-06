package provider

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/core/server"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

func TestAPITokenResource(t *testing.T) {

	userID := "87c570e2-c795-44b0-aefa-ebdcffd4d048"

	si := mocks.NewServerInterface(t)

	si.On("TeamsCreateApiToken", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(200, &models.TeamsCreateApiTokenResponse{
			ApiTokens: []struct {
				AccessKey   *string                                            "json:\"AccessKey,omitempty\""
				CreatedBy   *string                                            "json:\"CreatedBy\""
				CreatedTS   *time.Time                                         "json:\"CreatedTS,omitempty\""
				Description *string                                            "json:\"Description,omitempty\""
				ModifiedTS  *time.Time                                         "json:\"ModifiedTS,omitempty\""
				Name        *string                                            "json:\"Name,omitempty\""
				Role        *models.ApiRole                                    "json:\"Role,omitempty\""
				SecretKey   *string                                            "json:\"SecretKey,omitempty\""
				Status      *models.TeamsCreateApiTokenResponseApiTokensStatus "json:\"Status,omitempty\""
				Tags        *models.Tags                                       "json:\"Tags\""
			}{
				{
					AccessKey: aws.String(userID),
					Name:      aws.String("production"),
					Role:      (*models.ApiRole)(aws.String("api_readonly")),
					Tags:      &models.Tags{"owner": "test"},
				},
			},
		})
	})

	si.On("TeamsReadApiToken", mock.AnythingOfType("*echo.context"), userID, models.TeamsReadApiTokenParams{}).Return(
		func(c echo.Context, accessKey string, params models.TeamsReadApiTokenParams) error {
			return c.JSON(200, &models.TeamsReadApiTokens{
				ApiTokens: []struct {
					AccessKey   *string                                  "json:\"AccessKey,omitempty\""
					CreatedBy   *string                                  "json:\"CreatedBy\""
					CreatedTS   *time.Time                               "json:\"CreatedTS,omitempty\""
					Description string                                   "json:\"Description\""
					ModifiedTS  *time.Time                               "json:\"ModifiedTS,omitempty\""
					Name        string                                   "json:\"Name\""
					Role        models.ApiRole                           "json:\"Role\""
					Status      models.TeamsReadApiTokensApiTokensStatus "json:\"Status\""
					Tags        *models.Tags                             "json:\"Tags\""
				}{
					{
						AccessKey: aws.String(userID),
						Name:      "production",
						Role:      (models.ApiRole)("api_readonly"),
						Tags:      &models.Tags{"owner": "test"},
					},
				},
			})
		},
	)

	si.On("TeamsDeleteApiToken", mock.AnythingOfType("*echo.context"), userID).Return(func(c echo.Context, accessKey string) error {
		return c.JSON(200, &models.TeamsDeleteApiTokenResponse{})
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
				Config: testAccCheckStaxAPITokenConfig("production", "production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_api_token.production", "id", userID),
				),
			},
		},
	})

}

func testAccCheckStaxAPITokenConfig(label, name string) string {
	configTemplate := `
resource "stax_api_token" "${label}" {
	name = "${name}"
	role       = "api_readonly"
	tags = {
		"owner": "test",
	}
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label": label,
			"name":  name,
		},
	)
}
