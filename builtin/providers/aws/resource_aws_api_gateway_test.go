package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGateway_basic(t *testing.T) {
	var conf apigateway.RestApi

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSAPIGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayExists("aws_api_gateway.bar", &conf),
					testAccCheckAWSAPIGatewayAttributes(&conf),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayAttributes(conf *apigateway.RestApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.Id == "" {
			return fmt.Errorf("empty ID")
		}
		if *conf.Name == "" {
			return fmt.Errorf("empty Name")
		}

		return nil
	}
}

func testAccCheckAWSAPIGatewayExists(n string, res *apigateway.RestApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetRestApiInput{
			RestApiId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetRestApi(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway not found")
		}

		res.Id = describe.Id
		res.Name = describe.Name

		return nil
	}
}

func testAccCheckAWSAPIGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway" {
			continue
		}

		req := &apigateway.GetRestApisInput{}
		describe, err := conn.GetRestApis(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway still exists")
			}
		}

		return err
	}

	return nil
}

const testAccAWSAPIGatewayConfig = `
resource "aws_api_gateway" "bar" {
  name = "barf"
}
`
