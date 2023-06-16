package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/server"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

func TestAccountTypeResource(t *testing.T) {

	accountTypeID := "87c570e2-c795-44b0-aefa-ebdcffd4d048"

	si := mocks.NewServerInterface(t)

	si.On("AccountsCreateAccountType", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(200, &models.AccountsCreateAccountTypeResponse{Detail: struct {
			AccountType     models.AccountType     "json:\"AccountType\""
			Message         string                 "json:\"Message\""
			Operation       models.Operation       "json:\"Operation\""
			OperationStatus models.OperationStatus "json:\"OperationStatus\""
			Severity        string                 "json:\"Severity\""
		}{
			AccountType: models.AccountType{
				Id: aws.String(accountTypeID),
			},
		}})
	})

	si.On("AccountsReadAccountType", mock.AnythingOfType("*echo.context"), accountTypeID, mock.AnythingOfType("models.AccountsReadAccountTypeParams")).Return(func(c echo.Context, accountTypeID string, params models.AccountsReadAccountTypeParams) error {
		return c.JSON(200, &models.AccountsReadAccountTypes{
			AccountTypes: []models.AccountType{
				{
					Id:   aws.String(accountTypeID),
					Name: "production",
				},
			},
		})
	})

	si.On("AccountsDeleteAccountType", mock.AnythingOfType("*echo.context"), accountTypeID).Return(func(c echo.Context, accountTypeId string) error {
		return c.JSON(200, &models.AccountsDeleteAccountTypeResponse{
			AccountTypes: models.AccountType{
				Id:   aws.String(accountTypeID),
				Name: "production",
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
				Config: testAccCheckStaxAccountTypeConfig("production", "production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_account_type.production", "id", accountTypeID),
				),
			},
		},
	})

}

func testAccCheckStaxAccountTypeConfig(label, name string) string {
	configTemplate := `
resource "stax_account_type" "${label}" {
	name = "${name}"
}`
	return fasttemplate.ExecuteString(configTemplate, "${", "}",
		map[string]any{
			"label": label,
			"name":  name,
		},
	)
}
