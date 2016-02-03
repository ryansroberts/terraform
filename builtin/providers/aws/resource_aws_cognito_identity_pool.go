package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCognitoIdentityPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoIdentityPoolCreate,
		Read:   resourceAwsCognitoIdentityPoolRead,
		Update: resourceAwsCognitoIdentityPoolUpdate,
		Delete: resourceAwsCognitoIdentityPoolDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"allow_unauthenticated": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"login_providers": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"google": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"facebook": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"amazon": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"twitter": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"digits": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"developer_provider_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"openid_arns": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func providerMap(d *schema.ResourceData) map[string]string {
	m, p := make(map[string]string), d.Get("login_providers").(*schema.Set).List()[0].(map[string]interface{})

	if p["google"] != "" {
		m["accounts.google.com"] = p["google"].(string)
	}
	if p["facebook"] != "" {
		m["graph.facebook.com"] = p["facebook"].(string)
	}
	if p["twitter"] != "" {
		m["api.twitter.com"] = p["twitter"].(string)
	}
	if p["amazon"] != "" {
		m["www.amazon.com"] = p["amazon"].(string)
	}
	if p["digits"] != "" {
		m["www.digits.com"] = p["digits"].(string)
	}

	return m
}

func arnList(d *schema.ResourceData) []string {
	l := []string{}

	for _, value := range d.Get("openid_arns").([]interface{}) {
		l = append(l, value.(string))
	}

	return l
}

// resourceAwsLambdaFunction maps to:
// CreateFunction in the API / SDK
func resourceAwsCognitoIdentityPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	name := d.Get("name").(string)

	log.Printf("[DEBUG] Creating Cognito Identity with name %s", name)

	params := &cognitoidentity.CreateIdentityPoolInput{
		IdentityPoolName:               aws.String(name),
		DeveloperProviderName:          aws.String(d.Get("developer_provider_name").(string)),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated").(bool)),
		SupportedLoginProviders:        aws.StringMap(providerMap(d)),
		OpenIdConnectProviderARNs:      aws.StringSlice(arnList(d)),
	}

	resp, err := conn.CreateIdentityPool(params)
	if err != nil {
		return fmt.Errorf("Error creating cognito identity pool: %s", err)
	}

	d.SetId(*resp.IdentityPoolId)
	log.Printf("[DEBUG] Cognito identity pool ID: %s", d.Id())

	return resourceAwsCognitoIdentityPoolRead(d, meta)
}

func resourceAwsCognitoIdentityPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	resp, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error reading cognito identity pool %s : %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Received Cognito id pool : %s", resp)

	d.Set("name", resp.IdentityPoolName)
	d.Set("developer_provider_name", resp.DeveloperProviderName)
	d.Set("allow_unauthenticated", resp.IdentityPoolName)

	return nil
}

func resourceAwsCognitoIdentityPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	_, err := conn.UpdateIdentityPool(&cognitoidentity.IdentityPool{
		IdentityPoolId:                 aws.String(d.Id()),
		IdentityPoolName:               aws.String(d.Get("name").(string)),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated").(bool)),
		OpenIdConnectProviderARNs:      aws.StringSlice(arnList(d)),
		SupportedLoginProviders:        aws.StringMap(providerMap(d)),
	})

	if err != nil {
		return fmt.Errorf("Error updating cognito identity pool : %s", err)
	}

	return resourceAwsCognitoIdentityPoolRead(d, meta)
}

func resourceAwsCognitoIdentityPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidconn

	_, err := conn.DeleteIdentityPool(&cognitoidentity.DeleteIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})
	if err != nil {

		return err
	}

	return nil
}
