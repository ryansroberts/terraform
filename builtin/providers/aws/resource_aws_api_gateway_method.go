package aws

import (
	"fmt"
	"log"
	"strconv"
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
				ForceNew: true,
			},

			"api_key_required": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"request": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_method": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"models": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
						},
						"headers": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeBool},
						},
						"parameters": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeBool},
						},
						"templates": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"integration": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
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
						"credentials": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"response": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status_code": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"lambda_error_regex": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"models": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
						},
						"headers": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
						},
						"templates": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func boolMapFromState(d *schema.ResourceData, subtype string, field string) map[string]*bool {
	if v, ok := d.GetOk(subtype); ok {
		v := v.([]interface{})[0].(map[string]interface{})

		m := make(map[string]bool)

		if v[field] == nil {
			return nil
		}

		for k, value := range v[field].(map[string]interface{}) {
			m[k] = value.(bool)
		}

		return aws.BoolMap(m)
	}

	return nil
}

func stringMapFromState(d *schema.ResourceData, subtype string, field string) map[string]*string {
	if v, ok := d.GetOk(subtype); ok {
		v := v.([]interface{})[0].(map[string]interface{})
		m := make(map[string]string)

		if v[field] == nil {
			return nil
		}

		for k, value := range v[field].(map[string]interface{}) {
			m[k] = value.(string)
		}

		return aws.StringMap(m)
	}

	return nil
}

func resourceAwsApiGatewayMethodCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	_, err := conn.PutMethod(&apigateway.PutMethodInput{
		AuthorizationType: aws.String(d.Get("authorization").(string)),
		HttpMethod:        aws.String(d.Get("http_method").(string)),
		ResourceId:        aws.String(d.Get("resource_id").(string)),
		RestApiId:         aws.String(d.Get("api_id").(string)),
		RequestModels:     stringMapFromState(d, "request", "models"),
		RequestParameters: boolMapFromState(d, "request", "parameters"),
		ApiKeyRequired:    aws.Bool(d.Get("api_key_required").(bool)),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Method: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%s", d.Get("api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))
	log.Printf("[DEBUG] API Gateway Method ID: %s", d.Id())

	if integration, ok := d.GetOk("integration"); ok {
		integration := integration.([]interface{})[0].(map[string]interface{})

		var integrationHttpMethod *string
		if v, ok := integration["http_method"]; ok {
			integrationHttpMethod = aws.String(v.(string))
		}
		var uri *string
		if v, ok := integration["uri"]; ok {
			uri = aws.String(v.(string))
		}
		_, err := conn.PutIntegration(&apigateway.PutIntegrationInput{
			HttpMethod: aws.String(d.Get("http_method").(string)),
			ResourceId: aws.String(d.Get("resource_id").(string)),
			RestApiId:  aws.String(d.Get("api_id").(string)),
			Type:       aws.String(integration["type"].(string)),
			IntegrationHttpMethod: integrationHttpMethod,
			Uri:                uri,
			RequestTemplates:   stringMapFromState(d, "request", "templates"),
			CacheKeyParameters: nil,
			CacheNamespace:     nil,
			Credentials:        aws.String(integration["credentials"].(string)),
		})

		if err != nil {
			return fmt.Errorf("Error creating API Gateway Method Integration: %s", err)
		}
	}

	if v, ok := d.GetOk("response"); ok {
		responses := v.([]interface{})

		for _, response := range responses {
			response := response.(map[string]interface{})
			templates := make(map[string]string)
			if response["templates"] != nil {
				v := response["templates"].(map[string]interface{})
				for k, t := range v {
					templates[k] = t.(string)
				}
			}

			models := make(map[string]string)
			if response["models"] != nil {
				v := response["models"].(map[string]interface{})
				for k, t := range v {
					models[k] = t.(string)
				}
			}

			headers := make(map[string]string)
			headermap := make(map[string]bool)
			if response["headers"] != nil {
				v := response["headers"].(map[string]interface{})
				for k, t := range v {
					headers["method.response.header."+k] = t.(string)
					headermap["method.response.header."+k] = false
				}
			}

			_, err := conn.PutMethodResponse(&apigateway.PutMethodResponseInput{
				HttpMethod:         aws.String(d.Get("http_method").(string)),
				ResourceId:         aws.String(d.Get("resource_id").(string)),
				RestApiId:          aws.String(d.Get("api_id").(string)),
				StatusCode:         aws.String(strconv.Itoa(response["status_code"].(int))),
				ResponseModels:     aws.StringMap(models),
				ResponseParameters: aws.BoolMap(headermap),
			})

			if err != nil {
				return fmt.Errorf("Error creating API Gateway Method Responses: %s", err)
			}

			_, err = conn.PutIntegrationResponse(&apigateway.PutIntegrationResponseInput{
				HttpMethod:         aws.String(d.Get("http_method").(string)),
				ResourceId:         aws.String(d.Get("resource_id").(string)),
				RestApiId:          aws.String(d.Get("api_id").(string)),
				StatusCode:         aws.String(strconv.Itoa(response["status_code"].(int))),
				ResponseTemplates:  aws.StringMap(templates),
				ResponseParameters: aws.StringMap(headers),
			})

			if err != nil {
				return fmt.Errorf("Error creating API Gateway Method Integration Responses: %s", err)
			}

		}
	}

	return resourceAwsApiGatewayMethodRead(d, meta)
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
	resourceAwsApiGatewayMethodDelete(d, meta)
	return resourceAwsApiGatewayMethodCreate(d, meta)
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
