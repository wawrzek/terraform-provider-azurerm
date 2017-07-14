package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMMySQLFirewallRule_basic(t *testing.T) {
	resourceName := "azurerm_mysql_firewall_rule.test"
	ri := acctest.RandInt()
	config := testAccAzureRMMySQLFirewallRule_basic(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMMySQLFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMMySQLFirewallRuleExists(resourceName),
				),
			},
		},
	})
}

func testCheckAzureRMMySQLFirewallRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		serverName := rs.Primary.Attributes["server_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for MySQL Firewall Rule: %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).mysqlFirewallRulesClient

		resp, err := client.Get(resourceGroup, serverName, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on mysqlFirewallRulesClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: MySQL Firewall Rule %q (server %q resource group: %q) does not exist", name, serverName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMMySQLFirewallRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).mysqlDatabasesClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_mysql_firewall_rule" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		serverName := rs.Primary.Attributes["server_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(resourceGroup, serverName, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("MySQL Firewall Rule still exists:\n%#v", resp)
		}
	}

	return nil
}

func testAccAzureRMMySQLFirewallRule_basic(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
    name = "acctestRG-%d"
    location = "West US"
}

resource "azurerm_mysql_server" "test" {
  name = "acctestmysqlsvr-%d"
  location = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  sku {
    name = "MYSQLB50"
    capacity = 50
    tier = "Basic"
  }

  administrator_login = "acctestun"
  administrator_login_password = "H@Sh1CoR3!"
  version = "5.6"
  storage_mb = 51200
  ssl_enforcement = "Enabled"
}

resource "azurerm_mysql_firewall_rule" "test" {
  name                = "acctestfwrule-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  server_name         = "${azurerm_mysql_server.test.name}"
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "255.255.255.255"
}
`, rInt, rInt, rInt)
}
