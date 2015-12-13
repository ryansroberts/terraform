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

func resourceAwsApiGatewayMethod() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayMethodCreate,
		Read:   resourceAwsApiGatewayMethodRead,
		Update: resourceAwsApiGatewayMethodUpdate,
		Delete: resourceAwsApiGatewayMethodDelete,

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

			"authorization": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"api_key_required": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"integration": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true, // HTTP, AWS, MOCK
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(string)
								if value != "AWS" && value != "HTTP" && value != "MOCK" {
									errors = append(errors, fmt.Errorf(
										"%q has unsupported value %q", k, value))
								}
								return
							},
						},
						"http_method": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"uri": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsApiGatewayMethodCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	_, err := conn.PutMethod(&apigateway.PutMethodInput{
		AuthorizationType: aws.String(d.Get("authorization").(string)),
		HttpMethod:        aws.String(d.Get("http_method").(string)),
		ResourceId:        aws.String(d.Get("resource_id").(string)),
		RestApiId:         aws.String(d.Get("api_id").(string)),
		RequestModels:     nil,
		RequestParameters: nil,
		ApiKeyRequired:    aws.Bool(d.Get("api_key_required").(bool)),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Method: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))
	log.Printf("[DEBUG] API Gateway Method ID: %s", d.Id())

	if v, ok := d.GetOk("integration"); ok {
		integration := v.(map[string]interface{})

		var integrationHttpMethod *string
		if v, ok := integration["http_method"]; ok {
			integrationHttpMethod = aws.String(v.(string))
		}
		var uri *string
		if v, ok := integration["uri"]; ok {
			uri = aws.String(v.(string))
		}
		_, integrationErr := conn.PutIntegration(&apigateway.PutIntegrationInput{
			HttpMethod: aws.String(d.Get("http_method").(string)),
			ResourceId: aws.String(d.Get("resource_id").(string)),
			RestApiId:  aws.String(d.Get("api_id").(string)),
			Type:       aws.String(integration["type"].(string)),
			IntegrationHttpMethod: integrationHttpMethod,
			Uri:                uri,
			RequestParameters:  nil,
			RequestTemplates:   nil,
			CacheKeyParameters: nil,
			CacheNamespace:     nil,
			Credentials:        nil,
		})

		if integrationErr != nil {
			return fmt.Errorf("Error creating API Gateway Method Integration: %s", integrationErr)
		}
	}

	return nil
}

func resourceAwsApiGatewayMethodRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Method %s", d.Id())
	out, err := conn.GetMethod(&apigateway.GetMethodInput{
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

	return nil
}

func resourceAwsApiGatewayMethodUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Reading API Gateway Method %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)
	if o, n := d.GetChange("resource_id"); o.(string) != n.(string) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/resourceId"),
			Value: aws.String(d.Get("resource_id").(string)),
		})
	}

	// {op: "add", path: "/requestModels/application~1json", value: "user"}
	// {op: "remove", path: "/requestModels/application~1json"}
	out, err := conn.UpdateMethod(&apigateway.UpdateMethodInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		ResourceId:      aws.String(d.Get("resource_id").(string)),
		RestApiId:       aws.String(d.Get("api_id").(string)),
		PatchOperations: operations,
	})
	if err != nil {
		d.SetId("")
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Method: %s", out)
	d.SetId(fmt.Sprintf("%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))

	return nil
}

func resourceAwsApiGatewayMethodDelete(d *schema.ResourceData, meta interface{}) error {
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
		_, err := conn.DeleteMethod(&apigateway.DeleteMethodInput{
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
