package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayIntegrationCreate,
		Read:   resourceAwsApiGatewayIntegrationRead,
		Update: resourceAwsApiGatewayIntegrationUpdate,
		Delete: resourceAwsApiGatewayIntegrationDelete,

		Schema: map[string]*schema.Schema{
			"api_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_method": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"uri": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"credentials": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"integration_http_method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"request_templates": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func resourceAwsApiGatewayIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	var integrationHttpMethod *string
	if v, ok := d.GetOk("integration_http_method"); ok {
		integrationHttpMethod = aws.String(v.(string))
	}
	var uri *string
	if v, ok := d.GetOk("uri"); ok {
		uri = aws.String(v.(string))
	}
	templates := make(map[string]string)
	for k, v := range d.Get("request_templates").(map[string]interface{}) {
		templates[k] = v.(string)
	}

	var credentials *string
	if val, ok := d.GetOk("credentials"); ok {
		credentials = aws.String(val.(string))
	}

	_, err := conn.PutIntegration(&apigateway.PutIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("api_id").(string)),
		Type:       aws.String(d.Get("type").(string)),
		IntegrationHttpMethod: integrationHttpMethod,
		Uri:                uri,
		RequestParameters:  nil,
		RequestTemplates:   aws.StringMap(templates),
		Credentials:        credentials,
		CacheNamespace:     nil,
		CacheKeyParameters: nil,
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Method: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))

	return nil
}

func resourceAwsApiGatewayIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Method %s", d.Id())
	out, err := conn.GetIntegration(&apigateway.GetIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("api_id").(string)),
	})
	if err != nil {
		d.SetId("")
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Method: %s", out)
	d.SetId(fmt.Sprintf("%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))

	d.Set("request_templates", aws.StringValueMap(out.RequestTemplates))
	d.Set("credentials", aws.StringValue(out.Credentials))
	d.Set("type", aws.StringValue(out.Type))
	d.Set("uri", aws.StringValue(out.Uri))

	return nil
}

func resourceAwsApiGatewayIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsApiGatewayIntegrationCreate(d, meta)
}

func resourceAwsApiGatewayIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway Method: %s", d.Id())

	resourceId := d.Get("resource_id").(string)
	if o, n := d.GetChange("resource_id"); o.(string) != n.(string) {
		resourceId = o.(string)
	}
	httpMethod := d.Get("http_method").(string)
	if o, n := d.GetChange("http_method"); o.(string) != n.(string) {
		httpMethod = o.(string)
	}
	restApiID := d.Get("api_id").(string)
	if o, n := d.GetChange("api_id"); o.(string) != n.(string) {
		restApiID = o.(string)
	}

	return resource.Retry(5*time.Minute, func() error {
		log.Printf("[DEBUG] schema is %#v", d)
		_, err := conn.DeleteIntegration(&apigateway.DeleteIntegrationInput{
			HttpMethod: aws.String(httpMethod),
			ResourceId: aws.String(resourceId),
			RestApiId:  aws.String(restApiID),
		})
		if err == nil {
			return nil
		}

		apigatewayErr, ok := err.(awserr.Error)
		if apigatewayErr.Code() == "NotFoundException" {
			return nil
		}

		if !ok {
			return resource.RetryError{Err: err}
		}

		return resource.RetryError{Err: err}
	})
}
