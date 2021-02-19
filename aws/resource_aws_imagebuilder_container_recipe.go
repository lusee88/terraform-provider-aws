package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsImageBuilderContainerRecipe() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderContainerRecipeCreate,
		Read:   resourceAwsImageBuilderContainerRecipeRead,
		Update: resourceAwsImageBuilderContainerRecipeUpdate,
		Delete: resourceAwsImageBuilderContainerRecipeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"component": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"container_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"DOCKER"}, false),
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"dockerfile_template_data": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"dockerfile_template_data", "dockerfile_template_uri"},
				ValidateFunc: validation.StringLenBetween(1, 16000),
			},
			"dockerfile_template_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"dockerfile_template_data", "dockerfile_template_uri"},
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"image_os_version_override": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_override": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Windows", "Linux"}, false),
			},
			"semantic_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"tags": tagsSchema(),
			"target_repository": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"service": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"ECR"}, false),
						},
					},
				},
			},
			"working_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
		},
	}
}

func resourceAwsImageBuilderContainerRecipeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateContainerRecipeInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("component"); ok && len(v.([]interface{})) > 0 {
		input.Components = expandImageBuilderComponentConfigurations(v.([]interface{}))
	}

	if v, ok := d.GetOk("container_type"); ok {
		input.ContainerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dockerfile_template_data"); ok {
		input.DockerfileTemplateData = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parent_image"); ok {
		input.ParentImage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags()
	}

	if v, ok := d.GetOk("semantic_version"); ok {
		input.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_repository"); ok {
		input.TargetRepository = expandImageBuilderTargetContainerRepository(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := d.GetOk("working_directory"); ok {
		input.WorkingDirectory = aws.String(v.(string))
	}

	output, err := conn.CreateContainerRecipe(input)

	if err != nil {
		return fmt.Errorf("error creating Image Builder Container Recipe: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Container Recipe: empty response")
	}

	d.SetId(aws.StringValue(output.ContainerRecipeArn))

	return resourceAwsImageBuilderContainerRecipeRead(d, meta)
}

func resourceAwsImageBuilderContainerRecipeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetContainerRecipeInput{
		ContainerRecipeArn: aws.String(d.Id()),
	}

	output, err := conn.GetContainerRecipe(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Container Recipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Container Recipe (%s): %w", d.Id(), err)
	}

	if output == nil || output.ContainerRecipe == nil {
		return fmt.Errorf("error getting Image Builder Container Recipe (%s): empty response", d.Id())
	}

	containerRecipe := output.ContainerRecipe

	d.Set("arn", containerRecipe.Arn)
	d.Set("component", flattenImageBuilderComponentConfigurations(containerRecipe.Components))
	d.Set("container_type", containerRecipe.ContainerType)
	d.Set("date_created", containerRecipe.DateCreated)
	d.Set("description", containerRecipe.Description)
	d.Set("dockerfile_template_data", containerRecipe.DockerfileTemplateData)
	d.Set("name", containerRecipe.Name)
	d.Set("encrypted", containerRecipe.Encrypted)
	d.Set("kms_key_id", containerRecipe.KmsKeyId)
	d.Set("owner", containerRecipe.Owner)
	d.Set("parent_image", containerRecipe.ParentImage)
	d.Set("platform", containerRecipe.Platform)
	d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(containerRecipe.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	d.Set("semantic_version", containerRecipe.Version)
	d.Set("target_repository", []interface{}{flattenImageBuilderTargetContainerRepository(containerRecipe.TargetRepository)})
	d.Set("working_directory", containerRecipe.WorkingDirectory)

	return nil
}

func resourceAwsImageBuilderContainerRecipeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Image Recipe (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderContainerRecipeRead(d, meta)
}

func resourceAwsImageBuilderContainerRecipeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.DeleteContainerRecipeInput{
		ContainerRecipeArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteContainerRecipe(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Image Recipe (%s): %w", d.Id(), err)
	}

	return nil
}

func expandImageBuilderTargetContainerRepository(tfMap map[string]interface{}) *imagebuilder.TargetContainerRepository {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.TargetContainerRepository{}

	if v, ok := tfMap["repository_name"].(string); ok && v != "" {
		apiObject.RepositoryName = aws.String(v)
	}

	if v, ok := tfMap["service"].(string); ok && v != "" {
		apiObject.Service = aws.String(v)
	}

	return apiObject
}

func flattenImageBuilderTargetContainerRepository(apiObject *imagebuilder.TargetContainerRepository) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RepositoryName; v != nil {
		tfMap["repository_name"] = aws.StringValue(v)
	}

	if v := apiObject.Service; v != nil {
		tfMap["service"] = aws.StringValue(v)
	}

	return tfMap
}
