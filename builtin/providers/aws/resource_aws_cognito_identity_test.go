package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoIdentityPool(t *testing.T) {
	var conf cognitoidentity.CognitoIdentity

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoIdentityDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSCognitoIdentityConfig,
				Check:  testAccCheckAWSIdentityPoolExists("aws_cognitoidentity_pool.identitypool", &conf),
			},
			resource.TestStep{
				Config: testAccAWSCognitoIdentityConfig,
				Check:  testAccCheckAWSIDentityPoolRolesExist("aws_cognitoidentity_pool.identitypool", &conf),
			},
		},
	})
}

func testAccCheckAWSIdentityPoolExists(n string, res *cognitoidentity.CognitoIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Cognito Identity ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidconn

		req := &cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.DescribeIdentityPool(req)
		if err != nil {
			return err
		}

		if *describe.IdentityPoolId != rs.Primary.ID {
			return fmt.Errorf("Cognito identity not found")
		}

		return nil
	}
}

func testAccCheckAWSIDentityPoolRolesExist(n string, res *cognitoidentity.CognitoIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Cognito Identity ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidconn

		req := &cognitoidentity.GetIdentityPoolRolesInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetIdentityPoolRoles(req)
		if err != nil {
			return err
		}

		if *describe.IdentityPoolId != rs.Primary.ID {
			return fmt.Errorf("Cognito identity pool roles not found")
		}

		return nil
	}
}

func testAccCheckAWSCognitoIdentityDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognitoidentity_pool" {
			continue
		}

		req := &cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(s.RootModule().Resources["aws_cognitoidentity_pool.identitypool"].Primary.ID),
		}

		describe, err := conn.DescribeIdentityPool(req)

		if err == nil {
			if *describe.IdentityPoolId == rs.Primary.ID {
				return fmt.Errorf("Identity pool still exists")
			}
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

const testAccAWSCognitoIdentityConfig = `

resource "aws_iam_role" "iam_authenticated_by_cognito_role" {
  name = "iam_authenticated"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Federated": "cognito-identity.amazonaws.com"
        },
        "Action": "sts:AssumeRoleWithWebIdentity",
        "Condition": {
          "StringEquals": {
            "cognito-identity.amazonaws.com:aud": "${aws_cognitoidentity_pool.identitypool.id}"
          },
          "ForAnyValue:StringLike": {
            "cognito-identity.amazonaws.com:amr": "authenticated"
          }
        }
      }
    ]
}
EOF
}

resource "aws_iam_role" "iam_unauthenticated_by_cognito_role" {
  name = "iam_unauthenticated_by_cognito"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Federated": "cognito-identity.amazonaws.com"
        },
        "Action": "sts:AssumeRoleWithWebIdentity",
        "Condition": {
          "StringEquals": {
            "cognito-identity.amazonaws.com:aud": "${aws_cognitoidentity_pool.identitypool.id}"
          },
          "ForAnyValue:StringLike": {
            "cognito-identity.amazonaws.com:amr": "unauthenticated"
          }
        }
      }
    ]
}
EOF
}

resource "aws_cognitoidentity_pool" "identitypool" {
  name = "identitypool"
  allow_unauthenticated = false
  developer_provider_name = "testid"
  login_providers = {
    google = "902422948844-4k3k9cus7o929t25rckjvqma3ihp5l0u.apps.googleusercontent.com"
    facebook = "1003656116363712"
  }
}


resource "aws_cognitoidentity_pool_roles" "identitypool_roles" {
  identity_pool_id = "${aws_cognitoidentity_pool.identitypool.id}"
  authenticated = "${aws_iam_role.iam_authenticated_by_cognito_role.arn}"
  unauthenticated ="${aws_iam_role.iam_unauthenticated_by_cognito_role.arn}"
}
`
