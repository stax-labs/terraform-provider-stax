package provider

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labstack/echo/v4"
	"github.com/stax-labs/terraform-provider-stax/internal/api/mocks"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/models"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/server"
	"github.com/stax-labs/terraform-provider-stax/internal/api/staxsdk"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasttemplate"
)

const (
	accountTypeIDProduction = "87c570e2-c795-44b0-aefa-ebdcffd4d048"

	staxAccountResourceTemplate = `
resource "stax_account" "${accountLabel}" {
	name            = "${accountName}"
	account_type_id = "${accountTypeID}"
}`
)

func TestAccountResource(t *testing.T) {

	si := mocks.NewServerInterface(t)

	si.On("AccountsCreateAccount", mock.AnythingOfType("*echo.context")).Return(func(c echo.Context) error {
		return c.JSON(200, &models.AccountsCreateAccountResponse{TaskId: aws.String("4f2e7318-1e25-4b62-84b5-1cd701042760")})
	})

	si.On("TasksReadTask", mock.AnythingOfType("*echo.context"), mock.AnythingOfType("string")).Return(func(c echo.Context, taskId string) error {
		return c.JSON(200, &models.TasksReadTask{Status: staxsdk.TaskSucceeded, Accounts: &[]string{"f646e0cf-840c-401a-933c-1ef3432b5a37"}})
	})

	si.On("AccountsReadAccountTypes", mock.AnythingOfType("*echo.context"), mock.AnythingOfType("models.AccountsReadAccountTypesParams")).Return(func(c echo.Context, params models.AccountsReadAccountTypesParams) error {
		return c.JSON(200, &models.AccountsReadAccountTypes{
			AccountTypes: []models.AccountType{
				{
					Id:   aws.String(accountTypeIDProduction),
					Name: "production",
				},
			},
		})
	})

	si.On("AccountsReadAccounts", mock.AnythingOfType("*echo.context"), mock.AnythingOfType("models.AccountsReadAccountsParams")).Return(func(c echo.Context, params models.AccountsReadAccountsParams) error {
		return c.JSON(200, &models.AccountsReadAccounts{
			Accounts: []models.Account{
				{
					Id:          aws.String("f646e0cf-840c-401a-933c-1ef3432b5a37"),
					Name:        "presentation-dev",
					Status:      (*models.AccountStatus)(aws.String("ACTIVE")),
					AccountType: aws.String("production"),
					Tags:        &models.StaxTags{},
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
				Config: testAccCheckStaxAccountConfig("presentation-dev", "presentation-dev", accountTypeIDProduction),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaxAccountExists("stax_account.presentation-dev"),
				),
			},
		},
	})
}

func testAccCheckStaxAccountConfig(accountLabel, accountName, accountTypeID string) string {
	return fasttemplate.ExecuteString(staxAccountResourceTemplate, "${", "}",
		map[string]any{
			"accountLabel":  accountLabel,
			"accountName":   accountName,
			"accountTypeID": accountTypeID,
		},
	)
}

func testAccCheckStaxAccountExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No account id set")
		}

		return nil
	}
}
