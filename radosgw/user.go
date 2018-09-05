//user.go - implements user admin op API

package radosgw

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type CapType struct {
	Perm string `json:"perm"`
	Type string `json:"type"`
}

type KeyType struct {
	User      string `json:"user"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type QuotaType struct {
	MaxObjects int64 `json:"max_objects,omitempty"`
	MaxSize    int64 `json:"max_size,omitempty"`
	Enabled    bool  `json:"enabled,omitempty"`
}

type UserType struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Keys        []KeyType `json:"keys"`
	Caps        []CapType `json:"caps"`
	MaxBuckets  int64     `json:"max_buckets"`
	Suspended   int       `json:"suspended"`
}

// GetUser - get the user info by the specific uid
//
// PARAMS:
//     - uid: user id string
// RETURN:
//     - int: the response status code
//     - *UserType: the user infomation of the specific uid
//     - error: the request error
func (c *Client) GetUser(uid ...string) (int, *UserType, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(uid) != 0 {
		args.Add("uid", uid[0])
	}
	body, status, err := c.sendRequest("GET", "/user", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := &UserType{}
	if err := json.Unmarshal(body, result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// CreateUser - create the radosgw user
//
// PARAMS:
//     - uid: user id string
//     - dispName: display name string
//     - email: email of the user
//     - maxBuckets: the max bucket number for this user
// RETURN:
//     - int: the response status code
//     - *UserType: the new created user infomation
//     - error: the request error
func (c *Client) CreateUser(uid, dispName, email string, maxBuckets int64) (int, *UserType, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(uid) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	if len(dispName) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("display name should not be empty")
	}
	args.Add("display-name", dispName)
	if len(email) != 0 {
		args.Add("email", email)
	}
	if maxBuckets != -1 {
		args.Add("max-buckets", fmt.Sprintf("%d", maxBuckets))
	}

	body, status, err := c.sendRequest("PUT", "/user", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := &UserType{}
	if err := json.Unmarshal(body, result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// UpdateUser - update the radosgw user
//
// PARAMS:
//     - uid: user id string
//     - displayName: display name string
//     - email: email of the user
//     - maxBuckets: the max bucket number for this user
//     - suspended: set the user suspended of not
// RETURN:
//     - int: the response status code
//     - *UserType: the modified user information
//     - error: the request error
func (c *Client) UpdateUser(uid, displayName, email string,
	maxBuckets int64, suspended bool) (int, *UserType, error) {
	args := url.Values{}
	args.Add("format", "json")
	if len(uid) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	if len(displayName) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("display name should not be empty")
	}
	args.Add("display-name", displayName)
	args.Add("email", email)
	args.Add("max-buckets", fmt.Sprintf("%d", maxBuckets))
	if suspended {
		args.Add("suspended", "1")
	} else {
		args.Add("suspended", "0")
	}

	body, status, err := c.sendRequest("POST", "/user", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := &UserType{}
	if err := json.Unmarshal(body, result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// DeleteUser - delete the radosgw user by the user id
//
// PARAMS:
//     - uid: user id string
//     - purgeData: delete the data of the user or not
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) DeleteUser(uid string, purgeData bool) (int, error) {
	args := url.Values{}
	if len(uid) == 0 {
		return http.StatusBadRequest, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	if purgeData {
		args.Add("purge-data", "")
	}

	_, status, err := c.sendRequest("DELETE", "/user", args, nil, nil)
	return status, err
}

// CreateKey - creat a new ak/sk pair of the given user
//
// PARAMS:
//     - uid: user id of the ak/sk pair
// RETURN:
//     - int: the response status code
//     - *KeyType: the new create ak/sk pair
//     - error: the request error
func (c *Client) CreateKey(uid string) (int, []KeyType, error) {
	args := url.Values{}
	args.Add("format", "json")
	args.Add("key", "")
	if len(uid) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)

	body, status, err := c.sendRequest("PUT", "/user", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := make([]KeyType, 0)
	if err := json.Unmarshal(body, &result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// DeleteKey - delete the ak/sk pair by user id and ak
//
// PARAMS:
//     - uid: user id string
//     - ak: the access key id string
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) DeleteKey(uid, ak string) (int, error) {
	args := url.Values{}
	args.Add("key", "")
	if len(uid) == 0 || len(ak) == 0 {
		return http.StatusBadRequest, fmt.Errorf("user id or access key id should not be empty")
	}
	args.Add("uid", uid)
	args.Add("access-key", ak)

	_, status, err := c.sendRequest("DELETE", "/user", args, nil, nil)
	return status, err
}

// AddCaps - add a capability of a given user
//
// PARAMS:
//     - uid: user id of the ak/sk pair
//     - user: user caps of the user
//     - buckets: buckets caps of the user
//     - usage: usage caps of the user
// RETURN:
//     - int: the response status code
//     - []CapType: the current caps of the user
//     - error: the request error
func (c *Client) AddCaps(uid string, user, buckets, usage []string) (int, []CapType, error) {
	args := url.Values{}
	args.Add("format", "json")
	args.Add("caps", "")
	if len(uid) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	userCaps := make([]string, 0, 3)
	if user != nil && len(user) != 0 {
		userCaps = append(userCaps, fmt.Sprintf("user=%s", strings.Join(user, ",")))
	}
	if buckets != nil && len(buckets) != 0 {
		userCaps = append(userCaps, fmt.Sprintf("buckets=%s", strings.Join(buckets, ",")))
	}
	if usage != nil && len(usage) != 0 {
		userCaps = append(userCaps, fmt.Sprintf("usage=%s", strings.Join(usage, ",")))
	}
	args.Add("user-caps", strings.Join(userCaps, ";"))

	body, status, err := c.sendRequest("PUT", "/user", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := make([]CapType, 0)
	if err := json.Unmarshal(body, &result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// DeleteCaps - delete the capability of the given user id
//
// PARAMS:
//     - uid: user id string
//     - user: user caps of the user to be deleted
//     - buckets: buckets caps of the user to be deleted
//     - usage: usage caps of the user to be deleted
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) DeleteCaps(uid string, user, buckets, usage []string) (int, error) {
	args := url.Values{}
	args.Add("caps", "")
	if len(uid) == 0 {
		return http.StatusBadRequest, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	userCaps := make([]string, 0, 3)
	if user != nil && len(user) != 0 {
		userCaps = append(userCaps, fmt.Sprintf("user=%s", strings.Join(user, ",")))
	}
	if buckets != nil && len(buckets) != 0 {
		userCaps = append(userCaps, fmt.Sprintf("buckets=%s", strings.Join(buckets, ",")))
	}
	if usage != nil && len(usage) != 0 {
		userCaps = append(userCaps, fmt.Sprintf("usage=%s", strings.Join(usage, ",")))
	}
	args.Add("user-caps", strings.Join(userCaps, ";"))

	_, status, err := c.sendRequest("DELETE", "/user", args, nil, nil)
	return status, err
}

// GetQuota - get the quota of a given user
//
// PARAMS:
//     - uid: user id string
//     - quotaType: quota type, only "user" and "bucket" allowed
// RETURN:
//     - int: the response status code
//     - *QuotaType: the user quota setting object
//     - error: the request error
func (c *Client) GetQuota(uid, quotaType string) (int, *QuotaType, error) {
	args := url.Values{}
	args.Add("format", "json")
	args.Add("quota", "")
	if len(uid) == 0 {
		return http.StatusBadRequest, nil, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	if quotaType != "user" && quotaType != "bucket" {
		return http.StatusBadRequest, nil, fmt.Errorf("quota type is not valid")
	}
	args.Add("quota-type", quotaType)

	body, status, err := c.sendRequest("GET", "/user", args, nil, nil)
	if err != nil {
		return status, nil, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, nil, fmt.Errorf("%s", string(body))
	}
	result := &QuotaType{}
	if err := json.Unmarshal(body, result); err != nil {
		return status, nil, err
	}
	return status, result, nil
}

// SetQuota - set the quota of a given user
//
// PARAMS:
//     - uid: user id string
//     - quotaType: quota type, only "user" and "bucket" allowed
//     - maxObjects: max objects number, -1 means not set
//     - maxSize: max size can be used, -1 means not set
//     - enabled: enabled or not
//     - bucketName: the quota set on which bucket if quota type is "bucket"
// RETURN:
//     - int: the response status code
//     - error: the request error
func (c *Client) SetQuota(uid, quotaType string, maxObjects,
	maxSize int64, enabled bool, bucketName string) (int, error) {
	args := url.Values{}
	args.Add("format", "json")
	args.Add("quota", "")
	if len(uid) == 0 {
		return http.StatusBadRequest, fmt.Errorf("user id should not be empty")
	}
	args.Add("uid", uid)
	if quotaType != "user" && quotaType != "bucket" {
		return http.StatusBadRequest, fmt.Errorf("quota type is not valid")
	}
	args.Add("quota-type", quotaType)
	if maxObjects != -1 {
		args.Add("max-objects", fmt.Sprintf("%d", maxObjects))
	}
	if maxSize != -1 {
		args.Add("max-size-kb", fmt.Sprintf("%d", maxSize/1024))
	}
	args.Add("enabled", fmt.Sprintf("%v", enabled))
	if quotaType == "bucket" && len(bucketName) == 0 {
		return http.StatusBadRequest, fmt.Errorf("bucket name is empty for bucket quota type")
	}
	args.Add("bucket", bucketName)

	body, status, err := c.sendRequest("PUT", "/"+quotaType, args, nil, nil)
	if err != nil {
		return status, fmt.Errorf("%s: %s", err.Error(), string(body))
	}
	if status >= 400 {
		return status, fmt.Errorf("%s", string(body))
	}
	return status, nil
}
