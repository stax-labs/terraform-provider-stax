package provider

import (
	"fmt"
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
)

func TestAPITokensDataSource(t *testing.T) {

	apiTokenID := "28a6b88b-80d7-4ecd-8dad-2d956d5132e8"

	si := mocks.NewServerInterface(t)

	readAPITokensParams := models.TeamsReadApiTokensParams{
		IdFilter: aws.String(apiTokenID),
	}

	si.On("TeamsReadApiTokens",
		mock.AnythingOfType("*echo.context"),
		readAPITokensParams,
	).Return(func(c echo.Context, params models.TeamsReadApiTokensParams) error {
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
					AccessKey: aws.String(apiTokenID),
					Name:      "production",
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
				Config: fmt.Sprintf(`data "stax_api_tokens" "dedicated_dev" {id = "%s"}`, apiTokenID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.stax_api_tokens.dedicated_dev", "id", apiTokenID),
				),
			},
		},
	})
}
