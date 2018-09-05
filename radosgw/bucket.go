//bucket.go - implements bucket admin op API

package radosgw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type BucketStatsType struct {
	Bucket        string `json:"bucket"`
	Zonegroup     string `json:"zonegroup"`
	PlacementRule string `json:"placement_rule"`
	ID            string `json:"id"`
	Marker        string `json:"marker"`
	IndexPool     string `json:"index_pool"`
	Owner         string `json:"owner"`
	Ver           string `json:"ver"`
	MasterVer     string `json:"master_ver"`
	Mtime         string `json:"mtime"`
	MaxMarker     string `json:"max_marker"`
	Usage         struct {
		RgwMain struct {
			NumObjects int64 `json:"num_objects"`
			Size       int64 `json:"size"`
			SizeActual int64 `json:"size_actual"`
		} `json:"rgw.main"`
	} `json:"usage"`
	BucketQuota QuotaType `json:"bucket_quota"`
}

type BucketType struct {
	Name  string           `json:"name,omitempty"`
	Stats *BucketStatsType `json:"stats,omitempty"`
}

// GetBucket - get the bucket info
//
// PARAMS:
//     - bucket: the bucket name
//     - uid: the specific user id for the bucket
//     - stats: return the bucket stats data or not
// RETURN:
//     - int: the response status code
//     - []BucketType: the bucket infomation of the specific uid
//     - error: the request error
func (c *Client) GetBucket(bucket, uid string, stats bool) (int, []BucketType, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(bucket) != 0 {
		args.Add("bucket", bucket)
	}
	if len(uid) != 0 {
		args.Add("uid", uid)
	}
	args.Add("stats", fmt.Sprintf("%v", stats))

	body, status, err := c.sendRequest("GET", "/bucket", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}

	// Parse the result with different data
	result := make([]BucketType, 0)
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return status, nil, err
	}
	if arr, ok := data.([]interface{}); ok { // multiple bucket objects
		for _, v := range arr {
			if name, ok := v.(string); ok {
				result = append(result, BucketType{Name: name})
			} else {
				origin, _ := json.Marshal(v)
				resultStats := &BucketStatsType{}
				if err := json.Unmarshal(origin, resultStats); err != nil {
					return status, nil, err
				}
				result = append(result, BucketType{Stats: resultStats})
			}
		}
	} else { // only one bucket object
		resultStats := &BucketStatsType{}
		if err := json.Unmarshal(body, resultStats); err != nil {
			return status, nil, err
		}
		result = append(result, BucketType{Stats: resultStats})
	}
	return status, result, nil
}

// GetPolicy - get the bucket or object policy config
//
// PARAMS:
//     - bucket: the bucket name
//     - object: the object name
// RETURNS:
//     - int: the response status code
//     - []bytes: the policy config raw bytes
//     - error: the request error
func (c *Client) GetPolicy(bucket, object string) (int, []byte, error) {
	args := url.Values{}
	args.Add("format", "json")
	args.Add("policy", "")
	if len(bucket) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("bucket name should not be empty")
	}
	args.Add("bucket", bucket)
	if len(object) != 0 {
		args.Add("object", object)
	}

	body, status, err := c.sendRequest("GET", "/bucket", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	return status, body, nil
}

// DeleteBucket - delete the given bucket
//
// PARAMS:
//     - bucket: the bucket name to be deleted
//     - purgeObjects: remove a buckets objects before deletion
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) DeleteBucket(bucket string, purgeObjects bool) (int, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(bucket) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket name should not be empty")
	}
	args.Add("bucket", bucket)
	args.Add("purge-objects", fmt.Sprintf("%v", purgeObjects))

	body, status, err := c.sendRequest("DELETE", "/bucket", args, nil, nil)
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}

// DeleteObject - remove an existing object in the given bucket
//
// PARAMS:
//     - bucket: the bucket name
//     - object: the object name to be deleted
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) DeleteObject(bucket, object string) (int, error) {
	if len(bucket) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket name should not be empty")
	}
	if len(object) == 0 {
		return http.StatusBadRequest, fmt.Errorf("object name should not be empty")
	}
	uri := fmt.Sprintf("/%s/%s", bucket, object)

	c.SetPrefix("")
	body, status, err := c.sendRequest("DELETE", uri, nil, nil, nil)
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}

// LinkBucket - Link a bucket to a specified user and unlink it from any previous user
//
// PARAMS:
//     - bucket: the bucket name to link
//     - bucketId: the bucket id to link
//     - uid: the user id to link the bucket
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) LinkBucket(bucket, bucketId, uid string) (int, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(bucket) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket name should not be empty")
	}
	args.Add("bucket", bucket)
	if len(bucketId) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket id should not be empty")
	}
	args.Add("bucket-id", bucketId)
	if len(uid) == 0 {
		return http.StatusBadRequest, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)

	body, status, err := c.sendRequest("PUT", "/bucket", args, nil, nil)
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}

// UnlinkBucket - Unlink a bucket from a specified user for changing bucket ownership
//
// PARAMS:
//     - bucket: the bucket name to unlink
//     - uid: the user ID to unlink the bucket from
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) UnlinkBucket(bucket, uid string) (int, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(bucket) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket name should not be empty")
	}
	args.Add("bucket", bucket)
	if len(uid) == 0 {
		return http.StatusBadRequest, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)

	body, status, err := c.sendRequest("POST", "/bucket", args, nil, nil)
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}

// CreateBucket - create an bucket
//
// PARAMS:
//     - bucket: the bucket name to be created
//     - region: the bucket to put in this region
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) CreateBucket(bucket, region, acl string) (int, error) {
	if len(bucket) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket name should not be empty")
	}
	var aclHeader map[string]string
	if len(acl) != 0 {
		aclHeader = map[string]string{"x-amz-acl": acl}
	}
	inputBody := new(bytes.Buffer)
	if len(region) != 0 {
		inputBody.WriteString(fmt.Sprintf(`
<CreateBucketConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <LocationConstraint>%s</LocationConstraint>
</CreateBucketConfiguration>
`, region))
	}
	c.SetPrefix("")
	defer func() {
		c.SetPrefix(defaultAdminPrefix)
	}()
	body, status, err := c.sendRequest(
		"PUT", "/"+bucket, nil, aclHeader, ioutil.NopCloser(inputBody))
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}
