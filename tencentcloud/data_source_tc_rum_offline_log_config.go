/*
Use this data source to query detailed information of rum offlineLogConfig

Example Usage

```hcl
data "tencentcloud_rum_offline_log_config" "offlineLogConfig" {
  project_key = "ZEYrYfvaYQ30jRdmPx"
}
```
*/
package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	rum "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/rum/v20210622"
	"github.com/tencentcloudstack/terraform-provider-tencentcloud/tencentcloud/internal/helper"
)

func dataSourceTencentCloudRumOfflineLogConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudRumOfflineLogConfigRead,
		Schema: map[string]*schema.Schema{
			"project_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique project key for reporting.",
			},

			"unique_id_set": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Unique identifier of the user to be listened on(aid or uin).",
			},

			"msg": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "API call information.",
			},

			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
		},
	}
}

func dataSourceTencentCloudRumOfflineLogConfigRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.tencentcloud_rum_offline_log_config.read")()
	defer inconsistentCheck(d, meta)()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), logIdKey, logId)
	var projectKey string

	paramMap := make(map[string]interface{})
	if v, ok := d.GetOk("project_key"); ok {
		projectKey = v.(string)
		paramMap["project_key"] = helper.String(v.(string))
	}

	rumService := RumService{client: meta.(*TencentCloudClient).apiV3Conn}

	var logConfigs *rum.DescribeOfflineLogConfigsResponseParams
	err := resource.Retry(readRetryTimeout, func() *resource.RetryError {
		results, e := rumService.DescribeRumOfflineLogConfigByFilter(ctx, paramMap)
		if e != nil {
			return retryError(e)
		}
		logConfigs = results
		return nil
	})
	if err != nil {
		log.Printf("[CRITAL]%s read Rum uniqueIDSet failed, reason:%+v", logId, err)
		return err
	}

	if logConfigs == nil {
		return fmt.Errorf("Query by id %v is empty", projectKey)
	}

	var uniqueID []string
	if logConfigs.UniqueIDSet != nil && len(logConfigs.UniqueIDSet) > 0 {
		for _, v := range logConfigs.UniqueIDSet {
			uniqueID = append(uniqueID, *v)
		}
		_ = d.Set("unique_id_set", uniqueID)
	}

	if logConfigs.Msg != nil {
		_ = d.Set("msg", *logConfigs.Msg)
	}

	d.SetId(helper.DataResourceIdsHash(uniqueID))

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if e := writeToFile(output.(string), map[string]interface{}{
			"project_key":   projectKey,
			"unique_id_set": uniqueID,
			"msg":           *logConfigs.Msg,
		}); e != nil {
			return e
		}
	}

	return nil
}
