package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoIdentityPoolRoles() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoIdentityPoolRolesCreate,
		Read:   resourceAwsCognitoIdentityPoolRolesRead,
		Update: resourceAwsCognitoIdentityPoolRolesUpdate,
		Delete: resourceAwsCognitoIdentityPoolRolesDelete,

		Schema: map[string]*schema.Schema{
			"identity_pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"authenticated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"unauthenticated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsCognitoIdentityPoolRolesCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	m := make(map[string]string)
	if d.Get("authenticated") != "" {
		m["authenticated"] = d.Get("authenticated").(string)
	}
	if d.Get("unauthenticated") != "" {
		m["unauthenticated"] = d.Get("unauthenticated").(string)
	}

	d.SetId(d.Get("identity_pool_id").(string) + "_roles")

	_, err := conn.SetIdentityPoolRoles(&cognitoidentity.SetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Get("identity_pool_id").(string)),
		Roles:          aws.StringMap(m),
	})

	if err != nil {
		return fmt.Errorf("Error creating cognito identity pool roles: %s", err)
	}

	return resourceAwsCognitoIdentityPoolRolesRead(d, meta)
}

func resourceAwsCognitoIdentityPoolRolesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	pool, err := conn.GetIdentityPoolRoles(&cognitoidentity.GetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Get("identity_pool_id").(string)),
	})

	if err != nil {
		if err, ok := err.(awserr.Error); ok && err.Code() == "NotFound" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading cognito identity pool roles %s", err)
	}

	if pool.Roles != nil {
		if pool.Roles["authenticated"] != nil {
			d.Set("authenticated", pool.Roles["authenticated"])
		}
		if pool.Roles["unauthenticated"] != nil {
			d.Set("unauthenticated", pool.Roles["unauthenticated"])
		}
	}

	return nil
}

func resourceAwsCognitoIdentityPoolRolesUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn
	m := make(map[string]string)

	if d.Get("authenticated") != nil {
		m["authenticated"] = d.Get("authenticated").(string)
	}
	if d.Get("unauthenticated") != nil {
		m["unauthenticated"] = d.Get("unauthenticated").(string)
	}

	_, err := conn.SetIdentityPoolRoles(&cognitoidentity.SetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Get("identity_pool_id").(string)),
		Roles:          aws.StringMap(m),
	})

	if err != nil {
		if err, ok := err.(awserr.Error); ok && err.Code() == "NotFound" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error updating cognito identity pool roles %s", err)
	}

	return resourceAwsCognitoIdentityPoolRolesRead(d, meta)
}

func resourceAwsCognitoIdentityPoolRolesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	_, err := conn.SetIdentityPoolRoles(&cognitoidentity.SetIdentityPoolRolesInput{
		IdentityPoolId: aws.String(d.Get("identity_pool_id").(string)),
		Roles:          aws.StringMap(map[string]string{}),
	})

	if err != nil {
		if err, ok := err.(awserr.Error); ok && err.Code() == "NotFound" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error deleting cognito identity pool roles %s", err)
	}

	return nil
}
